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

func (c RPCClient) Capture(cfg core.CaptureConfig) (*core.Snapshot, error) {
	var snap core.Snapshot
	err := c.core.Call(RPCName+".Capture", cfg, &snap)
	return &snap, err
}

func (c RPCClient) ListConfig() ([]*core.SnapshotConfig, error) {
	var cfgs []*core.SnapshotConfig
	err := c.core.Call(RPCName+".ListConfig", true, &cfgs)
	return cfgs, err
}

func NewRPCClient() (RPCClient, error) {
	var err error
	client, err := rpc.Dial("unix", SOCKET)
	return RPCClient{core: client}, err
}
