package dockertest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPool(t *testing.T) {
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
	type args struct {
		opts RunOptions
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "no-healthcheck",
			args: args{
				opts: RunOptions{
					Image: "warashi/nginx:none",
				},
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(tt, err)
			},
		},
		{
			name: "success-healthcheck",
			args: args{
				opts: RunOptions{
					Image: "warashi/nginx:ok",
				},
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(tt, err)
			},
		},
		{
			name: "fail-healthcheck",
			args: args{
				opts: RunOptions{
					Image: "warashi/nginx:ng",
				},
			},
			assertion: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(tt, err)
			},
		},
		{
			name: "fail-pull",
			args: args{
				opts: RunOptions{
					Image: "warashi/nginx:notexist",
				},
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
			p, err := NewPool()
			require.NoError(t, err)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			t.Cleanup(cancel)
			got, err := p.Run(ctx, func(o *RunOptions) error { *o = tt.args.opts; return nil })
			t.Cleanup(func() { p.Purge(context.Background(), got) })
			tt.assertion(t, err)
		})
	}
}
