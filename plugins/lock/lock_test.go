package lock

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	lockPlugin "github.com/roadrunner-server/lock/v4"
	rpcPlugin "github.com/roadrunner-server/rpc/v4"
	mocklogger "github.com/roadrunner-server/rr-e2e-tests/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/exp/slog"
)

func TestLockInit(t *testing.T) {
	cont := endure.New(slog.LevelInfo)

	cfg := &config.Plugin{
		Version: "2023.1.0",
		Path:    "configs/.rr-lock-init.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		l,
		&rpcPlugin.Plugin{},
		&lockPlugin.Plugin{},
	)
	require.NoError(t, err)

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
	assert.True(t, lock(t, "127.0.0.1:6001", "foo", "bar", 5, 0))
	assert.True(t, lock(t, "127.0.0.1:6001", "foo", "bar", 0, 10))

	time.Sleep(time.Second)

	wg2 := &sync.WaitGroup{}
	wg2.Add(2)
	go func() {
		assert.False(t, lock(t, "127.0.0.1:6001", "foo", "bar", 0, 11))
		wg2.Done()
	}()

	go func() {
		assert.False(t, lock(t, "127.0.0.1:6001", "foo", "bar", 0, 11))
		wg2.Done()
	}()

	wg2.Wait()

	time.Sleep(time.Second)

	stopCh <- struct{}{}
	wg.Wait()

	require.Greater(t, oLogger.FilterMessageSnippet("acquire attempt failed, retrying in 1s").Len(), 10)
	require.Equal(t, 2, oLogger.FilterMessageSnippet("lock successfully acquired").Len())
	require.Equal(t, 4, oLogger.FilterMessageSnippet("lock request received").Len())
	require.Equal(t, 2, oLogger.FilterMessageSnippet("failed to acquire lock, wait timeout exceeded").Len())
}

func TestLockReadInit(t *testing.T) {
	cont := endure.New(slog.LevelInfo)

	cfg := &config.Plugin{
		Version: "2023.1.0",
		Path:    "configs/.rr-lock-init.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		l,
		&rpcPlugin.Plugin{},
		&lockPlugin.Plugin{},
	)
	require.NoError(t, err)

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
	assert.True(t, lock(t, "127.0.0.1:6001", "foo", "bar", 5, 0))
	assert.True(t, lockRead(t, "127.0.0.1:6001", "foo", "bar", 0, 10))

	time.Sleep(time.Second)

	wg2 := &sync.WaitGroup{}
	wg2.Add(2)
	go func() {
		assert.True(t, lockRead(t, "127.0.0.1:6001", "foo", "bar1", 0, 11))
		wg2.Done()
	}()

	go func() {
		assert.True(t, lockRead(t, "127.0.0.1:6001", "foo", "bar2", 0, 11))
		wg2.Done()
	}()

	wg2.Wait()

	time.Sleep(time.Second)
	assert.True(t, exists(t, "127.0.0.1:6001", "foo", "bar1"))
	assert.True(t, exists(t, "127.0.0.1:6001", "foo", "bar2"))

	assert.True(t, release(t, "127.0.0.1:6001", "foo", "bar"))
	assert.True(t, release(t, "127.0.0.1:6001", "foo", "bar1"))
	assert.True(t, release(t, "127.0.0.1:6001", "foo", "bar2"))

	stopCh <- struct{}{}
	wg.Wait()

	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("acquire attempt failed, retrying in 1s").Len(), 4)
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("no such resource, creating new").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("lock: expired, executing callback").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("lock: ttl expired").Len())
	assert.Equal(t, 4, oLogger.FilterMessageSnippet("lock successfully acquired").Len())
	assert.Equal(t, 3, oLogger.FilterMessageSnippet("read lock released").Len())
}

func TestLockUpdateTTL(t *testing.T) {
	cont := endure.New(slog.LevelInfo)

	cfg := &config.Plugin{
		Version: "2023.1.0",
		Path:    "configs/.rr-lock-init.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		l,
		&rpcPlugin.Plugin{},
		&lockPlugin.Plugin{},
	)
	require.NoError(t, err)

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
	assert.True(t, lock(t, "127.0.0.1:6001", "foo", "bar", 1000, 0))
	assert.True(t, updateTTL(t, "127.0.0.1:6001", "foo", "bar", 2))
	assert.True(t, lockRead(t, "127.0.0.1:6001", "foo", "bar1", 0, 10))

	time.Sleep(time.Second)

	assert.True(t, release(t, "127.0.0.1:6001", "foo", "bar1"))

	stopCh <- struct{}{}
	wg.Wait()

	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("acquire attempt failed, retrying in 1s").Len(), 1)
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("no such resource, creating new").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("lock: expired, executing callback").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("lock: ttl expired").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("updating lock TTL").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("lock successfully acquired").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("read lock released").Len())
}

func TestForceRelease(t *testing.T) {
	cont := endure.New(slog.LevelInfo)

	cfg := &config.Plugin{
		Version: "2023.1.0",
		Path:    "configs/.rr-lock-init.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		cfg,
		l,
		&rpcPlugin.Plugin{},
		&lockPlugin.Plugin{},
	)
	require.NoError(t, err)

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
	assert.True(t, lock(t, "127.0.0.1:6001", "foo", "bar", 1000, 0))
	assert.False(t, lockRead(t, "127.0.0.1:6001", "foo", "bar1", 0, 1))
	assert.True(t, forceRelease(t, "127.0.0.1:6001", "foo", "bar"))
	assert.True(t, lockRead(t, "127.0.0.1:6001", "foo", "bar1", 0, 10))

	time.Sleep(time.Second)

	assert.True(t, exists(t, "127.0.0.1:6001", "foo", "bar1"))
	assert.True(t, release(t, "127.0.0.1:6001", "foo", "bar1"))

	stopCh <- struct{}{}
	wg.Wait()

	assert.Equal(t, 3, oLogger.FilterMessageSnippet("lock request received").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("no such resource, creating new").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("lock successfully acquired").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("lock forcibly released").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("waiting for the lock to acquire").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("read lock released").Len())
}
