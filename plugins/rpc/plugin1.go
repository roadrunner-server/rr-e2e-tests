package rpc

import (
	"context"
	"fmt"
)

type Plugin1 struct {
	config Configurer
}

type Configurer interface {
	// UnmarshalKey takes a single key and unmarshal it into a Struct.
	UnmarshalKey(name string, out any) error
	// Has checks if config section exists.
	Has(name string) bool
}

func (p1 *Plugin1) Init(cfg Configurer) error {
	p1.config = cfg
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop(context.Context) error {
	return nil
}

func (p1 *Plugin1) Name() string {
	return "rpc_test.plugin1"
}

func (p1 *Plugin1) RPC() any {
	return &PluginRPC{srv: p1}
}

type PluginRPC struct {
	srv *Plugin1
}

func (r *PluginRPC) Hello(in string, out *string) error {
	*out = fmt.Sprintf("Hello, username: %s", in)
	return nil
}
