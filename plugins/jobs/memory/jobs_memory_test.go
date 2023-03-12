package memory

import (
	"io"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/goccy/go-json"
	jobState "github.com/roadrunner-server/api/v4/plugins/v1/jobs"
	"github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/roadrunner-server/informer/v4"
	"github.com/roadrunner-server/jobs/v4"
	"github.com/roadrunner-server/logger/v4"
	"github.com/roadrunner-server/memory/v4"
	"github.com/roadrunner-server/otel/v4"
	"github.com/roadrunner-server/resetter/v4"
	rpcPlugin "github.com/roadrunner-server/rpc/v4"
	mocklogger "github.com/roadrunner-server/rr-e2e-tests/mock"
	helpers "github.com/roadrunner-server/rr-e2e-tests/plugins/jobs"
	"github.com/roadrunner-server/server/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	jobsProto "go.buf.build/protocolbuffers/go/roadrunner-server/api/jobs/v1"
	"go.uber.org/zap"
	"golang.org/x/exp/slog"
)

func TestMemoryInit(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-memory-init.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&memory.Plugin{},
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

	time.Sleep(time.Second)
	out := &jobState.State{}
	t.Run("Stats", helpers.Stats("127.0.0.1:6001", out))

	assert.Equal(t, out.Active, int64(0))
	assert.Equal(t, out.Delayed, int64(0))
	assert.Equal(t, out.Reserved, int64(0))
	assert.Equal(t, uint64(13), out.Priority)

	helpers.DestroyPipelines("127.0.0.1:6001", "test-1", "test-2")

	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("plugin was started").Len())
}

func TestMemoryInitV27(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Path:    "configs/.rr-memory-init-v27.yaml",
		Prefix:  "rr",
		Version: "2.7.0",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&memory.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("PushPipeline", helpers.PushToPipe("test-1", false, "127.0.0.1:6001"))
	t.Run("PushPipeline", helpers.PushToPipe("test-2", false, "127.0.0.1:6001"))
	time.Sleep(time.Second * 1)

	helpers.DestroyPipelines("127.0.0.1:6001", "test-1", "test-2")

	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was started").Len())
	require.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
	require.Equal(t, 2, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	require.Equal(t, 2, oLogger.FilterMessageSnippet("job processing was started").Len())
}

func TestMemoryInitV27BadResp(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-memory-init-v27-br.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&memory.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("PushPipeline", helpers.PushToPipe("test-1", false, "127.0.0.1:6001"))
	t.Run("PushPipeline", helpers.PushToPipe("test-2", false, "127.0.0.1:6001"))
	time.Sleep(time.Second * 1)

	helpers.DestroyPipelines("127.0.0.1:6001", "test-1", "test-2")

	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 2, oLogger.FilterMessageSnippet("response handler error").Len())
}

func TestMemoryCreate(t *testing.T) {
	t.Skip("not for the CI")
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-memory-create.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&memory.Plugin{},
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

	time.Sleep(time.Second * 5)
	t.Run("PushPipeline", helpers.PushToPipe("example", false, "127.0.0.1:6001"))

	helpers.DestroyPipelines("127.0.0.1:6001", "test-1", "test-2")

	stopCh <- struct{}{}
	wg.Wait()
}

func TestMemoryDeclare(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-memory-declare.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&memory.Plugin{},
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

	t.Run("DeclarePipeline", declareMemoryPipe("10000"))
	t.Run("ConsumePipeline", consumeMemoryPipe)
	t.Run("PushPipeline", helpers.PushToPipe("test-3", false, "127.0.0.1:6001"))
	time.Sleep(time.Second)
	t.Run("PausePipeline", helpers.PausePipelines("127.0.0.1:6001", "test-3"))
	time.Sleep(time.Second)
	t.Run("DestroyPipeline", helpers.DestroyPipelines("127.0.0.1:6001", "test-3"))

	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("job processing was started").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("job was processed successfully").Len())
}

