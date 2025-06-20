package agent

type Metrics struct {
	gauges  map[string]float64
	counter int64
}

func newMetrics() Metrics {
	return Metrics{
		gauges: make(map[string]float64),
	}
}
