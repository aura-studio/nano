package message

import (
	"errors"
	"sync"

	"github.com/aura-studio/nano/cluster/clusterpb"
)

var (
	ErrDictionaryRouteNotFound = errors.New("dictionary: route not found")
	ErrDictionaryCodeNotFound  = errors.New("dictionary: code not found")
)

type dictionary struct {
	routeMap map[uint32]map[string]uint32
	codeMap  map[uint32]map[uint32]string
	rw       sync.RWMutex
}

func newDictionary() *dictionary {
	return &dictionary{
		routeMap: make(map[uint32]map[string]uint32),
		codeMap:  make(map[uint32]map[uint32]string),
	}
}

var Dictionary = newDictionary()

func (d *dictionary) Route(version uint32, route string) (uint32, error) {
	d.rw.RLock()
	defer d.rw.RUnlock()

	if _, ok := d.routeMap[version]; ok {
		if code, ok := d.routeMap[version][route]; ok {
			return code, nil
		}
	}
	if _, ok := d.routeMap[0]; ok {
		if code, ok := d.routeMap[version][route]; ok {
			return code, nil
		}
	}
	return 0, ErrDictionaryRouteNotFound
}

func (d *dictionary) Code(version uint32, code uint32) (string, error) {
	d.rw.RLock()
	defer d.rw.RUnlock()

	if _, ok := d.codeMap[version]; ok {
		if code, ok := d.codeMap[version][code]; ok {
			return code, nil
		}
	}
	if _, ok := d.codeMap[0]; ok {
		if code, ok := d.codeMap[version][code]; ok {
			return code, nil
		}
	}
	return "", ErrDictionaryCodeNotFound
}

func (d *dictionary) Register(items []*clusterpb.MessageItem) {
	d.rw.Lock()
	defer d.rw.Unlock()

	for _, item := range items {
		if item.Code > 0 {
			if _, ok := d.routeMap[item.VersionNum]; !ok {
				d.routeMap[item.VersionNum] = make(map[string]uint32)
			}
			d.routeMap[item.VersionNum][item.Route] = item.Code
			if _, ok := d.codeMap[item.VersionNum]; !ok {
				d.codeMap[item.VersionNum] = make(map[uint32]string)
			}
			d.codeMap[item.VersionNum][item.Code] = item.Route
		}
	}
}
