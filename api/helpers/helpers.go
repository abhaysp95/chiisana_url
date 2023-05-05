package helpers

import (
	"os"
	"strings"
)

func ResolveDomainError(url string) bool {
	if url == os.Getenv("DOMAIN") {
		return false
	}

	// somethings which may or may not give domain error
	fixedURL := strings.Replace(url, "http://", "", 1)
	fixedURL = strings.Replace(url, "https://", "", 1)
	fixedURL = strings.Replace(url, "www.", "", 1)
	fixedURL = strings.Split(url, "/")[0]

	if fixedURL == os.Getenv("DOMAIN") {
		return false
	}

	return true
}

func EnforceHTTP(url string) string {
	if url[:4] != "http" {
		return "http" + url
	}

	return url
}
