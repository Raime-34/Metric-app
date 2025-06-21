package main

import (
	"metricapp/internal/agent"
	"metricapp/internal/logger"
)

func main() {
	logger.InitLogger()
	collector := agent.NewCollector()

	logger.Info("Starting metrics collection")
	collector.Run()
}
