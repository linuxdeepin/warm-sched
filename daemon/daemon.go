package main

import (
	"../core"
	"fmt"
	"os"
)

type Daemon struct {
	cfgs []*SnapshotConfig

	history *History
}

func (d *Daemon) RunRPC(socket string) error {
	return RunRPCService(d, "unix", socket)
}

func RunDaemon(etc string, addr string) error {
	d := &Daemon{}

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
	err := RunDaemon("./etc", core.RPCSocket)
	if err != nil {
		fmt.Fprintf(os.Stderr, "E:%v\n", err)
		return
	}
}
