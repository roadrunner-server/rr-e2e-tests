package amqp

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roadrunner-server/config/v2"
	endure "github.com/roadrunner-server/endure/pkg/container"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/roadrunner-server/informer/v2"
	"github.com/roadrunner-server/jobs/v2"
	kp "github.com/roadrunner-server/kafka/v2"
	"github.com/roadrunner-server/resetter/v2"
	rpcPlugin "github.com/roadrunner-server/rpc/v2"
	mocklogger "github.com/roadrunner-server/rr-e2e-tests/mock"
	helpers "github.com/roadrunner-server/rr-e2e-tests/plugins/jobs"
	"github.com/roadrunner-server/server/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	jobsProto "go.buf.build/protocolbuffers/go/roadrunner-server/api/proto/jobs/v1"
	"go.uber.org/zap"
)

func TestKafkaInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.11.0",
		Path:    "configs/.rr-kafka-init.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&jobs.Plugin{},
		&kp.Plugin{},
		l,
		&resetter.Plugin{},
		&informer.Plugin{},
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

	time.Sleep(time.Second * 3)
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	req := &jobsProto.PushRequest{Job: &jobsProto.Job{
		Job:     "some/php/namespace",
		Id:      uuid.NewString(),
		Payload: `{"hello":"world"}`,
		Headers: map[string]*jobsProto.HeaderValue{"test": {Value: []string{"test2"}}},
		Options: &jobsProto.Options{
			AutoAck:   false,
			Priority:  1,
			Pipeline:  "test-1",
			Delay:     0,
			Topic:     "default",
			Partition: 0,
			Offset:    0,
		},
	}}

	wgg := &sync.WaitGroup{}
	wgg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer wgg.Done()
			er := &jobsProto.Empty{}
			errCall := client.Call("jobs.Push", req, er)
			require.NoError(t, errCall)
		}()
	}
	wgg.Wait()

	time.Sleep(time.Second * 10)
	t.Run("DestroyPipelines", helpers.DestroyPipelines("test-1"))
	time.Sleep(time.Second * 5)

	stopCh <- struct{}{}
	wg.Wait()

	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("message sent").Len(), 1000)
	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("job was pushed successfully").Len(), 1000)
	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("job was processed successfully").Len(), 1000)
}

func TestKafkaDeclare(t *testing.T) {
	t.Skip("not ready")
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.11.0",
		Path:    "configs/.rr-kafka-declare.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&jobs.Plugin{},
		&kp.Plugin{},
		l,
		&resetter.Plugin{},
		&informer.Plugin{},
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

	time.Sleep(time.Second * 3)
	t.Run("DeclarePipeline", declarePipe)
	time.Sleep(time.Second * 2)
	t.Run("ResumePipeline", helpers.ResumePipes("test-1"))
	time.Sleep(time.Second * 2)

	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	req := &jobsProto.PushRequest{Job: &jobsProto.Job{
		Job:     "some/php/namespace",
		Id:      uuid.NewString(),
		Payload: `{"hello":"world"}`,
		Headers: map[string]*jobsProto.HeaderValue{"test": {Value: []string{"test2"}}},
		Options: &jobsProto.Options{
			AutoAck:   false,
			Priority:  1,
			Pipeline:  "test-1",
			Topic:     "default",
			Partition: 0,
			Offset:    0,
		},
	}}

	for i := 0; i < 2; i++ {
		er := &jobsProto.Empty{}
		errCall := client.Call("jobs.Push", req, er)
		require.NoError(t, errCall)
	}

	time.Sleep(time.Second * 5)
	t.Run("PausePipeline", helpers.PausePipelines("test-1"))

	for i := 0; i < 2; i++ {
		er := &jobsProto.Empty{}
		errCall := client.Call("jobs.Push", req, er)
		require.NoError(t, errCall)
	}

	time.Sleep(time.Second * 2)
	t.Run("ResumePipeline", helpers.ResumePipes("test-1"))
	time.Sleep(time.Second * 5)

	t.Run("DestroyPipeline", helpers.DestroyPipelines("test-1"))
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 4, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	require.Equal(t, 4, oLogger.FilterMessageSnippet("job processing was started").Len())
	require.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was paused").Len())
}

func TestKafkaJobsError(t *testing.T) {
	t.Skip("not ready")
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.11.0",
		Path:    "configs/.rr-kafka-jobs-err.yaml",
		Prefix:  "rr",
	}

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		l,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&kp.Plugin{},
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

	time.Sleep(time.Second * 3)

	t.Run("DeclarePipeline", declarePipe)
	t.Run("ResumePipeline", helpers.ResumePipes("test-1"))
	t.Run("PushPipeline", helpers.PushToPipe("test-1", false))
	time.Sleep(time.Second * 25)
	t.Run("PausePipeline", helpers.PausePipelines("test-1"))
	t.Run("DestroyPipeline", helpers.DestroyPipelines("test-1"))

	t.Run("DestroyPipeline", helpers.DestroyPipelines("test-1"))

	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	require.Equal(t, 4, oLogger.FilterMessageSnippet("job processing was started").Len())
	require.Equal(t, 4, oLogger.FilterMessageSnippet("job was processed successfully").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was paused").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
	require.Equal(t, 3, oLogger.FilterMessageSnippet("jobs protocol error").Len())
}

func declarePipe(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	pipe := &jobsProto.DeclareRequest{Pipeline: map[string]string{
		"driver":                 "kafka",
		"name":                   "test-1",
		"priority":               "3",
		"number_of_partitions":   "3",
		"create_topics_on_start": "true",
		"topics":                 "default, test-1",
		"topics_config": `{
  			"compression.type": "snappy"
		}`,
		"consumer_config": `{
  			"bootstrap.servers": "127.0.0.1:9092",
  			"group.id": "default",
  			"enable.partition.eof": true,
  			"auto.offset.reset": "earliest"
		}`,
		"producer_config": `{
  			"bootstrap.servers": "127.0.0.1:9092"
		}`,
	}}

	er := &jobsProto.Empty{}
	err = client.Call("jobs.Declare", pipe, er)
	assert.NoError(t, err)
}

func reset(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	c := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	var ret bool
	err = c.Call("resetter.Reset", "jobs", &ret)
	assert.NoError(t, err)
	require.True(t, ret)
}
