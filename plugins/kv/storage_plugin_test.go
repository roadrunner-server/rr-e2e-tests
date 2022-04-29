package kv

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/roadrunner-server/boltdb/v2"
	"github.com/roadrunner-server/config/v2"
	endure "github.com/roadrunner-server/endure/pkg/container"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/roadrunner-server/kv/v2"
	"github.com/roadrunner-server/logger/v2"
	"github.com/roadrunner-server/memcached/v2"
	"github.com/roadrunner-server/memory/v2"
	"github.com/roadrunner-server/redis/v2"
	rpcPlugin "github.com/roadrunner-server/rpc/v2"
	mock_logger "github.com/roadrunner-server/rr-e2e-tests/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kvProto "go.buf.build/protocolbuffers/go/roadrunner-server/api/proto/kv/v1"
	"go.uber.org/zap"
)

func TestKVInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-kv-init.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&memory.Plugin{},
		&boltdb.Plugin{},
		&memcached.Plugin{},
		&redis.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&kv.Plugin{},
	)
	assert.NoError(t, err)

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

	time.Sleep(time.Second * 1)
	t.Run("KvSetTest", kvSetTest)
	t.Run("KvHasTest", kvHasTest)

	stopCh <- struct{}{}

	wg.Wait()

	_ = os.RemoveAll("rr.db")
	_ = os.RemoveAll("africa.db")
}

func TestKVNoInterval(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-kv-bolt-no-interval.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&boltdb.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&kv.Plugin{},
	)
	assert.NoError(t, err)

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

	time.Sleep(time.Second * 1)
	t.Run("KvSetTest", kvSetTest)
	t.Run("KvHasTest", kvHasTest)

	stopCh <- struct{}{}

	wg.Wait()

	_ = os.RemoveAll("rr.db")
	_ = os.RemoveAll("africa.db")
}

func TestKVCreateToReopenWithPerms(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-kv-bolt-perms.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&boltdb.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&kv.Plugin{},
	)
	assert.NoError(t, err)

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

	time.Sleep(time.Second * 1)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestKVCreateToReopenWithPerms2(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-kv-bolt-perms.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&boltdb.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&kv.Plugin{},
	)
	assert.NoError(t, err)

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

	time.Sleep(time.Second * 1)
	t.Run("KvSetTest", kvSetTest)
	t.Run("KvHasTest", kvHasTest)

	stopCh <- struct{}{}

	wg.Wait()

	_ = os.RemoveAll("rr.db")
	_ = os.RemoveAll("africa.db")
}

func kvSetTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	p := &kvProto.Request{
		Storage: "boltdb-south",
		Items: []*kvProto.Item{
			{
				Key:   "key",
				Value: []byte("val"),
			},
		},
	}

	resp := &kvProto.Response{}
	err = client.Call("kv.Set", p, resp)
	assert.NoError(t, err)
}

func kvHasTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	p := &kvProto.Request{
		Storage: "boltdb-south",
		Items: []*kvProto.Item{
			{
				Key:   "key",
				Value: []byte("val"),
			},
		},
	}

	ret := &kvProto.Response{}
	err = client.Call("kv.Has", p, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 1)
}

func TestBoltDb(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-boltdb.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&boltdb.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
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
	t.Run("BOLTDB", testRPCMethods)
	stopCh <- struct{}{}
	wg.Wait()

	_ = os.Remove("rr.db")
}

