package lock

import (
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	lockPlugin "github.com/roadrunner-server/lock/v4"
	"github.com/roadrunner-server/logger/v4"
	rpcPlugin "github.com/roadrunner-server/rpc/v4"
	mocklogger "github.com/roadrunner-server/rr-e2e-tests/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/exp/slog"
)

const SEC_MUL = 1000000

func TestLockInit(t *testing.T) {
	cont := endure.New(slog.LevelInfo)

	cfg := &config.Plugin{
		Version: "2023.1.0",
		Path:    "configs/.rr-lock-init.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	_ = l
	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		// l,
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
	// assert.True(t, lock(t, "127.0.0.1:6001", "foo", "bar", 1000000000, 0))
	//
	// time.Sleep(time.Second * 2)
	// updateTTL(t, "127.0.0.1:6001", "foo", "bar", 10000)
	// time.Sleep(time.Second * 2)
	// updateTTL(t, "127.0.0.1:6001", "foo", "bar", 10000)

	for i := 0; i < 1000; i++ {
		go func() {
			lock(t, "127.0.0.1:6001", "foo1", randomString(10), 1*SEC_MUL, 15*SEC_MUL)
		}()
		go func() {
			lock(t, "127.0.0.1:6001", "foo1", randomString(10), 1*SEC_MUL, 15*SEC_MUL)
		}()
		go func() {
			lock(t, "127.0.0.1:6001", "foo1", randomString(10), 1*SEC_MUL, 15*SEC_MUL)
		}()
		go func() {
			lock(t, "127.0.0.1:6001", "foo1", randomString(10), 1*SEC_MUL, 15*SEC_MUL)
		}()
		go func() {
			lockRead(t, "127.0.0.1:6001", "foo1", randomString(10), 1*SEC_MUL, 15*SEC_MUL)
		}()
		go func() {
			lockRead(t, "127.0.0.1:6001", "foo1", randomString(10), 1*SEC_MUL, 15*SEC_MUL)
		}()
		// go func() {
		// 	forceRelease(t, "127.0.0.1:6001", "foo1", "bar")
		// }()
	}
	// rs1 := randomString(10)
	// rs2 := randomString(10)
	// rs3 := randomString(10)
	// rs4 := randomString(10)
	//
	// lock(t, "127.0.0.1:6001", "foo1", rs1, 10*SEC_MUL, 100*SEC_MUL)
	// lockRead(t, "127.0.0.1:6001", "foo1", rs2, 10000, 2*SEC_MUL)
	// lockRead(t, "127.0.0.1:6001", "foo1", rs3, 10000, 1101010)
	// lockRead(t, "127.0.0.1:6001", "foo1", rs4, 10000, 7774499)

	// assert.True(t, lockRead(t, "127.0.0.1:6001", "foo1", "bar2", 1000000000, 0))
	// assert.True(t, lockRead(t, "127.0.0.1:6001", "foo1", "bar3", 1000000000, 0))
	// assert.True(t, lockRead(t, "127.0.0.1:6001", "foo1", "bar4", 1000000000, 0))
	// assert.True(t, lockRead(t, "127.0.0.1:6001", "foo1", "bar5", 1000000000, 0))
	// assert.True(t, lockRead(t, "127.0.0.1:6001", "foo1", "bar6", 1000000000, 0))
	// go func() {
	// 	time.Sleep(time.Second * 10)
	// 	release(t, "127.0.0.1:6001", "foo1", "bar1")
	// }()
	//
	// assert.True(t, lock(t, "127.0.0.1:6001", "foo1", "bar1", 100000000, 1000000))
	//
	// time.Sleep(time.Second * 2)

	// forceRelease(t, "127.0.0.1:6001", "foo1", "bar")

	// assert.True(t, lock(t, "127.0.0.1:6001", "foo1", "bar1", 10000000, 0))
	// assert.True(t, lock(t, "127.0.0.1:6001", "foo2", "bar2", 10000000, 0))
	// for i := 0; i < 10; i++ {
	// 	go func() {
	// 		lock(t, "127.0.0.1:6001", "foo", "arb", 10000000, 100000000)
	// 	}()
	// }
	//
	// go func() {
	// 	time.Sleep(time.Second)
	// 	release(t, "127.0.0.1:6001", "foo", "bar")
	// }()

	// forceRelease(t, "127.0.0.1:6001", "foo", "bar")
	// release(t, "127.0.0.1:6001", "foo1", "bar1")
	// release(t, "127.0.0.1:6001", "foo2", "bar2")

	// assert.True(t, lock(t, "127.0.0.1:6001", "foo", "bar", 0, 10))
	//
	// time.Sleep(time.Second)
	//
	// wg2 := &sync.WaitGroup{}
	// wg2.Add(2)
	// go func() {
	// 	assert.False(t, lock(t, "127.0.0.1:6001", "foo", "bar", 0, 11))
	// 	wg2.Done()
	// }()
	//
	// go func() {
	// 	assert.False(t, lock(t, "127.0.0.1:6001", "foo", "bar", 0, 11))
	// 	wg2.Done()
	// }()
	//
	// wg2.Wait()
	//
	time.Sleep(time.Minute)

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
	_ = l
	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		// l,
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
