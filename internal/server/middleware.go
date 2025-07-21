package server

import (
	"compress/gzip"
	"metricapp/internal/logger"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

func gzipHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Распаковка тела запроса, если оно в gzip
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		// Обработка gzip-ответа, если клиент поддерживает
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			grw := gzipResponseWriter{ResponseWriter: w, Writer: gz}
			next.ServeHTTP(grw, r)
			return
		}

		// Если клиент не поддерживает gzip
		next.ServeHTTP(w, r)
	})
}

func requestLogger(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		method := r.Method

		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)
		duration := time.Since(start)

		logger.Info(
			"Request log",
			zap.String("URI", uri),
			zap.String("Method", method),
			zap.Duration("Duration", duration),
			zap.Int("Status", responseData.status),
			zap.Int("Response size", responseData.size),
			zap.String("Resp", responseData.msg),
		)
	}

	return http.HandlerFunc(logFn)
}
