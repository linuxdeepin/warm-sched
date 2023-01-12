package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"

	"github.com/linuxdeepin/warm-sched/core"
	"github.com/linuxdeepin/warm-sched/events"
)

type Daemon struct {
	cfgs    []*SnapshotConfig
	history *History

	innerSource *innerSource

	userEnv map[string]string

	ctx    context.Context
	cancel func()

	scheduling bool

	retryCaptureCount map[string]int
	retryCaptureMu    sync.Mutex
}

func (d *Daemon) Status() []string {
	current, err := core.CaptureSnapshot(core.NewCaptureMethodMincores("/", "/home"))
	if err != nil {
		return nil
	}

	var ret []string
	head := fmt.Sprintf("%-20s%8s%7s%10s%15s%15s%10s",
		"ID",
		"CaptureC",
		"ApplyC",
		"LifetimeC",
		"SnapSize",
		"ContentSize",
		"Loaded%",
	)

	total := &core.Snapshot{}

	ret = append(ret, head)
	for _, cfg := range d.cfgs {
		ss, snap := d.history.Status(cfg.Id, current)

		v := fmt.Sprintf("%-20s%8d%7d%10d%15s%15s%8.2f",
			cfg.Id,
			ss.CaptureCount,
			ss.ApplyCount,
			ss.LifetimeCount,
			core.HumanSize(ss.SnapSize),
			core.HumanSize(ss.ContentSize),
			ss.LoadedPercentage*100,
		)
		ret = append(ret, v)

		if snap != nil {
			for _, i := range snap.Infos {
				total.Add(i)
			}
		}
	}

	tcs, tp := total.AnalyzeSnapshotLoad(current)
	ret = append(ret, fmt.Sprintf("%-20s%8s%7s%10s%15s%15s%8.2f",
		"TOTAL",
		"\u0078",
		"\u0078",
		"\u0078",
		"\u0078",
		core.HumanSize(tcs),
		tp*100,
	))

	return ret
}

func (d *Daemon) ForceEmitEvent(event string) error {
	events.Emit(event)
	return nil
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

func quitWhenLowMemory(ctx context.Context, cancel func(), threshold uint64) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 2):
			if MemFree() < threshold {
				Log("Quit because available memory is lower than %s.\n", core.HumanSize(int(threshold)))
				cancel()
				return
			}
		}
	}
}
func quitWhenTimeout(ctx context.Context, cancel func(), t time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(t):
		Log("Quit because timeout %v.\n", t)
		cancel()
		return
	}
}

func RunDaemon(etc string, cache, addr string, auto bool, lowMemory uint64, timeout time.Duration) error {
	ctx, cancel := context.WithCancel(context.Background())
	d := &Daemon{
		history:           NewHistory(ctx, cache),
		innerSource:       &innerSource{},
		ctx:               ctx,
		cancel:            cancel,
		retryCaptureCount: make(map[string]int),
	}

	if !FileExist(path.Join(cache, "debug")) {
		go quitWhenLowMemory(ctx, cancel, lowMemory)
		go quitWhenTimeout(ctx, cancel, timeout)
	} else {
		Log("In Debug Mode. Daemon wouldn't automatically quit.\n")
	}

	events.Register(d.innerSource)

	err := d.LoadConfigs(etc)
	if err != nil {
		return err
	}
	if auto {
		go d.Schedule(d.ctx)
	}
	return d.RunRPC(d.ctx, addr)
}

func main() {
	cfgDir := flag.String("etc", "./etc", "the directory of snapshot configures")
	cacheDir := flag.String("cache", "./cache", "the directory of caching")
	socket := flag.String("socket", core.RPCSocket, "the unix socket address.")
	auto := flag.Bool("auto", true, "automatically schedule")
	lowMemory := flag.Int("lowMemory", 200*1024, "The threshold of low memory in KB, when available memory is lower than the threshold, daemon will quit")

	timeout := flag.Int("timeout", 60*30, "Maximum seconds to wait")

	flag.Parse()

	defer func() {
		exec.Command("rmmod", "mincores").Run()
	}()

	t := time.Duration(*timeout) * time.Second
	err := RunDaemon(*cfgDir, *cacheDir, *socket, *auto, uint64(*lowMemory*1024), t)
	if err != nil {
		fmt.Fprintf(os.Stderr, "E main:%v\n", err)
		return
	}
}
