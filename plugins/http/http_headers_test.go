package http

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
	httpPlugin "github.com/roadrunner-server/http/v2"
	"github.com/roadrunner-server/logger/v2"
	"github.com/roadrunner-server/server/v2"
	"github.com/stretchr/testify/assert"
)

func TestXFF(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/headers/.rr-http-xff.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
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

	req, err := http.NewRequest("GET", "http://127.0.0.1:12311?hello=world", nil)
	assert.NoError(t, err)
	req.Header.Add("X-Forwarded-For", "127.0.0.1")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)

	err = r.Body.Close()
	assert.NoError(t, err)

	// ---

	req, err = http.NewRequest("GET", "http://127.0.0.1:12311?hello=world", nil)
	assert.NoError(t, err)
	req.Header.Add("X-Forwarded-For", "foo.workstation")

	r, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, r.StatusCode)

	err = r.Body.Close()
	assert.NoError(t, err)

	// ---

	req, err = http.NewRequest("GET", "http://127.0.0.1:12311?hello=world", nil)
	assert.NoError(t, err)
	req.Header.Add("X-Forwarded-For", "9.10.11.12")

	r, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, r.StatusCode)

	err = r.Body.Close()
	assert.NoError(t, err)

	stopCh <- struct{}{}
	wg.Wait()
}

func TestForwarded(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/headers/.rr-http-f.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
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

	req, err := http.NewRequest("GET", "http://127.0.0.1:12811?hello=world", nil)
	assert.NoError(t, err)
	req.Header.Add("Forwarded", "foo.workstation")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)

	err = r.Body.Close()
	assert.NoError(t, err)

	// --

	req, err = http.NewRequest("GET", "http://127.0.0.1:12811?hello=world", nil)
	assert.NoError(t, err)
	req.Header.Add("Forwarded", "by=foo;for=127.0.0.1;host=foo.workstation;proto=http")

	r, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)

	err = r.Body.Close()
	assert.NoError(t, err)

	// --

	req, err = http.NewRequest("GET", "http://127.0.0.1:12811?hello=world", nil)
	assert.NoError(t, err)
	req.Header.Add("Forwarded", "by=foo;for=127.0.0.1;host=foo.workstation;proto=http")

	r, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)

	err = r.Body.Close()
	assert.NoError(t, err)

	// --

	req, err = http.NewRequest("GET", "http://127.0.0.1:12811?hello=world", nil)
	assert.NoError(t, err)
	req.Header.Add("Forwarded", "by=foo;for=3.11.0.1;host=foo.workstation;proto=http")

	r, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, r.StatusCode)

	err = r.Body.Close()
	assert.NoError(t, err)

	stopCh <- struct{}{}
	wg.Wait()
}
