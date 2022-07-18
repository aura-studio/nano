package connector

import (
	"github.com/aura-studio/nano/codec"
	"github.com/aura-studio/nano/log"
)

type (
	// Options contains some configurations for connector
	Options struct {
		name        string      // component name
		wsPath      string      //websocket path
		isWebSocket bool        // is websocket
		logger      log.Logger  // logger
		codec       codec.Codec // codec
		branch      uint32
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

// WithWSPath set the websocket path
func WithWSPath(path string) Option {
	return func(opt *Options) {
		opt.wsPath = path
	}
}

func WithIsWebSocket(isWebSocket bool) Option {
	return func(opt *Options) {
		opt.isWebSocket = isWebSocket
	}
}

// WithLogger overrides the default logger
func WithLogger(l log.Logger) Option {
	return func(opt *Options) {
		opt.logger = l
	}
}

// WithCodec sets codec instead of default codec
func WithCodec(codec codec.Codec) Option {
	return func(opt *Options) {
		opt.codec = codec
	}
}

func WithBranch(branch uint32) Option {
	return func(opt *Options) {
		opt.branch = branch
	}
}
