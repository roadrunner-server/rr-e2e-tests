package informer

import (
	"context"
	"sync"
	"time"

	"github.com/roadrunner-server/api/v2/plugins/config"
	"github.com/roadrunner-server/api/v2/plugins/server"
	"github.com/roadrunner-server/api/v2/pool"
	"github.com/roadrunner-server/api/v2/state/process"
	processImpl "github.com/roadrunner-server/sdk/v2/state/process"
)

// Gauge //////////////
type Plugin2 struct {
	sync.Mutex

	config config.Configurer
	server server.Server

	pool pool.Pool
}

func (p2 *Plugin2) Init(cfg config.Configurer, server server.Server) error {
	p2.config = cfg
	p2.server = server
	return nil
}

func (p2 *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		time.Sleep(time.Second * 5)
		p2.Lock()
		defer p2.Unlock()
		var err error
		p2.pool, err = p2.server.NewWorkerPool(context.Background(), testPoolConfig, nil, nil)
		if err != nil {
			panic(err)
		}
	}()
	return errCh
}

func (p2 *Plugin2) Stop() error {
	return nil
}

func (p2 *Plugin2) Name() string {
	return "informer.plugin2"
}

func (p2 *Plugin2) Workers() []process.State {
	if p2.pool == nil {
		return nil
	}
	ps := make([]process.State, 0, len(p2.pool.Workers()))
	workers := p2.pool.Workers()
	for i := 0; i < len(workers); i++ {
		state, err := processImpl.WorkerProcessState(workers[i])
		if err != nil {
			return nil
		}
		ps = append(ps, state)
	}

	return ps
}
