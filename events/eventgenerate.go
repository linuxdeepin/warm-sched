package events

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Generator interface {
	Scope() string
	Check(ids []string) []string
}

type _Wait struct {
	events   []string
	callback func()
}

type _Manager struct {
	lock       sync.RWMutex //protect cache and generators
	cache      map[string]map[string]bool
	generators map[string]Generator

	waitlock sync.Mutex //protect waits and counts
	waits    map[int]_Wait
	counts   int
}

func (m *_Manager) newid() int {
	m.counts++
	v := m.counts
	return v
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
	waits:      make(map[int]_Wait),
}

func IsSupport(scope string) bool    { return _M_.isSupport(scope) }
func Pendings(scope string) []string { return _M_.Pendings(scope) }
func Emit(scope string, id string)   { _M_.Emit(scope, id) }
func Register(g Generator)           { _M_.Register(g) }
func Run() error                     { return _M_.Run() }

func Check(es []string) []string { return _M_.Check(es) }
func (m *_Manager) Check(es []string) []string {
	var ret []string
	for _, raw := range es {
		scope, id, ok := splitEvent(raw)
		if !ok {
			continue
		}
		g, ok := m.generators[scope]
		if !ok {
			continue
		}
		if len(g.Check([]string{id})) == 1 {
			ret = append(ret, raw)
		}
	}
	return ret
}

func Connect(es []string, callback func()) error { return _M_.Connect(es, callback) }

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

func (m *_Manager) Connect(es []string, callback func()) error {
	if len(es) == 0 {
		return fmt.Errorf("At least one event to connect")
	}
	err := m.setup(es)
	if err != nil {
		return err
	}
	m.waitlock.Lock()
	m.waits[m.newid()] = _Wait{es, callback}
	m.waitlock.Unlock()
	return nil
}

func (m *_Manager) setup(es []string) error {
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
	return nil
}

// Emit mark the event is appeared.
// Return the pendings and stop the generator if the scope hasn't any pendings
func (m *_Manager) Emit(scope string, id string) {
	m.lock.Lock()
	m.cache[scope][id] = true
	fmt.Printf("Emit \"%s:%s\"\n", scope, id)
	m.lock.Unlock()
}

func (m *_Manager) isDone(es []string) bool {
	m.lock.Lock()
	for _, raw := range es {
		scope, id, ok := splitEvent(raw)
		if !ok || !m.cache[scope][id] {
			m.lock.Unlock()
			return false
		}
	}
	m.lock.Unlock()
	return true
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

func (m *_Manager) Run() error {
	go m.poll()

	for {
		var dels []int
		m.waitlock.Lock()
		for id, w := range m.waits {
			if m.isDone(w.events) {
				dels = append(dels, id)
				if w.callback != nil {
					w.callback()
				}
			}
		}
		for _, id := range dels {
			delete(m.waits, id)
		}
		if len(m.waits) == 0 {
			m.waitlock.Unlock()
			return nil
		}
		m.waitlock.Unlock()
		fmt.Println("Waits...", m.waits)
		time.Sleep(time.Second)
	}
}

func (m *_Manager) poll() {
	fmt.Println("Start polling")
	for {
		anything := false
		for scope, g := range m.generators {
			pending := m.Pendings(scope)
			if len(pending) == 0 {
				continue
			}
			anything = true
			for _, id := range g.Check(pending) {
				m.Emit(scope, id)
			}
		}
		if !anything {
			break
		}
		time.Sleep(time.Second)
	}
	fmt.Println("Quit polling")
}
