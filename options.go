package dockertest

import (
	"github.com/docker/docker/api/types/container"
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
