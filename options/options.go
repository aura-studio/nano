package options

import (
	"time"

	"github.com/aura-studio/nano/log"

	"github.com/aura-studio/nano/codec"
	"github.com/aura-studio/nano/component"
	"github.com/aura-studio/nano/persist"
	"github.com/aura-studio/nano/pipeline"
)

// Options contains some configurations for current node
type Options struct {
	Pipeline       pipeline.Pipeline
	MasterPersist  persist.Persist
	IsMaster       bool
	AdvertiseAddr  string
	RetryInterval  time.Duration
	TCPAddr        string
	DebugAddr      string
	MemberAddr     string
	Components     *component.Components
	Label          string
	HttpAddr       string
	TSLCertificate string
	TSLKey         string
	Logger         log.Logger
	Codec          codec.Codec
	Etcd           bool
}

var Default = &Options{}
