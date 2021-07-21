package message

import (
	"errors"

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
}

var ErrIndexRouteNotFound = errors.New("dictionary: index route not found")
var ErrIndexCodeNotFound = errors.New("dictionary: index code not found")

var memberDictionary = &MemberDictionary{}

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

	return deepcopy.Copy(d).(*MemberDictionary)
}

func (d *MemberDictionary) write(route string, code uint32) {
	rw.RLock()
	defer rw.RUnlock()

	if code > 0 {
		memberDictionary.routes[route] = code
		memberDictionary.codes[code] = route
	}
}

func (d *MemberDictionary) writeClusterItems(items []*clusterpb.DictionaryItem) {
	rw.RLock()
	defer rw.RUnlock()

	for _, item := range items {
		if item.Code > 0 {
			memberDictionary.routes[item.Route] = item.Code
			memberDictionary.codes[item.Code] = item.Route
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
