package main

import (
	"../core"
	"net"
	"net/rpc"
)

type RPCService struct {
	daemon *Daemon
}

const RPCName = "daemon"

func RunRPCService(d *Daemon, netType string, addr string) error {
	l, err := net.Listen(netType, addr)
	if err != nil {
		return err
	}
	s := RPCService{d}
	err = rpc.RegisterName(RPCName, s)
	if err != nil {
		return err
	}
	rpc.Accept(l)
	return nil
}

func (s RPCService) EmitEvent(in core.EventSource, out *bool) error {
	s.daemon.events <- in
	return nil
}

func (s RPCService) ListConfig(_ bool, out *[]*core.SnapshotConfig) error {
	*out = s.daemon.cfgs
	return nil
}

func (RPCService) Capture(_ string, out *core.Snapshot) error {
	snap, err := core.CaptureSnapshot("", []string{"/"})
	*out = *snap
	return err
}

func (RPCService) Apply(_ core.Snapshot, out *bool) error {
	panic("Not implement")
}
