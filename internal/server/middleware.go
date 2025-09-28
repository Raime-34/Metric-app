package server

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"metricapp/internal/logger"
	"metricapp/internal/server/cfg"
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
			defer func() {
				err := gz.Close()
				if err != nil {
					logger.Error(
						"failed to close gzip reader",
						zap.Error(err),
					)
				}
			}()
			r.Body = gz
		}

		// Обработка gzip-ответа, если клиент поддерживает
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer func() {
				// Убрал спам логов
				gz.Close()
			}()
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

func hashChecker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.Cfg.Key == "" {
			next.ServeHTTP(w, r)
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(b))
		hash := r.Header.Get("HashSHA256")
		if hash == "" {
			next.ServeHTTP(w, r)
		} else {
			mac := hmac.New(sha256.New, []byte(cfg.Cfg.Key))
			mac.Write(b)
			calculatedHash := hex.EncodeToString(mac.Sum(nil))

			if hash != calculatedHash {
				http.Error(w, fmt.Sprintf("hash does not matched: %s and %s", hash, calculatedHash), http.StatusBadRequest)
				logger.Error(
					"hashes does not matched",
					zap.String("from request", hash),
					zap.String("calculated", calculatedHash),
				)
				return
			}

			next.ServeHTTP(w, r)
		}
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func setHash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		mac := hmac.New(sha256.New, []byte(cfg.Cfg.Key))
		mac.Write(rec.body.Bytes())
		sig := hex.EncodeToString(mac.Sum(nil))

		w.Header().Set("HashSHA256", sig)
		w.WriteHeader(rec.status)
		w.Write(rec.body.Bytes())
	})
}
