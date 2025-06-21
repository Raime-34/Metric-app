package agent

import (
	"fmt"
	"math/rand"
	"metricapp/internal/logger"
	models "metricapp/internal/model"
	"metricapp/internal/repository"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type MetricCollector struct {
	pollInterval   int
	reportInterval int
	reportHost     string
	repo           repository.Repo
}

func NewCollector(host string) MetricCollector {
	return MetricCollector{
		pollInterval:   2,
		reportInterval: 10,
		reportHost:     host,
		repo:           repository.NewInMemoryStorage(),
	}
}

func (mc *MetricCollector) Run() {
	collectTicker := time.NewTicker(time.Duration(mc.pollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(mc.reportInterval) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

loop:
	for {
		select {
		case <-collectTicker.C:
			mc.collect()
		case <-sendTicker.C:
			mc.sendMetrics()
		case <-sigs:
			break loop
		}
	}
}

func (mc *MetricCollector) collect() {
	var mStat runtime.MemStats
	runtime.ReadMemStats(&mStat)

	mc.repo.SetField("Alloc", composeValueMetric("Alloc", float64(mStat.Alloc)))
	mc.repo.SetField("BuckHashSys", composeValueMetric("BuckHashSys", float64(mStat.BuckHashSys)))
	mc.repo.SetField("Frees", composeValueMetric("Frees", float64(mStat.Frees)))
	mc.repo.SetField("GCCPUFraction", composeValueMetric("GCCPUFraction", mStat.GCCPUFraction))
	mc.repo.SetField("HeapAlloc", composeValueMetric("HeapAlloc", float64(mStat.HeapAlloc)))
	mc.repo.SetField("HeapIdle", composeValueMetric("HeapIdle", float64(mStat.HeapIdle)))
	mc.repo.SetField("HeapInuse", composeValueMetric("HeapInuse", float64(mStat.HeapInuse)))
	mc.repo.SetField("HeapObjects", composeValueMetric("HeapObjects", float64(mStat.HeapObjects)))
	mc.repo.SetField("HeapReleased", composeValueMetric("HeapReleased", float64(mStat.HeapReleased)))
	mc.repo.SetField("HeapSys", composeValueMetric("HeapSys", float64(mStat.HeapSys)))
	mc.repo.SetField("LastGC", composeValueMetric("LastGC", float64(mStat.LastGC)))
	mc.repo.SetField("Lookups", composeValueMetric("Lookups", float64(mStat.Lookups)))
	mc.repo.SetField("MCacheInuse", composeValueMetric("MCacheInuse", float64(mStat.MCacheInuse)))
	mc.repo.SetField("MCacheSys", composeValueMetric("MCacheSys", float64(mStat.MCacheSys)))
	mc.repo.SetField("MSpanInuse", composeValueMetric("MSpanInuse", float64(mStat.MSpanInuse)))
	mc.repo.SetField("MSpanSys", composeValueMetric("MSpanSys", float64(mStat.MSpanSys)))
	mc.repo.SetField("Mallocs", composeValueMetric("Mallocs", float64(mStat.Mallocs)))
	mc.repo.SetField("NextGC", composeValueMetric("NextGC", float64(mStat.NextGC)))
	mc.repo.SetField("NumForcedGC", composeValueMetric("NumForcedGC", float64(mStat.NumForcedGC)))
	mc.repo.SetField("NumGC", composeValueMetric("NumGC", float64(mStat.NumGC)))
	mc.repo.SetField("OtherSys", composeValueMetric("OtherSys", float64(mStat.OtherSys)))
	mc.repo.SetField("PauseTotalNs", composeValueMetric("PauseTotalNs", float64(mStat.PauseTotalNs)))
	mc.repo.SetField("StackInuse", composeValueMetric("StackInuse", float64(mStat.StackInuse)))
	mc.repo.SetField("StackSys", composeValueMetric("StackSys", float64(mStat.StackSys)))
	mc.repo.SetField("Sys", composeValueMetric("Sys", float64(mStat.Sys)))
	mc.repo.SetField("TotalAlloc", composeValueMetric("TotalAlloc", float64(mStat.TotalAlloc)))
	mc.repo.SetField("RandomValue", composeValueMetric("RandomValue", float64(rand.Int())))

	mc.repo.IncrementCounter()
}

func composeValueMetric(id string, v float64) models.Metrics {
	newMetric := models.Metrics{
		ID:    id,
		MType: models.Gauge,
		Value: &v,
	}

	return newMetric
}

func (mc *MetricCollector) sendMetrics() {
	logger.Info("Sending data to server...")
	metrics := mc.repo.GetFields()

	// Запрашивем изменение метрик типа gauge
	for _, metric := range metrics {
		if metric.MType == models.Counter {
			continue
		}

		url := fmt.Sprintf("%s/update/%s/%s/%v", mc.reportHost, metric.MType, metric.ID, *metric.Value)
		http.Post(url, "text/plain", nil)
	}

	// Отдельно отправляем счетчик
	pollCounter := metrics["PollCounter"]
	url := fmt.Sprintf("%s/update/%s/%s/%v", mc.reportHost, pollCounter.MType, pollCounter.ID, *pollCounter.Delta)
	http.Post(url, "text/plain", nil)
}
