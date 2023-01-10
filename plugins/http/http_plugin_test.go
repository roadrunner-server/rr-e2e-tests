package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/rpc"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/roadrunner-server/gzip/v4"
	httpPlugin "github.com/roadrunner-server/http/v4"
	"github.com/roadrunner-server/informer/v4"
	"github.com/roadrunner-server/logger/v4"
	"github.com/roadrunner-server/resetter/v4"
	rpcPlugin "github.com/roadrunner-server/rpc/v4"
	mocklogger "github.com/roadrunner-server/rr-e2e-tests/mock"
	"github.com/roadrunner-server/sdk/v4/state/process"
	"github.com/roadrunner-server/send/v4"
	"github.com/roadrunner-server/server/v4"
	"github.com/roadrunner-server/static/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yookoala/gofast"
	"go.uber.org/zap"
	"golang.org/x/exp/slog"
	"golang.org/x/net/http2"
)

func TestHTTPInit(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-init.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
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
	stopCh <- struct{}{}
	wg.Wait()
}

func TestHTTPAccessLogs(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-access-logs.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
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

	t.Run("AccessLogsEcho", echoAccessLogs)

	stopCh <- struct{}{}
	wg.Wait()
}

func echoAccessLogs(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:58332", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NotNil(t, r)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "hello world", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func TestHTTPXSendFile(t *testing.T) {
	cont := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Minute))

	cfg := &config.Plugin{
		Version: "2.12.2",
		Path:    "configs/http/.rr-http-sendfile.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&send.Plugin{},
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
	t.Run("X-Sendfile", xsendfile)
	stopCh <- struct{}{}
	wg.Wait()
}

func xsendfile(t *testing.T) {
	parsedURL, _ := url.Parse("http://127.0.0.1:44444")
	client := http.Client{}
	pwd, _ := os.Getwd()
	req := &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	b, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	require.True(t, len(b) > 0)
	require.Empty(t, resp.Header.Get("X-Sendfile"))

	file, err := os.ReadFile(fmt.Sprintf("%s/../../php_test_files/well", pwd))
	require.NoError(t, err)
	assert.True(t, len(b) == len(file))
	require.NoError(t, resp.Body.Close())
	_, _ = io.Discard.Write(file)
}

func TestHTTPNoConfigSection(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-no-http.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
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
	stopCh <- struct{}{}
	wg.Wait()
}

func TestHTTPInformerReset(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-resetter.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&send.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&informer.Plugin{},
		&resetter.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("HTTPInformerTest", informerTest("127.0.0.1:6008"))
	t.Run("HTTPEchoTestBefore", echoHTTP)
	t.Run("HTTPResetTest", resetTest("127.0.0.1:6008"))
	t.Run("HTTPEchoTestAfter", echoHTTP)

	stopCh <- struct{}{}

	wg.Wait()
}

func TestSSL(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-ssl.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&send.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("SSLEcho", sslEcho)
	t.Run("SSLNoRedirect", sslNoRedirect)
	t.Run("FCGEcho", fcgiEcho)

	stopCh <- struct{}{}
	wg.Wait()
}

