package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
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

// Checks a string to see if it contains a valid date. Returns error if not valid
func CheckDate(t string) error {

	_, err := time.Parse(time.DateOnly, t)
	if err != nil {
		return fmt.Errorf("the date must in the format of YYYY-MM-DD")
	}

	return nil

}

// Regex expression for a valid tenant Id
// Doing it here ensures it is compiled once to improve performance
var uuidTenantIdRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ValidateTenantID checks if the tenant ID is valid and returns an error if not
func ValidateTenantID(tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID must not be empty")
	}
	if !uuidTenantIdRegex.MatchString(tenantID) {
		return fmt.Errorf("tenant ID must be a valid UUID format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)")
	}
	return nil
}

// Regex for instance Id
var uuidInstanceIdRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}$`)

// ValidateInstanceID checks if the instance ID is valid and returns an error if not
func ValidateInstanceID(InstanceID string) error {
	if InstanceID == "" {
		return fmt.Errorf("instance ID must not be empty")
	}
	if !uuidInstanceIdRegex.MatchString(InstanceID) {
		return fmt.Errorf("instance ID must be a valid UUID format (xxxxxxxx)")
	}
	return nil
}
