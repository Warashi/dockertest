package dockertest

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(tt, err)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewPool()
			tt.assertion(t, err)
		})
	}
}

func TestPool_Run(t *testing.T) {
	t.Parallel()

	pool, err := NewPool()
	require.NoError(t, err)
	type args struct {
		image string
		opts  []Option
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "no-healthcheck",
			args: args{
				image: "warashi/nginx:none",
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(tt, err)
			},
		},
		{
			name: "success-healthcheck",
			args: args{
				image: "warashi/nginx:ok",
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(tt, err)
			},
		},
		{
			name: "fail-healthcheck",
			args: args{
				image: "warashi/nginx:ng",
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(tt, err)
			},
		},
		{
			name: "fail-pull",
			args: args{
				image: "warashi/nginx:notexist",
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(tt, err)
			},
		},
		{
			name: "with-platform",
			args: args{
				image: "warashi/nginx:ok",
				opts:  []Option{WithPlatform("amd64")},
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(tt, err)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			t.Cleanup(cancel)
			got, err := pool.Run(ctx, tt.args.image, tt.args.opts...)
			t.Cleanup(func() { pool.Purge(context.Background(), got) })
			tt.assertion(t, err)
		})
	}
}

func TestPool_Purge(t *testing.T) {
	t.Parallel()

	type fields struct {
		client *client.Client
	}
	type args struct {
		ctx context.Context
		r   *Resource
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &Pool{
				client: tt.fields.client,
			}
			tt.assertion(t, p.Purge(tt.args.ctx, tt.args.r))
		})
	}
}

func TestResource_GetHostPort(t *testing.T) {
	t.Parallel()

	pool, err := NewPool()
	require.NoError(t, err)
	type args struct {
		proto string
		port  string
	}
	tests := []struct {
		name      string
		image     string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:  "found",
			image: "nginx:latest",
			args: args{
				proto: "tcp",
				port:  "80",
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(tt, err)
			},
		},
		{
			name:  "not-found/8080/tcp",
			image: "nginx:latest",
			args: args{
				proto: "tcp",
				port:  "8080",
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(tt, err)
			},
		},
		{
			name:  "not-found/80/udp",
			image: "nginx:latest",
			args: args{
				proto: "udp",
				port:  "8080",
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(tt, err)
			},
		},
		{
			name:  "protocol-error",
			image: "nginx:latest",
			args: args{
				proto: "http",
				port:  "80",
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(tt, err)
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			t.Cleanup(cancel)
			r, err := pool.Run(ctx, tt.image)
			require.NoError(t, err)
			t.Cleanup(func() { pool.Purge(context.Background(), r) })
			_, err = r.GetHostPort(tt.args.proto, tt.args.port)
			tt.assertion(t, err)
		})
	}
}
