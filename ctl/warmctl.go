package main

import (
	"../core"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

type AppFlags struct {
	capture bool
	apply   bool
}

func InitFlags() AppFlags {
	var af AppFlags

	flag.BoolVar(&af.capture, "c", false, "capture a snapshot")
	flag.BoolVar(&af.apply, "a", false, "apply the snapshot")

	flag.Parse()
	return af
}

func CGroupPIDs(cpath string) ([]int, error) {
	bs, err := ioutil.ReadFile(path.Join(cpath, "cgroup.procs"))
	if err != nil {
		return nil, err
	}
	var ret []int
	for _, str := range strings.Fields(string(bs)) {
		v, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		ret = append(ret, v)
	}
	return ret, nil
}

func doActions(c RPCClient, af AppFlags, args []string) error {
	var err error
	switch {
	case af.capture:
		// pids, err := CGroupPIDs("/sys/fs/cgroup/pids/user.slice/user-1000.slice/session-2.scope")
		// if err != nil {
		// 	return err
		// }
		test_cfg := core.CaptureConfig{
			Method: []core.CaptureMethod{
				//core.CaptureMethodMincores("/"),
				core.CaptureMethodPIDs(os.Getpid()),
				// core.CaptureMethodPIDs(pids...),
			},
		}

		snap, err := c.Capture(test_cfg)
		if err != nil {
			return err
		}
		DumpSnapshot(snap)
	case af.apply:
		panic("not impement")
	default:
		cfgs, err := c.ListConfig()
		if err != nil {
			return err
		}
		for _, cfg := range cfgs {
			fmt.Println(cfg.Id, cfg.Description, cfg.TryFile)
		}
	}
	return err
}

func main() {
	c, err := NewRPCClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cant Init RPC:", err)
		return
	}

	af := InitFlags()
	err = doActions(c, af, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "E:", err)
	}
}
