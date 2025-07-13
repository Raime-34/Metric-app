package main

import (
	"flag"
	"metricapp/internal/agent"
	"metricapp/internal/logger"
)

func main() {
	logger.InitLogger()
	collector := agent.NewCollector()
	flag.Parse()

	logger.Info("Starting metrics collection")
	collector.Run()
}
