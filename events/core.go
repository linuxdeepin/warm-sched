package events

import (
	"context"
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
	callback EventCallback
}

type EventCallback func() bool

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

func Pendings(scope string) []string { return _M_.Pendings(scope) }
func Check(es []string) []string     { return _M_.Check(es) }

// Connect 连接事件，callback 返回 true 时当事件发生后不会消除事件处理器，反之会消除事件处理器。
func Connect(es []string, callback EventCallback) error { return _M_.Connect(es, callback) }

func Register(g Generator)          { _M_.Register(g) }
func Run(ctx context.Context) error { return _M_.Run(ctx) }

func Scopes() []string { return _M_.Scopes() }

func Emit(event string) {
	scope, id, ok := splitEvent(event)
	if !ok {
		return
	}
	// 仅仅允许 x11 scope
	if scope != "x11" {
		return
	}
	_M_.Emit(scope, id)
}

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

func splitEvent(raw string) (string, string, bool) {
	fs := strings.SplitN(raw, ":", 2)
	if len(fs) != 2 {
		return "", "", false
	}
	return fs[0], fs[1], true
}

func SplitEvent(raw string) (string, string, bool) {
	return splitEvent(raw)
}

func (m *_Manager) Scopes() []string {
	var ret []string
	m.lock.RLock()
	for s := range m.generators {
		ret = append(ret, s)
	}
	m.lock.RUnlock()
	return ret
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

func (m *_Manager) Connect(es []string, callback EventCallback) error {
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

// reset 重置事件为未发生状态，仅会处理 x11 scope 的事件，其它忽略。
// 如果参数 es 包含 x11 scope 时间，reset 之后 isDone 会返回 false。
func (m *_Manager) reset(es []string) {
	m.lock.Lock()
	for _, event := range es {
		scope, id, ok := splitEvent(event)
		if ok {
			// 仅允许 reset x11 scope 事件
			if scope == "x11" {
				if _, exists := m.cache[scope][id]; exists {
					m.cache[scope][id] = false
					// 设为 false 后，事件恢复为未发生状态，继续被监控。
				}
			}
		}
	}
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

func (m *_Manager) Run(ctx context.Context) error {
	go m.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var dels []int
			m.waitlock.Lock()
			for id, w := range m.waits {
				if m.isDone(w.events) {
					shouldDel := true
					if w.callback != nil {
						// callback 返回 true 时不删除这个事件处理器
						if w.callback() {
							shouldDel = false
							m.reset(w.events)
						}
					}
					if shouldDel {
						dels = append(dels, id)
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
			time.Sleep(time.Second)
		}
	}
}

func (m *_Manager) poll(ctx context.Context) {
	fmt.Println("Start polling")
	for {
		select {
		case <-ctx.Done():
			goto end
		default:
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
				goto end
			}
			time.Sleep(time.Second)
		}
	}
end:
	fmt.Println("Quit polling")
}
