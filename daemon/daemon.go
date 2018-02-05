package main

import (
	"../core"
	"fmt"
	"os"
)

type Daemon struct {
	cfgs []*core.SnapshotConfig

	events chan core.EventSource

	storage *Storeage

	history *History

	addr string
}

func (d *Daemon) Run() error {
	return RunRPCService(d, "unix", d.addr)
}

func NewDaemon(etc string, addr string) (*Daemon, error) {
	cfgs, err := ScanConfigs(etc)
	if err != nil {
		return nil, err
	}
	return &Daemon{
		addr: addr,
		cfgs: cfgs,
	}, nil
}

func RunDaemon() error {
	d, err := NewDaemon("./etc", core.RPCSocket)
	if err != nil {
		return err
	}
	return d.Run()
}

func main() {
	err := RunDaemon()
	if err != nil {
		fmt.Fprintf(os.Stderr, "E:%v\n", err)
		return
	}
}
