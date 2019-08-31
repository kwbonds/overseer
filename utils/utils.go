// Package utils contains utility functions
package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// GetMD5Hash generates a simple MD5 hash from a string
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
