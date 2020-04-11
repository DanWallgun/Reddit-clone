package utils

import (
	"crypto/rand"
	"encoding/base64"
	prand "math/rand"
)

var IsRandom bool = true
var seed int64 = 0

func GenerateRandomString(length int) string {
	bytes := make([]byte, length)
	if !IsRandom {
		prand.NewSource(seed)
		prand.Read(bytes)
	} else {
		rand.Read(bytes)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
