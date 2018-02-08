package main

import (
	"../core"
	"flag"
	"fmt"
	"os"
	"time"
)

type Daemon struct {
	cfgs    []*SnapshotConfig
	history *History
}

func (d *Daemon) RunRPC(socket string) error {
	Log("RunRPC at %q\n", socket)
	return RunRPCService(d, "unix", socket)
}

func RunDaemon(etc string, cache, addr string, auto bool) error {
	d := &Daemon{
		history: NewHistory(cache),
	}
	err := d.LoadConfigs(etc)
	if err != nil {
		return err
	}
	if auto {
		go d.Schedule()
	}
	return d.RunRPC(addr)
}

func main() {
	cfgDir := flag.String("etc", "./etc", "the directory of snapshot configures")
	cacheDir := flag.String("cache", "./cache", "the directory of caching")
	socket := flag.String("socket", core.RPCSocket, "the unix socket address.")
	auto := flag.Bool("auto", true, "automatically schedule")

	timeout := flag.Int("timeout", 60*10, "Maximum seconds to wait")

	flag.Parse()

	time.AfterFunc(time.Duration(*timeout)*time.Second, func() {
		Log("Timeout, so normal quitting daemon.\n")
		os.Exit(0)
	})

	err := RunDaemon(*cfgDir, *cacheDir, *socket, *auto)
	if err != nil {
		fmt.Fprintf(os.Stderr, "E:%v\n", err)
		return
	}
}
