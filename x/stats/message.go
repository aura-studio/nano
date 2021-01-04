package stats

import (
	"sync"

	"github.com/lonng/nano/message"
	"github.com/lonng/nano/session"
	"github.com/lonng/nano/x/virtualtime"
)

type (
	// MessageStatsItem stores category info for MessageStatsInfo
	MessageStatsItem struct {
		Count            int64   //消息数
		TotalBytes       int64   //总字节数
		AvgBytes         float64 //平均一个请求字节数
		TotalProcessTime int64   //总处理时间
		AvgProcessTime   float64 //平均处理时间
	}

	// MessageStatsInfo stores packet info
	MessageStatsInfo struct {
		StartTime  int64
		EndTime    int64
		Summary    *MessageStatsItem
		RouteStats map[string]*MessageStatsItem //按route分类的统计
		TypeStats  map[byte]*MessageStatsItem   //按类型分类的统计
	}

	// MessageStats used to swap MessageStatsInfo
	MessageStats struct {
		m    sync.Mutex
		info *MessageStatsInfo
	}
)

var (
	messageStats *MessageStats
)

func init() {
	messageStats = NewMessageStats()
}

// NewMessageStats creates a new message stats
func NewMessageStats() *MessageStats {
	return &MessageStats{
		info: &MessageStatsInfo{
			StartTime:  virtualtime.Now().UnixNano(),
			Summary:    &MessageStatsItem{},
			RouteStats: make(map[string]*MessageStatsItem),
			TypeStats:  make(map[byte]*MessageStatsItem),
		},
	}
}

// Info swaps a new info with old messageStats info
func (ms *MessageStats) Info() *MessageStatsInfo {
	messageStats.m.Lock()
	defer messageStats.m.Unlock()

	messageStats.info.EndTime = virtualtime.Now().UnixNano()
	info := messageStats.info
	messageStats.info = ms.info

	return info
}

// MessageStatsInboundProcessor is the handler for inbound
func MessageStatsInboundProcessor(s *session.Session, msg *message.Message) error {
	dataLength := int64(len(msg.Data))

	messageStats.m.Lock()
	defer messageStats.m.Unlock()

	messageStats.info.Summary.Count++
	messageStats.info.Summary.TotalBytes += dataLength

	route := msg.Route
	_, ok := messageStats.info.RouteStats[route]
	if !ok {
		messageStats.info.RouteStats[route] = &MessageStatsItem{}
	}
	messageStats.info.RouteStats[route].Count++
	messageStats.info.RouteStats[route].TotalBytes += dataLength

	typ := byte(msg.Type)
	_, ok = messageStats.info.TypeStats[typ]
	if !ok {
		messageStats.info.TypeStats[typ] = &MessageStatsItem{}
	}
	messageStats.info.TypeStats[typ].Count++
	messageStats.info.TypeStats[typ].TotalBytes += dataLength
	return nil
}

//  MessageStatsOutboundProcessor processor it he handler for
func MessageStatsOutboundProcessor(s *session.Session, msg *message.Message) error {
	dataLength := int64(len(msg.Data))

	messageStats.m.Lock()
	defer messageStats.m.Unlock()

	messageStats.info.Summary.Count++
	messageStats.info.Summary.TotalBytes += dataLength

	route := msg.Route
	_, ok := messageStats.info.RouteStats[route]
	if !ok {
		messageStats.info.RouteStats[route] = &MessageStatsItem{}
	}
	messageStats.info.RouteStats[route].Count++
	messageStats.info.RouteStats[route].TotalBytes += dataLength

	typ := byte(msg.Type)
	_, ok = messageStats.info.TypeStats[typ]
	if !ok {
		messageStats.info.TypeStats[typ] = &MessageStatsItem{}
	}
	messageStats.info.TypeStats[typ].Count++
	messageStats.info.TypeStats[typ].TotalBytes += dataLength
	return nil
}
