package agent

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"metricapp/internal/logger"
	models "metricapp/internal/model"
	"metricapp/internal/repository"
	"metricapp/internal/zip"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

type MetricCollector struct {
	pollInterval   int
	reportInterval int
	reportHost     string
	repo           Repo[models.Metrics]
}

type Repo[T any] interface {
	SetField(string, T)
	GetFields() map[string]T
	IncrementCounter(...struct {
		Name  string
		Delta int64
	})
}

func NewCollector() *MetricCollector {
	newCollector := MetricCollector{
		repo: repository.NewAgentMemoryStorage(),
	}

	var cfg struct {
		Address        string `env:"ADDRESS"`
		ReportInterval int    `env:"REPORT_INTERVAL"`
		PollInterval   int    `env:"POLL_INTERVAL"`
	}

	err := env.Parse(&cfg)
	if err == nil {
		newCollector.reportHost = cfg.Address
		newCollector.reportInterval = cfg.ReportInterval
		newCollector.pollInterval = cfg.PollInterval
	}

	if newCollector.reportHost == "" {
		flag.StringVar(&newCollector.reportHost, "a", "localhost:8080", "URL адрес сервера сбора метрик")
	}
	if newCollector.pollInterval == 0 {
		flag.IntVar(&newCollector.pollInterval, "p", 2, "Промежуток времени сбора метрик")
	}
	if newCollector.reportInterval == 0 {
		flag.IntVar(&newCollector.reportInterval, "r", 10, "Промежуток времени отправки данных на сервер")
	}

	return &newCollector
}

