package durability

import (
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/roadrunner-server/amqp/v2"
	"github.com/roadrunner-server/beanstalk/v2"
	"github.com/roadrunner-server/config/v2"
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/informer/v2"
	"github.com/roadrunner-server/jobs/v2"
	"github.com/roadrunner-server/kafka/v2"
	"github.com/roadrunner-server/logger/v2"
	"github.com/roadrunner-server/nats/v2"
	"github.com/roadrunner-server/resetter/v2"
	rpcPlugin "github.com/roadrunner-server/rpc/v2"
	mocklogger "github.com/roadrunner-server/rr-e2e-tests/mock"
	helpers "github.com/roadrunner-server/rr-e2e-tests/plugins/jobs"
	"github.com/roadrunner-server/server/v2"
	"github.com/roadrunner-server/sqs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDurabilityAMQP(t *testing.T) {
	client := toxiproxy.NewClient("127.0.0.1:8474")

	_, err := client.CreateProxy("redial", "127.0.0.1:23679", "127.0.0.1:5672")
	require.NoError(t, err)
	defer helpers.DeleteProxy("redial", t)

	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-amqp-durability-redial.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&amqp.Plugin{},
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
	helpers.DisableProxy("redial", t)
	time.Sleep(time.Second * 3)

	go func() {
		time.Sleep(time.Second * 5)
		helpers.EnableProxy("redial", t)
	}()

	t.Run("PushPipelineWhileRedialing-1", helpers.PushToPipeErr("test-1"))
	t.Run("PushPipelineWhileRedialing-2", helpers.PushToPipeErr("test-2"))

	time.Sleep(time.Second * 15)
	t.Run("PushPipelineWhileRedialing-1", helpers.PushToPipe("test-1", false))
	t.Run("PushPipelineWhileRedialing-2", helpers.PushToPipe("test-2", false))

	time.Sleep(time.Second * 5)

	stopCh <- struct{}{}
	wg.Wait()

	t.Cleanup(func() {
		helpers.DestroyPipelines("test-1", "test-2")
	})
}

func TestDurabilitySQS(t *testing.T) {
	client := toxiproxy.NewClient("127.0.0.1:8474")

	_, err := client.CreateProxy("redial", "127.0.0.1:19324", "127.0.0.1:9324")
	require.NoError(t, err)
	defer helpers.DeleteProxy("redial", t)

	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-sqs-durability-redial.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&sqs.Plugin{},
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
	helpers.DisableProxy("redial", t)

	go func() {
		time.Sleep(time.Second)
		t.Run("PushPipelineWhileRedialing-1", helpers.PushToPipe("test-1", false))
		time.Sleep(time.Second)
		t.Run("PushPipelineWhileRedialing-2", helpers.PushToPipe("test-2", false))
	}()

	time.Sleep(time.Second * 5)
	helpers.EnableProxy("redial", t)
	time.Sleep(time.Second * 5)

	t.Run("PushPipelineWhileRedialing-3", helpers.PushToPipe("test-1", false))
	t.Run("PushPipelineWhileRedialing-4", helpers.PushToPipe("test-2", false))

	time.Sleep(time.Second * 10)

	stopCh <- struct{}{}
	wg.Wait()

	t.Cleanup(func() {
		helpers.DestroyPipelines("test-1", "test-2")
	})
}

func TestDurabilityBeanstalk(t *testing.T) {
	client := toxiproxy.NewClient("127.0.0.1:8474")

	_, err := client.CreateProxy("redial", "127.0.0.1:11400", "127.0.0.1:11300")
	require.NoError(t, err)
	defer helpers.DeleteProxy("redial", t)

	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*60))
	require.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-beanstalk-durability-redial.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&beanstalk.Plugin{},
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
	helpers.DisableProxy("redial", t)

	go func() {
		time.Sleep(time.Second * 2)
		t.Run("PushPipelineWhileRedialing-1", helpers.PushToPipe("test-1", false))
		t.Run("PushPipelineWhileRedialing-2", helpers.PushToPipe("test-2", false))
	}()

	time.Sleep(time.Second * 5)
	helpers.EnableProxy("redial", t)
	time.Sleep(time.Second * 2)

	t.Run("PushPipelineWhileRedialing-1", helpers.PushToPipe("test-1", false))
	t.Run("PushPipelineWhileRedialing-2", helpers.PushToPipe("test-2", false))

	time.Sleep(time.Second * 10)

	stopCh <- struct{}{}
	wg.Wait()

	t.Cleanup(func() {
		helpers.DestroyPipelines("test-1", "test-2")
	})
}