func sslNoRedirect(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("../../test-certs/localhost+2-client.pem", "../../test-certs/localhost+2-client-key.pem")
	require.NoError(t, err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	req, err := http.NewRequest("GET", "http://127.0.0.1:8085?hello=world", nil)
	assert.NoError(t, err)

	r, err := client.Do(req)
	assert.NoError(t, err)

	assert.Nil(t, r.TLS)

	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
}

func sslEcho(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("../../test-certs/localhost+2-client.pem", "../../test-certs/localhost+2-client-key.pem")
	require.NoError(t, err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	req, err := http.NewRequest("GET", "https://127.0.0.1:8893?hello=world", nil)
	assert.NoError(t, err)

	r, err := client.Do(req)
	assert.NoError(t, err)

	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
}

func fcgiEcho(t *testing.T) {
	fcgiConnFactory := gofast.SimpleConnFactory("tcp", "127.0.0.1:16920")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/?hello=world", nil)
	fcgiHandler.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Result().Body) //nolint:bodyclose

	defer func() {
		_ = w.Result().Body.Close()
		w.Body.Reset()
	}()

	assert.NoError(t, err)
	assert.Equal(t, 201, w.Result().StatusCode) //nolint:bodyclose
	assert.Equal(t, "WORLD", string(body))
}

func TestSSLRedirect(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-ssl-redirect.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&send.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("SSLRedirect", sslRedirect)

	stopCh <- struct{}{}
	wg.Wait()
}

func sslRedirect(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("../../test-certs/localhost+2-client.pem", "../../test-certs/localhost+2-client-key.pem")
	require.NoError(t, err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	req, err := http.NewRequest("GET", "http://127.0.0.1:8087?hello=world", nil)
	assert.NoError(t, err)

	r, err := client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, r.TLS)

	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
}

func TestSSLPushPipes(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-ssl-push.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("SSLPush", sslPush)

	stopCh <- struct{}{}
	wg.Wait()
}

func sslPush(t *testing.T) {
	cert, err := tls.LoadX509KeyPair("../../test-certs/localhost+2-client.pem", "../../test-certs/localhost+2-client-key.pem")
	require.NoError(t, err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	req, err := http.NewRequest("GET", "https://127.0.0.1:8894?hello=world", nil)
	assert.NoError(t, err)

	r, err := client.Do(req)
	assert.NoError(t, err)

	assert.NotNil(t, r.TLS)

	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, "", r.Header.Get("Http2-Release"))

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
}

func TestFastCGI_Echo(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-fcgi.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&static.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("FastCGIEcho", fcgiEcho1)

	stopCh <- struct{}{}
	wg.Wait()
}

func fcgiEcho1(t *testing.T) {
	time.Sleep(time.Second * 2)
	fcgiConnFactory := gofast.SimpleConnFactory("tcp", "127.0.0.1:6920")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/hello-world", nil)
	fcgiHandler.ServeHTTP(w, req)

	_, err := io.ReadAll(w.Result().Body) //nolint:bodyclose
	assert.NoError(t, err)
	assert.Equal(t, 200, w.Result().StatusCode) //nolint:bodyclose
}

func TestFastCGI_EchoUnix(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-fcgi-unix.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&static.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("FastCGIEcho", fcgiEchoUnix)

	stopCh <- struct{}{}
	wg.Wait()

	t.Cleanup(func() {
		_ = os.RemoveAll("rr.sock")
	})
}

func fcgiEchoUnix(t *testing.T) {
	time.Sleep(time.Second * 2)
	fcgiConnFactory := gofast.SimpleConnFactory("unix", "rr.sock")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/hello-world", nil)
	fcgiHandler.ServeHTTP(w, req)

	_, err := io.ReadAll(w.Result().Body) //nolint:bodyclose
	assert.NoError(t, err)
	assert.Equal(t, 200, w.Result().StatusCode) //nolint:bodyclose
}

func TestFastCGI_RequestUri(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-fcgi-reqUri.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
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

	time.Sleep(time.Second * 1)
	t.Run("FastCGIServiceRequestUri", fcgiReqURI)

	stopCh <- struct{}{}
	wg.Wait()
}

func fcgiReqURI(t *testing.T) {
	time.Sleep(time.Second * 2)
	fcgiConnFactory := gofast.SimpleConnFactory("tcp", "127.0.0.1:6921")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/hello-world", nil)
	fcgiHandler.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Result().Body) //nolint:bodyclose
	assert.NoError(t, err)
	assert.Equal(t, 200, w.Result().StatusCode) //nolint:bodyclose
	assert.Contains(t, string(body), "ddddd")
}

func TestHTTP2Req(t *testing.T) {
	cont := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Second*5))

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-h2-ssl.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		l,
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

	time.Sleep(time.Second * 1)

	cert, err := tls.LoadX509KeyPair("../../test-certs/localhost+2-client.pem", "../../test-certs/localhost+2-client-key.pem")
	require.NoError(t, err)

	tr := &http2.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}

	client := &http.Client{
		Transport:     tr,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}

	req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1:23452?hello=world", nil)
	require.NoError(t, err)

	r, err := client.Do(req)
	require.NoError(t, err)

	assert.Equal(t, r.StatusCode, http.StatusCreated)
	data, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	require.Equal(t, data, []byte("WORLD"))
	require.NoError(t, r.Body.Close())

	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("http server was started").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("http log").Len())
}

func TestH2CUpgrade(t *testing.T) {
	cont := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Second*5))

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-h2c.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		l,
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

	time.Sleep(time.Second * 1)

	req, err := http.NewRequest("PRI", "http://127.0.0.1:8083", nil)
	require.NoError(t, err)

	req.Header.Add("Upgrade", "h2c")
	req.Header.Add("Connection", "HTTP2-Settings")
	req.Header.Add("Connection", "Upgrade")
	req.Header.Add("HTTP2-Settings", "AAMAAABkAARAAAAAAAIAAAAA")

	r, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, "101 Switching Protocols", r.Status)
	require.NoError(t, r.Body.Close())

	assert.Equal(t, http.StatusSwitchingProtocols, r.StatusCode)

	resp, err := http.Get("http://127.0.0.1:8083?hello=world")
	require.NoError(t, err)
	require.Equal(t, 1, resp.ProtoMajor)
	_ = resp.Body.Close()

	time.Sleep(time.Second * 2)
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("http server was started").Len())
}

