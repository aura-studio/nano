package message

import (
	"sync"

	"github.com/aura-studio/nano/cluster/clusterpb"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/serialize"
	"github.com/aura-studio/nano/serialize/json"
	"github.com/aura-studio/nano/serialize/protobuf"
	"github.com/aura-studio/nano/serialize/rawstring"
)

const (
	Unknown uint32 = iota
	JSON
	Protobuf
	RawString
)

var (
	jsonSerializer      = json.NewSerializer()
	protobufSerializer  = protobuf.NewSerializer()
	rawStringSerializer = rawstring.NewSerializer()
)

var (
	// Serializers is a map from route to serializer
	Serializers = make(map[string]serialize.Serializer)

	rw sync.RWMutex
)

func GetSerializerType(s serialize.Serializer) uint32 {
	switch s.(type) {
	case *json.Serializer:
		return JSON
	case *protobuf.Serializer:
		return Protobuf
	case *rawstring.Serializer:
		return RawString
	default:
		return Unknown
	}
}

func GetSerializer(typ uint32) serialize.Serializer {
	switch typ {
	case JSON:
		return jsonSerializer
	case Protobuf:
		return protobufSerializer
	case RawString:
		return rawStringSerializer
	default:
		return env.Serializer
	}
}

// DuplicateSerializers returns serializers for compressed route.
func DuplicateSerializers() map[string]serialize.Serializer {
	rw.RLock()
	defer rw.RUnlock()

	return Serializers
}

// WriteSerializerItem is to set serializer item when server registers.
func WriteSerializerItem(route string, typ uint32) map[string]serialize.Serializer {
	rw.Lock()
	defer rw.Unlock()

	Serializers[route] = GetSerializer(typ)

	return Serializers
}

// WriteSerializers is to set serializers when new serializer dictionary is found.
func WriteSerializers(items []*clusterpb.DictionaryItem) map[string]serialize.Serializer {
	rw.Lock()
	defer rw.Unlock()

	for _, item := range items {
		Serializers[item.Route] = GetSerializer(item.Serializer)
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

func RouteSerialize(serializers map[string]serialize.Serializer, route string, v interface{}) ([]byte, error) {
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
