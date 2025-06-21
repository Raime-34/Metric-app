package agent

import (
	"math/rand"
	models "metricapp/internal/model"
	"metricapp/internal/repository"
	"runtime"
)

type MetricCollector struct {
	repo repository.Repo
}

func NewCollector() MetricCollector {
	return MetricCollector{
		repo: repository.NewInMemoryStorage(),
	}
}

func (mc *MetricCollector) Run() {
	collectTicker := time.NewTicker(2 * time.Second)
	sendTicker := time.NewTicker(10 * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

loop:
	for {
		select {
		case <-collectTicker.C:
			mc.collect()
		case <-sendTicker.C:

		case <-sigs:
			break loop
		}
	}
}

func (mc *MetricCollector) collect() {
	var mStat runtime.MemStats
	runtime.ReadMemStats(&mStat)

	mc.repo.SetField("Alloc", addField("Alloc", models.Gauge, mStat.Alloc))
	mc.repo.SetField("BuckHashSys", addField("BuckHashSys", models.Gauge, mStat.BuckHashSys))
	mc.repo.SetField("Frees", addField("Frees", models.Gauge, mStat.Frees))
	mc.repo.SetField("GCCPUFraction", addField("GCCPUFraction", models.Gauge, mStat.GCCPUFraction))
	mc.repo.SetField("HeapAlloc", addField("HeapAlloc", models.Gauge, mStat.HeapAlloc))
	mc.repo.SetField("HeapIdle", addField("HeapIdle", models.Gauge, mStat.HeapIdle))
	mc.repo.SetField("HeapInuse", addField("HeapInuse", models.Gauge, mStat.HeapInuse))
	mc.repo.SetField("HeapObjects", addField("HeapObjects", models.Gauge, mStat.HeapObjects))
	mc.repo.SetField("HeapReleased", addField("HeapReleased", models.Gauge, mStat.HeapReleased))
	mc.repo.SetField("HeapSys", addField("HeapSys", models.Gauge, mStat.HeapSys))
	mc.repo.SetField("LastGC", addField("LastGC", models.Gauge, mStat.LastGC))
	mc.repo.SetField("Lookups", addField("Lookups", models.Gauge, mStat.Lookups))
	mc.repo.SetField("MCacheInuse", addField("MCacheInuse", models.Gauge, mStat.MCacheInuse))
	mc.repo.SetField("MCacheSys", addField("MCacheSys", models.Gauge, mStat.MCacheSys))
	mc.repo.SetField("MSpanInuse", addField("MSpanInuse", models.Gauge, mStat.MSpanInuse))
	mc.repo.SetField("MSpanSys", addField("MSpanSys", models.Gauge, mStat.MSpanSys))
	mc.repo.SetField("Mallocs", addField("Mallocs", models.Gauge, mStat.Mallocs))
	mc.repo.SetField("NextGC", addField("NextGC", models.Gauge, mStat.NextGC))
	mc.repo.SetField("NumForcedGC", addField("NumForcedGC", models.Gauge, mStat.NumForcedGC))
	mc.repo.SetField("NumGC", addField("NumGC", models.Gauge, mStat.NumGC))
	mc.repo.SetField("OtherSys", addField("OtherSys", models.Gauge, mStat.OtherSys))
	mc.repo.SetField("PauseTotalNs", addField("PauseTotalNs", models.Gauge, mStat.PauseTotalNs))
	mc.repo.SetField("StackInuse", addField("StackInuse", models.Gauge, mStat.StackInuse))
	mc.repo.SetField("StackSys", addField("StackSys", models.Gauge, mStat.StackSys))
	mc.repo.SetField("Sys", addField("Sys", models.Gauge, mStat.Sys))
	mc.repo.SetField("TotalAlloc", addField("TotalAlloc", models.Gauge, mStat.TotalAlloc))
	mc.repo.SetField("RandomValue", addField("RandomValue", models.Gauge, rand.Int()))

	mc.repo.IncrementCounter()
}

func addField(id string, mType string, n any) models.Metrics {
	newMetric := models.Metrics{
		ID:    id,
		MType: mType,
	}

	switch n.(type) {
	case float64:
		newV := n.(float64)
		newMetric.Value = &newV
	case int64:
		newC := n.(int64)
		newMetric.Delta = &newC
	}

	return newMetric
}
