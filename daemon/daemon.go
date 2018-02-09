package main

import (
	"../core"
	"../events"
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

type Daemon struct {
	cfgs    []*SnapshotConfig
	history *History

	innerSource *innerSource

	userEnv map[string]string
}

func (d *Daemon) SwitchUserSession(envs map[string]string) error {
	if len(envs) == 0 {
		return fmt.Errorf("Empty envs")
	}
	if d.userEnv != nil {
		return fmt.Errorf("There already has a session switched")
	}
	d.innerSource.MarkUser()
	d.userEnv = envs
	for _, cfg := range d.cfgs {
		for _, m := range cfg.Capture.Method {
			m.SetEnvs(d.userEnv)
		}
	}
	for k, v := range d.userEnv {
		switch k {
		case "DISPLAY", "XAUTHORITY":
			os.Setenv(k, v)
		}
	}
	return nil
}

func (d *Daemon) RunRPC(ctx context.Context, socket string) error {
	Log("RunRPC at %q\n", socket)
	return RunRPCService(ctx, d, "unix", socket)
}

func RunDaemon(etc string, cache, addr string, auto bool) error {
	d := &Daemon{
		history:     NewHistory(cache),
		innerSource: &innerSource{},
	}
	events.Register(d.innerSource)

	err := d.LoadConfigs(etc)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if auto {
		go func() {
			v := d.Schedule(ctx)
			fmt.Println("HHHH:", v)
		}()
	}
	return d.RunRPC(ctx, addr)
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
		fmt.Fprintf(os.Stderr, "E main:%v\n", err)
		return
	}
}
