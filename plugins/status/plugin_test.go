package status

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

	"log/slog"

	statusv1beta1 "github.com/roadrunner-server/api/v4/build/status/v1"
	"github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	httpPlugin "github.com/roadrunner-server/http/v4"
	"github.com/roadrunner-server/jobs/v4"
	"github.com/roadrunner-server/logger/v4"
	"github.com/roadrunner-server/memory/v4"
	rpcPlugin "github.com/roadrunner-server/rpc/v4"
	"github.com/roadrunner-server/server/v4"
	"github.com/roadrunner-server/status/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const resp = `plugin: http, status: 200
plugin: rpc not found`

func TestStatusHttp(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-status-init.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&status.Plugin{},
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
	t.Run("CheckerGetStatus", checkHTTPStatus)

	stopCh <- struct{}{}
	wg.Wait()
}

func TestStatusRPC(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-status-init.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&status.Plugin{},
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
	t.Run("CheckerGetStatusRpc", func(t *testing.T) {
		checkRPCStatus(t, "http", 200, "6005")
	})
	stopCh <- struct{}{}
	wg.Wait()
}

func TestReadyHttp(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-status-init.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&status.Plugin{},
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
	t.Run("CheckerGetReadiness", checkHTTPReadiness)

	stopCh <- struct{}{}
	wg.Wait()
}

func TestReadinessRPCWorkerNotReady(t *testing.T) {
	cont := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Second))

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-ready-init.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&status.Plugin{},
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
				// timeout, error here is OK, because in the PHP we are sleeping for the 300s
				_ = cont.Stop()
				return
			}
		}
	}()

	time.Sleep(time.Second)
	t.Run("DoHttpReq", doHTTPReq)
	time.Sleep(time.Second * 5)
	t.Run("CheckerGetReadiness2", checkHTTPReadiness2)
	t.Run("CheckerGetRpcReadiness", func(t *testing.T) {
		checkRPCReadiness(t, "http", 503, "6006")
	})
	stopCh <- struct{}{}
	wg.Wait()
}

func TestJobsStatus(t *testing.T) {
	cont := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Second))

	cfg := &config.Plugin{
		Version: "2023.1.0",
		Path:    "configs/.rr-jobs-status.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&server.Plugin{},
		&status.Plugin{},
		&jobs.Plugin{},
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
				// timeout, error here is OK, because in the PHP we are sleeping for the 300s
				_ = cont.Stop()
				return
			}
		}
	}()

	time.Sleep(time.Second)

	req, err := http.NewRequest("GET", "http://127.0.0.1:35544/jobs", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	require.NotNil(t, r)

	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, r.StatusCode)

	assert.Contains(t, string(b), "test-1")
	assert.Contains(t, string(b), "test-2")

	err = r.Body.Close()
	assert.NoError(t, err)

	stopCh <- struct{}{}
	wg.Wait()
}

func TestJobsReadiness(t *testing.T) {
	cont := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Second))

	cfg := &config.Plugin{
		Version: "2023.2.0",
		Path:    "configs/.rr-jobs-status.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&server.Plugin{},
		&status.Plugin{},
		&jobs.Plugin{},
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
				// timeout, error here is OK, because in the PHP we are sleeping for the 300s
				_ = cont.Stop()
				return
			}
		}
	}()

	time.Sleep(time.Second)
	t.Run("checkJobsReadiness", checkJobsReadiness)
	t.Run("checkJobsRPC", func(t *testing.T) {
		checkRPCReadiness(t, "jobs", 200, "6007")
		checkRPCStatus(t, "jobs", 200, "6007")
	})

	stopCh <- struct{}{}
	wg.Wait()
}

func checkJobsReadiness(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:35544/ready?plugin=jobs", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "plugin: jobs, status: 200\n", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func checkHTTPStatus(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:34333/health?plugin=http&plugin=rpc", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, resp, string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func doHTTPReq(t *testing.T) {
	go func() {
		client := &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", "http://127.0.0.1:11933", nil)
		assert.NoError(t, err)

		_, err = client.Do(req) //nolint:bodyclose
		assert.Error(t, err)
	}()
}

func checkHTTPReadiness2(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:34334/ready?plugin=http&plugin=rpc", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 503, r.StatusCode)
	assert.Equal(t, "plugin: http, status: 503\n", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func checkHTTPReadiness(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:34333/ready?plugin=http&plugin=rpc", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, resp, string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func checkRPCReadiness(t *testing.T, plugin string, status int64, port string) {
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &statusv1beta1.Request{
		Plugin: plugin,
	}

	rsp := &statusv1beta1.Response{}

	err = client.Call("status.Ready", req, rsp)
	assert.NoError(t, err)
	assert.Equal(t, rsp.GetCode(), status)
}

func checkRPCStatus(t *testing.T, plugin string, status int64, port string) {
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &statusv1beta1.Request{
		Plugin: plugin,
	}

	rsp := &statusv1beta1.Response{}

	err = client.Call("status.Status", req, rsp)
	assert.NoError(t, err)
	assert.Equal(t, rsp.Code, status)
}
