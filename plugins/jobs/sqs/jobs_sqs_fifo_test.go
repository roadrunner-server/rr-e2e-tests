package sqs

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/roadrunner-server/config/v2"
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/informer/v2"
	"github.com/roadrunner-server/jobs/v2"
	"github.com/roadrunner-server/logger/v2"
	"github.com/roadrunner-server/resetter/v2"
	rpcPlugin "github.com/roadrunner-server/rpc/v2"
	mocklogger "github.com/roadrunner-server/rr-e2e-tests/mock"
	helpers "github.com/roadrunner-server/rr-e2e-tests/plugins/jobs"
	"github.com/roadrunner-server/server/v2"
	"github.com/roadrunner-server/sqs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSQSInitFifo(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-sqs-init_fifo.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&sqs.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)
	t.Run("PushPipelineFifo", helpers.PushToPipe("test-1", false))
	t.Run("PushPipelineFifo", helpers.PushToPipe("test-2", false))
	time.Sleep(time.Second * 2)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestSQSInitFifoAutoAck(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	cfg := &config.Plugin{
		Version: "2.9.2",
		Path:    "configs/.rr-sqs-init_fifo.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&sqs.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)
	t.Run("PushPipelineFifo", helpers.PushToPipe("test-1", true))
	t.Run("PushPipelineFifo", helpers.PushToPipe("test-2", true))
	time.Sleep(time.Second * 2)
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 2, oLogger.FilterMessageSnippet("auto ack is turned on, message acknowledged").Len())
}

func TestSQSInitV27BadRespFifo(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:    "configs/.rr-sqs-init-v27-br_fifo.yaml",
		Prefix:  "rr",
		Version: "2.7.0",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&sqs.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)
	t.Run("PushPipelineFifo", helpers.PushToPipe("test-1", false))
	t.Run("PushPipelineFifo", helpers.PushToPipe("test-2", false))
	time.Sleep(time.Second)

	stopCh <- struct{}{}
	wg.Wait()

	require.GreaterOrEqual(t, oLogger.FilterMessageSnippet("response handler error").Len(), 2)
}

func TestSQSDeclareFifo(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-sqs-declare_fifo.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&sqs.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)

	t.Run("DeclarePipelineFifo", declareSQSPipeFifo("default-decl.fifo"))
	t.Run("ConsumePipelineFifo", helpers.ResumePipes("test-3"))
	t.Run("PushPipelineFifo", helpers.PushToPipe("test-3", false))
	time.Sleep(time.Second)
	t.Run("PausePipelineFifo", helpers.PausePipelines("test-3"))
	time.Sleep(time.Second)
	t.Run("DestroyPipelineFifo", helpers.DestroyPipelines("test-3"))

	time.Sleep(time.Second * 5)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestSQSJobsErrorFifo(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-sqs-jobs-err_fifo.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&sqs.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)

	t.Run("DeclarePipelineFifo", declareSQSPipeFifo("default-err.fifo"))
	t.Run("ConsumePipelineFifo", helpers.ResumePipes("test-3"))
	t.Run("PushPipelineFifo", helpers.PushToPipe("test-3", false))
	time.Sleep(time.Second * 25)
	t.Run("PausePipelineFifo", helpers.PausePipelines("test-3"))
	time.Sleep(time.Second)
	t.Run("DestroyPipelineFifo", helpers.DestroyPipelines("test-3"))

	time.Sleep(time.Second * 5)
	stopCh <- struct{}{}
	wg.Wait()

	time.Sleep(time.Second * 5)
}

func TestSQSRespondFifo(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-sqs-respond_fifo.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&sqs.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)

	t.Run("DeclarePipelineFifo", declareSQSPipe("default"))
	t.Run("ConsumePipelineFifo", helpers.ResumePipes("test-3"))
	t.Run("PushPipelineFifo", helpers.PushToPipe("test-3", false))
	time.Sleep(time.Second)
	t.Run("DestroyPipelineFifo", helpers.DestroyPipelines("test-3"))
	t.Run("DestroyPipelineFifo", helpers.DestroyPipelines("test-1"))

	time.Sleep(time.Second * 5)
	stopCh <- struct{}{}
	wg.Wait()
}
