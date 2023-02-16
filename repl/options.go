package repl

import (
	"github.com/aura-studio/nano/codec"
	"github.com/aura-studio/nano/codec/plaincodec"
	"github.com/aura-studio/nano/component"
	"github.com/aura-studio/nano/env"
	"github.com/aura-studio/nano/message"
	"github.com/aura-studio/nano/serializer"
	"github.com/go-redis/redis"
)

// ErrorReader is the type for ErrorReader
type ErrorReader func(int64) string
type ReasonReader func(int64) string

// Options contains some configurations for current node
type Options struct {
	IsWebSocket          bool
	WebSocketPath        string
	WebSocketCompression bool
	PrettyJSON           bool
	SerializerType       serializer.SerializerType // serializer for connector
	ErrorReader          ErrorReader
	ReasonReader         ReasonReader
	Components           *component.Components
	RedisOptions         *redis.Options
	Codec                codec.Codec
	Dictionary           message.Dictionary
	Branch               uint32
	EnableUnexpected     bool
}

// Option defines a type for option, an option is a func operate cluster.Options
type Option func(*Options)

var (
	options = &Options{
		SerializerType: serializer.Protobuf,
		Codec:          plaincodec.NewCodec(),
	}
)

// WithIsWebSocket indicates whether current node WebSocket is enabled
func WithIsWebSocket(isWebSocket bool) Option {
	return func(o *Options) {
		o.IsWebSocket = isWebSocket
	}
}

// WithWebSocketPath sets root path for ws
func WithWebSocketPath(wsPath string) Option {
	return func(o *Options) {
		o.WebSocketPath = wsPath
	}
}

// WithPrettyJSON sets replied JSON pretty
func WithPrettyJSON(prettyJSON bool) Option {
	return func(o *Options) {
		o.PrettyJSON = prettyJSON
	}
}

// WithSerializerType customizes application serializer, which automatically Marshal
// and UnMarshal handler payload
func WithSerializerType(serializer serializer.SerializerType) Option {
	return func(o *Options) {
		o.SerializerType = serializer
	}
}

// WithErrorReader customize error reader, which can read error msg from code
func WithErrorReader(errorReader ErrorReader) Option {
	return func(o *Options) {
		o.ErrorReader = errorReader
	}
}

// WithReasonReader 显示错误码对应的文本的option
func WithReasonReader(reasonReader ReasonReader) Option {
	return func(o *Options) {
		o.ReasonReader = reasonReader
	}
}

// WithComponents sets the Components
func WithComponents(components *component.Components) Option {
	return func(o *Options) {
		o.Components = components
	}
}

// WithRedisClient sets the redis client options
func WithRedisClient(redisOptions *redis.Options) Option {
	return func(o *Options) {
		o.RedisOptions = redisOptions
	}
}

func WithVersion(version string) Option {
	return func(o *Options) {
		env.Version = version
		env.ShortVersion = message.ShortVersion(version)
	}
}

func WithCodec(codec codec.Codec) Option {
	return func(o *Options) {
		o.Codec = codec
	}
}

func WithDictionary(dictionary message.Dictionary) Option {
	return func(o *Options) {
		o.Dictionary = dictionary
	}
}

func WithBranch(branch uint32) Option {
	return func(o *Options) {
		o.Branch = branch
	}
}

func WithEnableUnexpected(enableUnexpected bool) Option {
	return func(o *Options) {
		o.EnableUnexpected = enableUnexpected
	}
}

func WithWebSocketCompression(webSocketCompression bool) Option {
	return func(o *Options) {
		o.WebSocketCompression = webSocketCompression
	}
}
