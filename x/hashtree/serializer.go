package hashtree

import (
	"github.com/lonng/nano/serialize"
	"github.com/lonng/nano/serialize/json"
	"github.com/lonng/nano/serialize/protobuf"
)

var (
	jsonSerializer  serialize.Serializer = json.NewSerializer()
	protoSerializer serialize.Serializer = protobuf.NewSerializer()
)
