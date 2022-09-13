package monitor

var (
	TotalRequests = NewCounterWithLabel(
		"request_total",
		"Total number of processed requests",
		[]string{"cluster", "method", "status_code"},
	)

	ProcessingRequests = NewGaugeWithLabel(
		"request_processing",
		"Current number of processing requests",
		[]string{"cluster", "method"},
	)

	RequestLatency = NewHistogramWithLabel(
		"request_latency_ms",
		"Histogram of latency for requests",
		[]float64{200.0, 400.0, 600.0, 800.0, 1000.0, 1500.0, 2000.0,
			2500.0, 3000.0, 5000.0, 10000.0, 20000.0, 30000.0, 45000.0, 60000.0},
		[]string{"cluster", "method"},
	)

	RealTimeRequestLatency = NewGaugeWithLabel(
		"realtime_request_latency_ms",
		"Histogram of max latency for requests",
		[]string{"cluster", "method"},
	)

	RealTimeRequestBodySize = NewGaugeWithLabel(
		"realtime_request_body_size",
		"Max request body size of every request",
		[]string{"cluster", "method"},
	)

	cluster = "defaultCluster"
)

func InitMonitor(c string) {
	cluster = c
}

func GetCluster() string {
	return cluster
}
