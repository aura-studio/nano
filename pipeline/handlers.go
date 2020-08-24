package pipeline

import (
	"github.com/lonng/nano/message"
	"github.com/lonng/nano/session"
)


func HandlerStatistic(s *session.Session, msg *message.Message) error {
	dataLen := len(msg.Data)
	dataLen += 1 + 8 + 2//1 byte for compressed & message type, 8 bytes for message id, 2 bytes for route code/route length, others for payload
	if !msg.Compressed() {
		dataLen += len(msg.Route)//if not compressed len(msg.Route) bytes for msg.Route
	}

	PipeInfo.Statistic.Summary.ReqCount++
	PipeInfo.Statistic.Summary.TotalBytes += int64(dataLen)

	route := msg.Route
	_, ok := PipeInfo.Statistic.RouteStatistic[route]
	if !ok {
		PipeInfo.Statistic.RouteStatistic[route] = &StatisticItem{}
	}
	PipeInfo.Statistic.RouteStatistic[route].ReqCount++
	PipeInfo.Statistic.RouteStatistic[route].TotalBytes += int64(dataLen)

	tp := byte(msg.Type)
	_, ok = PipeInfo.Statistic.TypeStatistic[tp]
	if !ok {
		PipeInfo.Statistic.TypeStatistic[tp] = &StatisticItem{}
	}
	PipeInfo.Statistic.TypeStatistic[tp].ReqCount++
	PipeInfo.Statistic.TypeStatistic[tp].TotalBytes += int64(dataLen)

	return nil
}
