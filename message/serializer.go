package message

import (
	"sync"

	"github.com/aura-studio/nano/cluster/clusterpb"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/serializer"
)

var (
	// SerializerTypeMap is a map from route to serializer
	SerializerTypeMap = make(map[string]serializer.SerializerType)

	rw sync.RWMutex
)

// DuplicateSerializerTypeMap returns serializers for compressed route.
func DuplicateSerializerTypeMap() map[string]serializer.SerializerType {
	rw.RLock()
	defer rw.RUnlock()

	return SerializerTypeMap
}

// WriteSerializerItem is to set serializer item when server registers.
func WriteSerializerItem(route string, serializerType serializer.SerializerType) map[string]serializer.SerializerType {
	rw.Lock()
	defer rw.Unlock()

	SerializerTypeMap[route] = serializerType

	return SerializerTypeMap
}

// WriteSerializers is to set serializers when new serializer dictionary is found.
func WriteSerializers(items []*clusterpb.DictionaryItem) map[string]serializer.SerializerType {
	rw.Lock()
	defer rw.Unlock()

	for _, item := range items {
		SerializerTypeMap[item.Route] = serializer.SerializerType(item.Serializer)
	}

	return SerializerTypeMap
}

func Serialize(v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	data, err := env.SerializerType.Serializer().Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func RouteSerialize(serializerTypes map[string]serializer.SerializerType, route string, v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	serializerType, ok := serializerTypes[route]
	if !ok {
		serializerType = env.SerializerType
	}
	data, err := serializerType.Serializer().Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}