func testRPCMethods(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	keys := &kvProto.Request{
		Storage: "boltdb-rr",
		Items: []*kvProto.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "c",
			},
		},
	}

	data := &kvProto.Request{
		Storage: "boltdb-rr",
		Items: []*kvProto.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:     "c",
				Value:   []byte("cc"),
				Timeout: tt,
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	ret := &kvProto.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", data, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // should be 2

	ret = &kvProto.Response{}
	err = client.Call("kv.MGet", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // c is expired

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)

	data2 := &kvProto.Request{
		Storage: "boltdb-rr",
		Items: []*kvProto.Item{
			{
				Key:     "a",
				Timeout: tt2,
			},
			{
				Key:     "b",
				Timeout: tt2,
			},
			{
				Key:     "d",
				Timeout: tt2,
			},
		},
	}

	// MEXPIRE
	ret = &kvProto.Response{}
	err = client.Call("kv.MExpire", data2, ret)
	assert.NoError(t, err)

	// TTL
	keys2 := &kvProto.Request{
		Storage: "boltdb-rr",
		Items: []*kvProto.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "d",
			},
		},
	}

	ret = &kvProto.Response{}
	err = client.Call("kv.TTL", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	// DELETE
	keysDel := &kvProto.Request{
		Storage: "boltdb-rr",
		Items: []*kvProto.Item{
			{
				Key: "e",
			},
		},
	}

	ret = &kvProto.Response{}
	err = client.Call("kv.Delete", keysDel, ret)
	assert.NoError(t, err)

	// HAS AFTER DELETE
	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keysDel, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	dataClear := &kvProto.Request{
		Storage: "boltdb-rr",
		Items: []*kvProto.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:   "c",
				Value: []byte("cc"),
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	clear := &kvProto.Request{Storage: "boltdb-rr"}

	ret = &kvProto.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", dataClear, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 5) // should be 5

	ret = &kvProto.Response{}
	err = client.Call("kv.Clear", clear, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0) // should be 5
}

func TestMemcached(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-memcached.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&memcached.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
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
	t.Run("MEMCACHED", testRPCMethodsMemcached)
	stopCh <- struct{}{}
	wg.Wait()
}

func testRPCMethodsMemcached(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)

	keys := &kvProto.Request{
		Storage: "memcached-rr",
		Items: []*kvProto.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "c",
			},
		},
	}

	data := &kvProto.Request{
		Storage: "memcached-rr",
		Items: []*kvProto.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:     "c",
				Value:   []byte("cc"),
				Timeout: tt,
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	ret := &kvProto.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", data, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // should be 2

	ret = &kvProto.Response{}
	err = client.Call("kv.MGet", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // c is expired

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)

	data2 := &kvProto.Request{
		Storage: "memcached-rr",
		Items: []*kvProto.Item{
			{
				Key:     "a",
				Timeout: tt2,
			},
			{
				Key:     "b",
				Timeout: tt2,
			},
			{
				Key:     "d",
				Timeout: tt2,
			},
		},
	}

	// MEXPIRE
	ret = &kvProto.Response{}
	err = client.Call("kv.MExpire", data2, ret)
	assert.NoError(t, err)

	// TTL call is not supported for the memcached driver
	keys2 := &kvProto.Request{
		Storage: "memcached-rr",
		Items: []*kvProto.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "d",
			},
		},
	}

	ret = &kvProto.Response{}
	err = client.Call("kv.TTL", keys2, ret)
	assert.Error(t, err)
	assert.Len(t, ret.GetItems(), 0)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	// DELETE
	keysDel := &kvProto.Request{
		Storage: "memcached-rr",
		Items: []*kvProto.Item{
			{
				Key: "e",
			},
		},
	}

	ret = &kvProto.Response{}
	err = client.Call("kv.Delete", keysDel, ret)
	assert.NoError(t, err)

	// HAS AFTER DELETE
	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keysDel, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	dataClear := &kvProto.Request{
		Storage: "memcached-rr",
		Items: []*kvProto.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:   "c",
				Value: []byte("cc"),
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	clear := &kvProto.Request{Storage: "memcached-rr"}

	ret = &kvProto.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", dataClear, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 5) // should be 5

	ret = &kvProto.Response{}
	err = client.Call("kv.Clear", clear, ret)
	assert.NoError(t, err)

	time.Sleep(time.Second * 2)
	ret = &kvProto.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0) // should be 5
}

func TestInMemory(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-in-memory.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&memory.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
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
	t.Run("INMEMORY", testRPCMethodsInMemory)
	stopCh <- struct{}{}
	wg.Wait()
}

