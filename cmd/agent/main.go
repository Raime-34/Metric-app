package main

import (
	"metricapp/internal/agent"
	"metricapp/internal/logger"
)

func main() {
	logger.InitLogger()

	logger.Info("Starting metrics collection")
	agent.Run()
}
