package nano

import (
	"time"

	"github.com/lonng/nano/cluster"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/env"
	"github.com/lonng/nano/log"
	"github.com/lonng/nano/message"
	"github.com/lonng/nano/persistence"
	"github.com/lonng/nano/pipeline"
	"github.com/lonng/nano/serialize"
	"github.com/lonng/nano/upgrader"
	"google.golang.org/grpc"
)

// Option defines a type for option, an option is a func operate cluster.Options
type Option func(*cluster.Options)

// WithPipeline sets the pipeline option.
func WithPipeline(pipeline pipeline.Pipeline) Option {
	return func(opt *cluster.Options) {
		opt.Pipeline = pipeline
	}
}

// WithConvention sets the convention between
func WithConvention(convention cluster.Convention) Option {
	return func(opt *cluster.Options) {
		opt.Convention = convention
	}
}

// WithAdvertiseAddr sets the advertise address option, it will be the listen address in
// master node and an advertise address which cluster member to connect
func WithAdvertiseAddr(addr string, retryInterval ...time.Duration) Option {
	return func(opt *cluster.Options) {
		opt.AdvertiseAddr = addr
		if len(retryInterval) > 0 {
			opt.RetryInterval = retryInterval[0]
		}
	}
}

// WithClientAddr sets the listen address which is used to establish connection between
// cluster members. Will select an available port automatically if no member address
// setting and panic if no available port
func WithClientAddr(addr string) Option {
	return func(opt *cluster.Options) {
		opt.ClientAddr = addr
	}
}

// WithMaster sets the option to indicate whether the current node is master node
func WithMaster() Option {
	return func(opt *cluster.Options) {
		opt.IsMaster = true
	}
}

// WithGrpcOptions sets the grpc dial options
func WithGrpcOptions(opts ...grpc.DialOption) Option {
	return func(_ *cluster.Options) {
		env.GrpcOptions = append(env.GrpcOptions, opts...)
	}
}

// WithComponents sets the Components
func WithComponents(components *component.Components) Option {
	return func(opt *cluster.Options) {
		opt.Components = components
	}
}

// WithDebugMode makes 'nano' run under Debug mode.
func WithDebugMode(debug bool) Option {
	return func(_ *cluster.Options) {
		env.Debug = debug
	}
}

// WithTimerPrecision sets the ticker precision, and time precision can not less
// than a Millisecond, and can not change after application running. The default
// precision is time.Second
func WithTimerPrecision(precision time.Duration) Option {
	if precision < time.Millisecond {
		panic("time precision can not less than a Millisecond")
	}
	return func(_ *cluster.Options) {
		env.TimerPrecision = precision
	}
}

// WithSerializer customizes application serializer, which automatically Marshal
// and UnMarshal handler payload
func WithSerializer(serializer serialize.Serializer) Option {
	return func(opt *cluster.Options) {
		env.Serializer = serializer
	}
}

// WithLabel sets the current node label in cluster
func WithLabel(label string) Option {
	return func(opt *cluster.Options) {
		opt.Label = label
	}
}

// WithVersion sets the current node version in cluster
func WithVersion(version string) Option {
	return func(opt *cluster.Options) {
		env.Version = version
		env.ShortVersion = message.ShortVersion(version)
	}
}

// WithHttpUpgrader sets the http upgrader for socket
func WithHttpUpgrader(upgrader upgrader.Upgrader) Option {
	return func(opt *cluster.Options) {
		opt.HttpUpgrader = upgrader
	}
}

// WithHttpAddr sets the independent http address
func WithHttpAddr(httpAddr string) Option {
	return func(opt *cluster.Options) {
		opt.HttpAddr = httpAddr
	}
}

// WithMasterPersist sets the persistence of cluster
func WithMasterPersist(persistence persistence.Persistence) Option {
	return func(opt *cluster.Options) {
		opt.MasterPersist = persistence
	}
}

// WithTSLConfig sets the `key` and `certificate` of TSL
func WithTSLConfig(certificate, key string) Option {
	return func(opt *cluster.Options) {
		opt.TSLCertificate = certificate
		opt.TSLKey = key
	}
}

// WithLogger overrides the default logger
func WithLogger(l log.Logger) Option {
	return func(opt *cluster.Options) {
		opt.Logger = l
	}
}
