package main

import (
	"flag"
	"metricapp/internal/agent"
	"metricapp/internal/logger"
)

func main() {
	flag.Parse()
	logger.InitLogger()
	collector := agent.NewCollector()

	logger.Info("Starting metrics collection")
	collector.Run()
}
