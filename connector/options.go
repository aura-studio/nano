package connector

import (
	"github.com/lonng/nano/serialize"
)

type (
	// Options contains some configurations for connector
	Options struct {
		name       string               // component name
		dictionary map[string]uint16    // Dictionary info
		serializer serialize.Serializer // serializer for connector
		wsPath     string               //websocket path
	}

	// Option used to customize handler
	Option func(options *Options)
)

// WithName is used to name connector
func WithName(name string) Option {
	return func(opt *Options) {
		opt.name = name
	}
}

// WithDictionary is used to set compressed flag
func WithDictionary(dictionary map[string]uint16) Option {
	return func(opt *Options) {
		opt.dictionary = dictionary
	}
}

// WithSerializer customizes application serializer, which automatically Marshal
// and UnMarshal handler payload
func WithSerializer(serializer serialize.Serializer) Option {
	return func(opt *Options) {
		opt.serializer = serializer
	}
}

// WithWSPath set the websocket path
func WithWSPath(path string) Option {
	return func(opt *Options) {
		opt.wsPath = path
	}
}
