package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/linuxdeepin/warm-sched/core"
)

type AppFlags struct {
	capture      bool
	schedule     bool
	switchToUser bool
	apply        bool
	test         bool

	dump bool

	diffAdded bool
	diff      bool
}

func InitFlags() AppFlags {
	var af AppFlags
	flag.BoolVar(&af.test, "t", false, "build a test draft snapshot")
	flag.BoolVar(&af.capture, "c", false, "capture a snapshot")
	flag.BoolVar(&af.apply, "a", false, "apply a snapshot")
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

var defaultMincores = []string{"/", "/home"}

func Draft(name string) error {
	c1, err := core.CaptureSnapshot(core.NewCaptureMethodMincores(defaultMincores...))
	if err != nil {
		return err
	}
	fmt.Println("Please run manually the test target application.")
	fmt.Println("Press Ctrl+C when the target has been launched.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	select {
	case <-c:
	case <-time.After(time.Second * 600):
		return fmt.Errorf("Timeout")
	}
	c2, err := core.CaptureSnapshot(core.NewCaptureMethodMincores(defaultMincores...))
	if err != nil {
		return err
	}

	diffs := core.CompareSnapshot(c1, c2)
	err = core.StoreTo(name+".snap", core.Snapshot{Infos: diffs.Added})
	if err != nil {
		return err
	}
	fmt.Printf("See the draft by run \"warmctl -dump %s.snap\".\n", name)
	return nil
}

func doActions(af AppFlags, args []string) error {
	switch {
	case af.test:
		if len(args) < 1 {
			return fmt.Errorf("Please specify the draft name")
		}
		return Draft(args[0])
	case af.apply:
		if len(args) < 1 {
			return fmt.Errorf("Please specify the snapshot file path")
		}
		var s1 core.Snapshot
		err := core.LoadFrom(args[0], &s1)
		if err != nil {
			return err
		}
		return core.ApplySnapshot(&s1, true)
	case af.capture:
		current, err := core.CaptureSnapshot(core.NewCaptureMethodMincores(defaultMincores...))
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
		return err
	case af.dump:
		var s1 core.Snapshot
		err := core.LoadFrom(args[0], &s1)
		if err != nil {
			return err
		}
		return core.DumpSnapshot(&s1)
	case af.schedule:
		return Schedule()
	case af.switchToUser:
		return SwitchUserSession()
	case af.diffAdded:
		if len(args) < 3 {
			return fmt.Errorf("Please specify three snapshot file")
		}
		diffs, err := CompareSnapshot(args[0], args[1])
		if err != nil {
			return err
		}
		return core.StoreTo(args[2], core.Snapshot{Infos: diffs.Added})
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
		cfgs, err := ListConfig()
		if err != nil {
			return err
		}
		for _, cfg := range cfgs {
			fmt.Println(cfg)
		}
	}
	return nil
}

func main() {
	af := InitFlags()
	err := doActions(af, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "E:", err)
	}
}
