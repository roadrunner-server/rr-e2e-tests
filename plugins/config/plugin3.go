package config

import (
	"context"
	"time"

	"github.com/roadrunner-server/errors"
)

type Foo3 struct {
	configProvider Configurer
}

// Depends on S2 and DB (S3 in the current case)
func (f *Foo3) Init(p Configurer) error {
	f.configProvider = p
	return nil
}

func (f *Foo3) Serve() chan error {
	const op = errors.Op("foo_plugin_serve")
	errCh := make(chan error, 1)

	if f.configProvider.GracefulTimeout() != time.Second*10 {
		errCh <- errors.E(op, errors.Str("GracefulTimeout should be eq to 10 seconds"))
		return errCh
	}

	return errCh
}

func (f *Foo3) Stop(context.Context) error {
	return nil
}