func testRPCMethodsInMemory(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl

	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	keys := &kvProto.Request{
		Storage: "memory-rr",
		Items: []*kvProto.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "c",
			},
		},
	}

	data := &kvProto.Request{
		Storage: "memory-rr",
		Items: []*kvProto.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:     "c",
				Value:   []byte("cc"),
				Timeout: tt,
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	ret := &kvProto.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", data, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // should be 2

	ret = &kvProto.Response{}
	err = client.Call("kv.MGet", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // c is expired

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)

	data2 := &kvProto.Request{
		Storage: "memory-rr",
		Items: []*kvProto.Item{
			{
				Key:     "a",
				Timeout: tt2,
			},
			{
				Key:     "b",
				Timeout: tt2,
			},
			{
				Key:     "d",
				Timeout: tt2,
			},
		},
	}

	// MEXPIRE
	ret = &kvProto.Response{}
	err = client.Call("kv.MExpire", data2, ret)
	assert.NoError(t, err)

	// TTL
	keys2 := &kvProto.Request{
		Storage: "memory-rr",
		Items: []*kvProto.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "d",
			},
		},
	}

	ret = &kvProto.Response{}
	err = client.Call("kv.TTL", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	// DELETE
	keysDel := &kvProto.Request{
		Storage: "memory-rr",
		Items: []*kvProto.Item{
			{
				Key: "e",
			},
		},
	}

	ret = &kvProto.Response{}
	err = client.Call("kv.Delete", keysDel, ret)
	assert.NoError(t, err)

	// HAS AFTER DELETE
	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keysDel, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	dataClear := &kvProto.Request{
		Storage: "memory-rr",
		Items: []*kvProto.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:   "c",
				Value: []byte("cc"),
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	clear := &kvProto.Request{Storage: "memory-rr"}

	ret = &kvProto.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", dataClear, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 5) // should be 5

	ret = &kvProto.Response{}
	err = client.Call("kv.Clear", clear, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0) // should be 5
}

func TestRedis(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-redis.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&redis.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
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
	t.Run("REDIS", testRPCMethodsRedis)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestRedisGlobalSection(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-redis-global.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&kv.Plugin{},
		&redis.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
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
	t.Run("REDIS", testRPCMethodsRedis)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestRedisNoConfig(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-redis-no-config.yaml", // should be used default
		Prefix:  "rr",
	}

	l, oLogger := mock_logger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&kv.Plugin{},
		&redis.Plugin{},
		&rpcPlugin.Plugin{},
		&memory.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	_, err = cont.Serve()
	assert.NoError(t, err)
	_ = cont.Stop()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("plugin was started").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("can't find local or global configuration, this section will be skipped").Len())
}

func testRPCMethodsRedis(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	// add 5 second ttl
	tt := time.Now().Add(time.Second * 5).Format(time.RFC3339)
	keys := &kvProto.Request{
		Storage: "redis-rr",
		Items: []*kvProto.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "c",
			},
		},
	}

	data := &kvProto.Request{
		Storage: "redis-rr",
		Items: []*kvProto.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:     "c",
				Value:   []byte("cc"),
				Timeout: tt,
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	ret := &kvProto.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", data, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3) // should be 3

	// key "c" should be deleted
	time.Sleep(time.Second * 7)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // should be 2

	ret = &kvProto.Response{}
	err = client.Call("kv.MGet", keys, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 2) // c is expired

	tt2 := time.Now().Add(time.Second * 10).Format(time.RFC3339)

	data2 := &kvProto.Request{
		Storage: "redis-rr",
		Items: []*kvProto.Item{
			{
				Key:     "a",
				Timeout: tt2,
			},
			{
				Key:     "b",
				Timeout: tt2,
			},
			{
				Key:     "d",
				Timeout: tt2,
			},
		},
	}

	// MEXPIRE
	ret = &kvProto.Response{}
	err = client.Call("kv.MExpire", data2, ret)
	assert.NoError(t, err)

	// TTL
	keys2 := &kvProto.Request{
		Storage: "redis-rr",
		Items: []*kvProto.Item{
			{
				Key: "a",
			},
			{
				Key: "b",
			},
			{
				Key: "d",
			},
		},
	}

	ret = &kvProto.Response{}
	err = client.Call("kv.TTL", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 3)

	// HAS AFTER TTL
	time.Sleep(time.Second * 15)
	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	ret = &kvProto.Response{}
	err = client.Call("kv.TTL", keys2, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	// DELETE
	keysDel := &kvProto.Request{
		Storage: "redis-rr",
		Items: []*kvProto.Item{
			{
				Key: "e",
			},
		},
	}

	ret = &kvProto.Response{}
	err = client.Call("kv.Delete", keysDel, ret)
	assert.NoError(t, err)

	// HAS AFTER DELETE
	ret = &kvProto.Response{}
	err = client.Call("kv.Has", keysDel, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0)

	dataClear := &kvProto.Request{
		Storage: "redis-rr",
		Items: []*kvProto.Item{
			{
				Key:   "a",
				Value: []byte("aa"),
			},
			{
				Key:   "b",
				Value: []byte("bb"),
			},
			{
				Key:   "c",
				Value: []byte("cc"),
			},
			{
				Key:   "d",
				Value: []byte("dd"),
			},
			{
				Key:   "e",
				Value: []byte("ee"),
			},
		},
	}

	clear := &kvProto.Request{Storage: "redis-rr"}

	ret = &kvProto.Response{}
	// Register 3 keys with values
	err = client.Call("kv.Set", dataClear, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 5) // should be 5

	ret = &kvProto.Response{}
	err = client.Call("kv.Clear", clear, ret)
	assert.NoError(t, err)

	ret = &kvProto.Response{}
	err = client.Call("kv.Has", dataClear, ret)
	assert.NoError(t, err)
	assert.Len(t, ret.GetItems(), 0) // should be 5
}
