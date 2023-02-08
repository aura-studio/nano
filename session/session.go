package session

import (
	"net"
	"sync/atomic"

	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/snowflake"
	"github.com/mohae/deepcopy"
)

type Session struct {
	*KernelSession
	lastMid uint64
}

func (s *Session) LastMid() uint64 {
	return s.lastMid
}

func (s *Session) RPC(route string, v interface{}) error {
	return s.KernelSession.entity.RPC(s.lastMid, route, v)
}

func (s *Session) Response(route string, v interface{}) error {
	return s.KernelSession.entity.Response(s.lastMid, route, v)
}

// New returns a new session instance
// a NetworkEntity is a low-level network instance
func New(entity NetworkEntity, id uint64) *Session {
	return &Session{
		KernelSession: NewKernelSession(entity, id),
	}
}

func Context(s *Session, lastMid uint64) *Session {
	return &Session{
		KernelSession: s.KernelSession,
		lastMid:       lastMid,
	}
}

// ID returns the session id
func (s *Session) ID() uint64 {
	return s.id
}

func (s *Session) SID() uint64 {
	if env.SnowflakeNode != nil {
		id := snowflake.ID(s.id)
		return uint64(id.Time(env.SnowflakeNode)<<10 + id.Step(env.SnowflakeNode))
	}
	return s.id % (1 << 32) // SID is low 32 bit of id
}

func (s *Session) NID() uint64 { // NodeID
	if env.SnowflakeNode != nil {
		id := snowflake.ID(s.id)
		return uint64(id.Node(env.SnowflakeNode))
	}
	return s.id >> 32 // SID is high 32 bit of id
}

// UID returns uid that bind to current session
func (s *Session) UID() int64 {
	return atomic.LoadInt64(&s.uid)
}

func (s *Session) Version() string {
	return s.version
}

func (s *Session) BindVersion(version string) {
	s.version = version
}

func (s *Session) ShortVer() uint32 {
	return s.shortVer
}

func (s *Session) BindShortVer(shortVer uint32) {
	s.shortVer = shortVer
}

// Bind bind UID to current session
func (s *Session) BindUID(uid int64) {
	atomic.StoreInt64(&s.uid, uid)
}

func (s *Session) BindBranch(branch uint32) {
	s.branch = branch
}

func (s *Session) Branch() uint32 {
	return s.branch
}

// Close terminate current session, session related data will not be released,
// all related data should be Clear explicitly in Session closed callback
func (s *Session) Close() {
	s.entity.Close()
}

// RemoteAddr returns the remote network address.
func (s *Session) RemoteAddr() net.Addr {
	return s.entity.RemoteAddr()
}

// Remove delete data associated with the key from session storage
func (s *Session) Remove(key string) {
	s.Lock()
	defer s.Unlock()

	delete(s.data, key)
}

// Set associates value with the key in session storage
func (s *Session) Set(key string, value interface{}) {
	s.Lock()
	defer s.Unlock()

	s.data[key] = value
}

// HasKey decides whether a key has associated value
func (s *Session) HasKey(key string) bool {
	s.RLock()
	defer s.RUnlock()

	_, has := s.data[key]
	return has
}

// Int returns the value associated with the key as a int.
func (s *Session) Int(key string) int {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(int)
	if !ok {
		return 0
	}
	return value
}

// Int8 returns the value associated with the key as a int8.
func (s *Session) Int8(key string) int8 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(int8)
	if !ok {
		return 0
	}
	return value
}

// Int16 returns the value associated with the key as a int16.
func (s *Session) Int16(key string) int16 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(int16)
	if !ok {
		return 0
	}
	return value
}

// Int32 returns the value associated with the key as a int32.
func (s *Session) Int32(key string) int32 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(int32)
	if !ok {
		return 0
	}
	return value
}

// Int64 returns the value associated with the key as a int64.
func (s *Session) Int64(key string) int64 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(int64)
	if !ok {
		return 0
	}
	return value
}

// Uint returns the value associated with the key as a uint.
func (s *Session) Uint(key string) uint {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(uint)
	if !ok {
		return 0
	}
	return value
}

// Uint8 returns the value associated with the key as a uint8.
func (s *Session) Uint8(key string) uint8 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(uint8)
	if !ok {
		return 0
	}
	return value
}

// Uint16 returns the value associated with the key as a uint16.
func (s *Session) Uint16(key string) uint16 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(uint16)
	if !ok {
		return 0
	}
	return value
}

// Uint32 returns the value associated with the key as a uint32.
func (s *Session) Uint32(key string) uint32 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(uint32)
	if !ok {
		return 0
	}
	return value
}

// Uint64 returns the value associated with the key as a uint64.
func (s *Session) Uint64(key string) uint64 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(uint64)
	if !ok {
		return 0
	}
	return value
}

// Float32 returns the value associated with the key as a float32.
func (s *Session) Float32(key string) float32 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(float32)
	if !ok {
		return 0
	}
	return value
}

// Float64 returns the value associated with the key as a float64.
func (s *Session) Float64(key string) float64 {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return 0
	}

	value, ok := v.(float64)
	if !ok {
		return 0
	}
	return value
}

// Bool returns the value associated with the key as a boolean.
func (s *Session) Bool(key string) bool {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return false
	}

	value, ok := v.(bool)
	if !ok {
		return false
	}
	return value
}

// String returns the value associated with the key as a string.
func (s *Session) String(key string) string {
	s.RLock()
	defer s.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return ""
	}

	value, ok := v.(string)
	if !ok {
		return ""
	}
	return value
}

// Value returns the value associated with the key as a interface{}.
func (s *Session) Value(key string) interface{} {
	s.RLock()
	defer s.RUnlock()

	return s.data[key]
}

// State returns all session state
func (s *Session) State() map[string]interface{} {
	s.RLock()
	defer s.RUnlock()

	return s.data
}

// Restore session state after reconnect
func (s *Session) Restore(data map[string]interface{}) {
	s.Lock()
	defer s.Unlock()

	s.data = data
}

// Clear releases all data related to current session
func (s *Session) Clear() {
	s.Lock()
	defer s.Unlock()

	s.uid = 0
	s.data = map[string]interface{}{}
}

// On is to register callback on events
func (s *Session) On(ev string, f func(s *Session)) {
	s.onEvents[ev] = append(s.onEvents[ev], func(s *Session, i ...interface{}) {
		f(s)
	})
}

// OnData is to register callback on events
func (s *Session) OnData(ev string, f func(s *Session, i ...interface{})) {
	s.onEvents[ev] = append(s.onEvents[ev], f)
}

// Trigger is to trigger an event with args
func (s *Session) Trigger(ev string, i ...interface{}) {
	data := deepcopy.Copy(i).([]interface{})
	for _, f := range s.onEvents[ev] {
		copyData := deepcopy.Copy(data).([]interface{})
		f(s, copyData...)
	}
}