func TestDurabilityNATS(t *testing.T) {
	client := toxiproxy.NewClient("127.0.0.1:8474")

	_, err := client.CreateProxy("redial", "127.0.0.1:19224", "127.0.0.1:4222")
	require.NoError(t, err)
	defer helpers.DeleteProxy("redial", t)

	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.9.0",
		Path:    "configs/.rr-nats-durability-redial.yaml",
		Prefix:  "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.Plugin{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&nats.Plugin{},
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
	helpers.DisableProxy("redial", t)
	time.Sleep(time.Second * 3)

	go func() {
		time.Sleep(time.Second)
		t.Run("PushPipelineWhileRedialing-1", helpers.PushToPipe("test-1", false))
		time.Sleep(time.Second)
		t.Run("PushPipelineWhileRedialing-2", helpers.PushToPipe("test-2", false))
	}()

	time.Sleep(time.Second * 5)
	helpers.EnableProxy("redial", t)
	time.Sleep(time.Second * 2)

	t.Run("PushPipelineWhileRedialing-3", helpers.PushToPipe("test-1", false))
	t.Run("PushPipelineWhileRedialing-4", helpers.PushToPipe("test-2", false))

	time.Sleep(time.Second * 2)

	stopCh <- struct{}{}
	wg.Wait()

	t.Cleanup(func() {
		helpers.DestroyPipelines("test-1", "test-2")
	})
}

func TestDurabilityKafka(t *testing.T) {
	cmd := exec.Command("docker-compose", "-f", "../../../env/docker-compose-kafka.yaml", "up")
	err := cmd.Start()
	require.NoError(t, err)

	go func() {
		_ = cmd.Wait()
	}()

	time.Sleep(time.Second * 40)

	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*30))
	require.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.11.0",
		Path:    "configs/.rr-kafka-durability-redial.yaml",
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
		&kafka.Plugin{},
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

	time.Sleep(time.Second * 5)
	cmd2 := exec.Command("docker-compose", "-f", "../../../env/docker-compose-kafka.yaml", "down")
	err = cmd2.Start()
	require.NoError(t, err)

	go func() {
		_ = cmd2.Wait()
	}()

	time.Sleep(time.Second * 10)

	t.Run("PushPipelineWhileRedialing-1", helpers.PushToPipeErr("test-1"))
	t.Run("PushPipelineWhileRedialing-2", helpers.PushToPipeErr("test-2"))

	cmd3 := exec.Command("docker-compose", "-f", "../../../env/docker-compose-kafka.yaml", "up")
	err = cmd3.Start()
	require.NoError(t, err)

	go func() {
		_ = cmd3.Wait()
	}()

	time.Sleep(time.Second * 25)

	t.Run("PushPipelineWhileRedialing-1", helpers.PushToPipe("test-1", false))
	t.Run("PushPipelineWhileRedialing-2", helpers.PushToPipe("test-2", false))
	time.Sleep(time.Second * 5)
	t.Run("DestroyPipelines", helpers.DestroyPipelines("test-1", "test-2"))
	time.Sleep(time.Second * 20)

	stopCh <- struct{}{}
	wg.Wait()

	assert.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was started").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("job processing was started").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("job push error").Len())

	t.Cleanup(func() {
		cmd4 := exec.Command("docker-compose", "-f", "../../../env/docker-compose-kafka.yaml", "down")
		err = cmd4.Start()
		require.NoError(t, err)

		go func() {
			_ = cmd4.Wait()
		}()

		time.Sleep(time.Second * 15)
	})
}

func TestDurabilityKafkaCG(t *testing.T) {
	cmd := exec.Command("docker-compose", "-f", "../../../env/docker-compose-kafka.yaml", "up")
	err := cmd.Start()
	require.NoError(t, err)

	go func() {
		_ = cmd.Wait()
	}()

	time.Sleep(time.Second * 40)

	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*30))
	require.NoError(t, err)

	cfg := &config.Plugin{
		Version: "2.11.0",
		Path:    "configs/.rr-kafka-durability-redial-cg.yaml",
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
		&kafka.Plugin{},
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

	time.Sleep(time.Second * 5)
	cmd2 := exec.Command("docker-compose", "-f", "../../../env/docker-compose-kafka.yaml", "down")
	err = cmd2.Start()
	require.NoError(t, err)

	go func() {
		_ = cmd2.Wait()
	}()

	time.Sleep(time.Second * 10)

	t.Run("PushPipelineWhileRedialing-1", helpers.PushToPipeErr("test-1"))
	t.Run("PushPipelineWhileRedialing-2", helpers.PushToPipeErr("test-2"))

	cmd3 := exec.Command("docker-compose", "-f", "../../../env/docker-compose-kafka.yaml", "up")
	err = cmd3.Start()
	require.NoError(t, err)

	go func() {
		_ = cmd3.Wait()
	}()

	time.Sleep(time.Second * 25)

	t.Run("PushPipelineWhileRedialing-1", helpers.PushToPipe("test-1", false))
	t.Run("PushPipelineWhileRedialing-2", helpers.PushToPipe("test-2", false))
	time.Sleep(time.Second * 5)
	t.Run("DestroyPipelines", helpers.DestroyPipelines("test-1", "test-2"))
	time.Sleep(time.Second * 20)

	stopCh <- struct{}{}
	wg.Wait()

	assert.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was started").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("pipeline was stopped").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("job processing was started").Len())
	assert.Equal(t, 2, oLogger.FilterMessageSnippet("job push error").Len())

	t.Cleanup(func() {
		cmd4 := exec.Command("docker-compose", "-f", "../../../env/docker-compose-kafka.yaml", "down")
		err = cmd4.Start()
		require.NoError(t, err)

		go func() {
			_ = cmd4.Wait()
		}()

		time.Sleep(time.Second * 15)
	})
}
