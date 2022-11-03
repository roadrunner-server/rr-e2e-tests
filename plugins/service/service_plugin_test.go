//go:build linux

package service

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/roadrunner-server/config/v3"
	endure "github.com/roadrunner-server/endure/pkg/container"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/roadrunner-server/informer/v3"
	"github.com/roadrunner-server/logger/v3"
	"github.com/roadrunner-server/reload/v3"
	"github.com/roadrunner-server/resetter/v3"
	rpcPlugin "github.com/roadrunner-server/rpc/v3"
	mocklogger "github.com/roadrunner-server/rr-e2e-tests/mock"
	"github.com/roadrunner-server/sdk/v3/state/process"
	"github.com/roadrunner-server/service/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	serviceProto "go.buf.build/protocolbuffers/go/roadrunner-server/api/service/v1"
	"go.uber.org/zap"
)

func TestServiceInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-init.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
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
				return
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

	time.Sleep(time.Second * 10)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceTrimOutput(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-newlines.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&service.Plugin{},
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
				return
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
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("stdout write").Len())
}

func TestServiceWorkers(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-workers.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
		&informer.Plugin{},
		&rpcPlugin.Plugin{},
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
				return
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
	t.Run("workers", workers("service"))
	time.Sleep(time.Second)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceInitStdout(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-init-stdout.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
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
				return
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
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceEnv(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-env.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&service.Plugin{},
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
				return
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
	stopCh <- struct{}{}
	wg.Wait()
	require.Equal(t, 0, oLogger.FilterMessageSnippet("faillll").Len())
}

func TestServiceError(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-error.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)
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
				return
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

	time.Sleep(time.Second * 10)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceRestarts(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-restarts.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
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
				return
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

	time.Sleep(time.Second * 10)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceCreate(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-create.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
		&rpcPlugin.Plugin{},
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
				return
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

	in := &serviceProto.Create{
		Name:            "foo",
		Command:         "php test_files/loop.php",
		ProcessNum:      1,
		ExecTimeout:     3,
		RemainAfterExit: false,
		Env:             map[string]string{"foo": "bar"},
		RestartSec:      0,
	}

	out := &serviceProto.Response{}

	t.Run("create", create(in, out))

	time.Sleep(time.Second * 2)

	out = &serviceProto.Response{}
	t.Run("terminate", terminate(&serviceProto.Service{Name: "foo"}, out))

	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceCreateEmptyConfig(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-create-empty.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
		&rpcPlugin.Plugin{},
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
				return
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

	in := &serviceProto.Create{
		Name:            "foo",
		Command:         "php test_files/loop.php",
		ProcessNum:      1,
		ExecTimeout:     3,
		RemainAfterExit: false,
		Env:             map[string]string{"foo": "bar"},
		RestartSec:      0,
	}

	out := &serviceProto.Response{}

	t.Run("create", create(in, out))

	time.Sleep(time.Second * 3)
	l := &serviceProto.List{}
	t.Run("list", list(&serviceProto.Service{}, l))

	for i := 0; i < len(l.GetServices()); i++ {
		cmd := &serviceProto.Service{
			Name: l.GetServices()[i],
		}

		out = &serviceProto.Response{}

		t.Run("terminate", terminate(cmd, out))
	}

	time.Sleep(time.Second * 2)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceRestart(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-create-empty.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
		&rpcPlugin.Plugin{},
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
				return
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

	in := &serviceProto.Create{
		Name:            "foo",
		Command:         "php test_files/loop.php",
		ProcessNum:      1,
		ExecTimeout:     3,
		RemainAfterExit: false,
		Env:             map[string]string{"foo": "bar"},
		RestartSec:      0,
	}

	out := &serviceProto.Response{}

	t.Run("create", create(in, out))

	time.Sleep(time.Second * 3)
	l := &serviceProto.List{}
	t.Run("list", list(&serviceProto.Service{}, l))

	for i := 0; i < len(l.GetServices()); i++ {
		cmd := &serviceProto.Service{
			Name: l.GetServices()[i],
		}

		out = &serviceProto.Response{}

		t.Run("restart", restart(cmd, out))

		time.Sleep(time.Second * 5)
	}

	time.Sleep(time.Second * 2)
	out = &serviceProto.Response{}
	t.Run("terminate", terminate(&serviceProto.Service{Name: "foo"}, out))
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceRestartConcurrent(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-create-empty.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
		&rpcPlugin.Plugin{},
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
				return
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

	in := &serviceProto.Create{
		Name:            "foo",
		Command:         "php test_files/loop.php",
		ProcessNum:      1,
		ExecTimeout:     3,
		RemainAfterExit: false,
		Env:             map[string]string{"foo": "bar"},
		RestartSec:      0,
	}

	out := &serviceProto.Response{}

	t.Run("create", create(in, out))

	time.Sleep(time.Second)

	l := &serviceProto.List{}
	t.Run("list", list(nil, l))

	for jj := 0; jj < 100; jj++ {
		go func() {
			for i := 0; i < len(l.GetServices()); i++ {
				cmd := &serviceProto.Service{
					Name: l.GetServices()[i],
				}

				out1 := &serviceProto.Response{}

				t.Run("restart", restart(cmd, out1))

				time.Sleep(time.Second)
			}
		}()

		go func() {
			for i := 0; i < len(l.GetServices()); i++ {
				cmd := &serviceProto.Service{
					Name: l.GetServices()[i],
				}

				out2 := &serviceProto.Response{}

				t.Run("restart", restart(cmd, out2))

				time.Sleep(time.Second)
			}
		}()
	}

	time.Sleep(time.Second * 10)
	out = &serviceProto.Response{}
	t.Run("terminate", terminate(&serviceProto.Service{Name: "foo"}, out))
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceListConcurrent(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*5))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-create-empty.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
		&rpcPlugin.Plugin{},
	)
	assert.NoError(t, err)

	require.NoError(t, cont.Init())

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
				return
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

	in := &serviceProto.Create{
		Name:            "foo",
		Command:         "php test_files/loop.php",
		ProcessNum:      1,
		ExecTimeout:     3,
		RemainAfterExit: false,
		Env:             map[string]string{"foo": "bar"},
		RestartSec:      0,
	}

	out := &serviceProto.Response{}

	t.Run("create", create(in, out))

	time.Sleep(time.Second)

	l := &serviceProto.List{}
	t.Run("list", list(nil, l))

	for jj := 0; jj < 100; jj++ {
		go func() {
			for i := 0; i < len(l.GetServices()); i++ {
				cmd := &serviceProto.Service{
					Name: l.GetServices()[i],
				}

				out1 := &serviceProto.Response{}

				t.Run("restart", restart(cmd, out1))
				ll := &serviceProto.List{}
				t.Run("list", list(nil, ll))
				require.Len(t, ll.GetServices(), 1)

				time.Sleep(time.Millisecond * 100)
			}
		}()

		go func() {
			for i := 0; i < len(l.GetServices()); i++ {
				cmd := &serviceProto.Service{
					Name: l.GetServices()[i],
				}

				out2 := &serviceProto.Response{}
				t.Run("restart", restart(cmd, out2))
				ll := &serviceProto.List{}
				t.Run("list", list(nil, ll))
				require.Len(t, ll.GetServices(), 1)

				time.Sleep(time.Millisecond * 100)
			}
		}()
	}

	time.Sleep(time.Second * 15)
	out = &serviceProto.Response{}
	t.Run("terminate", terminate(&serviceProto.Service{Name: "foo"}, out))
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceStatus(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-create-empty.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
		&rpcPlugin.Plugin{},
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
				return
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

	in := &serviceProto.Create{
		Name:            "foo",
		Command:         "php test_files/loop.php",
		ProcessNum:      1,
		ExecTimeout:     10,
		RemainAfterExit: false,
		Env:             map[string]string{"foo": "bar"},
		RestartSec:      0,
	}

	out := &serviceProto.Response{}

	t.Run("create", create(in, out))

	time.Sleep(time.Second)

	l := &serviceProto.List{}
	t.Run("list", list(nil, l))
	require.Len(t, l.GetServices(), 1)

	inStat := &serviceProto.Service{
		Name: l.GetServices()[0],
	}

	outStat := &serviceProto.Statuses{}
	t.Run("stats", status(inStat, outStat))
	require.NotEmpty(t, outStat.Status[0].GetCommand())
	require.NotZero(t, outStat.Status[0].GetMemoryUsage())
	require.NotZero(t, outStat.Status[0].GetPid())

	out = &serviceProto.Response{}
	t.Run("terminate", terminate(&serviceProto.Service{Name: l.GetServices()[0]}, out))

	time.Sleep(time.Second * 2)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceInitRemain(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-init-remain.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&service.Plugin{},
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
				return
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

	time.Sleep(time.Second * 10)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestServiceReset(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-reset.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&resetter.Plugin{},
		l,
		&service.Plugin{},
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
				return
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

	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	var ok bool
	err = client.Call("resetter.Reset", "service", &ok)
	require.NoError(t, err)
	require.True(t, ok)

	time.Sleep(time.Second * 5)
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 20, oLogger.FilterMessageSnippet("The number is: 0").Len())
	require.Equal(t, 20, oLogger.FilterMessageSnippet("Hello 0").Len())
}

func TestServiceReset2(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-reload.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&reload.Plugin{},
		&rpcPlugin.Plugin{},
		&resetter.Plugin{},
		&service.Plugin{},
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
				return
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

	file, err := os.Create("foo.txt")
	require.NoError(t, err)

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		require.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		var ok bool
		err = client.Call("resetter.Reset", "service", &ok)
		require.NoError(t, err)
		require.True(t, ok)
	}()

	time.Sleep(time.Second * 10)
	stopCh <- struct{}{}
	wg.Wait()

	assert.LessOrEqual(t, oLogger.FilterMessageSnippet("The number is: 0").Len(), 30)
	assert.LessOrEqual(t, oLogger.FilterMessageSnippet("Hello 0").Len(), 30)

	t.Cleanup(func() {
		_ = file.Close()
		_ = os.Remove("foo.txt")
	})
}

func TestServiceReset3(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-reload.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&rpcPlugin.Plugin{},
		&resetter.Plugin{},
		&reload.Plugin{},
		&service.Plugin{},
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
				return
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

	file, err := os.Create("foo.txt")
	require.NoError(t, err)

	time.Sleep(time.Second * 4)
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 20, oLogger.FilterMessageSnippet("The number is: 0").Len())
	require.Equal(t, 20, oLogger.FilterMessageSnippet("Hello 0").Len())

	t.Cleanup(func() {
		_ = file.Close()
		_ = os.Remove("foo.txt")
	})
}

func TestServiceReset4(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-service-reload.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&rpcPlugin.Plugin{},
		&resetter.Plugin{},
		&reload.Plugin{},
		&service.Plugin{},
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
				return
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

	file, err := os.Create("foo2.txt")
	require.NoError(t, err)

	go func() {
		l1 := &serviceProto.List{}
		t.Run("list", list(nil, l1))
		require.Len(t, l1.GetServices(), 2)
	}()

	go func() {
		l2 := &serviceProto.List{}
		t.Run("list", list(nil, l2))
		require.Len(t, l2.GetServices(), 2)
		for i := 0; i < len(l2.GetServices()); i++ {
			cmd := &serviceProto.Service{
				Name: l2.GetServices()[i],
			}

			out := &serviceProto.Response{}

			t.Run("terminate", terminate(cmd, out))
		}
	}()

	time.Sleep(time.Second * 10)
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 20, oLogger.FilterMessageSnippet("service have started").Len())
	t.Cleanup(func() {
		_ = file.Close()
		_ = os.Remove("foo2.txt")
	})
}

func create(in *serviceProto.Create, out *serviceProto.Response) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		require.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		err = client.Call("service.Create", in, out)
		require.NoError(t, err)
	}
}

func terminate(in *serviceProto.Service, out *serviceProto.Response) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		require.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		err = client.Call("service.Terminate", in, out)
		require.NoError(t, err)
	}
}

func restart(in *serviceProto.Service, out *serviceProto.Response) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		require.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		err = client.Call("service.Restart", in, out)
		require.NoError(t, err)
	}
}

func status(in *serviceProto.Service, out *serviceProto.Statuses) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		require.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		err = client.Call("service.Status", in, out)
		require.NoError(t, err)
	}
}

func list(in *serviceProto.Service, out *serviceProto.List) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		require.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		err = client.Call("service.List", in, out)
		require.NoError(t, err)
	}
}

func workers(service string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		require.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		// WorkerList contains list of workers.
		lst := struct {
			// Workers is list of workers.
			Workers []process.State `json:"workers"`
		}{}

		err = client.Call("informer.Workers", service, &lst)
		require.NoError(t, err)
		require.Len(t, lst.Workers, 20)
	}
}
