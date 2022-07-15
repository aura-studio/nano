package state

import (
	"sync"

	"github.com/spf13/cast"
)

const (
	Set    = "Set"
	Append = "Append"
)

type StateMod struct {
	Action string
	Key    string
	Value  string
}

type State struct {
	data sync.Map

	mods []StateMod
	mu   sync.RWMutex

	facade *Facade
}

func NewState() *State {
	state := &State{
		data: sync.Map{},
		mods: make([]StateMod, 0),
	}
	state.facade = NewFacade(state)
	return state
}

func (s *State) Facade() *Facade {
	return s.facade
}

func (s *State) Set(k string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mods = append(s.mods, StateMod{
		Key:    k,
		Action: Set,
		Value:  cast.ToString(v),
	})

	s.facade.set(k, v)
}

func (s *State) Append(k string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mods = append(s.mods, StateMod{
		Key:    k,
		Action: Append,
		Value:  cast.ToString(v),
	})

	s.facade.append(k, v)
}

func (s *State) Dump() []StateMod {
	s.mu.Lock()
	defer s.mu.Unlock()

	defer func() {
		s.mods = nil
	}()
	return s.mods
}

func (s *State) Data() map[string]string {
	data := make(map[string]string)
	s.data.Range(func(key, value interface{}) bool {
		data[cast.ToString(key)] = cast.ToString(value)
		return true
	})
	return data
}

func (s *State) Update(data map[string]string) {
	for k, v := range data {
		s.data.Store(k, v)
	}
	s.facade.update(data)
}

func (s *State) Int(key string) int {
	v, ok := s.data.Load(key)
	if !ok {
		return 0
	}
	return cast.ToInt(v)
}

func (s *State) Int64(key string) int64 {
	v, ok := s.data.Load(key)
	if !ok {
		return 0
	}
	return cast.ToInt64(v)
}

func (s *State) Float64(key string) float64 {
	v, ok := s.data.Load(key)
	if !ok {
		return 0
	}
	return cast.ToFloat64(v)
}

func (s *State) String(key string) string {
	v, ok := s.data.Load(key)
	if !ok {
		return ""
	}
	return cast.ToString(v)
}

func (s *State) Bool(key string) bool {
	v, ok := s.data.Load(key)
	if !ok {
		return false
	}
	return cast.ToBool(v)
}

type Facade struct {
	data  sync.Map
	state *State
}

func NewFacade(state *State) *Facade {
	return &Facade{
		data:  sync.Map{},
		state: state,
	}
}

func (f *Facade) Facade() *Facade {
	return f
}

func (f *Facade) Set(k string, v interface{}) {
	f.state.Set(k, v)
}

func (f *Facade) set(k string, v interface{}) {
	f.data.Store(k, v)
}

func (f *Facade) Append(k string, v interface{}) {
	f.state.Append(k, v)
}

func (f *Facade) append(k string, v interface{}) {
	if _, ok := f.data.Load(k); !ok {
		f.data.Store(k, "0")
	}
	vOrigin, _ := f.data.Load(k)
	f.data.Store(k, cast.ToString(cast.ToFloat64(vOrigin)+cast.ToFloat64(v)))
}

func (f *Facade) Dump() []StateMod {
	return f.state.Dump()
}

func (f *Facade) Data() map[string]string {
	data := make(map[string]string)
	f.data.Range(func(key, value interface{}) bool {
		data[cast.ToString(key)] = cast.ToString(value)
		return true
	})
	return data
}

func (f *Facade) Update(data map[string]string) {
	f.state.Update(data)
}

func (f *Facade) update(data map[string]string) {
	for k, v := range data {
		f.data.Store(k, v)
	}
}

func (f *Facade) Int(key string) int {
	v, ok := f.data.Load(key)
	if !ok {
		return 0
	}
	return cast.ToInt(v)
}

func (f *Facade) Int64(key string) int64 {
	v, ok := f.data.Load(key)
	if !ok {
		return 0
	}
	return cast.ToInt64(v)
}

func (f *Facade) Float64(key string) float64 {
	v, ok := f.data.Load(key)
	if !ok {
		return 0
	}
	return cast.ToFloat64(v)
}

func (f *Facade) String(key string) string {
	v, ok := f.data.Load(key)
	if !ok {
		return ""
	}
	return cast.ToString(v)
}

func (f *Facade) Bool(key string) bool {
	v, ok := f.data.Load(key)
	if !ok {
		return false
	}
	return cast.ToBool(v)
}

func (f *Facade) Value(key string) interface{} {
	v, ok := f.data.Load(key)
	if !ok {
		return nil
	}
	return v
}
