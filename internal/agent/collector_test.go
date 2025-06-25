package agent

import (
	"flag"
	"metricapp/internal/logger"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestMetricCollector_Run(t *testing.T) {
	flag.Parse()
	done := make(chan bool)
	var n atomic.Uint64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		if n.Load() == 28 {
			done <- true
		}
	}))
	flag.Set("a", strings.TrimPrefix(server.URL, "http://"))

	logger.InitLogger()
	go Run()

	deadline := time.NewTimer(20 * time.Second)

	select {
	case <-done:
		return
	case <-deadline.C:
		t.Fatal("Timeout to get response")
	}
}
