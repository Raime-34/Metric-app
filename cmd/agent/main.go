package main

import "metricapp/internal/agent"

func main() {
	collector := agent.NewCollector()
	collector.Run()
}
