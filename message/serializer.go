package message

import (
	"github.com/lonng/nano/env"
	"github.com/lonng/nano/serialize"
	"github.com/lonng/nano/serialize/json"
	"github.com/lonng/nano/serialize/protobuf"
	"github.com/lonng/nano/serialize/rawstring"
)

const (
	Unknown uint16 = iota
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
)

func GetSerializerType(s serialize.Serializer) uint16 {
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

func GetSerializer(typ uint16) serialize.Serializer {
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

// ReadSerializers returns serializers for compressed route.
func ReadSerializers() map[string]serialize.Serializer {
	rw.RLock()
	defer rw.RUnlock()

	return Serializers
}

// WriteSerializerItem is to set serializer item when server registers.
func WriteSerializerItem(route string, typ uint16) map[string]serialize.Serializer {
	rw.Lock()
	defer rw.Unlock()

	Serializers[route] = GetSerializer(typ)

	return Serializers
}

// WriteSerializers is to set serializers when new serializer dictionary is found.
func WriteSerializers(serializers map[string]uint16) map[string]serialize.Serializer {
	rw.Lock()
	defer rw.Unlock()

	for route, typ := range serializers {
		Serializers[route] = GetSerializer(typ)
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
