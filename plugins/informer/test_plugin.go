package informer

import (
	"context"
	"log"
	"os/exec"
	"time"

	"github.com/roadrunner-server/sdk/v4/payload"
	"github.com/roadrunner-server/sdk/v4/pool"
	staticPool "github.com/roadrunner-server/sdk/v4/pool/static_pool"
	"github.com/roadrunner-server/sdk/v4/state/process"
	"github.com/roadrunner-server/sdk/v4/worker"
	"go.uber.org/zap"
)

var testPoolConfig = &pool.Config{ //nolint:gochecknoglobals
	NumWorkers:      10,
	MaxJobs:         100,
	AllocateTimeout: time.Second * 10,
	DestroyTimeout:  time.Second * 10,
	Supervisor: &pool.SupervisorConfig{
		WatchTick:       60 * time.Second,
		TTL:             1000 * time.Second,
		IdleTTL:         10 * time.Second,
		ExecTTL:         10 * time.Second,
		MaxWorkerMemory: 1000,
	},
}

type Configurer interface {
	// UnmarshalKey takes a single key and unmarshal it into a Struct.
	UnmarshalKey(name string, out any) error
	// Has checks if config section exists.
	Has(name string) bool
}

// Server creates workers for the application.
type Server interface {
	CmdFactory(env map[string]string) func() *exec.Cmd
	NewPool(ctx context.Context, cfg *pool.Config, env map[string]string, _ *zap.Logger) (*staticPool.Pool, error)
	NewWorker(ctx context.Context, env map[string]string) (*worker.Process, error)
}

type Pool interface {
	// Workers returns worker list associated with the pool.
	Workers() (workers []*worker.Process)
	// Exec payload
	Exec(ctx context.Context, p *payload.Payload) (*payload.Payload, error)
	// Reset kill all workers inside the watcher and replaces with new
	Reset(ctx context.Context) error
	// Destroy all underlying stack (but let them complete the task).
	Destroy(ctx context.Context)
}

// Gauge //////////////

type Plugin1 struct {
	config Configurer
	server Server
}

func (p1 *Plugin1) Init(cfg Configurer, server Server) error {
	p1.config = cfg
	p1.server = server
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
	return "informer.plugin1"
}

func (p1 *Plugin1) Workers() []*process.State {
	p, err := p1.server.NewPool(context.Background(), testPoolConfig, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	ps := make([]*process.State, 0, len(p.Workers()))
	workers := p.Workers()
	for i := 0; i < len(workers); i++ {
		state, err := process.WorkerProcessState(workers[i])
		if err != nil {
			return nil
		}
		ps = append(ps, state)
	}

	return ps
}
