package utils

import (
	"encoding/base64"
	"encoding/json"
)

// Returns base64 encoding of two strings
// for use with Basic Auth
func Base64Encode(s1, s2 string) string {
	auth := s1 + ":" + s2
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func Unmarshal[T any](payload []byte) (T, error) {
	var result T
	err := json.Unmarshal(payload, &result)
	return result, err
}

func Marshall(payload any) ([]byte, error) {
	result, err := json.Marshal(payload)

	return result, err
}
