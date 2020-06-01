package message

import (
	"sync"
)

var (
	// Routes is a map from route to code
	Routes = make(map[string]uint16)
	// Codes is a map from code to route
	Codes = make(map[uint16]string)
	rw    sync.RWMutex
)

// ReadDictionary returns dictionary for compressed route.
func ReadDictionary() (map[string]uint16, map[uint16]string) {
	rw.RLock()
	defer rw.RUnlock()

	return Routes, Codes
}

// WriteDictionaryItem is to set dictionary item when server registers.
func WriteDictionaryItem(route string, code uint16) (map[string]uint16, map[uint16]string) {
	rw.Lock()
	defer rw.Unlock()

	if code > 0 {
		Routes[route] = code
		Codes[code] = route
	}

	return Routes, Codes
}

// WriteDictionary is to set dictionary when new route dictionray is found.
func WriteDictionary(dict map[string]uint16) (map[string]uint16, map[uint16]string) {
	rw.Lock()
	defer rw.Unlock()

	for route, code := range dict {
		if code > 0 {
			Routes[route] = code
			Codes[code] = route
		}
	}

	return Routes, Codes
}

// ParseDictionary parses dictionary into routes and codes independently
func ParseDictionary(dict map[string]uint16) (map[string]uint16, map[uint16]string) {
	routes := make(map[string]uint16)
	codes := make(map[uint16]string)
	for route, code := range dict {
		if code > 0 {
			routes[route] = code
			codes[code] = route
		}
	}
	return routes, codes
}
