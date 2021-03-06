package server

import (
	"context"
	"time"

	"github.com/roadrunner-server/api/v2/payload"
	"github.com/roadrunner-server/api/v2/plugins/config"
	"github.com/roadrunner-server/api/v2/plugins/server"
	"github.com/roadrunner-server/api/v2/pool"
	"github.com/roadrunner-server/errors"
	poolImpl "github.com/roadrunner-server/sdk/v2/pool"
	"github.com/roadrunner-server/sdk/v2/worker"
	serverImpl "github.com/roadrunner-server/server/v2"
)

const ConfigSection = "server"
const Response = "test"

var testPoolConfig = &poolImpl.Config{ //nolint:gochecknoglobals
	NumWorkers:      10,
	MaxJobs:         100,
	AllocateTimeout: time.Second * 10,
	DestroyTimeout:  time.Second * 10,
	Supervisor: &poolImpl.SupervisorConfig{
		WatchTick:       60 * time.Second,
		TTL:             1000 * time.Second,
		IdleTTL:         10 * time.Second,
		ExecTTL:         10 * time.Second,
		MaxWorkerMemory: 1000,
	},
}

type Foo struct {
	configProvider config.Configurer
	wf             server.Server
	pool           pool.Pool
}

func (f *Foo) Init(p config.Configurer, workerFactory server.Server) error {
	f.configProvider = p
	f.wf = workerFactory
	return nil
}

func (f *Foo) Serve() chan error {
	const op = errors.Op("serve")

	// test payload for echo
	r := &payload.Payload{
		Context: nil,
		Body:    []byte(Response),
	}

	errCh := make(chan error, 1)

	conf := &serverImpl.Config{}
	var err error
	err = f.configProvider.UnmarshalKey(ConfigSection, conf)
	if err != nil {
		errCh <- err
		return errCh
	}

	// test CMDFactory
	cmd := f.wf.CmdFactory(nil)
	if cmd == nil {
		errCh <- errors.E(op, errors.Str("command is nil"))
		return errCh
	}

	// test worker creation
	w, err := f.wf.NewWorker(context.Background(), nil)
	if err != nil {
		errCh <- err
		return errCh
	}

	go func() {
		_ = w.Wait()
	}()

	// test that our worker is functional
	sw := worker.From(w)

	rsp, err := sw.Exec(r)
	if err != nil {
		errCh <- err
		return errCh
	}

	if string(rsp.Body) != Response {
		errCh <- errors.E("response from worker is wrong", errors.Errorf("response: %s", rsp.Body))
		return errCh
	}

	// should not be errors
	err = sw.Stop()
	if err != nil {
		errCh <- err
		return errCh
	}

	// test pool
	f.pool, err = f.wf.NewWorkerPool(context.Background(), testPoolConfig, nil, nil)
	if err != nil {
		errCh <- err
		return errCh
	}

	// test pool execution
	rsp, err = f.pool.Exec(r)
	if err != nil {
		errCh <- err
		return errCh
	}

	// echo of the "test" should be -> test
	if string(rsp.Body) != Response {
		errCh <- errors.E("response from worker is wrong", errors.Errorf("response: %s", rsp.Body))
		return errCh
	}

	return errCh
}

func (f *Foo) Stop() error {
	return nil
}