func TestH2C(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-h2c.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		l,
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

	tr := &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			// use the http dial (w/o tls)
			return net.Dial(network, addr)
		},
	}
	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8083?hello=world", nil)
	require.NoError(t, err)

	r, err := client.Do(req)
	require.NoError(t, err)

	assert.Equal(t, "201 Created", r.Status)
	data, err := io.ReadAll(r.Body)
	require.NoError(t, err)

	require.Equal(t, []byte("WORLD"), data)

	require.Equal(t, 2, r.ProtoMajor)
	require.NoError(t, r.Body.Close())

	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("http server was started").Len())
}

func TestHttpMiddleware(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&PluginMiddleware{},
		&PluginMiddleware2{},
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

	time.Sleep(time.Second * 1)
	t.Run("MiddlewareTest", middleware)

	stopCh <- struct{}{}
	wg.Wait()
}

func middleware(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:18903?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)

	req, err = http.NewRequest("GET", "http://127.0.0.1:18903/halt", nil)
	assert.NoError(t, err)

	r, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err = io.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 500, r.StatusCode)
	assert.Equal(t, "halted", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func TestHttpEchoErr(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	rIn := `
rpc:
  listen: tcp://127.0.0.1:6003
  disabled: false

server:
  command: "php ../../php_test_files/http/client.php echoerr pipes"
  relay: "pipes"
  relay_timeout: "20s"

http:
  debug: true
  address: 127.0.0.1:34999
  max_request_size: 1024
  middleware: [ "pluginMiddleware", "pluginMiddleware2" ]
  uploads:
    forbid: [ "" ]
  trusted_subnets: [ "10.0.0.0/8", "127.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "::1/128", "fc00::/7", "fe80::/10" ]
  pool:
    num_workers: 2
    max_jobs: 0
    allocate_timeout: 60s
    destroy_timeout: 60s

logs:
  mode: development
  level: debug
`

	cfg := &config.Plugin{
		Version:   "2.9.0",
		Type:      "yaml",
		ReadInCfg: []byte(rIn),
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&PluginMiddleware{},
		&PluginMiddleware2{},
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

	time.Sleep(time.Second * 3)

	t.Run("HttpEchoError", echoError)

	stopCh <- struct{}{}
	wg.Wait()
}

func echoError(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:34999?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	require.NotNil(t, r)

	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))
	err = r.Body.Close()
	assert.NoError(t, err)
}

func TestHttpBrokenPipes(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-broken-pipes.yaml",
		Prefix:  "rr",
		Type:    "yaml",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&PluginMiddleware{},
		&PluginMiddleware2{},
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
			// should be error from the plugin
			case e := <-ch:
				assert.Error(t, e.Error)
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

	wg.Wait()
}

func TestHTTPSupervisedPool(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-supervised-pool.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&informer.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("HTTPEchoRunActivateWorker", echoHTTP2)
	// bigger timeout to handle idle_ttl on slow systems
	time.Sleep(time.Second * 10)
	t.Run("HTTPInformerCompareWorkersTestBefore", informerTestBefore)
	t.Run("HTTPEchoShouldBeNewWorker", echoHTTP2)
	// worker should be destructed (idle_ttl)
	t.Run("HTTPInformerCompareWorkersTestAfter", informerTestAfter)

	stopCh <- struct{}{}
	wg.Wait()
}

func echoHTTP2(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:18888?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

// get worker
// sleep
// supervisor destroy worker
// compare pid's
var workerPid = 0 //nolint:gochecknoglobals

func informerTestBefore(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:15432")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	list := struct {
		// Workers is list of workers.
		Workers []process.State `json:"workers"`
	}{}

	err = client.Call("informer.Workers", "http", &list)
	assert.NoError(t, err)
	assert.Len(t, list.Workers, 1)
	// save the pid
	workerPid = int(list.Workers[0].Pid)
}

