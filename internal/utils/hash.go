package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func CalculateHash(b []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(b)
	hash := hex.EncodeToString(mac.Sum(nil))

	return hash
}
