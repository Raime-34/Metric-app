package agent

import (
	"math/rand"
	"runtime"
)

type MetricCollector struct{}

func (mc *MetricCollector) Collect() Metrics {
	var mStat runtime.MemStats
	runtime.ReadMemStats(&mStat)

	metrics := newMetrics()
	metrics.gauges["Alloc"] = float64(mStat.Alloc)
	metrics.gauges["BuckHashSys"] = float64(mStat.BuckHashSys)
	metrics.gauges["Frees"] = float64(mStat.Frees)
	metrics.gauges["GCCPUFraction"] = float64(mStat.GCCPUFraction)
	metrics.gauges["HeapAlloc"] = float64(mStat.HeapAlloc)
	metrics.gauges["HeapIdle"] = float64(mStat.HeapIdle)
	metrics.gauges["HeapInuse"] = float64(mStat.HeapInuse)
	metrics.gauges["HeapObjects"] = float64(mStat.HeapObjects)
	metrics.gauges["HeapReleased"] = float64(mStat.HeapReleased)
	metrics.gauges["HeapSys"] = float64(mStat.HeapSys)
	metrics.gauges["LastGC"] = float64(mStat.LastGC)
	metrics.gauges["Lookups"] = float64(mStat.Lookups)
	metrics.gauges["MCacheInuse"] = float64(mStat.MCacheInuse)
	metrics.gauges["MCacheSys"] = float64(mStat.MCacheSys)
	metrics.gauges["MSpanInuse"] = float64(mStat.MSpanInuse)
	metrics.gauges["MSpanSys"] = float64(mStat.MSpanSys)
	metrics.gauges["Mallocs"] = float64(mStat.Mallocs)
	metrics.gauges["NextGC"] = float64(mStat.NextGC)
	metrics.gauges["NumForcedGC"] = float64(mStat.NumForcedGC)
	metrics.gauges["NumGC"] = float64(mStat.NumGC)
	metrics.gauges["OtherSys"] = float64(mStat.OtherSys)
	metrics.gauges["PauseTotalNs"] = float64(mStat.PauseTotalNs)
	metrics.gauges["StackInuse"] = float64(mStat.StackInuse)
	metrics.gauges["StackSys"] = float64(mStat.StackSys)
	metrics.gauges["Sys"] = float64(mStat.Sys)
	metrics.gauges["TotalAlloc"] = float64(mStat.TotalAlloc)
	metrics.gauges["RandomValue"] = float64(rand.Int())

	return metrics
}
