package main

import (
	"log"
	"net"
	"net/rpc"

	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	jobsProto "go.buf.build/protocolbuffers/go/roadrunner-server/api/jobs/v1"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		log.Fatal(err)
	}
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	list(client)
}

func list(client *rpc.Client) { //nolint:unused,deadcode
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

	pipe := &jobsProto.Pipelines{Pipelines: l}

	er = &jobsProto.Empty{}
	err = client.Call("jobs.Destroy", pipe, er)
	if err != nil {
		log.Println(err)
	}
}
