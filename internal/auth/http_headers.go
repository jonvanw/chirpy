package auth

import (
	"net/http"
	"errors"
	"regexp"
)

func extractAuthValue(headers http.Header, key string, notFoundError string) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	pattern := `\b` + regexp.QuoteMeta(key) + `\s+(\S+)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(authHeader)
	if len(matches) < 2 {
		return "", errors.New(notFoundError)
	}
	return matches[1], nil
}

func GetBearerToken(headers http.Header) (string, error) {
	return extractAuthValue(headers, "Bearer", "bearer token not found")
}

func GetAPIKey(headers http.Header) (string, error) {
	return extractAuthValue(headers, "ApiKey", "API key not found")
}