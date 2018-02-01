package main

import (
	"testing"
)

type DummyModule struct {
}

func (DummyModule) Name() string { return "DummyModule" }

func (DummyModule) ActionSupport() []string {
	return []string{"TEST"}
}

func (m DummyModule) HandleAction(name string, args ...interface{}) error {
	return args[0].(*_Loop).Quit()
}

func TestLoop(t *testing.T) {
	l := NewLoop(nil)
	d := &DummyModule{}
	l.InstallModule(d)

	err := l.Emit("FOOBAR")
	if err == nil {
		t.Fatalf("The loop shouldn't handle FOOBAR")
	}

	go l.Emit("TEST", l)

	err = l.Start()
	if err != nil {
		t.Fatalf("The loop terminated with error %v", err)
	}
}
