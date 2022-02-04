package gzip

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/roadrunner-server/config/v2"
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/gzip/v2"
	httpPlugin "github.com/roadrunner-server/http/v2"
	"github.com/roadrunner-server/logger/v2"
	mock_logger "github.com/roadrunner-server/rr-e2e-tests/mock"
	"github.com/roadrunner-server/server/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGzipPlugin(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-http-withGzip.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
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

	time.Sleep(time.Second * 2)
	t.Run("GzipCheckHeader", headerCheck)

	stopCh <- struct{}{}
	wg.Wait()
}

func headerCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:18953", nil)
	assert.NoError(t, err)
	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: false,
		},
	}

	r, err := client.Do(req)
	require.NoError(t, err)
	require.True(t, r.Uncompressed)

	err = r.Body.Close()
	require.NoError(t, err)
}

func TestMiddlewareNotExist(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-http-middlewareNotExist.yaml",
		Prefix: "rr",
	}

	l, oLogger := mock_logger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
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
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("http server was started").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("requested middleware does not exist").Len())
}
