package memory

//
//import (
//	"net"
//	"net/rpc"
//	"os"
//	"os/signal"
//	"sync"
//	"syscall"
//	"testing"
//	"time"
//
//	jobState "github.com/roadrunner-server/api/v3/plugins/v1/jobs"
//	endure "github.com/roadrunner-server/endure/pkg/container"
//	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
//	rpcPlugin "github.com/roadrunner-server/rpc/v3"
//	mocklogger "github.com/roadrunner-server/rr-e2e-tests/mock"
//	helpers "github.com/roadrunner-server/rr-e2e-tests/plugins/jobs"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	jobsProto "go.buf.build/protocolbuffers/go/roadrunner-server/api/jobs/v1"
//	"go.uber.org/zap"
//)
//
//func TestMemoryInit(t *testing.T) {
//	cont := endure.New(slog.LevelDebug)
//	assert.NoError(t, err)
//
//	cfg := &config.Plugin{
//		Version: "2.9.0",
//		Path:    "configs/.rr-memory-init.yaml",
//		Prefix:  "rr",
//	}
//
//	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
//	err = cont.RegisterAll(
//		cfg,
//		&server.Plugin{},
//		&rpcPlugin.Plugin{},
//		l,
//		&jobs.Plugin{},
//		&resetter.Plugin{},
//		&informer.Plugin{},
//		&memory.Plugin{},
//	)
//	assert.NoError(t, err)
//
//	err = cont.Init()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ch, err := cont.Serve()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//
//	stopCh := make(chan struct{}, 1)
//
//	go func() {
//		defer wg.Done()
//		for {
//			select {
//			case e := <-ch:
//				assert.Fail(t, "error", e.Error.Error())
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//			case <-sig:
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			case <-stopCh:
//				// timeout
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			}
//		}
//	}()
//
//	time.Sleep(time.Second)
//	out := &jobState.State{}
//	t.Run("Stats", helpers.Stats(out))
//
//	assert.Equal(t, out.Active, int64(0))
//	assert.Equal(t, out.Delayed, int64(0))
//	assert.Equal(t, out.Reserved, int64(0))
//	assert.Equal(t, uint64(13), out.Priority)
//
//	stopCh <- struct{}{}
//	wg.Wait()
//
//	require.Equal(t, 1, oLogger.FilterMessageSnippet("plugin was started").Len())
//}
//
//func TestMemoryInitV27(t *testing.T) {
//	cont := endure.New(slog.LevelDebug)
//	assert.NoError(t, err)
//
//	cfg := &config.Plugin{
//		Path:    "configs/.rr-memory-init-v27.yaml",
//		Prefix:  "rr",
//		Version: "2.7.0",
//	}
//
//	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
//	err = cont.RegisterAll(
//		cfg,
//		&server.Plugin{},
//		&rpcPlugin.Plugin{},
//		l,
//		&jobs.Plugin{},
//		&resetter.Plugin{},
//		&informer.Plugin{},
//		&memory.Plugin{},
//	)
//	assert.NoError(t, err)
//
//	err = cont.Init()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ch, err := cont.Serve()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//
//	stopCh := make(chan struct{}, 1)
//
//	go func() {
//		defer wg.Done()
//		for {
//			select {
//			case e := <-ch:
//				assert.Fail(t, "error", e.Error.Error())
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//			case <-sig:
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			case <-stopCh:
//				// timeout
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			}
//		}
//	}()
//
//	time.Sleep(time.Second * 1)
//	t.Run("PushPipeline", helpers.PushToPipe("test-1", false))
//	t.Run("PushPipeline", helpers.PushToPipe("test-2", false))
//	time.Sleep(time.Second * 1)
//
//	stopCh <- struct{}{}
//	wg.Wait()
//
//	require.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was started").Len())
//	require.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
//	require.Equal(t, 2, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
//	require.Equal(t, 2, oLogger.FilterMessageSnippet("job processing was started").Len())
//}
//
//func TestMemoryInitV27BadResp(t *testing.T) {
//	cont := endure.New(slog.LevelDebug)
//	assert.NoError(t, err)
//
//	cfg := &config.Plugin{
//		Version: "2.9.0",
//		Path:    "configs/.rr-memory-init-v27-br.yaml",
//		Prefix:  "rr",
//	}
//
//	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
//	err = cont.RegisterAll(
//		cfg,
//		&server.Plugin{},
//		&rpcPlugin.Plugin{},
//		l,
//		&jobs.Plugin{},
//		&resetter.Plugin{},
//		&informer.Plugin{},
//		&memory.Plugin{},
//	)
//	assert.NoError(t, err)
//
//	err = cont.Init()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ch, err := cont.Serve()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//
//	stopCh := make(chan struct{}, 1)
//
//	go func() {
//		defer wg.Done()
//		for {
//			select {
//			case e := <-ch:
//				assert.Fail(t, "error", e.Error.Error())
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//			case <-sig:
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			case <-stopCh:
//				// timeout
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			}
//		}
//	}()
//
//	time.Sleep(time.Second * 1)
//	t.Run("PushPipeline", helpers.PushToPipe("test-1", false))
//	t.Run("PushPipeline", helpers.PushToPipe("test-2", false))
//	time.Sleep(time.Second * 1)
//
//	stopCh <- struct{}{}
//	wg.Wait()
//
//	require.Equal(t, 2, oLogger.FilterMessageSnippet("response handler error").Len())
//}
//
//func TestMemoryCreate(t *testing.T) {
//	t.Skip("not for the CI")
//	cont := endure.New(slog.LevelDebug)
//	assert.NoError(t, err)
//
//	cfg := &config.Plugin{
//		Version: "2.9.0",
//		Path:    "configs/.rr-memory-create.yaml",
//		Prefix:  "rr",
//	}
//
//	err = cont.RegisterAll(
//		cfg,
//		&server.Plugin{},
//		&rpcPlugin.Plugin{},
//		&logger.Plugin{},
//		&jobs.Plugin{},
//		&resetter.Plugin{},
//		&informer.Plugin{},
//		&memory.Plugin{},
//	)
//	assert.NoError(t, err)
//
//	err = cont.Init()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ch, err := cont.Serve()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//
//	stopCh := make(chan struct{}, 1)
//
//	go func() {
//		defer wg.Done()
//		for {
//			select {
//			case e := <-ch:
//				assert.Fail(t, "error", e.Error.Error())
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//			case <-sig:
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			case <-stopCh:
//				// timeout
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			}
//		}
//	}()
//
//	time.Sleep(time.Second * 5)
//	t.Run("PushPipeline", helpers.PushToPipe("example", false))
//	stopCh <- struct{}{}
//	wg.Wait()
//}
//
//func TestMemoryDeclare(t *testing.T) {
//	cont := endure.New(slog.LevelDebug)
//	assert.NoError(t, err)
//
//	cfg := &config.Plugin{
//		Version: "2.9.0",
//		Path:    "configs/.rr-memory-declare.yaml",
//		Prefix:  "rr",
//	}
//
//	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
//	err = cont.RegisterAll(
//		cfg,
//		&server.Plugin{},
//		&rpcPlugin.Plugin{},
//		l,
//		&jobs.Plugin{},
//		&resetter.Plugin{},
//		&informer.Plugin{},
//		&memory.Plugin{},
//	)
//	assert.NoError(t, err)
//
//	err = cont.Init()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ch, err := cont.Serve()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//
//	stopCh := make(chan struct{}, 1)
//
//	go func() {
//		defer wg.Done()
//		for {
//			select {
//			case e := <-ch:
//				assert.Fail(t, "error", e.Error.Error())
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//			case <-sig:
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			case <-stopCh:
//				// timeout
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			}
//		}
//	}()
//
//	time.Sleep(time.Second * 3)
//
//	t.Run("DeclarePipeline", declareMemoryPipe("10000"))
//	t.Run("ConsumePipeline", consumeMemoryPipe)
//	t.Run("PushPipeline", helpers.PushToPipe("test-3", false))
//	time.Sleep(time.Second)
//	t.Run("PausePipeline", helpers.PausePipelines("test-3"))
//	time.Sleep(time.Second)
//	t.Run("DestroyPipeline", helpers.DestroyPipelines("test-3"))
//
//	stopCh <- struct{}{}
//	wg.Wait()
//
//	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
//	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
//	require.Equal(t, 1, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
//	require.Equal(t, 1, oLogger.FilterMessageSnippet("job processing was started").Len())
//	require.Equal(t, 1, oLogger.FilterMessageSnippet("job was processed successfully").Len())
//}
//
//func TestMemoryPauseResume(t *testing.T) {
//	cont := endure.New(slog.LevelDebug)
//	assert.NoError(t, err)
//
//	cfg := &config.Plugin{
//		Version: "2.9.0",
//		Path:    "configs/.rr-memory-pause-resume.yaml",
//		Prefix:  "rr",
//	}
//
//	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
//	err = cont.RegisterAll(
//		cfg,
//		&server.Plugin{},
//		&rpcPlugin.Plugin{},
//		l,
//		&jobs.Plugin{},
//		&resetter.Plugin{},
//		&informer.Plugin{},
//		&memory.Plugin{},
//	)
//
//	assert.NoError(t, err)
//
//	err = cont.Init()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ch, err := cont.Serve()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//
//	stopCh := make(chan struct{}, 1)
//
//	go func() {
//		defer wg.Done()
//		for {
//			select {
//			case e := <-ch:
//				assert.Fail(t, "error", e.Error.Error())
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//			case <-sig:
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			case <-stopCh:
//				// timeout
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			}
//		}
//	}()
//
//	time.Sleep(time.Second * 3)
//
//	t.Run("Pause", helpers.PausePipelines("test-local"))
//	t.Run("pushToDisabledPipe", helpers.PushToDisabledPipe("test-local"))
//	t.Run("Resume", helpers.ResumePipes("test-local"))
//	t.Run("pushToEnabledPipe", helpers.PushToPipe("test-local", false))
//	time.Sleep(time.Second * 1)
//
//	stopCh <- struct{}{}
//	wg.Wait()
//
//	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
//	require.Equal(t, 3, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
//	require.Equal(t, 2, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
//	require.Equal(t, 2, oLogger.FilterMessageSnippet("job processing was started").Len())
//	require.Equal(t, 2, oLogger.FilterMessageSnippet("job was processed successfully").Len())
//}
//
//func TestMemoryJobsError(t *testing.T) {
//	cont := endure.New(slog.LevelDebug)
//	assert.NoError(t, err)
//
//	cfg := &config.Plugin{
//		Version: "2.9.0",
//		Path:    "configs/.rr-memory-jobs-err.yaml",
//		Prefix:  "rr",
//	}
//
//	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
//	err = cont.RegisterAll(
//		cfg,
//		&server.Plugin{},
//		&rpcPlugin.Plugin{},
//		l,
//		&jobs.Plugin{},
//		&resetter.Plugin{},
//		&informer.Plugin{},
//		&memory.Plugin{},
//	)
//	assert.NoError(t, err)
//
//	err = cont.Init()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ch, err := cont.Serve()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//
//	stopCh := make(chan struct{}, 1)
//
//	go func() {
//		defer wg.Done()
//		for {
//			select {
//			case e := <-ch:
//				assert.Fail(t, "error", e.Error.Error())
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//			case <-sig:
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			case <-stopCh:
//				// timeout
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			}
//		}
//	}()
//
//	time.Sleep(time.Second * 3)
//
//	t.Run("DeclarePipeline", declareMemoryPipe("10000"))
//	t.Run("ConsumePipeline", helpers.ResumePipes("test-3"))
//	t.Run("PushPipeline", helpers.PushToPipe("test-3", false))
//	time.Sleep(time.Second * 25)
//	t.Run("PausePipeline", helpers.PausePipelines("test-3"))
//	time.Sleep(time.Second)
//	t.Run("DestroyPipeline", helpers.DestroyPipelines("test-3"))
//
//	stopCh <- struct{}{}
//	wg.Wait()
//
//	assert.Equal(t, 1, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
//	assert.Equal(t, 4, oLogger.FilterMessageSnippet("job processing was started").Len())
//	assert.Equal(t, 4, oLogger.FilterMessageSnippet("job was processed successfully").Len())
//	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was paused").Len())
//	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
//	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
//	assert.Equal(t, 3, oLogger.FilterMessageSnippet("jobs protocol error").Len())
//}
//
//func TestMemoryStats(t *testing.T) {
//	cont := endure.New(slog.LevelDebug)
//	assert.NoError(t, err)
//
//	cfg := &config.Plugin{
//		Version: "2.9.0",
//		Path:    "configs/.rr-memory-declare.yaml",
//		Prefix:  "rr",
//	}
//
//	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
//	err = cont.RegisterAll(
//		cfg,
//		&server.Plugin{},
//		&rpcPlugin.Plugin{},
//		l,
//		&jobs.Plugin{},
//		&resetter.Plugin{},
//		&informer.Plugin{},
//		&memory.Plugin{},
//	)
//	assert.NoError(t, err)
//
//	err = cont.Init()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ch, err := cont.Serve()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//
//	stopCh := make(chan struct{}, 1)
//
//	go func() {
//		defer wg.Done()
//		for {
//			select {
//			case e := <-ch:
//				assert.Fail(t, "error", e.Error.Error())
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//			case <-sig:
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			case <-stopCh:
//				// timeout
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			}
//		}
//	}()
//
//	time.Sleep(time.Second * 3)
//
//	t.Run("DeclarePipeline", declareMemoryPipe("10000"))
//	t.Run("ConsumePipeline", consumeMemoryPipe)
//	t.Run("PushPipeline", helpers.PushToPipe("test-3", false))
//	time.Sleep(time.Second)
//	t.Run("PausePipeline", helpers.PausePipelines("test-3"))
//	time.Sleep(time.Second)
//
//	t.Run("PushPipeline", helpers.PushToPipeDelayed("test-3", 5))
//	t.Run("PushPipeline", helpers.PushToPipe("test-3", false))
//
//	time.Sleep(time.Second)
//	out := &jobState.State{}
//	t.Run("Stats", helpers.Stats(out))
//
//	assert.Equal(t, "test-3", out.Pipeline)
//	assert.Equal(t, "memory", out.Driver)
//	assert.Equal(t, "test-3", out.Queue)
//
//	assert.Equal(t, int64(0), out.Active)
//	assert.Equal(t, int64(1), out.Delayed)
//	assert.Equal(t, int64(0), out.Reserved)
//	assert.Equal(t, uint64(33), out.Priority)
//
//	time.Sleep(time.Second)
//	t.Run("ConsumePipeline", consumeMemoryPipe)
//	time.Sleep(time.Second * 7)
//
//	out = &jobState.State{}
//	t.Run("Stats", helpers.Stats(out))
//
//	assert.Equal(t, out.Pipeline, "test-3")
//	assert.Equal(t, out.Driver, "memory")
//	assert.Equal(t, out.Queue, "test-3")
//
//	assert.Equal(t, int64(0), out.Active)
//	assert.Equal(t, int64(0), out.Delayed)
//	assert.Equal(t, int64(0), out.Reserved)
//	assert.Equal(t, uint64(33), out.Priority)
//
//	t.Run("DestroyEphemeralPipeline", helpers.DestroyPipelines("test-3"))
//
//	stopCh <- struct{}{}
//	wg.Wait()
//
//	require.Equal(t, 3, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
//	require.Equal(t, 3, oLogger.FilterMessageSnippet("job processing was started").Len())
//	require.Equal(t, 3, oLogger.FilterMessageSnippet("job was processed successfully").Len())
//	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was paused").Len())
//	require.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
//	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
//}
//
//func TestMemoryPrefetch(t *testing.T) {
//	cont := endure.New(slog.LevelDebug)
//	assert.NoError(t, err)
//
//	cfg := &config.Plugin{
//		Version: "2.12.1",
//		Path:    "configs/.rr-memory-prefetch.yaml",
//		Prefix:  "rr",
//	}
//
//	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
//	err = cont.RegisterAll(
//		cfg,
//		&server.Plugin{},
//		&rpcPlugin.Plugin{},
//		l,
//		&jobs.Plugin{},
//		&resetter.Plugin{},
//		&informer.Plugin{},
//		&memory.Plugin{},
//	)
//	assert.NoError(t, err)
//
//	err = cont.Init()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	ch, err := cont.Serve()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//
//	stopCh := make(chan struct{}, 1)
//
//	go func() {
//		defer wg.Done()
//		for {
//			select {
//			case e := <-ch:
//				assert.Fail(t, "error", e.Error.Error())
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//			case <-sig:
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			case <-stopCh:
//				// timeout
//				err = cont.Stop()
//				if err != nil {
//					assert.FailNow(t, "error", err.Error())
//				}
//				return
//			}
//		}
//	}()
//
//	time.Sleep(time.Second * 3)
//
//	t.Run("DeclarePipeline", declareMemoryPipe("1"))
//	t.Run("ConsumePipeline", consumeMemoryPipe)
//	for i := 0; i < 10; i++ {
//		t.Run("PushPipeline", helpers.PushToPipe("test-3", false))
//	}
//
//	time.Sleep(time.Second * 15)
//
//	t.Run("DestroyEphemeralPipeline", helpers.DestroyPipelines("test-3"))
//
//	stopCh <- struct{}{}
//	wg.Wait()
//
//	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("prefetch limit was reached, waiting for the jobs to be processed").Len(), 1)
//	assert.Equal(t, 10, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
//	assert.Equal(t, 10, oLogger.FilterMessageSnippet("job processing was started").Len())
//	assert.Equal(t, 10, oLogger.FilterMessageSnippet("job was processed successfully").Len())
//	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
//	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
//	assert.Equal(t, 1, oLogger.FilterMessageSnippet("destroy signal received").Len())
//}
//
//func declareMemoryPipe(prefetch string) func(t *testing.T) {
//	return func(t *testing.T) {
//		conn, err := net.Dial("tcp", "127.0.0.1:6001")
//		assert.NoError(t, err)
//		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
//
//		pipe := &jobsProto.DeclareRequest{Pipeline: map[string]string{
//			"driver":   "memory",
//			"name":     "test-3",
//			"prefetch": prefetch,
//			"priority": "33",
//		}}
//
//		er := &jobsProto.Empty{}
//		err = client.Call("jobs.Declare", pipe, er)
//		assert.NoError(t, err)
//	}
//}
//
//func consumeMemoryPipe(t *testing.T) {
//	conn, err := net.Dial("tcp", "127.0.0.1:6001")
//	assert.NoError(t, err)
//	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
//
//	pipe := &jobsProto.Pipelines{Pipelines: make([]string, 1)}
//	pipe.GetPipelines()[0] = "test-3"
//
//	er := &jobsProto.Empty{}
//	err = client.Call("jobs.Resume", pipe, er)
//	assert.NoError(t, err)
//}
