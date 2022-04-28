package dockertest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Pool struct {
	client *client.Client
}

type Resource struct {
	ID    string
	ports nat.PortMap
}

func NewPool() (*Pool, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("client.NewClientWithOpts: %w", err)
	}
	return &Pool{client: c}, nil
}

func (p *Pool) Run(ctx context.Context, image string, opts ...Option) (*Resource, error) {
	opt := new(RunOptions)
	for _, o := range opts {
		o.Apply(opt)
	}

	if opt.Config == nil {
		opt.Config = &container.Config{Image: image}
	} else {
		opt.Config.Image = image
	}
	if opt.HostConfig == nil {
		opt.HostConfig = &container.HostConfig{PublishAllPorts: true, AutoRemove: true}
	} else {
		opt.HostConfig.PublishAllPorts = true
		opt.HostConfig.AutoRemove = true
	}

	if _, _, err := p.client.ImageInspectWithRaw(ctx, opt.Config.Image); err != nil {
		r, err := p.client.ImagePull(ctx, opt.Config.Image, types.ImagePullOptions{Platform: opt.Platform})
		if err != nil {
			return nil, fmt.Errorf("p.client.ImagePull: %w", err)
		}
		io.Copy(io.Discard, r)
	}

	resp, err := p.client.ContainerCreate(ctx, opt.Config, opt.HostConfig, opt.NetworkingConfig, nil, opt.ContainerName)
	if err != nil {
		return nil, fmt.Errorf("p.client.ContainerCreate: %w", err)
	}

	if err := p.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("p.client.ContainerStart: %w", err)
	}

	c, err := p.client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("p.client.ContainerInspect: %w", err)
	}
	if health := c.State.Health; health == nil || health.Status == types.NoHealthcheck || health.Status == types.Healthy {
		return &Resource{ID: c.ID, ports: c.NetworkSettings.Ports}, nil
	}

	msgCh, errCh := p.client.Events(ctx, types.EventsOptions{Filters: filters.NewArgs(
		filters.Arg("type", events.ContainerEventType),
		filters.Arg("container", resp.ID),
		filters.Arg("event", "health_status"),
	)})
	for {
		select {
		case msg := <-msgCh:
			if msg.Action == "health_status: healthy" {
				return &Resource{ID: resp.ID, ports: c.NetworkSettings.Ports}, nil
			}
		case err := <-errCh:
			p.Purge(context.Background(), &Resource{ID: resp.ID})
			return nil, fmt.Errorf("p.client.Events: %w", err)
		}
	}
}

func (p *Pool) Purge(ctx context.Context, r *Resource) error {
	if r == nil {
		return nil
	}
	if err := p.client.ContainerKill(ctx, r.ID, ""); err != nil {
		return fmt.Errorf("r.client.ContainerKill: %w", err)
	}
	return nil
}

func (r *Resource) GetHostPort(proto, port string) (string, error) {
	p, err := nat.NewPort(proto, port)
	if err != nil {
		return "", fmt.Errorf("nat.NewPort: %w", err)
	}
	m, ok := r.ports[p]
	if !ok || len(m) == 0 {
		return "", errors.New("port not found")
	}
	return net.JoinHostPort(m[0].HostIP, m[0].HostPort), nil
}
