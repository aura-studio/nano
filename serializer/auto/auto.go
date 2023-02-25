package auto

import (
	"encoding/json"

	"google.golang.org/protobuf/proto"
)

type serializer struct{}

func NewSerializer() *serializer {
	return &serializer{}
}

var Serializer = NewSerializer()

func (s *serializer) Marshal(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case *string:
		return []byte(*v), nil
	case proto.Message:
		return proto.Marshal(v)
	default:
		return json.Marshal(v)
	}
}

func (s *serializer) Unmarshal(data []byte, v interface{}) error {
	switch v := v.(type) {
	case *string:
		*v = string(data)
		return nil
	case proto.Message:
		return proto.Unmarshal(data, v)
	default:
		return json.Unmarshal(data, v)
	}
}
