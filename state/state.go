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
}

func NewState() *State {
	return &State{
		data: sync.Map{},
		mods: make([]StateMod, 0),
	}
}

func (s *State) Set(k string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mods = append(s.mods, StateMod{
		Key:    k,
		Action: Set,
		Value:  cast.ToString(v),
	})
}

func (s *State) Append(k string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mods = append(s.mods, StateMod{
		Key:    k,
		Action: Append,
		Value:  cast.ToString(v),
	})
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
