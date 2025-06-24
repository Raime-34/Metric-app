package main

import (
	"flag"
	"metricapp/internal/agent"
	"metricapp/internal/logger"
)

func main() {
	flag.Parse()
	logger.InitLogger()

	logger.Info("Starting metrics collection")
	agent.Run()
}