func TestMemoryPauseResume(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-memory-pause-resume.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&memory.Plugin{},
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

	t.Run("Pause", helpers.PausePipelines("127.0.0.1:6001", "test-local"))
	t.Run("pushToDisabledPipe", helpers.PushToDisabledPipe("127.0.0.1:6001", "test-local"))
	t.Run("Resume", helpers.ResumePipes("127.0.0.1:6001", "test-local"))
	t.Run("pushToEnabledPipe", helpers.PushToPipe("test-local", false, "127.0.0.1:6001"))
	time.Sleep(time.Second * 1)

	helpers.DestroyPipelines("127.0.0.1:6001", "test-local")

	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
	require.Equal(t, 3, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
	require.Equal(t, 2, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	require.Equal(t, 2, oLogger.FilterMessageSnippet("job processing was started").Len())
	require.Equal(t, 2, oLogger.FilterMessageSnippet("job was processed successfully").Len())
}

func TestMemoryJobsError(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-memory-jobs-err.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&memory.Plugin{},
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

	t.Run("DeclarePipeline", declareMemoryPipe("10000"))
	t.Run("ConsumePipeline", helpers.ResumePipes("127.0.0.1:6001", "test-3"))
	t.Run("PushPipeline", helpers.PushToPipe("test-3", false, "127.0.0.1:6001"))
	time.Sleep(time.Second * 25)
	t.Run("PausePipeline", helpers.PausePipelines("127.0.0.1:6001", "test-3"))
	time.Sleep(time.Second)
	t.Run("DestroyPipeline", helpers.DestroyPipelines("127.0.0.1:6001", "test-3"))

	stopCh <- struct{}{}
	wg.Wait()

	assert.Equal(t, 1, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	assert.Equal(t, 4, oLogger.FilterMessageSnippet("job processing was started").Len())
	assert.Equal(t, 4, oLogger.FilterMessageSnippet("job was processed successfully").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was paused").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
	assert.Equal(t, 3, oLogger.FilterMessageSnippet("jobs protocol error").Len())
}

func TestMemoryStats(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-memory-declare.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&memory.Plugin{},
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

	t.Run("DeclarePipeline", declareMemoryPipe("10000"))
	t.Run("ConsumePipeline", consumeMemoryPipe)
	t.Run("PushPipeline", helpers.PushToPipe("test-3", false, "127.0.0.1:6001"))
	time.Sleep(time.Second)
	t.Run("PausePipeline", helpers.PausePipelines("127.0.0.1:6001", "test-3"))
	time.Sleep(time.Second)

	t.Run("PushPipeline", helpers.PushToPipeDelayed("127.0.0.1:6001", "test-3", 5))
	t.Run("PushPipeline", helpers.PushToPipe("test-3", false, "127.0.0.1:6001"))

	time.Sleep(time.Second)
	out := &jobState.State{}
	t.Run("Stats", helpers.Stats("127.0.0.1:6001", out))

	assert.Equal(t, "test-3", out.Pipeline)
	assert.Equal(t, "memory", out.Driver)
	assert.Equal(t, "test-3", out.Queue)

	assert.Equal(t, int64(0), out.Active)
	assert.Equal(t, int64(1), out.Delayed)
	assert.Equal(t, int64(0), out.Reserved)
	assert.Equal(t, uint64(33), out.Priority)

	time.Sleep(time.Second)
	t.Run("ConsumePipeline", consumeMemoryPipe)
	time.Sleep(time.Second * 7)

	out = &jobState.State{}
	t.Run("Stats", helpers.Stats("127.0.0.1:6001", out))

	assert.Equal(t, out.Pipeline, "test-3")
	assert.Equal(t, out.Driver, "memory")
	assert.Equal(t, out.Queue, "test-3")

	assert.Equal(t, int64(0), out.Active)
	assert.Equal(t, int64(0), out.Delayed)
	assert.Equal(t, int64(0), out.Reserved)
	assert.Equal(t, uint64(33), out.Priority)

	t.Run("DestroyPipeline", helpers.DestroyPipelines("127.0.0.1:6001", "test-3"))

	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 3, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	require.Equal(t, 3, oLogger.FilterMessageSnippet("job processing was started").Len())
	require.Equal(t, 3, oLogger.FilterMessageSnippet("job was processed successfully").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was paused").Len())
	require.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
}

func TestMemoryPrefetch(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.12.1",
		Path:    "configs/.rr-memory-prefetch.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&memory.Plugin{},
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

	t.Run("DeclarePipeline", declareMemoryPipe("1"))
	t.Run("ConsumePipeline", consumeMemoryPipe)
	for i := 0; i < 10; i++ {
		t.Run("PushPipeline", helpers.PushToPipe("test-3", false, "127.0.0.1:6001"))
	}

	time.Sleep(time.Second * 15)

	t.Run("DestroyPipeline", helpers.DestroyPipelines("127.0.0.1:6001", "test-3"))

	stopCh <- struct{}{}
	wg.Wait()

	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("prefetch limit was reached, waiting for the jobs to be processed").Len(), 1)
	assert.Equal(t, 10, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	assert.Equal(t, 10, oLogger.FilterMessageSnippet("job processing was started").Len())
	assert.Equal(t, 10, oLogger.FilterMessageSnippet("job was processed successfully").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("destroy signal received").Len())
}

func TestMemoryTracer(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2023.1.0",
		Path:    "configs/.rr-memory-tracer.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&otel.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&memory.Plugin{},
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

	time.Sleep(time.Second)
	t.Run("PushPipeline", helpers.PushToPipe("test-1", false, "127.0.0.1:6001"))
	time.Sleep(time.Second * 2)

	t.Run("DestroyPipeline", helpers.DestroyPipelines("127.0.0.1:6001", "test-1"))

	stopCh <- struct{}{}
	wg.Wait()

	resp, err := http.Get("http://127.0.0.1:9411/api/v2/spans?serviceName=rr_test")
	assert.NoError(t, err)

	buf, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var spans []string
	err = json.Unmarshal(buf, &spans)
	assert.NoError(t, err)

	sort.Slice(spans, func(i, j int) bool {
		return spans[i] < spans[j]
	})

	expected := []string{
		"destroy_pipeline",
		"in_memory_listener",
		"in_memory_push",
		"in_memory_stop",
		"jobs_listener",
		"push",
	}

	assert.Equal(t, expected, spans)

	require.Equal(t, 1, oLogger.FilterMessageSnippet("plugin was started").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("job processing was started").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())

	t.Cleanup(func() {
		_ = resp.Body.Close()
	})
}

func declareMemoryPipe(prefetch string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		pipe := &jobsProto.DeclareRequest{Pipeline: map[string]string{
			"driver":   "memory",
			"name":     "test-3",
			"prefetch": prefetch,
			"priority": "33",
		}}

		er := &jobsProto.Empty{}
		err = client.Call("jobs.Declare", pipe, er)
		assert.NoError(t, err)
	}
}

func consumeMemoryPipe(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	pipe := &jobsProto.Pipelines{Pipelines: make([]string, 1)}
	pipe.GetPipelines()[0] = "test-3"

	er := &jobsProto.Empty{}
	err = client.Call("jobs.Resume", pipe, er)
	assert.NoError(t, err)
}
