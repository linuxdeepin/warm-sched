package main

import (
	"../core"
	"fmt"
	"os"
)

type Daemon struct {
	cfgs    []*SnapshotConfig
	history *History
}

func (d *Daemon) RunRPC(socket string) error {
	return RunRPCService(d, "unix", socket)
}

func RunDaemon(etc string, cache, addr string) error {
	d := &Daemon{
		history: NewHistory(cache),
	}

	err := d.LoadConfigs(etc)
	if err != nil {
		return err
	}

	err = d.Schedule()
	if err != nil {
		return err
	}

	return d.RunRPC(addr)
}

func main() {
	err := RunDaemon("./etc", "./cache", core.RPCSocket)
	if err != nil {
		fmt.Fprintf(os.Stderr, "E:%v\n", err)
		return
	}
}
