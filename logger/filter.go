package logger

import "strings"

var sensitiveKeywords = []string{
	"password",
	"salt",
	"token", "access", "refresh",
	"secret", "key", "private",
	"email", "mail",
	"phone", "mobile", "tel",
	"username",
}

func containsSensitiveWord(s string) bool {
	lower := strings.ToLower(s)
	for _, key := range sensitiveKeywords {
		if strings.Contains(lower, key) {
			return true
		}
	}
	return false
}
