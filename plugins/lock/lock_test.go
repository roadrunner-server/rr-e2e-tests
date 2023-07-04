package lock

import (
	"crypto/rand"
	"math/big"
	"os"
	"os/signal"
	"sort"
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

const secMult = 1000000

// race condition test, all methods are involved
func TestLockInit(t *testing.T) {
	cont := endure.New(slog.LevelInfo)

	cfg := &config.Plugin{
		Version: "2023.2.0",
		Path:    "configs/.rr-lock-init.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
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

	resources := map[int]string{0: "foo", 1: "foo1", 2: "foo2", 3: "foo3", 4: "foo4", 5: "foo5"}

	for i := 0; i < 1000; i++ {
		rs := randomString(10)
		go func() {
			_, err := lock("127.0.0.1:6001", resources[genRandNum(6)], rs, (genRandNum(5)+1)*secMult, (genRandNum(15)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lock("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(4)+1)*secMult, (genRandNum(11)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lock("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(2)+1)*secMult, (genRandNum(90)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lock("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(10)+1)*secMult, (genRandNum(10)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lock("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(20)+1)*secMult, (genRandNum(13)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lock("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(80)+1)*secMult, (genRandNum(10)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lock("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(20)+1)*secMult, (genRandNum(19)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := updateTTL("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(5))*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := exists("127.0.0.1:6001", "foo1", rs)
			assert.NoError(t, err)
		}()

		go func() {
			_, err := release("127.0.0.1:6001", "foo1", rs)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lockRead("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(20)+1)*secMult, (genRandNum(15)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lockRead("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(2)+1)*secMult, (genRandNum(34)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lockRead("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(20)+1)*secMult, (genRandNum(13)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lockRead("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(25)+1)*secMult, (genRandNum(15)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lockRead("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(20)+1)*secMult, (genRandNum(76)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := lockRead("127.0.0.1:6001", resources[genRandNum(6)], randomString(10), (genRandNum(20)+1)*secMult, (genRandNum(15)+1)*secMult)
			assert.NoError(t, err)
		}()
		go func() {
			_, err := forceRelease("127.0.0.1:6001", "foo1", "bar")
			assert.NoError(t, err)
		}()
	}

	time.Sleep(time.Minute * 3)

	stopCh <- struct{}{}
	wg.Wait()
	time.Sleep(time.Second * 5)
}

func TestLockFromSeveralProcesses(t *testing.T) {
	cont := endure.New(slog.LevelInfo)

	cfg := &config.Plugin{
		Version: "2023.1.0",
		Path:    "configs/.rr-lock-init.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
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

	time.Sleep(time.Second * 2)
	answ := make([]int, 0, 4)
	mu := &sync.Mutex{}

	go func() {
		res, err := lock("127.0.0.1:6001", "foo", "bar", 5*secMult, 1*secMult)
		assert.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()
		switch res {
		case true:
			answ = append(answ, 1)
		case false:
			answ = append(answ, 0)
		}
	}()
	go func() {
		res, err := lock("127.0.0.1:6001", "foo", "bar", 5*secMult, 1*secMult)
		assert.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()
		switch res {
		case true:
			answ = append(answ, 1)
		case false:
			answ = append(answ, 0)
		}
	}()
	go func() {
		res, err := lock("127.0.0.1:6001", "foo", "bar", 5*secMult, 1*secMult)
		assert.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()
		switch res {
		case true:
			answ = append(answ, 1)
		case false:
			answ = append(answ, 0)
		}
	}()
	go func() {
		res, err := lock("127.0.0.1:6001", "foo", "bar", 5*secMult, 1*secMult)
		assert.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()
		switch res {
		case true:
			answ = append(answ, 1)
		case false:
			answ = append(answ, 0)
		}
	}()

	time.Sleep(time.Second * 10)

	stopCh <- struct{}{}
	wg.Wait()
	time.Sleep(time.Second * 2)

	mu.Lock()
	sort.Ints(answ)
	assert.Equal(t, []int{0, 0, 0, 1}, answ)
	mu.Unlock()
}

func TestLockReadInit(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	cfg := &config.Plugin{
		Version: "2023.1.0",
		Path:    "configs/.rr-lock-init.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		l,
		cfg,
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
	res, err := lock("127.0.0.1:6001", "foo", "bar", 5*secMult, 0)
	assert.True(t, res)
	assert.NoError(t, err)

	res, err = lockRead("127.0.0.1:6001", "foo", "bar", 0, 10*secMult)
	assert.True(t, res)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	wg2 := &sync.WaitGroup{}
	wg2.Add(2)
	go func() {
		res2, err2 := lockRead("127.0.0.1:6001", "foo", "bar1", 0, 11*secMult)
		assert.True(t, res2)
		assert.NoError(t, err2)
		wg2.Done()
	}()

	go func() {
		res3, err3 := lockRead("127.0.0.1:6001", "foo", "bar2", 0, 11*secMult)
		assert.True(t, res3)
		assert.NoError(t, err3)
		wg2.Done()
	}()

	wg2.Wait()
	time.Sleep(time.Second)

	res, err = exists("127.0.0.1:6001", "foo", "bar1")
	assert.True(t, res)
	assert.NoError(t, err)
	res, err = exists("127.0.0.1:6001", "foo", "bar2")
	assert.True(t, res)
	assert.NoError(t, err)

	res, err = release("127.0.0.1:6001", "foo", "bar")
	assert.True(t, res)
	assert.NoError(t, err)
	res, err = release("127.0.0.1:6001", "foo", "bar1")
	assert.True(t, res)
	assert.NoError(t, err)
	res, err = release("127.0.0.1:6001", "foo", "bar2")
	assert.True(t, res)
	assert.NoError(t, err)

	stopCh <- struct{}{}
	wg.Wait()
	time.Sleep(time.Second * 2)

	assert.Equal(t, 1, oLogger.FilterMessageSnippet("waiting to acquire a lock, w==1, r==0").Len())
	assert.Equal(t, 10, oLogger.FilterMessageSnippet("releaseMuCh lock returned").Len())
}

func TestLockUpdateTTL(t *testing.T) {
	cont := endure.New(slog.LevelInfo)

	cfg := &config.Plugin{
		Version: "2023.2.0",
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

	res, err := lock("127.0.0.1:6001", "foo", "bar", 1000*secMult, 0)
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = updateTTL("127.0.0.1:6001", "foo", "bar", 2*secMult)
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = lockRead("127.0.0.1:6001", "foo", "bar1", 0, 10*secMult)
	assert.NoError(t, err)
	assert.True(t, res)

	time.Sleep(time.Second * 3)

	res, err = release("127.0.0.1:6001", "foo", "bar1")
	assert.NoError(t, err)
	assert.True(t, res)

	stopCh <- struct{}{}
	wg.Wait()

	assert.Equal(t, 1, oLogger.FilterMessageSnippet("updateTTL request received").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("updating r/lock ttl").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("lock successfully released").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("r/lock: ttl removed, callback call").Len())
}

func TestForceRelease(t *testing.T) {
	cont := endure.New(slog.LevelInfo)

	cfg := &config.Plugin{
		Version: "2023.2.0",
		Path:    "configs/.rr-lock-init.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err := cont.RegisterAll(
		l,
		cfg,
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
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)
	res, err := lock("127.0.0.1:6001", "foo", "bar", 1000*secMult, 0)
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = lockRead("127.0.0.1:6001", "foo", "bar1", 0, 1*secMult)
	assert.NoError(t, err)
	assert.False(t, res)

	res, err = forceRelease("127.0.0.1:6001", "foo", "bar")
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = lockRead("127.0.0.1:6001", "foo", "bar1", 0, 10*secMult)
	assert.NoError(t, err)
	assert.True(t, res)

	time.Sleep(time.Second)

	res, err = exists("127.0.0.1:6001", "foo", "bar1")
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = release("127.0.0.1:6001", "foo", "bar1")
	assert.NoError(t, err)
	assert.True(t, res)

	stopCh <- struct{}{}
	wg.Wait()

	assert.Equal(t, 1, oLogger.FilterMessageSnippet("failed to acquire a readlock, timeout exceeded, w==1, r==0").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("all force-release messages were sent").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("lock successfully released").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("r/lock: ttl removed, callback call").Len())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[genRandNum(len(letterBytes))]
	}
	return string(b)
}

func genRandNum(max int) int {
	bg := big.NewInt(int64(max))

	n, err := rand.Int(rand.Reader, bg)
	if err != nil {
		panic(err)
	}

	return int(n.Int64())
}
