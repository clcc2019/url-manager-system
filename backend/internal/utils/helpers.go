package utils
package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int, charset string) (string, error) {
	b := make([]byte, length)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b), nil
}

// StringPtr 返回字符串指针
func StringPtr(s string) *string {
	return &s
}

// IntPtr 返回整数指针
func IntPtr(i int) *int {
	return &i
}

// BoolPtr 返回布尔指针
func BoolPtr(b bool) *bool {
	return &b
}