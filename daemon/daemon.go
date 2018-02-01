package main

import (
	"fmt"
	"os"
)

var SOCKET = os.ExpandEnv("${XDG_RUNTIME_DIR}/warm-sched.socket")

type Message struct {
	Name string
	Body interface{}
}

func main() {
	err := NewRPCService("unix", SOCKET)
	if err != nil {
		fmt.Fprintf(os.Stderr, "E:%v\n", err)
		return
	}
}
