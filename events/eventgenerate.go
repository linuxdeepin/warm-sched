package events

import (
	"fmt"
	"strings"
	"sync"
)

type Generator interface {
	Scope() string
	Prepare(ids []string) error
	Run() error
	Stop()
}

type _Manager struct {
	lock       sync.RWMutex
	cache      map[string]map[string]bool
	generators map[string]Generator
}

func (m *_Manager) Register(g Generator) {
	m.lock.Lock()
	scope := g.Scope()
	m.cache[scope] = make(map[string]bool)
	m.generators[scope] = g
	m.lock.Unlock()
}

var _M_ = &_Manager{
	cache:      make(map[string]map[string]bool),
	generators: make(map[string]Generator),
}

func IsSupport(scope string) bool           { return _M_.isSupport(scope) }
func WaitAll(es ...string) error            { return _M_.WaitAll(es...) }
func Pendings(scope string) []string        { return _M_.Pendings(scope) }
func Emit(scope string, id string) []string { return _M_.Emit(scope, id) }
func Register(g Generator)                  { _M_.Register(g) }

func splitEvent(raw string) (string, string, bool) {
	fs := strings.SplitN(raw, ":", 2)
	if len(fs) != 2 {
		return "", "", false
	}
	return fs[0], fs[1], true
}

func (m *_Manager) isSupport(raw string) bool {
	scope, _, ok := splitEvent(raw)
	if !ok {
		return false
	}
	m.lock.Lock()
	_, ok = m.generators[scope]
	m.lock.Unlock()
	return ok
}

func (m *_Manager) WaitAll(es ...string) error {
	for _, e := range es {
		scope, id, ok := splitEvent(e)
		if !ok {
			return fmt.Errorf("Illegal event format %q", e)
		}
		m.lock.Lock()
		if _, ok := m.generators[scope]; !ok {
			m.lock.Unlock()
			return fmt.Errorf("Doesn't support scope of %q events", scope)
		}
		m.cache[scope][id] = false
		m.lock.Unlock()
	}
	return m.run()
}

// Emit mark the event is appeared.
// Return the pendings and stop the generator if the scope hasn't any pendings
func (m *_Manager) Emit(scope string, id string) []string {
	m.lock.Lock()
	g, ok := m.generators[scope]
	if !ok {
		m.lock.Unlock()
		panic("BUG ON Emit Event")
	}
	m.cache[scope][id] = true
	fmt.Printf("Emit \"%s:%s\"\n", scope, id)
	p := m.pendings(scope)
	if len(p) == 0 {
		go g.Stop()
	}
	m.lock.Unlock()
	return p
}

// Pendings return the number of floats
func (m *_Manager) Pendings(scope string) []string {
	m.lock.Lock()
	ret := m.pendings(scope)
	m.lock.Unlock()
	return ret
}
func (m *_Manager) pendings(scope string) []string {
	var ret []string
	for k, v := range m.cache[scope] {
		if !v {
			ret = append(ret, k)
		}
	}
	return ret
}

func (m *_Manager) run() error {
	var g sync.WaitGroup
	for scope := range m.cache {
		g.Add(1)
		go func(s string) {
			err := m.startScope(s)
			if err != nil {
				fmt.Printf("Error when monitor %q -> %v\n", s, err)
			}
			g.Done()
		}(scope)
	}
	g.Wait()
	return nil
}

func (m *_Manager) startScope(scope string) error {
	pendings := m.Pendings(scope)
	if len(pendings) == 0 {
		return nil
	}

	m.lock.Lock()
	g := m.generators[scope]
	m.lock.Unlock()

	err := g.Prepare(pendings)
	if err != nil {
		fmt.Println("ERROR:", err)
		return err
	}
	fmt.Printf("START scope %q\n", scope)
	err = g.Run()
	fmt.Printf("Stop scope %q\n", scope)
	return err
}
