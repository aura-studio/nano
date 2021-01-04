package hook

// Handler processes a byte slice to a fixed byte slice
type Handler func([]byte) []byte

// Processor is to fix line after the log generated.
type Processor struct {
	Handler Handler
}

// ProcessBytes processes a byte slice to a fixed byte slice
func (p *Processor) Process(b []byte) []byte {
	return p.Handler(b)
}
