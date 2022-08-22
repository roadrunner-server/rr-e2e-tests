//go:build nightly

package http

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/roadrunner-server/http/v2/handler"
	"github.com/roadrunner-server/http/v2/uploads"
	"github.com/roadrunner-server/sdk/v2/ipc/pipe"
	"github.com/roadrunner-server/sdk/v2/pool"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var mockLog = zap.NewNop() //nolint:gochecknoglobals

func TestHandler_Error(t *testing.T) {
	p, err := pool.NewStaticPool(context.Background(),
		func(cmd string) *exec.Cmd {
			return exec.Command("php", "../../php_test_files/http/client.php", "error", "pipes")
		},
		pipe.NewPipeFactory(mockLog),
		&pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	cfg := &config.CommonOptions{
		MaxRequestSize:    1024,
		InternalErrorCode: 500,
	}

	upldCfg := &uploads.Uploads{
		Dir:       os.TempDir(),
		Forbidden: map[string]struct{}{},
		Allowed:   map[string]struct{}{},
	}

	h, err := handler.NewHandler(cfg, upldCfg, p, mockLog)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	_, r, err := get("http://127.0.0.1:8177/?hello=world")
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()
	assert.Equal(t, 500, r.StatusCode)
}

func TestHandler_Error2(t *testing.T) {
	p, err := pool.NewStaticPool(context.Background(),
		func(cmd string) *exec.Cmd {
			return exec.Command("php", "../../php_test_files/http/client.php", "error2", "pipes")
		},
		pipe.NewPipeFactory(mockLog),
		&pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	cfg := &config.CommonOptions{
		MaxRequestSize:    1024,
		InternalErrorCode: 500,
	}

	upldCfg := &uploads.Uploads{
		Dir:       os.TempDir(),
		Forbidden: map[string]struct{}{},
		Allowed:   map[string]struct{}{},
	}

	h, err := handler.NewHandler(cfg, upldCfg, p, mockLog)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8178", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	_, r, err := get("http://127.0.0.1:8178/?hello=world")
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()
	assert.Equal(t, 500, r.StatusCode)
}

func TestHandler_ErrorDuration(t *testing.T) {
	p, err := pool.NewStaticPool(context.Background(),
		func(cmd string) *exec.Cmd {
			return exec.Command("php", "../../php_test_files/http/client.php", "error", "pipes")
		},
		pipe.NewPipeFactory(mockLog),
		&pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	cfg := &config.CommonOptions{
		MaxRequestSize:    1024,
		InternalErrorCode: 500,
	}

	upldCfg := &uploads.Uploads{
		Dir:       os.TempDir(),
		Forbidden: map[string]struct{}{},
		Allowed:   map[string]struct{}{},
	}

	h, err := handler.NewHandler(cfg, upldCfg, p, mockLog)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8182", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	_, r, err := get("http://127.0.0.1:8182/?hello=world")
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()

	assert.Equal(t, 500, r.StatusCode)
}
