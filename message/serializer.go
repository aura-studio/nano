package message

import (
	"sync"

	"github.com/aura-studio/nano/cluster/clusterpb"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/serializer"
)

var (
	// Serializers is a map from route to serializer
	Serializers = make(map[string]serializer.Serializer)

	rw sync.RWMutex
)

// DuplicateSerializers returns serializers for compressed route.
func DuplicateSerializers() map[string]serializer.Serializer {
	rw.RLock()
	defer rw.RUnlock()

	return Serializers
}

// WriteSerializerItem is to set serializer item when server registers.
func WriteSerializerItem(route string, typ serializer.SerializerType) map[string]serializer.Serializer {
	rw.Lock()
	defer rw.Unlock()

	Serializers[route] = typ.Serializer()

	return Serializers
}

// WriteSerializers is to set serializers when new serializer dictionary is found.
func WriteSerializers(items []*clusterpb.DictionaryItem) map[string]serializer.Serializer {
	rw.Lock()
	defer rw.Unlock()

	for _, item := range items {
		Serializers[item.Route] = serializer.SerializerType(item.Serializer).Serializer()
	}

	return Serializers
}

func Serialize(v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	data, err := env.Serializer.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func RouteSerialize(serializers map[string]serializer.Serializer, route string, v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	serializer, ok := serializers[route]
	if !ok {
		serializer = env.Serializer
	}
	data, err := serializer.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}
