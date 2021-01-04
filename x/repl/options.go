package repl

import (
	"strings"

	"github.com/go-redis/redis"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/env"
	"github.com/lonng/nano/message"
	"github.com/lonng/nano/serialize"
	"github.com/lonng/nano/serialize/json"
	"github.com/lonng/nano/serialize/protobuf"
)

// ErrorReader is the type for ErrorReader
type ErrorReader func(int64) string
type ReasonReader func(int64) string

// Options contains some configurations for current node
type Options struct {
	IsWebSocket    bool
	WSPath         string
	PrettyJSON     bool
	Serializer     serialize.Serializer // serializer for connector
	SerializerName string
	ErrorReader    ErrorReader
	ReasonReader   ReasonReader
	Components     *component.Components
	RedisOptions   *redis.Options
}

// Option defines a type for option, an option is a func operate cluster.Options
type Option func(*Options)

var (
	opt = &Options{
		IsWebSocket:    false,
		WSPath:         "",
		PrettyJSON:     false,
		Serializer:     protobuf.NewSerializer(),
		SerializerName: "",
		ErrorReader:    nil,
		ReasonReader:   nil,
		Components:     nil,
	}
)

// WithIsWebsocket indicates whether current node WebSocket is enabled
func WithIsWebsocket(isWebSocket bool) Option {
	return func(opt *Options) {
		opt.IsWebSocket = isWebSocket
	}
}

// WithWSPath sets root path for ws
func WithWSPath(wsPath string) Option {
	return func(opt *Options) {
		opt.WSPath = wsPath
	}
}

// WithPrettyJSON sets replied JSON pretty
func WithPrettyJSON(prettyJSON bool) Option {
	return func(opt *Options) {
		opt.PrettyJSON = prettyJSON
	}
}

// WithSerializer customizes application serializer, which automatically Marshal
// and UnMarshal handler payload
func WithSerializer(serializer string) Option {
	return func(opt *Options) {
		if strings.ToUpper(serializer) == "JSON" {
			opt.Serializer = json.NewSerializer()
			opt.SerializerName = "JSON"
		} else {
			opt.Serializer = protobuf.NewSerializer()
			opt.SerializerName = "Protobuf"
		}
	}
}

// WithErrorReader customize error reader, which can read error msg from code
func WithErrorReader(errorReader ErrorReader) Option {
	return func(opt *Options) {
		opt.ErrorReader = errorReader
	}
}

// WithReasonReader 显示错误码对应的文本的option
func WithReasonReader(reasonReader ReasonReader) Option {
	return func(opt *Options) {
		opt.ReasonReader = reasonReader
	}
}

// WithComponents sets the Components
func WithComponents(components *component.Components) Option {
	return func(opt *Options) {
		opt.Components = components
	}
}

// WithRedisClient sets the redis client options
func WithRedisClient(redisOptions *redis.Options) Option {
	return func(opt *Options) {
		opt.RedisOptions = redisOptions
	}
}

func WithVersion(version string) Option {
	return func(opt *Options) {
		env.Version = version
		env.ShortVersion = message.ShortVersion(version)
	}
}
