package lock

import (
	"net"
	"net/rpc"

	lockApi "github.com/roadrunner-server/api/v4/build/lock/v1beta1"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
)

const (
	lockRPC         string = "lock.Lock"
	rlockRPC        string = "lock.LockRead"
	releaseRPC      string = "lock.Release"
	updateTTLRPC    string = "lock.UpdateTTL"
	forceReleaseRPC string = "lock.ForceRelease"
	existsRPC       string = "lock.Exists"
)

func lock(address string, resource, id string, ttl, wait int) (bool, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false, err
	}
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
		Ttl:      ptrTo(int64(ttl)),
		Wait:     ptrTo(int64(wait)),
	}

	resp := &lockApi.Response{}
	err = client.Call(lockRPC, req, resp)
	if err != nil {
		return false, err
	}
	return resp.Ok, nil
}

func lockRead(address string, resource, id string, ttl, wait int) (bool, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false, err
	}
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
		Ttl:      ptrTo(int64(ttl)),
		Wait:     ptrTo(int64(wait)),
	}

	resp := &lockApi.Response{}
	err = client.Call(rlockRPC, req, resp)
	if err != nil {
		return false, err
	}
	return resp.Ok, nil
}

func release(address string, resource, id string) (bool, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false, err
	}
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
	}

	resp := &lockApi.Response{}
	err = client.Call(releaseRPC, req, resp)
	if err != nil {
		return false, err
	}
	return resp.Ok, nil
}

func updateTTL(address string, resource, id string, ttl int) (bool, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false, nil
	}
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
		Ttl:      ptrTo(int64(ttl)),
	}

	resp := &lockApi.Response{}
	err = client.Call(updateTTLRPC, req, resp)
	if err != nil {
		return false, nil
	}
	return resp.Ok, nil
}

func forceRelease(address string, resource, id string) (bool, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false, nil
	}
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
	}

	resp := &lockApi.Response{}
	err = client.Call(forceReleaseRPC, req, resp)
	if err != nil {
		return false, nil
	}
	return resp.Ok, nil
}

func exists(address string, resource, id string) (bool, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false, nil
	}
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	req := &lockApi.Request{
		Resource: resource,
		Id:       id,
	}

	resp := &lockApi.Response{}
	err = client.Call(existsRPC, req, resp)
	if err != nil {
		return false, nil
	}
	return resp.Ok, nil
}

func ptrTo[T any](val T) *T {
	return &val
}
