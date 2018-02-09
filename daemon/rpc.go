package main

import (
	"../core"
	"context"
	"fmt"
	"net"
	"net/rpc"
	"time"
)

type RPCService struct {
	daemon *Daemon
}

func RunRPCService(ctx context.Context, d *Daemon, netType string, addr string) error {
	l, err := net.Listen(netType, addr)
	if err != nil {
		return err
	}
	s := RPCService{d}
	err = rpc.RegisterName(core.RPCName, s)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				l.Close()
				return
			case <-time.After(time.Second):
			}
		}
	}()
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

func (s RPCService) SwitchUserSession(env map[string]string, out *bool) error {
	return s.daemon.SwitchUserSession(env)
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

func (s RPCService) Schedule(_ string, out *bool) error {
	err := s.daemon.Schedule(context.TODO())
	if err != nil {
		return err
	}
	return nil
}
func (s RPCService) SchedulePendings(_ string, out *map[string][]string) error {
	*out = EventWaits()
	return nil
}

func (s RPCService) SnapStatus(_ string, out *[]SnapStatus) error {
	v := s.daemon.history.SnapStatus()
	*out = v
	return nil
}

func (RPCService) Apply(_ core.Snapshot, out *bool) error {
	panic("Not implement")
}
