package main

import (
	"fmt"
	"net/rpc"
	"os"
	"time"

	"github.com/linuxdeepin/warm-sched/core"
)

func SwitchUserSession() error {
	c, err := rpc.Dial("unix", core.RPCSocket)
	if err != nil {
		return err
	}

	var noused bool
	env := map[string]string{
		"HOME":       os.Getenv("HOME"),
		"DISPLAY":    os.Getenv("DISPLAY"),
		"XAUTHORITY": os.Getenv("XAUTHORITY"),
	}
	return c.Call(core.RPCName+".SwitchUserSession", env, &noused)
}

func ForceEmitEvent(event string) error {
	c, err := rpc.Dial("unix", core.RPCSocket)
	if err != nil {
		return err
	}

	var noused bool
	return c.Call(core.RPCName+".ForceEmitEvent", event, &noused)
}

func Schedule() error {
	c, err := rpc.Dial("unix", core.RPCSocket)
	if err != nil {
		return err
	}

	var serr chan error
	go func() {
		serr <- c.Call(core.RPCName+".Schedule", "", nil)
	}()

	for {
		select {
		case <-time.After(time.Second):
			var list map[string][]string
			err := c.Call(core.RPCName+".SchedulePendings", "", &list)
			if err != nil {
				return err
			}
			if len(list) == 0 {
				return nil
			}
			for scope, vs := range list {
				fmt.Printf("Wait %v@%q\n", vs, scope)
			}
		case e := <-serr:
			fmt.Println("Schedule Done")
			return e
		}
	}
}

func ListConfig() ([]string, error) {
	c, err := rpc.Dial("unix", core.RPCSocket)
	if err != nil {
		return nil, err
	}
	var cfgs []string
	err = c.Call(core.RPCName+".ListConfig", true, &cfgs)
	return cfgs, err
}
