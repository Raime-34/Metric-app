package main

import (
	"metricapp/internal/server"
)

func main() {
	server := server.MetricServer{}
	server.Start()
}