func (mc *MetricCollector) Run() {
	collectTicker := time.NewTicker(time.Duration(mc.pollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(mc.reportInterval) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// "Нинада goto. Почему не continue?"
	//
	// Тут не особо понял, так что решил просто коммент оставить
	//
	// Использовал goto для выхода из внешнего лупа
	// Просто брейк же сработает на select
	// а continue это не то, что мне нужно было
	// так как я обрабатываю сис кол на завершение работы программы
loop:
	for {
		select {
		case <-collectTicker.C:
			mc.collect()
		case <-sendTicker.C:
			mc.sendMetricsAsBatch()
		case <-sigs:
			break loop
		}
	}
}

func (mc *MetricCollector) collect() {
	var mStat runtime.MemStats
	runtime.ReadMemStats(&mStat)

	mc.repo.SetField("Alloc", models.ComposeMetrics("Alloc", models.Gauge, float64(mStat.Alloc), 0))
	mc.repo.SetField("BuckHashSys", models.ComposeMetrics("BuckHashSys", models.Gauge, float64(mStat.BuckHashSys), 0))
	mc.repo.SetField("Frees", models.ComposeMetrics("Frees", models.Gauge, float64(mStat.Frees), 0))
	mc.repo.SetField("GCCPUFraction", models.ComposeMetrics("GCCPUFraction", models.Gauge, mStat.GCCPUFraction, 0))
	mc.repo.SetField("HeapAlloc", models.ComposeMetrics("HeapAlloc", models.Gauge, float64(mStat.HeapAlloc), 0))
	mc.repo.SetField("HeapIdle", models.ComposeMetrics("HeapIdle", models.Gauge, float64(mStat.HeapIdle), 0))
	mc.repo.SetField("HeapInuse", models.ComposeMetrics("HeapInuse", models.Gauge, float64(mStat.HeapInuse), 0))
	mc.repo.SetField("HeapObjects", models.ComposeMetrics("HeapObjects", models.Gauge, float64(mStat.HeapObjects), 0))
	mc.repo.SetField("HeapReleased", models.ComposeMetrics("HeapReleased", models.Gauge, float64(mStat.HeapReleased), 0))
	mc.repo.SetField("HeapSys", models.ComposeMetrics("HeapSys", models.Gauge, float64(mStat.HeapSys), 0))
	mc.repo.SetField("LastGC", models.ComposeMetrics("LastGC", models.Gauge, float64(mStat.LastGC), 0))
	mc.repo.SetField("Lookups", models.ComposeMetrics("Lookups", models.Gauge, float64(mStat.Lookups), 0))
	mc.repo.SetField("MCacheInuse", models.ComposeMetrics("MCacheInuse", models.Gauge, float64(mStat.MCacheInuse), 0))
	mc.repo.SetField("MCacheSys", models.ComposeMetrics("MCacheSys", models.Gauge, float64(mStat.MCacheSys), 0))
	mc.repo.SetField("MSpanInuse", models.ComposeMetrics("MSpanInuse", models.Gauge, float64(mStat.MSpanInuse), 0))
	mc.repo.SetField("MSpanSys", models.ComposeMetrics("MSpanSys", models.Gauge, float64(mStat.MSpanSys), 0))
	mc.repo.SetField("Mallocs", models.ComposeMetrics("Mallocs", models.Gauge, float64(mStat.Mallocs), 0))
	mc.repo.SetField("NextGC", models.ComposeMetrics("NextGC", models.Gauge, float64(mStat.NextGC), 0))
	mc.repo.SetField("NumForcedGC", models.ComposeMetrics("NumForcedGC", models.Gauge, float64(mStat.NumForcedGC), 0))
	mc.repo.SetField("NumGC", models.ComposeMetrics("NumGC", models.Gauge, float64(mStat.NumGC), 0))
	mc.repo.SetField("OtherSys", models.ComposeMetrics("OtherSys", models.Gauge, float64(mStat.OtherSys), 0))
	mc.repo.SetField("PauseTotalNs", models.ComposeMetrics("PauseTotalNs", models.Gauge, float64(mStat.PauseTotalNs), 0))
	mc.repo.SetField("StackInuse", models.ComposeMetrics("StackInuse", models.Gauge, float64(mStat.StackInuse), 0))
	mc.repo.SetField("StackSys", models.ComposeMetrics("StackSys", models.Gauge, float64(mStat.StackSys), 0))
	mc.repo.SetField("Sys", models.ComposeMetrics("Sys", models.Gauge, float64(mStat.Sys), 0))
	mc.repo.SetField("TotalAlloc", models.ComposeMetrics("TotalAlloc", models.Gauge, float64(mStat.TotalAlloc), 0))
	mc.repo.SetField("RandomValue", models.ComposeMetrics("RandomValue", models.Gauge, float64(rand.Int()), 0))
	mc.repo.SetField("GCSys", models.ComposeMetrics("GCSys", models.Gauge, float64(mStat.GCSys), 0))

	mc.repo.IncrementCounter()
}

func (mc *MetricCollector) sendMetrics() {
	logger.Info("Sending data to server...")
	metrics := mc.repo.GetFields()

	// Запрашивем изменение метрик типа gauge
	for _, metric := range metrics {
		if metric.MType == models.Counter {
			continue
		}

		err := deliverMetric(metric, mc.reportHost)
		if err != nil {
			logger.Error(
				"failed to send metric",
				zap.String("ID", metric.ID),
				zap.Error(err),
			)
		}
	}

	// Отдельно отправляем счетчик
	pCount := metrics["PollCounter"]
	pCount.ID = "PollCount"
	err := deliverMetric(pCount, mc.reportHost)
	logger.Error(
		"failed to send metric",
		zap.String("ID", pCount.ID),
		zap.Error(err),
	)
}

func (mc *MetricCollector) sendMetricsAsBatch() {
	var req []models.Metrics

	metrics := mc.repo.GetFields()

	for _, m := range metrics {
		req = append(req, m)
	}

	pCount := metrics["PollCounter"]
	pCount.ID = "PollCount"
	req = append(req, pCount)

	err := deliverMetrics(req, mc.reportHost)
	if err != nil {
		logger.Error("failed to send batch", zap.Error(err))
	}
}

func deliverMetric(metric models.Metrics, reportHost string) error {
	b, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	b, err = zip.GzipCompress(b)
	if err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	r := bytes.NewReader(b)

	url := fmt.Sprintf("http://%s/update/", reportHost)
	req, err := http.NewRequest(http.MethodPost, url, r)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Error(
				"failed to close response body",
				zap.Error(err),
			)
		}
	}()

	return nil
}

func deliverMetrics(metrics []models.Metrics, reportHost string) error {
	if len(metrics) == 0 {
		return nil
	}

	b, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	b, err = zip.GzipCompress(b)
	if err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	r := bytes.NewReader(b)

	url := fmt.Sprintf("http://%s/updates/", reportHost)
	req, err := http.NewRequest(http.MethodPost, url, r)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Error("failed to close response body", zap.Error(cerr))
		}
	}()

	return nil
}
