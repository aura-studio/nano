package message

import (
	"errors"
	"sync"

	"github.com/aura-studio/nano/cluster/clusterpb"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/serialize"
	"github.com/aura-studio/nano/serialize/json"
	"github.com/aura-studio/nano/serialize/protobuf"
	"github.com/aura-studio/nano/serialize/rawstring"
)

var (
	ErrSerializerRouteNotFound = errors.New("serializer: route not found")
)

const (
	Unknown uint32 = iota
	JSON
	Protobuf
	RawString
)

type serializer struct {
	serializerMap map[uint32]map[string]uint32
	rw            sync.RWMutex
}

func newSerializer() *serializer {
	return &serializer{
		serializerMap: make(map[uint32]map[string]uint32),
	}
}

var Serializer = newSerializer()

func (s *serializer) Deserialize(route string, payload []byte, v interface{}) error {
	serializer, err := s.Route(route)
	if err != nil {
		return err
	}
	return serializer.Unmarshal(payload, v)
}

func (s *serializer) Serialize(route string, v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}
	serializer, err := s.Route(route)
	if err != nil {
		return nil, err
	}
	data, err := serializer.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *serializer) Register(items []*clusterpb.MessageItem) {
	s.rw.Lock()
	defer s.rw.Unlock()

	for _, item := range items {
		if _, ok := s.serializerMap[item.VersionNum]; !ok {
			s.serializerMap[item.VersionNum] = make(map[string]uint32)
		}
		s.serializerMap[item.VersionNum][item.Route] = item.Serializer
	}
}

func (s *serializer) Route(route string) (serialize.Serializer, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	if _, ok := s.serializerMap[env.VersionNum]; ok {
		if serializer, ok := s.serializerMap[env.VersionNum][route]; ok {
			return s.Instance(serializer), nil
		}
	}
	if _, ok := s.serializerMap[0]; ok {
		if serializer, ok := s.serializerMap[0][route]; ok {
			return s.Instance(serializer), nil
		}
	}
	return nil, ErrDictionaryRouteNotFound
}

func (s *serializer) Instance(serializer uint32) serialize.Serializer {
	switch serializer {
	case JSON:
		return json.NewSerializer()
	case Protobuf:
		return protobuf.NewSerializer()
	case RawString:
		return rawstring.NewSerializer()
	default:
		return rawstring.NewSerializer()
	}
}
