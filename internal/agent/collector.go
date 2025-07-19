package agent

import (
	"flag"
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

	"github.com/caarlos0/env/v11"
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
	if newCollector.reportInterval == 0 {
		flag.IntVar(&newCollector.pollInterval, "p", 2, "Промежуток времени сбора метрик")
	}
	if newCollector.pollInterval == 0 {
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
			mc.sendMetrics()
		case <-sigs:
			break loop
		}
	}
}

func (mc *MetricCollector) collect() {
	var mStat runtime.MemStats
	runtime.ReadMemStats(&mStat)

	mc.repo.SetField("Alloc", models.ComposeMetrics("Alloc", float64(mStat.Alloc)))
	mc.repo.SetField("BuckHashSys", models.ComposeMetrics("BuckHashSys", float64(mStat.BuckHashSys)))
	mc.repo.SetField("Frees", models.ComposeMetrics("Frees", float64(mStat.Frees)))
	mc.repo.SetField("GCCPUFraction", models.ComposeMetrics("GCCPUFraction", mStat.GCCPUFraction))
	mc.repo.SetField("HeapAlloc", models.ComposeMetrics("HeapAlloc", float64(mStat.HeapAlloc)))
	mc.repo.SetField("HeapIdle", models.ComposeMetrics("HeapIdle", float64(mStat.HeapIdle)))
	mc.repo.SetField("HeapInuse", models.ComposeMetrics("HeapInuse", float64(mStat.HeapInuse)))
	mc.repo.SetField("HeapObjects", models.ComposeMetrics("HeapObjects", float64(mStat.HeapObjects)))
	mc.repo.SetField("HeapReleased", models.ComposeMetrics("HeapReleased", float64(mStat.HeapReleased)))
	mc.repo.SetField("HeapSys", models.ComposeMetrics("HeapSys", float64(mStat.HeapSys)))
	mc.repo.SetField("LastGC", models.ComposeMetrics("LastGC", float64(mStat.LastGC)))
	mc.repo.SetField("Lookups", models.ComposeMetrics("Lookups", float64(mStat.Lookups)))
	mc.repo.SetField("MCacheInuse", models.ComposeMetrics("MCacheInuse", float64(mStat.MCacheInuse)))
	mc.repo.SetField("MCacheSys", models.ComposeMetrics("MCacheSys", float64(mStat.MCacheSys)))
	mc.repo.SetField("MSpanInuse", models.ComposeMetrics("MSpanInuse", float64(mStat.MSpanInuse)))
	mc.repo.SetField("MSpanSys", models.ComposeMetrics("MSpanSys", float64(mStat.MSpanSys)))
	mc.repo.SetField("Mallocs", models.ComposeMetrics("Mallocs", float64(mStat.Mallocs)))
	mc.repo.SetField("NextGC", models.ComposeMetrics("NextGC", float64(mStat.NextGC)))
	mc.repo.SetField("NumForcedGC", models.ComposeMetrics("NumForcedGC", float64(mStat.NumForcedGC)))
	mc.repo.SetField("NumGC", models.ComposeMetrics("NumGC", float64(mStat.NumGC)))
	mc.repo.SetField("OtherSys", models.ComposeMetrics("OtherSys", float64(mStat.OtherSys)))
	mc.repo.SetField("PauseTotalNs", models.ComposeMetrics("PauseTotalNs", float64(mStat.PauseTotalNs)))
	mc.repo.SetField("StackInuse", models.ComposeMetrics("StackInuse", float64(mStat.StackInuse)))
	mc.repo.SetField("StackSys", models.ComposeMetrics("StackSys", float64(mStat.StackSys)))
	mc.repo.SetField("Sys", models.ComposeMetrics("Sys", float64(mStat.Sys)))
	mc.repo.SetField("TotalAlloc", models.ComposeMetrics("TotalAlloc", float64(mStat.TotalAlloc)))
	mc.repo.SetField("RandomValue", models.ComposeMetrics("RandomValue", float64(rand.Int())))

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

		url := fmt.Sprintf("http://%s/update/%s/%s/%v", mc.reportHost, metric.MType, metric.ID, *metric.Value)
		resp, err := http.Post(url, "text/plain", nil)
		if err == nil {
			defer resp.Body.Close()
		}
	}

	// Отдельно отправляем счетчик
	pollCounter := metrics["PollCounter"]
	url := fmt.Sprintf("http://%s/update/%s/%s/%v", mc.reportHost, pollCounter.MType, pollCounter.ID, *pollCounter.Delta)
	resp, err := http.Post(url, "text/plain", nil)
	if err == nil {
		defer resp.Body.Close()
	}
}
