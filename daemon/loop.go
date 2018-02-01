package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type _Loop struct {
	buf chan _Event

	logger io.Writer

	callbacksLock sync.RWMutex
	callbacks     map[string][]EventHandle
}

func NewLoop(logger io.Writer) *_Loop {
	if logger == nil {
		logger = os.Stderr
	}

	return &_Loop{
		logger:    logger,
		buf:       make(chan _Event),
		callbacks: make(map[string][]EventHandle),
	}
}

type _Event struct {
	Name string
	Args []interface{}
}

type EventHandle func(ename string, args ...interface{}) error

type Module interface {
	Name() string
	ActionSupport() []string
	HandleAction(name string, args ...interface{}) error
}

func (l *_Loop) log(fmtStr string, args ...interface{}) {
	fmt.Fprintf(l.logger, fmtStr, args...)
}

func (l *_Loop) Start() error {
	l.log("Starting Loop\n")
	return l.loop()
}

func (l *_Loop) Quit() error {
	close(l.buf)
	return nil
}

func (l *_Loop) Emit(name string, args ...interface{}) error {
	l.callbacksLock.RLock()
	n := len(l.callbacks[name])
	l.callbacksLock.RUnlock()
	if n == 0 {
		return fmt.Errorf("There hasn't module support handle event %s", name)
	}
	l.buf <- _Event{name, args}
	return nil
}

func (l *_Loop) loop() error {
	for e := range l.buf {
		l.callbacksLock.Lock()
		for _, c := range l.callbacks[e.Name] {
			go c(e.Name, e.Args...)
		}
		l.callbacksLock.Unlock()
	}
	return nil
}
func (l *_Loop) connect(name string, cb EventHandle) {
	l.callbacksLock.Lock()
	l.callbacks[name] = append(l.callbacks[name], cb)
	l.callbacksLock.Unlock()
}

func (l *_Loop) InstallModule(mod Module) error {
	as := mod.ActionSupport()
	if len(as) == 0 {
		return fmt.Errorf("Module %s hasn't any useful", mod.Name())
	}
	l.log("Init module %s %v\n", mod.Name(), as)

	for _, a := range as {
		l.connect(a, mod.HandleAction)
	}
	return nil
}
