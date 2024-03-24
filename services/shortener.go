package services

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	counter   uint64
	counterMu sync.Mutex
)

func base64ToBase62(base64Encoded string) string {
	base62Chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	base64Chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/"

	var base62Encoded string
	for _, char := range base64Encoded {
		index := strings.IndexRune(base64Chars, char)
		if index != -1 {
			base62Encoded += string(base62Chars[index])
		}
	}

	return base62Encoded[:10]
}

func GenerateShortURL(longURL string) string {
	urlStr :=  fmt.Sprintf(
		"%s-%d-%d", longURL, time.Now().UnixNano(), getUniqueCounter(),
	)
	hash := sha256.Sum256([]byte(urlStr))

	base64EncodedString := base64.URLEncoding.EncodeToString(hash[:])

	shortURL := base64ToBase62(base64EncodedString)

	return shortURL
}

func getUniqueCounter() uint64 {
	counterMu.Lock()
	defer counterMu.Unlock()

	counter++
	return counter
}
