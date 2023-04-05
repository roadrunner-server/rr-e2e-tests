package config

import (
	"context"

	"github.com/roadrunner-server/errors"
)

type Foo5 struct {
	configProvider Configurer
}

func (f *Foo5) Init(p Configurer) error {
	f.configProvider = p
	return nil
}

func (f *Foo5) Serve() chan error {
	const op = errors.Op("foo_plugin_serve")
	errCh := make(chan error, 1)

	r := &ReloadConfig{}
	err := f.configProvider.UnmarshalKey("reload", r)
	if err != nil {
		errCh <- err
	}

	var allCfg AllConfig
	err = f.configProvider.Unmarshal(&allCfg)
	if err != nil {
		errCh <- errors.E(op, errors.Str("should be at least one pattern, but got 0"))
		return errCh
	}

	if allCfg.RPC.Listen != "tcp://127.0.0.1:6001" {
		errCh <- errors.E(op, errors.Str("RPC.Listen should be overwritten"))
		return errCh
	}

	if allCfg.Logs.Mode != "development" {
		errCh <- errors.E(op, errors.Str("Logs.Mode failed to parse"))
		return errCh
	}

	if allCfg.Logs.Level != "" {
		errCh <- errors.E(op, errors.Str("Logs.Level failed to parse"))
		return errCh
	}

	return errCh
}

func (f *Foo5) Stop(context.Context) error {
	return nil
}
