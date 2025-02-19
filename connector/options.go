package connector

import (
	"github.com/aura-studio/nano/codec"
	"github.com/aura-studio/nano/log"
	"github.com/aura-studio/nano/message"
	"github.com/aura-studio/nano/serializer"
)

type (
	// Options contains some configurations for connector
	Options struct {
		name                 string                // component name
		Serializer           serializer.Serializer // serializer for connector
		WebSocketPath        string                // websocket path
		IsWebSocket          bool                  // is websocket
		Logger               log.Logger            // logger
		Codec                codec.Codec           // codec
		Dictionary           message.Dictionary    // dictionary
		Branch               uint32
		ShortVersion         uint32
		WebSocketCompression bool
		CryptoType           uint32
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

// WithSerializerType customizes application serializer, which automatically Marshal
// and UnMarshal handler payload
func WithSerializer(serializer serializer.Serializer) Option {
	return func(opt *Options) {
		opt.Serializer = serializer
	}
}

// WithWebSocketPath set the websocket path
func WithWebSocketPath(path string) Option {
	return func(opt *Options) {
		opt.WebSocketPath = path
	}
}

func WithIsWebSocket(isWebSocket bool) Option {
	return func(opt *Options) {
		opt.IsWebSocket = isWebSocket
	}
}

// WithLogger overrides the default logger
func WithLogger(l log.Logger) Option {
	return func(opt *Options) {
		opt.Logger = l
	}
}

// WithCodec sets codec instead of default codec
func WithCodec(codec codec.Codec) Option {
	return func(opt *Options) {
		opt.Codec = codec
	}
}

func WithDictionary(dictionary message.Dictionary) Option {
	return func(opt *Options) {
		opt.Dictionary = dictionary
	}
}

func WithBranch(branch uint32) Option {
	return func(opt *Options) {
		opt.Branch = branch
	}
}

func WithShortVersion(shortVersion uint32) Option {
	return func(opt *Options) {
		opt.ShortVersion = shortVersion
	}
}

func WithWebSocketCompression(webSocketCompression bool) Option {
	return func(opt *Options) {
		opt.WebSocketCompression = webSocketCompression
	}
}
