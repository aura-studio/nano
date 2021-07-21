package message

import (
	"errors"
	"sync"

	"github.com/aura-studio/nano/cluster/clusterpb"
	"github.com/mohae/deepcopy"
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
	// RW lock
	rw sync.RWMutex
}

var ErrIndexRouteNotFound = errors.New("dictionary: index route not found")
var ErrIndexCodeNotFound = errors.New("dictionary: index code not found")

var memberDictionary = &MemberDictionary{
	routes: make(map[string]uint32),
	codes:  make(map[uint32]string),
}

var EmptyDictionary = &MemberDictionary{
	routes: make(map[string]uint32),
	codes:  make(map[uint32]string),
}

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
	d.rw.RLock()
	defer d.rw.RUnlock()

	return deepcopy.Copy(d).(*MemberDictionary)
}

func (d *MemberDictionary) write(route string, code uint32) {
	d.rw.Lock()
	defer d.rw.Unlock()

	if code > 0 {
		d.routes[route] = code
		d.codes[code] = route
	}
}

func (d *MemberDictionary) writeClusterItems(items []*clusterpb.DictionaryItem) {
	d.rw.Lock()
	defer d.rw.Unlock()

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
