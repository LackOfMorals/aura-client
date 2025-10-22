package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
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
func Marshall(payload any) ([]byte, error) {
	result, err := json.Marshal(payload)

	return result, err
}

// Takes a payload and returns indented JSON
func MarshallIndent(payload any) ([]byte, error) {

	result, err := json.MarshalIndent(payload, "", "  ")

	return result, err

}

func CheckDate(t string) error {

	_, err := time.Parse(time.DateOnly, t)
	if err != nil {
		return fmt.Errorf("date must in the format of YYYY-MM-DD")
	}

	return nil

}
