package main

import (
	"../core"
	"fmt"
	"net"
	"net/rpc"
)

type RPCService struct {
	daemon *Daemon
}

func RunRPCService(d *Daemon, netType string, addr string) error {
	l, err := net.Listen(netType, addr)
	if err != nil {
		return err
	}

	s := RPCService{d}
	err = rpc.RegisterName(core.RPCName, s)
	if err != nil {
		return err
	}
	rpc.Accept(l)
	return nil
}

func (s RPCService) ListConfig(_ bool, out *[]string) error {
	var ret []string
	for _, cfg := range s.daemon.cfgs {
		ret = append(ret, cfg.Id)
	}
	*out = ret
	return nil
}

func (s RPCService) Capture(id string, out *core.Snapshot) error {
	for _, cfg := range s.daemon.cfgs {
		if cfg.Id == id {
			snap, err := core.CaptureSnapshot(cfg.Capture.Method...)
			if err != nil {
				return err
			}
			*out = *snap
			return nil
		}
	}
	return fmt.Errorf("Not Found Configure of %q", id)
}

func (RPCService) Apply(_ core.Snapshot, out *bool) error {
	panic("Not implement")
}
