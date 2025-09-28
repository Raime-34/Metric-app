package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"metricapp/internal/server/cfg"
)

func CalculateHash(b []byte) string {
	mac := hmac.New(sha256.New, []byte(cfg.Cfg.Key))
	mac.Write(b)
	hash := hex.EncodeToString(mac.Sum(nil))

	return hash
}
