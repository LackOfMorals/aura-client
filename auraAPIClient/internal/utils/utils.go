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

// Takes a JSON payload and copies it into a struct T
func Unmarshal[T any](payload []byte) (T, error) {
	var result T
	err := json.Unmarshal(payload, &result)
	return result, err
}

// Takes a payload and returns JSON
func Marshal(payload any) ([]byte, error) {
	result, err := json.Marshal(payload)

	return result, err
}

// Takes a payload and returns indented JSON
func MarshalIndent(payload any) ([]byte, error) {
	result, err := json.MarshalIndent(payload, "", "  ")
	return result, err
}
