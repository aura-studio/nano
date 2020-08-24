package pipeline

import (
	"github.com/lonng/nano/cluster/clusterpb"
	"sync"

	"github.com/lonng/nano/message"
	"github.com/lonng/nano/session"
)

type (
	// Message is the alias of `message.Message`
	Message = message.Message

	Func func(s *session.Session, msg *message.Message) error

	Pipeline interface {
		Outbound() Channel
		Inbound() Channel
	}

	Info struct {
		Statistic *Statistic
	}

	StatisticItem struct {
		ReqCount int64//请求数
		TotalBytes int64//总字节数
		BytesPerReq float64//平均一个请求字节数
		TotalProcessTime int64//总处理时间
		AvgProcessTime float64//平均处理时间
	}

	Statistic struct {
		Summary *StatisticItem
		RouteStatistic map[string]*StatisticItem//按route分类的统计
		TypeStatistic map[byte]*StatisticItem//按类型分类的统计
	}

	pipeline struct {
		outbound, inbound *pipelineChannel
	}

	Channel interface {
		PushFront(h Func)
		PushBack(h Func)
		Process(s *session.Session, msg *message.Message) error
	}

	pipelineChannel struct {
		mu       sync.RWMutex
		handlers []Func
	}
)

var PipeInfo *Info

func init() {
	PipeInfo = &Info{
		Statistic: &Statistic{
			Summary:       &StatisticItem{},
			RouteStatistic: make(map[string]*StatisticItem),
			TypeStatistic:  make(map[byte]*StatisticItem),
		},
	}
}

func New(opts ...Option) Pipeline {
	p := &pipeline{
		outbound: &pipelineChannel{},
		inbound:  &pipelineChannel{},
	}

	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *pipeline) Outbound() Channel { return p.outbound }
func (p *pipeline) Inbound() Channel  { return p.inbound }

// PushFront push a function to the front of the pipeline
func (p *pipelineChannel) PushFront(h Func) {
	p.mu.Lock()
	defer p.mu.Unlock()
	handlers := make([]Func, len(p.handlers)+1)
	handlers[0] = h
	copy(handlers[1:], p.handlers)
	p.handlers = handlers
}

// PushFront push a function to the end of the pipeline
func (p *pipelineChannel) PushBack(h Func) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers = append(p.handlers, h)
}

// Process process message with all pipeline functions
func (p *pipelineChannel) Process(s *session.Session, msg *message.Message) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.handlers) < 1 {
		return nil
	}
	for _, h := range p.handlers {
		err := h(s, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (info *Info) ToProto() *clusterpb.QueryStatsResponse {
	result := &clusterpb.QueryStatsResponse{
		PipeInfo: &clusterpb.PipeInfo{
			Item:       &clusterpb.StatisticItem{
				ReqCount:         info.Statistic.Summary.ReqCount,
				TotalBytes:       info.Statistic.Summary.TotalBytes,
				BytesPerReq:      info.Statistic.Summary.BytesPerReq,
				TotalProcessTime: info.Statistic.Summary.TotalProcessTime,
				AvgProcessTime:   info.Statistic.Summary.AvgProcessTime,
			},
			RouteItems: make([]*clusterpb.RouteStatistic, 0),
			TypeItems:  make([]*clusterpb.TypeStatistic, 0),
		},
	}
	for route, item := range info.Statistic.RouteStatistic {
		result.PipeInfo.RouteItems = append(result.PipeInfo.RouteItems, &clusterpb.RouteStatistic{
			Route: route,
			Item:  &clusterpb.StatisticItem{
				ReqCount:         item.ReqCount,
				TotalBytes:       item.TotalBytes,
				BytesPerReq:      item.BytesPerReq,
				TotalProcessTime: item.TotalProcessTime,
				AvgProcessTime:   item.AvgProcessTime,
			},
		})
	}
	for tp, item := range info.Statistic.TypeStatistic {
		result.PipeInfo.TypeItems = append(result.PipeInfo.TypeItems, &clusterpb.TypeStatistic{
			Type: int64(tp),
			Item: &clusterpb.StatisticItem{
				ReqCount:         item.ReqCount,
				TotalBytes:       item.TotalBytes,
				BytesPerReq:      item.BytesPerReq,
				TotalProcessTime: item.TotalProcessTime,
				AvgProcessTime:   item.AvgProcessTime,
			},
		})
	}
	return result
}
