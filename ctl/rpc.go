package main

import (
	"../core"
	"fmt"
	"net/rpc"
	"time"
)

type RPCClient struct {
	core *rpc.Client
}

func (c RPCClient) Capture(id string) (*core.Snapshot, error) {
	var snap core.Snapshot
	err := c.core.Call(core.RPCName+".Capture", id, &snap)
	return &snap, err
}

func (c RPCClient) Schedule() error {
	var serr chan error
	go func() {
		serr <- c.core.Call(core.RPCName+".Schedule", "", nil)
	}()

	for {
		select {
		case <-time.After(time.Second):
			var list map[string][]string
			err := c.core.Call(core.RPCName+".SchedulePendings", "", &list)
			if err != nil {
				return err
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

func (c RPCClient) ListConfig() ([]string, error) {
	var cfgs []string
	err := c.core.Call(core.RPCName+".ListConfig", true, &cfgs)
	return cfgs, err
}

func NewRPCClient() (RPCClient, error) {
	var err error
	client, err := rpc.Dial("unix", core.RPCSocket)
	return RPCClient{core: client}, err
}