func informerTestAfter(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:15432")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	list := struct {
		// Workers is list of workers.
		Workers []process.State `json:"workers"`
	}{}

	assert.NotZero(t, workerPid)

	time.Sleep(time.Second * 5)

	err = client.Call("informer.Workers", "http", &list)
	assert.NoError(t, err)
	assert.Len(t, list.Workers, 1)
	assert.NotEqual(t, workerPid, list.Workers[0].Pid)
}

func TestHTTPBigRequestSize(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-big-req-size.yaml",
		Prefix:  "rr",
		Type:    "yaml",
	}

	err := cont.RegisterAll(
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

	buf := make([]byte, 1024*1024*10)

	_, err = rand.Read(buf)
	assert.NoError(t, err)

	bt := bytes.NewBuffer(buf)

	req, err := http.NewRequest("GET", "http://127.0.0.1:10085?hello=world", bt)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
	assert.Equal(t, "serve_http: http: request body too large\n", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)

	stopCh <- struct{}{}
	wg.Wait()
}

func TestStaticEtagPlugin(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-static.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
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
	t.Run("ServeSampleEtag", serveStaticSampleEtag)
	t.Run("NoStaticHeaders", noStaticHeaders)

	stopCh <- struct{}{}
	wg.Wait()
}

func serveStaticSampleEtag(t *testing.T) {
	// OK 200 response
	b, r, err := get("http://127.0.0.1:21603/sample.txt")
	assert.NoError(t, err)
	assert.Contains(t, b, "sample")
	assert.Equal(t, r.StatusCode, http.StatusOK)
	etag := r.Header.Get("Etag")

	_ = r.Body.Close()

	// Should be 304 response with same etag
	c := http.Client{
		Timeout: time.Second * 5,
	}

	parsedURL, _ := url.Parse("http://127.0.0.1:21603/sample.txt")

	req := &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
		Header: map[string][]string{"If-None-Match": {etag}},
	}

	resp, err := c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotModified, resp.StatusCode)
	_ = resp.Body.Close()
}

// regular request should not contain static headers
func noStaticHeaders(t *testing.T) {
	// OK 200 response
	_, r, err := get("http://127.0.0.1:21603")
	assert.NoError(t, err)
	assert.NotContains(t, r.Header["input"], "custom-header")  //nolint:staticcheck
	assert.NotContains(t, r.Header["output"], "output-header") //nolint:staticcheck
	assert.Equal(t, r.StatusCode, http.StatusOK)

	_ = r.Body.Close()
}

func TestStaticPluginSecurity(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-static-security.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
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
	t.Run("ServeSampleNotAllowedPath", serveStaticSampleNotAllowedPath)

	stopCh <- struct{}{}
	wg.Wait()
}

func serveStaticSampleNotAllowedPath(t *testing.T) {
	// Should be 304 response with same etag
	c := http.Client{
		Timeout: time.Second * 5,
	}

	parsedURL := &url.URL{
		Scheme: "http",
		User:   nil,
		Host:   "127.0.0.1:21603",
		Path:   "%2e%2e%/tests/",
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
	}

	resp, err := c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()

	parsedURL = &url.URL{
		Scheme: "http",
		User:   nil,
		Host:   "127.0.0.1:21603",
		Path:   "%2e%2e%5ctests/",
	}

	req = &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
	}

	resp, err = c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()

	parsedURL = &url.URL{
		Scheme: "http",
		User:   nil,
		Host:   "127.0.0.1:21603",
		Path:   "..%2ftests/",
	}

	req = &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
	}

	resp, err = c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()

	parsedURL = &url.URL{
		Scheme: "http",
		User:   nil,
		Host:   "127.0.0.1:21603",
		Path:   "%2e%2e%2ftests/",
	}

	req = &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
	}

	resp, err = c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()

	_, r, err := get("http://127.0.0.1:21603/../../sample.txt")
	assert.NoError(t, err)
	assert.Equal(t, 403, r.StatusCode)
	_ = r.Body.Close()
}

func TestStaticPlugin(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-static.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
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
	t.Run("ServeSample", serveStaticSample)
	t.Run("StaticNotForbid", staticNotForbid)
	t.Run("StaticHeaders", staticHeaders)

	stopCh <- struct{}{}
	wg.Wait()
}

func staticHeaders(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:21603/php_test_files/client.php", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Header.Get("Output") != "output-header" {
		t.Fatal("can't find output header in response")
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, all("../../php_test_files/client.php"), string(b))
	require.Equal(t, all("../../php_test_files/client.php"), string(b))
}

