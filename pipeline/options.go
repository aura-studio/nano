package pipeline

type Option func(*pipeline)

func WithStatistic() Option {
	return func(p *pipeline) {
		p.inbound.PushBack(HandlerStatistic)
	}
}
