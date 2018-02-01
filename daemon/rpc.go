package main

import (
	"../core"
	"net"
	"net/rpc"
)

type RPCService struct {
}

const RPCName = "daemon"

func NewRPCService(netType string, addr string) error {
	l, err := net.Listen(netType, addr)
	if err != nil {
		return err
	}
	s := RPCService{}
	err = rpc.RegisterName(RPCName, s)
	if err != nil {
		return err
	}
	rpc.Accept(l)
	return nil
}

type Event struct {
	Scope string
	Id    string
}

func (RPCService) EmitEvent(in Event, out *bool) error {
	panic("Not implement")
}

func (RPCService) LoadCaptureConfig(in CaptureConfig, out *bool) error {
	panic("Not implement")
}

func (RPCService) LoadApplyConfig(in ApplyConfig, out *bool) error {
	panic("Not implement")
}

func (RPCService) Capture(_ string, out *core.Snapshot) error {
	snap, err := core.CaptureSnapshot("", []string{"/"})
	*out = *snap
	return err
}

func (RPCService) Apply(_ core.Snapshot, out *bool) error {
	panic("Not implement")
}
