package connector

import (
	"github.com/lonng/nano/log"
	"github.com/lonng/nano/serialize"
)

type (
	// Options contains some configurations for connector
	Options struct {
		name       string               // component name
		dictionary map[string]uint16    // Dictionary info
		serializer serialize.Serializer // serializer for connector
		wsPath     string               //websocket path
		logger     log.Logger           // logger
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

// WithLogger overrides the default logger
func WithLogger(l log.Logger) Option {
	return func(opt *Options) {
		opt.logger = l
	}
}
