package dockertest

import (
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

type RunOptions struct {
	Platform         string
	ContainerName    string
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

type Option interface {
	Apply(*RunOptions)
}

type (
	RunOption        func(*RunOptions)
	ContainerOption  func(*container.Config)
	HealthOption     func(*container.HealthConfig)
	HostOption       func(*container.HostConfig)
	NetworkingOption func(*network.NetworkingConfig)
)

func (f RunOption) Apply(o *RunOptions) {
	f(o)
}

func (f ContainerOption) Apply(o *RunOptions) {
	if o.Config == nil {
		o.Config = new(container.Config)
	}
	f(o.Config)
}

func (f HostOption) Apply(o *RunOptions) {
	if o.HostConfig == nil {
		o.HostConfig = new(container.HostConfig)
	}
	f(o.HostConfig)
}

func (f HealthOption) Apply(o *RunOptions) {
	if o.Config == nil {
		o.Config = new(container.Config)
	}
	if o.Config.Healthcheck == nil {
		o.Config.Healthcheck = new(container.HealthConfig)
	}
	f(o.Config.Healthcheck)
}

func (f NetworkingOption) Apply(o *RunOptions) {
	if o.NetworkingConfig == nil {
		o.NetworkingConfig = new(network.NetworkingConfig)
	}
	f(o.NetworkingConfig)
}

func WithContainerName(name string) RunOption {
	return func(o *RunOptions) { o.ContainerName = name }
}

func WithPlatform(name string) RunOption {
	return func(o *RunOptions) { o.Platform = name }
}

func WithMount(m mount.Mount) HostOption {
	return func(o *container.HostConfig) { o.Mounts = append(o.Mounts, m) }
}

func WithTmpfs(target string) HostOption {
	return WithMount(mount.Mount{Type: mount.TypeTmpfs, Target: target})
}

func WithBind(target, source string) HostOption {
	return WithMount(mount.Mount{Type: mount.TypeBind, Target: target, Source: source})
}

func WithNoHealthcheck() HealthOption {
	return func(hc *container.HealthConfig) { hc.Test = []string{"NONE"} }
}

func WithHealthcheck(cmd string) HealthOption {
	return func(hc *container.HealthConfig) { hc.Test = []string{"CMD-SHELL", cmd} }
}

func WithHealthcheckInterval(d time.Duration) HealthOption {
	return func(hc *container.HealthConfig) { hc.Interval = d }
}

func WithCommand(cmd []string) ContainerOption {
	return func(c *container.Config) { c.Cmd = cmd }
}

func WithShellCommand(cmd string) ContainerOption {
	return func(c *container.Config) { c.Cmd = []string{"sh", "-c", cmd} }
}
