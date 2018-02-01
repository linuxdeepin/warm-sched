package main

import (
	"../core"
	"net/rpc"
	"os"
)

var SOCKET = os.ExpandEnv("${XDG_RUNTIME_DIR}/warm-sched.socket")
var RPCName = "daemon"

type RPCClient struct {
	core *rpc.Client
}

func (c RPCClient) Capture() (*core.Snapshot, error) {
	var snap core.Snapshot
	err := c.core.Call(RPCName+".Capture", "", &snap)
	return &snap, err
}

func NewRPCClient() (RPCClient, error) {
	var err error
	client, err := rpc.Dial("unix", SOCKET)
	return RPCClient{core: client}, err
}
