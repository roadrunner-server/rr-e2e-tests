package tcp

import (
	"encoding/json"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"log/slog"

	"github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/roadrunner-server/logger/v4"
	rpcPlugin "github.com/roadrunner-server/rpc/v4"
	"github.com/roadrunner-server/server/v4"
	"github.com/roadrunner-server/tcp/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTCPInit(t *testing.T) {
	cont := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Second*30))

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-tcp-init.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&tcp.Plugin{},
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
	c, err := net.Dial("tcp", "127.0.0.1:7777")
	require.NoError(t, err)
	_, err = c.Write([]byte("wuzaaaa\n\r\n"))
	require.NoError(t, err)

	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	require.NoError(t, err)

	var d1 map[string]any
	err = json.Unmarshal(buf[:n], &d1)
	fmt.Println(d1)
	require.NoError(t, err)

	require.Equal(t, d1["remote_addr"].(string), c.LocalAddr().String())

	_ = c.Close()

	// ---

	time.Sleep(time.Second * 1)
	c, err = net.Dial("tcp", "127.0.0.1:8889")
	require.NoError(t, err)
	_, err = c.Write([]byte("helooooo\r\n"))
	require.NoError(t, err)

	buf = make([]byte, 1024)
	n, err = c.Read(buf)
	require.NoError(t, err)

	var d2 map[string]any
	err = json.Unmarshal(buf[:n], &d2)
	fmt.Println(d2)
	require.NoError(t, err)

	require.Equal(t, d2["remote_addr"].(string), c.LocalAddr().String())

	_ = c.Close()

	// ---

	time.Sleep(time.Second * 1)
	c, err = net.Dial("tcp", "127.0.0.1:8810")
	require.NoError(t, err)
	_, err = c.Write([]byte("HEEEEEEEEEEEEEYYYYYYYYYYYYY\r\n"))
	require.NoError(t, err)

	buf = make([]byte, 1024)
	n, err = c.Read(buf)
	require.NoError(t, err)

	var d3 map[string]any
	err = json.Unmarshal(buf[:n], &d3)
	fmt.Println(d3)
	require.NoError(t, err)

	require.Equal(t, d3["remote_addr"].(string), c.LocalAddr().String())

	_ = c.Close()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestTCPEmptySend(t *testing.T) {
	cont := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Minute))

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-tcp-empty.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&tcp.Plugin{},
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
	c, err := net.Dial("tcp", "127.0.0.1:7779")
	require.NoError(t, err)
	_, err = c.Write([]byte(""))
	require.NoError(t, err)

	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	require.NoError(t, err)

	var d map[string]any
	err = json.Unmarshal(buf[:n], &d)
	require.NoError(t, err)

	require.Equal(t, d["remote_addr"].(string), c.LocalAddr().String())

	// ---

	stopCh <- struct{}{}
	wg.Wait()
}

func TestTCPConnClose(t *testing.T) {
	cont := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Minute))

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-tcp-close.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&server.Plugin{},
		&tcp.Plugin{},
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
	c, err := net.Dial("tcp", "127.0.0.1:7788")
	require.NoError(t, err)
	_, err = c.Write([]byte("hello \r\n"))
	require.NoError(t, err)

	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	require.NoError(t, err)

	var d map[string]any
	err = json.Unmarshal(buf[:n], &d)
	require.NoError(t, err)

	require.NotEmpty(t, d["uuid"].(string))

	t.Run("CloseConnection", closeConn(d["uuid"].(string), "127.0.0.1:6001"))
	// ---

	stopCh <- struct{}{}
	wg.Wait()
}

func TestTCPFull(t *testing.T) {
	cont := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Minute))

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-tcp-full.yaml",
		Prefix:  "rr",
	}

	err := cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&tcp.Plugin{},
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
	waitCh := make(chan struct{}, 3)

	go func() {
		c, err := net.Dial("tcp", "127.0.0.1:7778")
		require.NoError(t, err)

		buf := make([]byte, 1024)
		n, err := c.Read(buf)
		require.NoError(t, err)

		require.Equal(t, []byte("hello \r\n"), buf[:n])

		var d map[string]any
		for i := 0; i < 100; i++ {
			_, err = c.Write([]byte("foo \r\n"))
			require.NoError(t, err)

			n, err = c.Read(buf)
			require.NoError(t, err)

			err = json.Unmarshal(buf[:n], &d)
			require.NoError(t, err)

			require.Equal(t, d["remote_addr"].(string), "foo1")
			require.Equal(t, d["body"].(string), "foo \r\n")
		}

		_ = c.Close()
		waitCh <- struct{}{}
	}()

	go func() {
		c, err := net.Dial("tcp", "127.0.0.1:8811")
		require.NoError(t, err)

		buf := make([]byte, 1024)
		n, err := c.Read(buf)
		require.NoError(t, err)

		require.Equal(t, []byte("hello \r\n"), buf[:n])

		var d map[string]any
		for i := 0; i < 100; i++ {
			_, err = c.Write([]byte("bar \r\n"))
			require.NoError(t, err)

			n, err = c.Read(buf)
			require.NoError(t, err)

			err = json.Unmarshal(buf[:n], &d)
			require.NoError(t, err)

			require.Equal(t, d["remote_addr"].(string), "foo2")
			require.Equal(t, d["body"].(string), "bar \r\n")
		}
		_ = c.Close()
		waitCh <- struct{}{}
	}()

	go func() {
		c, err := net.Dial("tcp", "127.0.0.1:8812")
		require.NoError(t, err)

		buf := make([]byte, 1024)
		n, err := c.Read(buf)
		require.NoError(t, err)

		require.Equal(t, []byte("hello \r\n"), buf[:n])

		var d map[string]any
		for i := 0; i < 100; i++ {
			_, err = c.Write([]byte("baz \r\n"))
			require.NoError(t, err)

			n, err = c.Read(buf)
			require.NoError(t, err)

			err = json.Unmarshal(buf[:n], &d)
			require.NoError(t, err)

			require.Equal(t, d["remote_addr"].(string), "foo3")
			require.Equal(t, d["body"].(string), "baz \r\n")
		}

		_ = c.Close()
		waitCh <- struct{}{}
	}()

	// ---

	<-waitCh
	<-waitCh
	<-waitCh

	stopCh <- struct{}{}
	wg.Wait()
}

func closeConn(uuid string, address string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", address)
		require.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		var ret bool
		err = client.Call("tcp.Close", uuid, &ret)
		require.NoError(t, err)
		require.True(t, ret)
	}
}
