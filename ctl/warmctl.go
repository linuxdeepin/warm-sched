package main

import (
	"../core"
	"flag"
	"fmt"
	"os"
)

type AppFlags struct {
	capture      bool
	schedule     bool
	switchToUser bool

	dump bool

	diffAdded bool
	diff      bool
}

func InitFlags() AppFlags {
	var af AppFlags

	flag.BoolVar(&af.capture, "c", false, "capture a snapshot")
	flag.BoolVar(&af.schedule, "s", false, "schedule handle snapshot by configures")
	flag.BoolVar(&af.switchToUser, "u", false, "let warm-daemon switch on this session")
	flag.BoolVar(&af.dump, "dump", false, "dump content of the snapshot")
	flag.BoolVar(&af.diff, "d", false, "show difference current pagecache with captured $id")
	flag.BoolVar(&af.diffAdded, "da", false, "same as flag d, except this will save added snapshot to the $file")

	flag.Parse()
	return af
}

func CompareSnapshot(f1 string, f2 string) (*core.SnapshotDiff, error) {
	var s1, s2 core.Snapshot
	err := core.LoadFrom(f1, &s1)
	if err != nil {
		return nil, err
	}
	err = core.LoadFrom(f2, &s2)
	if err != nil {
		return nil, err
	}
	diffs := core.CompareSnapshot(&s1, &s2)
	return diffs, nil
}

func doActions(af AppFlags, args []string) error {
	c, err := NewRPCClient()
	if err != nil {
		return fmt.Errorf("Cant Init RPC:%v", err)
	}

	switch {
	case af.capture:
		current, err := core.CaptureSnapshot(core.NewCaptureMethodMincores("/", "/home"))
		if err != nil {
			return err
		}
		switch len(args) {
		case 0:
			err = core.DumpSnapshot(current)
		case 1:
			err = core.StoreTo(args[0], current)
		default:
			err = fmt.Errorf("Too many arguments")
		}
	case af.dump:
		var s1 core.Snapshot
		err = core.LoadFrom(args[0], &s1)
		if err != nil {
			return err
		}
		err = core.DumpSnapshot(&s1)
	case af.schedule:
		err = c.Schedule()
	case af.switchToUser:
		err = c.SwitchUserSession()
	case af.diffAdded:
		if len(args) < 3 {
			return fmt.Errorf("Please specify three snapshot file")
		}
		diffs, err := CompareSnapshot(args[0], args[1])
		if err != nil {
			return err
		}
		err = core.StoreTo(args[2], core.Snapshot{Infos: diffs.Added})
	case af.diff:
		if len(args) < 2 {
			return fmt.Errorf("Please specify tow snapshot file")
		}
		diffs, err := CompareSnapshot(args[0], args[1])
		if err != nil {
			return err
		}
		fmt.Println(diffs)
	default:
		cfgs, err := c.ListConfig()
		if err != nil {
			return err
		}
		for _, cfg := range cfgs {
			fmt.Println(cfg)
		}
	}
	return err
}

func main() {
	af := InitFlags()
	err := doActions(af, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "E:", err)
	}
}
