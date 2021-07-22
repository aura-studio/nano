package message

import (
	"errors"

	"github.com/aura-studio/nano/cluster/clusterpb"
	lua "github.com/yuin/gopher-lua"
)

type Dictionary interface {
	IndexRoute(string) (uint32, error)
	IndexCode(uint32) (string, error)
}

type MemberDictionary struct {
	// Routes is a map from route to code
	routes map[string]uint32
	// Codes is a map from code to route
	codes map[uint32]string
}

func NewMemberDictionary() *MemberDictionary {
	return &MemberDictionary{
		routes: make(map[string]uint32),
		codes:  make(map[uint32]string),
	}
}

var ErrIndexRouteNotFound = errors.New("dictionary: index route not found")
var ErrIndexCodeNotFound = errors.New("dictionary: index code not found")

var (
	memberDictionary = NewMemberDictionary()
	EmptyDictionary  = NewMemberDictionary()
)

func (d *MemberDictionary) IndexRoute(route string) (uint32, error) {
	if d == nil {
		return 0, ErrIndexRouteNotFound
	}
	code, ok := d.routes[route]
	if !ok {
		return 0, ErrIndexRouteNotFound
	}
	return code, nil
}

func (d *MemberDictionary) IndexCode(code uint32) (string, error) {
	if d == nil {
		return "", ErrIndexRouteNotFound
	}
	route, ok := d.codes[code]
	if !ok {
		return "", ErrIndexCodeNotFound
	}
	return route, nil
}

func (d *MemberDictionary) duplicate() *MemberDictionary {
	rw.RLock()
	defer rw.RUnlock()

	memberDictionary := NewMemberDictionary()

	for route, code := range d.routes {
		memberDictionary.routes[route] = code
	}

	for code, route := range d.codes {
		memberDictionary.codes[code] = route
	}

	return memberDictionary
}

func (d *MemberDictionary) write(route string, code uint32) {
	rw.Lock()
	defer rw.Unlock()

	if code > 0 {
		d.routes[route] = code
		d.codes[code] = route
	}
}

func (d *MemberDictionary) writeClusterItems(items []*clusterpb.DictionaryItem) {
	rw.Lock()
	defer rw.Unlock()

	for _, item := range items {
		if item.Code > 0 {
			d.routes[item.Route] = item.Code
			d.codes[item.Code] = item.Route
		}
	}
}

// DuplicateDictionary returns dictionary for compressed route.
func DuplicateDictionary() *MemberDictionary {
	return memberDictionary.duplicate()
}

// WriteDictionaryItem is to set dictionary item when server registers.
func WriteDictionaryItem(route string, code uint32) {
	memberDictionary.write(route, code)
}

// WriteDictionary is to set dictionary when new route dictionary is found.
func WriteDictionary(items []*clusterpb.DictionaryItem) {
	memberDictionary.writeClusterItems(items)
}

type LuaDictionary struct {
	// Routes is a map from route to code
	routes map[string]uint32
	// Codes is a map from code to route
	codes map[uint32]string
}

func NewLuaDictionary(file string) *LuaDictionary {
	d := &LuaDictionary{
		routes: make(map[string]uint32),
		codes:  make(map[uint32]string),
	}
	d.LoadLua(file)
	return d
}

func (d *LuaDictionary) IndexRoute(route string) (uint32, error) {
	if d == nil {
		return 0, ErrIndexRouteNotFound
	}
	code, ok := d.routes[route]
	if !ok {
		return 0, ErrIndexRouteNotFound
	}
	return code, nil
}

func (d *LuaDictionary) IndexCode(code uint32) (string, error) {
	if d == nil {
		return "", ErrIndexRouteNotFound
	}
	route, ok := d.codes[code]
	if !ok {
		return "", ErrIndexCodeNotFound
	}
	return route, nil
}

func (d *LuaDictionary) write(route string, code uint32) {
	if code > 0 {
		d.routes[route] = code
		d.codes[code] = route
	}
}

func (d *LuaDictionary) InitModule(l *lua.LState) int {
	var exports = map[string]lua.LGFunction{
		"Write": func(l *lua.LState) int {
			route := l.ToString(1)
			code := uint32(l.ToInt(2))
			d.write(route, code)
			return 0
		},
	}
	mod := l.SetFuncs(l.NewTable(), exports)
	l.Push(mod)
	return 1
}

func (d *LuaDictionary) LoadLua(file string) {
	l := lua.NewState()
	defer l.Close()
	l.PreloadModule("GoDictionary", d.InitModule)
	if err := l.DoFile(file); err != nil {
		panic(err)
	}
}
