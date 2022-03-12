package server

import (
	"github.com/roadrunner-server/api/v2/plugins/config"
	"github.com/roadrunner-server/api/v2/plugins/server"
	"github.com/roadrunner-server/api/v2/pool"
)

type Foo4 struct {
	configProvider config.Configurer
	wf             server.Server
	pool           pool.Pool
}

func (f *Foo4) Init(p config.Configurer, workerFactory server.Server) error {
	f.configProvider = p
	f.wf = workerFactory
	return nil
}

func (f *Foo4) Serve() chan error {
	errCh := make(chan error, 1)

	return errCh
}

func (f *Foo4) Stop() error {
	return nil
}
