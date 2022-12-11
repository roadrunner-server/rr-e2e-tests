package app_logger //nolint:stylecheck

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	applogger "github.com/roadrunner-server/app-logger/v3"
	configImpl "github.com/roadrunner-server/config/v3"
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/logger/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppLogger(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Version = "2.12.0"

	err = container.RegisterAll(
		&applogger.Plugin{},
		&logger.Plugin{},
		vp,
	)

	require.NoError(t, err)

	err = container.Init()
	require.NoError(t, err)

	ch, err := container.Serve()
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
				err = container.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-sig:
				err = container.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = container.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)
	stopCh <- struct{}{}
	wg.Wait()
}
