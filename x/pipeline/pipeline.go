package pipeline

import (
	"github.com/lonng/nano/pipeline"
	"github.com/lonng/nano/x/stats"
)

// GatePipeline implements Pipeline
type GatePipeline struct {
	pipeline.Pipeline
}

// NewGatePipeline creates a new gate pipeline
func NewGatePipeline() *GatePipeline {
	p := &GatePipeline{
		Pipeline: pipeline.New(),
	}
	p.Inbound().PushBack(stats.MessageStatsInboundProcessor)
	p.Outbound().PushBack(stats.MessageStatsOutboundProcessor)
	return p
}
