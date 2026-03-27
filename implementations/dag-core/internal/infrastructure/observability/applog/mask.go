package applog

import "strings"

var sensitiveKeys = []string{
	"token",
	"secret",
	"password",
	"authorization",
	"api_key",
	"apikey",
}

func MaskSensitive(input string) string {
	normalized := strings.ToLower(input)
	for _, key := range sensitiveKeys {
		if strings.Contains(normalized, key) {
			return "[redacted]"
		}
	}
	return input
}
