package general

import (
	"io"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/roadrunner-server/amqp/v4"
	jobsProto "github.com/roadrunner-server/api/v4/build/jobs/v1"
	"github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/roadrunner-server/informer/v4"
	"github.com/roadrunner-server/jobs/v4"
	"github.com/roadrunner-server/logger/v4"
	"github.com/roadrunner-server/memory/v4"
	"github.com/roadrunner-server/metrics/v4"
	"github.com/roadrunner-server/resetter/v4"
	rpcPlugin "github.com/roadrunner-server/rpc/v4"
	helpers "github.com/roadrunner-server/rr-e2e-tests/plugins/jobs"
	"github.com/roadrunner-server/server/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestJobsInit(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-jobs-init.yaml",
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
		&amqp.Plugin{},
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
	t.Run("DestroyPipeline", helpers.DestroyPipelines("127.0.0.1:6001", "test-local", "test-local-2", "test-local-3", "test-1", "test-2-amqp", "test-3"))

	stopCh <- struct{}{}
	wg.Wait()
}

func TestJOBSMetrics(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Prefix:  "rr",
		Path:    "configs/.rr-jobs-metrics.yaml",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&server.Plugin{},
		&jobs.Plugin{},
		&logger.Plugin{},
		&metrics.Plugin{},
		&memory.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	tt := time.NewTimer(time.Minute * 3)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer tt.Stop()
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
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)

	t.Run("DeclareEphemeralPipeline", declareMemoryPipe)
	t.Run("ConsumeEphemeralPipeline", consumeMemoryPipe)

	genericOut, err := get()
	assert.NoError(t, err)

	assert.Contains(t, genericOut, `rr_jobs_jobs_err 0`)
	assert.Contains(t, genericOut, `rr_jobs_jobs_ok 0`)
	assert.Contains(t, genericOut, `rr_jobs_push_err 0`)
	assert.Contains(t, genericOut, `rr_jobs_push_ok 0`)
	assert.Contains(t, genericOut, `workers_memory_bytes`)
	assert.Contains(t, genericOut, `state="ready"}`)
	assert.Contains(t, genericOut, `{pid=`)
	assert.Contains(t, genericOut, `rr_jobs_total_workers 1`)
	assert.NotContains(t, genericOut, `rr_jobs_requests_total`)
	assert.NotContains(t, genericOut, `rr_jobs_push_latency`)

	t.Run("PushInMemoryPipeline", helpers.PushToPipe("test-3", false, "127.0.0.1:6001"))
	time.Sleep(time.Second)
	t.Run("PushInMemoryPipeline", helpers.PushToPipeDelayed("127.0.0.1:6001", "test-3", 5))
	time.Sleep(time.Second)
	t.Run("PushInMemoryPipeline", helpers.PushToPipe("test-3", false, "127.0.0.1:6001"))
	time.Sleep(time.Second * 5)

	genericOut, err = get()
	assert.NoError(t, err)

	assert.Contains(t, genericOut, `rr_jobs_jobs_err 0`)
	assert.Contains(t, genericOut, `rr_jobs_jobs_ok 3`)
	assert.Contains(t, genericOut, `rr_jobs_push_err 0`)
	assert.Contains(t, genericOut, `rr_jobs_push_ok 3`)
	assert.Contains(t, genericOut, `rr_jobs_requests_total{driver="memory",job="test-3",source="single"} 3`)
	assert.NotContains(t, genericOut, `rr_jobs_requests_total{driver="memory",job="test-3",source="batch"}`)

	t.Run("PushInMemoryPipeline", helpers.PushToPipeBatch("127.0.0.1:6001", "test-3", 2, false))
	t.Run("PushInMemoryPipeline", helpers.PushToPipeBatch("127.0.0.1:6001", "test-3", 5, false))

	time.Sleep(time.Second)

	genericOut, err = get()
	assert.NoError(t, err)

	assert.Contains(t, genericOut, `rr_jobs_jobs_err 0`)
	assert.Contains(t, genericOut, `rr_jobs_jobs_ok 10`)
	assert.Contains(t, genericOut, `rr_jobs_push_err 0`)
	assert.Contains(t, genericOut, `rr_jobs_push_ok 10`)
	assert.Contains(t, genericOut, `rr_jobs_requests_total{driver="memory",job="test-3",source="single"} 3`)
	assert.Contains(t, genericOut, `rr_jobs_requests_total{driver="memory",job="test-3",source="batch"} 7`)
	assert.Contains(t, genericOut, `rr_jobs_push_latency_bucket{driver="memory",job="test-3"`)

	t.Run("DestroyPipeline", helpers.DestroyPipelines("127.0.0.1:6001", "test-3"))

	close(sig)

	wg.Wait()
}

const getAddr = "http://127.0.0.1:2112/metrics"

// get request and return body
func get() (string, error) {
	r, err := http.Get(getAddr)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	err = r.Body.Close()
	if err != nil {
		return "", err
	}
	// unsafe
	return string(b), err
}

func declareMemoryPipe(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	pipe := &jobsProto.DeclareRequest{Pipeline: map[string]string{
		"driver":   "memory",
		"name":     "test-3",
		"prefetch": "10000",
	}}

	er := &jobsProto.Empty{}
	err = client.Call("jobs.Declare", pipe, er)
	assert.NoError(t, err)
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
