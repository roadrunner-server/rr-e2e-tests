package lock

import (
	"net"
	"net/rpc"
	"testing"

	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	"github.com/stretchr/testify/require"
	lockApi "go.buf.build/protocolbuffers/go/roadrunner-server/api/lock/v1beta1"
)

const (
	lockRPC         string = "lock.Lock"
	rlockRPC        string = "lock.LockRead"
	releaseRPC      string = "lock.Release"
	updateTTLRPC    string = "lock.UpdateTTL"
	forceReleaseRPC string = "lock.ForceRelease"
	existsRPC       string = "lock.Exists"
)

func lock(t *testing.T, address string, resource, id string, ttl, wait int) bool {
	conn, err := net.Dial("tcp", address)
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
		Ttl:      ptrTo(int64(ttl)),
		Wait:     ptrTo(int64(wait)),
	}

	resp := &lockApi.Response{}
	err = client.Call(lockRPC, req, resp)
	require.NoError(t, err)
	return resp.Ok
}

func lockRead(t *testing.T, address string, resource, id string, ttl, wait int) bool {
	conn, err := net.Dial("tcp", address)
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
		Ttl:      ptrTo(int64(ttl)),
		Wait:     ptrTo(int64(wait)),
	}

	resp := &lockApi.Response{}
	err = client.Call(rlockRPC, req, resp)
	require.NoError(t, err)
	return resp.Ok
}

func release(t *testing.T, address string, resource, id string) bool {
	conn, err := net.Dial("tcp", address)
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
	}

	resp := &lockApi.Response{}
	err = client.Call(releaseRPC, req, resp)
	require.NoError(t, err)
	return resp.Ok
}

func updateTTL(t *testing.T, address string, resource, id string, ttl int) bool {
	conn, err := net.Dial("tcp", address)
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
		Ttl:      ptrTo(int64(ttl)),
	}

	resp := &lockApi.Response{}
	err = client.Call(updateTTLRPC, req, resp)
	require.NoError(t, err)
	return resp.Ok
}

func forceRelease(t *testing.T, address string, resource, id string) bool {
	conn, err := net.Dial("tcp", address)
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
	}

	resp := &lockApi.Response{}
	err = client.Call(forceReleaseRPC, req, resp)
	require.NoError(t, err)
	return resp.Ok
}

func exists(t *testing.T, address string, resource, id string) bool {
	conn, err := net.Dial("tcp", address)
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
	}

	resp := &lockApi.Response{}
	err = client.Call(existsRPC, req, resp)
	require.NoError(t, err)
	return resp.Ok
}

func ptrTo[T any](val T) *T {
	return &val
}