func staticNotForbid(t *testing.T) {
	b, r, err := get("http://127.0.0.1:21603/php_test_files/client.php")
	require.NoError(t, err)
	require.Equal(t, all("../../php_test_files/client.php"), b)
	require.Equal(t, all("../../php_test_files/client.php"), b)
	_ = r.Body.Close()
}

func serveStaticSample(t *testing.T) {
	b, r, err := get("http://127.0.0.1:21603/sample.txt")
	require.NoError(t, err)
	require.Contains(t, b, "sample")
	_ = r.Body.Close()
}

func TestStaticDisabled_Error(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-static-disabled.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
	)
	assert.NoError(t, err)
	assert.Error(t, cont.Init())
}

func TestStaticFilesDisabled(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-static-files-disable.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
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
	t.Run("StaticFilesDisabled", staticFilesDisabled)

	stopCh <- struct{}{}
	wg.Wait()
}

func staticFilesDisabled(t *testing.T) {
	b, r, err := get("http://127.0.0.1:45877/php_test_files/client.php?hello=world")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "WORLD", b)
	_ = r.Body.Close()
}

func TestStaticFilesForbid(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-static-files.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		l,
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
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
	t.Run("StaticTestFilesDir", staticTestFilesDir)
	t.Run("StaticNotFound", staticNotFound)
	t.Run("StaticFilesForbid", staticFilesForbid)

	stopCh <- struct{}{}
	wg.Wait()
	time.Sleep(time.Second)

	o1 := oLogger.FilterMessageSnippet("http server was started")
	o3 := oLogger.FilterMessageSnippet("http log")

	require.Equal(t, 1, o1.Len())
	require.Equal(t, 3, o3.Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("file extension is forbidden").Len())
}

func staticTestFilesDir(t *testing.T) {
	b, r, err := get("http://127.0.0.1:34653/http?hello=world")
	assert.NoError(t, err)
	assert.Equal(t, "WORLD", b)
	_ = r.Body.Close()
}

func staticNotFound(t *testing.T) {
	b, _, _ := get("http://127.0.0.1:34653/client.XXX?hello=world") //nolint:bodyclose
	assert.Equal(t, "WORLD", b)
}

func staticFilesForbid(t *testing.T) {
	b, r, err := get("http://127.0.0.1:34653/client.php?hello=world")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "WORLD", b)
	_ = r.Body.Close()
}

func TestHTTPIssue659(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-issue659.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("HTTPIssue659", echoIssue659)

	stopCh <- struct{}{}

	wg.Wait()
}

func TestHTTPIPv6Long(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-ipv6.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("HTTPEchoIPv6-long", echoHTTPIPv6Long)

	stopCh <- struct{}{}

	wg.Wait()
}

func TestHTTPIPv6Short(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/http/.rr-http-ipv6-2.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("HTTPEchoIPv6-short", echoHTTPIPv6Short)

	stopCh <- struct{}{}

	wg.Wait()
}

func echoIssue659(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:32552", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Empty(t, b)
	assert.Equal(t, 444, r.StatusCode)

	err = r.Body.Close()
	assert.NoError(t, err)
}

func echoHTTP(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:10084?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func echoHTTPIPv6Long(t *testing.T) {
	req, err := http.NewRequest("GET", "http://[0:0:0:0:0:0:0:1]:10684?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func echoHTTPIPv6Short(t *testing.T) {
	req, err := http.NewRequest("GET", "http://[::1]:10784?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func resetTest(address string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", address)
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		// WorkerList contains list of workers.

		var ret bool
		err = client.Call("resetter.Reset", "http", &ret)
		assert.NoError(t, err)
		assert.True(t, ret)
		ret = false

		var services []string
		err = client.Call("resetter.List", nil, &services)
		assert.NoError(t, err)
		if services[0] != "http" {
			t.Fatal("no enough services")
		}
	}
}

func informerTest(address string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", address)
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		// WorkerList contains list of workers.
		list := struct {
			// Workers is list of workers.
			Workers []process.State `json:"workers"`
		}{}

		err = client.Call("informer.Workers", "http", &list)
		assert.NoError(t, err)
		assert.Len(t, list.Workers, 2)
	}
}

// HELPERS
func get(url string) (string, *http.Response, error) {
	r, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", nil, err
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return "", nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return "", nil, err
	}

	return string(b), r, err
}

func all(fn string) string {
	f, err := os.Open(fn)
	if err != nil {
		panic(err)
	}

	b := new(bytes.Buffer)
	_, err = io.Copy(b, f)
	if err != nil {
		return ""
	}

	err = f.Close()
	if err != nil {
		return ""
	}

	return b.String()
}
