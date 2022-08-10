package kafka

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

func TestKafkaInitCG(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.11.0",
		Path:    "configs/.rr-kafka-init-cg.yaml",
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
			Priority: 1,
			Pipeline: "test-1",
			Topic:    "test-2",
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

	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("job was pushed successfully").Len(), 1000)
	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("job was pushed successfully").Len(), 1000)
	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("job was processed successfully").Len(), 1000)
}

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
			Priority: 1,
			Pipeline: "test-1",
			Topic:    "test-1",
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

	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("job was pushed successfully").Len(), 1000)
	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("job was pushed successfully").Len(), 1000)
	assert.GreaterOrEqual(t, oLogger.FilterMessageSnippet("job was processed successfully").Len(), 1000)
}

func TestKafkaDeclareCG(t *testing.T) {
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
	t.Run("DeclarePipeline", declarePipeCG("test-33"))
	time.Sleep(time.Second * 2)
	t.Run("ResumePipeline", helpers.ResumePipes("test-33"))
	time.Sleep(time.Second * 2)

	t.Run("PushPipeline", helpers.PushToPipe("test-33", false))

	for i := 0; i < 2; i++ {
		t.Run("PushPipeline", helpers.PushToPipe("test-33", false))
	}

	time.Sleep(time.Second * 5)
	t.Run("PausePipeline", helpers.PausePipelines("test-33"))

	for i := 0; i < 2; i++ {
		t.Run("PushPipeline", helpers.PushToPipe("test-33", false))
	}

	time.Sleep(time.Second * 2)
	t.Run("ResumePipeline", helpers.ResumePipes("test-33"))
	time.Sleep(time.Second * 5)

	t.Run("DestroyPipeline", helpers.DestroyPipelines("test-33"))
	stopCh <- struct{}{}
	wg.Wait()

	assert.Equal(t, 5, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	assert.Equal(t, 5, oLogger.FilterMessageSnippet("job processing was started").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was paused").Len())
}

func TestKafkaDeclare(t *testing.T) {
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
	t.Run("DeclarePipeline", declarePipe("test-22"))
	time.Sleep(time.Second * 2)
	t.Run("ResumePipeline", helpers.ResumePipes("test-22"))
	time.Sleep(time.Second * 2)

	t.Run("PushPipeline", helpers.PushToPipe("test-22", false))

	for i := 0; i < 2; i++ {
		t.Run("PushPipeline", helpers.PushToPipe("test-22", false))
	}

	time.Sleep(time.Second * 5)
	t.Run("PausePipeline", helpers.PausePipelines("test-22"))

	for i := 0; i < 2; i++ {
		t.Run("PushPipeline", helpers.PushToPipe("test-22", false))
	}

	time.Sleep(time.Second * 2)
	t.Run("ResumePipeline", helpers.ResumePipes("test-22"))
	time.Sleep(time.Second * 5)

	t.Run("DestroyPipeline", helpers.DestroyPipelines("test-22"))
	stopCh <- struct{}{}
	wg.Wait()

	assert.Equal(t, 5, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	assert.Equal(t, 5, oLogger.FilterMessageSnippet("job processing was started").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was paused").Len())
}

func TestKafkaJobsError(t *testing.T) {
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
	t.Run("DeclarePipeline", declarePipe("test-3"))
	time.Sleep(time.Second)
	t.Run("ResumePipeline", helpers.ResumePipes("test-3"))
	t.Run("PushPipeline", helpers.PushToPipe("test-3", false))
	time.Sleep(time.Second * 25)
	t.Run("PausePipeline", helpers.PausePipelines("test-3"))
	t.Run("DestroyPipeline", helpers.DestroyPipelines("test-3"))

	t.Run("DestroyPipeline", helpers.DestroyPipelines("test-3"))

	stopCh <- struct{}{}
	wg.Wait()

	assert.Equal(t, 1, oLogger.FilterMessageSnippet("job was pushed successfully").Len())
	assert.Equal(t, 4, oLogger.FilterMessageSnippet("job processing was started").Len())
	assert.Equal(t, 4, oLogger.FilterMessageSnippet("job was processed successfully").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was paused").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was resumed").Len())
	assert.Equal(t, 1, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
	assert.Equal(t, 3, oLogger.FilterMessageSnippet("jobs protocol error").Len())
}

func declarePipe(topic string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		pipe := &jobsProto.DeclareRequest{
			Pipeline: map[string]string{
				"driver":   "kafka",
				"name":     topic,
				"priority": "3",
				"topic":    topic,
				"partitions_offsets": `
				{
				     "0": "0",
				     "1": "0",
				     "2": "0"
				}`,
				"max_open_requests": "100",
				"client_id":         "roadrunner",
				"version":           "3.2.0.0",

				// create_topics
				"replication_factor": "1",
				// "replica_assignment": `
				// {
				//      "0": [1,2,3],
				//      "1": [1,2,3],
				//      "2": [1,2,3]
				// }
				// `,
				"config_entries": `
    			{
                      "max.message.bytes": "100000"
				}
				`,

				// producer
				"max_message_bytes": "10000",
				"required_acks":     "-1",
				"timeout":           "10",
				"compression_codec": "snappy",
				"compression_level": "100",
				"idempotent":        "false",

				// consumer
				// "heartbeat_interval": "3",
				// "session_timeout":    "10",
			},
		}

		er := &jobsProto.Empty{}
		err = client.Call("jobs.Declare", pipe, er)
		assert.NoError(t, err)
	}
}

func declarePipeCG(topic string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

		pipe := &jobsProto.DeclareRequest{
			Pipeline: map[string]string{
				"driver":   "kafka",
				"name":     topic,
				"priority": "3",
				"topic":    topic,
				"group_id": "foo",
				"partitions_offsets": `
				{
				     "0": "0",
				     "1": "0",
				     "2": "0"
				}`,
				"max_open_requests": "100",
				"client_id":         "roadrunner",
				"version":           "3.2.0.0",

				// create_topics
				"replication_factor": "1",
				// "replica_assignment": `
				// {
				//      "0": [1,2,3],
				//      "1": [1,2,3],
				//      "2": [1,2,3]
				// }
				// `,
				"config_entries": `
    			{
                      "max.message.bytes": "100000"
				}
				`,

				// producer
				"max_message_bytes": "10000",
				"required_acks":     "-1",
				"timeout":           "10",
				"compression_codec": "snappy",
				"compression_level": "100",
				"idempotent":        "false",

				// consumer
				// "heartbeat_interval": "3",
				// "session_timeout":    "10",
			},
		}

		er := &jobsProto.Empty{}
		err = client.Call("jobs.Declare", pipe, er)
		assert.NoError(t, err)
	}
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
