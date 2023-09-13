package main

import (
	"fmt"
	"log"
	rand2 "math/rand"
	"net"
	"net/rpc"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	jobsProto "github.com/roadrunner-server/api/v4/build/jobs/v1"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
)

const (
	push    string = "jobs.Push"
	pause   string = "jobs.Pause"
	destroy string = "jobs.Destroy"
	declare string = "jobs.Declare"
	resume  string = "jobs.Resume"
	stat    string = "jobs.Stat"
)

func main() {
	wg := &sync.WaitGroup{}
	wg.Add(10)

	rate := uint64(0)
	delayedCloseCh := make(chan string, 1000000)

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 10000; j++ {
					n := uuid.NewString()
					declareAMQPPipe(client, n)
					startPipelines(client, n)
					push100(client, n)
					atomic.AddUint64(&rate, 1)
					delayedCloseCh <- n
				}
				wg.Done()
			}()
		}
	}()

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 10; j++ {
					n := uuid.NewString()
					declareBeanstalkPipe(client, n)
					startPipelines(client, n)
					push100(client, n)
					atomic.AddUint64(&rate, 1)
					delayedCloseCh <- n
				}
				wg.Done()
			}()
		}
	}()

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for i := 0; i < 3; i++ {
			go func() {
				for j := 0; j < 10; j++ {
					n := uuid.NewString()
					declareBoltDBPipe(client, n, n)
					startPipelines(client, n)
					push100(client, n)
					atomic.AddUint64(&rate, 1)
					delayedCloseCh <- n

					cur, err := os.Getwd()
					if err != nil {
						panic(err)
					}
					_ = os.Remove(path.Join(cur, n))
				}
				wg.Done()
			}()
		}
	}()

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 1000; j++ {
					n := uuid.NewString()
					declareMemoryPipe(client, n)
					startPipelines(client, n)
					push100(client, n)
					atomic.AddUint64(&rate, 1)
					delayedCloseCh <- n
				}
				wg.Done()
			}()
		}
	}()

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 1000; j++ {
					n := uuid.NewString()
					declareSQSPipe(client, n)
					startPipelines(client, n)
					push100(client, n)
					atomic.AddUint64(&rate, 1)
					delayedCloseCh <- n
				}
				wg.Done()
			}()
		}
	}()

	go func() {
		tt := time.NewTicker(time.Second)
		for { //nolint:gosimple
			select {
			case <-tt.C:
				fmt.Printf("-- RATE: %d --\n", atomic.LoadUint64(&rate)*100)
				atomic.StoreUint64(&rate, 0)
			}
		}
	}()

	stopCh := make(chan struct{})

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for {
			select {
			case k := <-delayedCloseCh:
				pausePipelines(client, k)
				destroyPipelines(delayedCloseCh, client, k)
			case <-stopCh:
				return
			}
		}
	}()

	wg.Wait()
	stopCh <- struct{}{}
}

func push100(client *rpc.Client, pipe string) {
	for j := 0; j < 100; j++ {
		payloads := &jobsProto.PushRequest{
			Job: &jobsProto.Job{
				Job:     "Some/Super/PHP/Class",
				Id:      uuid.NewString(),
				Payload: []byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularized in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum."),
				Headers: map[string]*jobsProto.HeaderValue{"test": {Value: []string{"hello"}}},
				Options: &jobsProto.Options{
					Priority: int64(rand2.Intn(100) + 1), //nolint:gosec
					Pipeline: pipe,
				},
			},
		}

		resp := jobsProto.Empty{}
		err := client.Call(push, payloads, &resp)
		if err != nil {
			log.Println(err)
		}
	}
}

func startPipelines(client *rpc.Client, pipes ...string) {
	pipe := &jobsProto.Pipelines{Pipelines: make([]string, len(pipes))}

	for i := 0; i < len(pipes); i++ {
		pipe.GetPipelines()[i] = pipes[i]
	}

	er := &jobsProto.Empty{}
	err := client.Call(resume, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func pausePipelines(client *rpc.Client, pipes ...string) {
	pipe := &jobsProto.Pipelines{Pipelines: make([]string, len(pipes))}

	for i := 0; i < len(pipes); i++ {
		pipe.GetPipelines()[i] = pipes[i]
	}

	er := &jobsProto.Empty{}
	err := client.Call(pause, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func destroyPipelines(rest chan<- string, client *rpc.Client, p string) {
	pipe := &jobsProto.Pipelines{Pipelines: make([]string, 1)}

	pipe.GetPipelines()[0] = p

	er := &jobsProto.Empty{}
	err := client.Call(destroy, pipe, er)
	if err != nil {
		log.Println(err)
		if !strings.Contains(err.Error(), "no such pipeline") {
			rest <- p
		}
	}
}

func list(client *rpc.Client) []string { //nolint:unused,deadcode
	resp := &jobsProto.Pipelines{}
	er := &jobsProto.Empty{}
	err := client.Call("jobs.List", er, resp)
	if err != nil {
		log.Println(err)
	}

	l := make([]string, len(resp.GetPipelines()))

	for i := 0; i < len(resp.GetPipelines()); i++ {
		l[i] = resp.GetPipelines()[i]
	}
	return l
}

func declareSQSPipe(client *rpc.Client, n string) {
	pipe := &jobsProto.DeclareRequest{Pipeline: map[string]string{
		"driver":             "sqs",
		"name":               n,
		"queue":              n,
		"prefetch":           "10",
		"priority":           "3",
		"visibility_timeout": "0",
		"wait_time_seconds":  "3",
		"tags":               `{"purpose":"roadrunner-load-testing"}`,
	}}

	er := &jobsProto.Empty{}
	err := client.Call(declare, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func declareAMQPPipe(client *rpc.Client, p string) {
	pipe := &jobsProto.DeclareRequest{Pipeline: map[string]string{
		"driver":          "amqp",
		"name":            p,
		"routing_key":     "test-3",
		"queue":           "default",
		"exchange_type":   "direct",
		"exchange":        "amqp.default",
		"prefetch":        "1000",
		"priority":        "1",
		"exclusive":       "false",
		"multiple_ack":    "false",
		"requeue_on_fail": "false",
	}}

	er := &jobsProto.Empty{}
	err := client.Call(declare, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func declareBeanstalkPipe(client *rpc.Client, n string) {
	pipe := &jobsProto.DeclareRequest{Pipeline: map[string]string{
		"driver":          "beanstalk",
		"name":            n,
		"tube":            n,
		"reserve_timeout": "1",
		"priority":        "3",
		"tube_priority":   "10",
	}}

	er := &jobsProto.Empty{}
	err := client.Call(declare, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func declareBoltDBPipe(client *rpc.Client, n, file string) {
	pipe := &jobsProto.DeclareRequest{Pipeline: map[string]string{
		"driver":   "boltdb",
		"name":     n,
		"prefetch": "100",
		"priority": "2",
		"file":     file,
	}}

	er := &jobsProto.Empty{}
	err := client.Call(declare, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func declareMemoryPipe(client *rpc.Client, p string) {
	pipe := &jobsProto.DeclareRequest{Pipeline: map[string]string{
		"driver":   "memory",
		"name":     p,
		"prefetch": "10000",
		"priority": "1",
	}}

	er := &jobsProto.Empty{}
	err := client.Call(declare, pipe, er)
	if err != nil {
		log.Println(err)
	}
}
