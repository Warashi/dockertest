package dockertest

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type Pool struct {
	client *client.Client
}

type RunOptions struct {
	Image       string
	Cmd         []string
	Mounts      []mount.Mount
	Healthcheck *container.HealthConfig
	Platform    *v1.Platform
}

type Option func(*RunOptions) error

type Resource struct {
	ID string
}

func NewPool() (*Pool, error) {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("client.NewClientWithOpts: %w", err)
	}
	return &Pool{client: c}, nil
}

func (p *Pool) Run(ctx context.Context, opts ...Option) (*Resource, error) {
	opt := new(RunOptions)
	for i := range opts {
		if err := opts[i](opt); err != nil {
			return nil, fmt.Errorf("opts[%d]: %w", i, err)
		}
	}
	//var platform string
	//if opt.Platform != nil {
	//	platform = opt.Platform.Architecture
	//}
	//r, err := p.client.ImagePull(ctx, opt.Image, types.ImagePullOptions{Platform: platform})
	//if err != nil {
	//	return nil, fmt.Errorf("p.client.ImagePull: %w", err)
	//}
	//io.Copy(io.Discard, r)

	resp, err := p.client.ContainerCreate(ctx,
		&container.Config{
			Image:       opt.Image,
			Cmd:         opt.Cmd,
			Healthcheck: opt.Healthcheck,
		},
		&container.HostConfig{
			Mounts:     opt.Mounts,
			AutoRemove: true,
		},
		&network.NetworkingConfig{},
		opt.Platform,
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("p.client.ContainerCreate: %w", err)
	}

	if err := p.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("p.client.ContainerStart: %w", err)
	}

	if opt.Healthcheck == nil {
		return &Resource{ID: resp.ID}, nil
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
				return &Resource{ID: resp.ID}, nil
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
