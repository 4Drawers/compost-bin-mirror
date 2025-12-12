package test

import (
	"net/http"
	"testing"
	"time"
)

const (
	baseUrl = "http://localhost:17890"
	timeout = 10 * time.Second
)

func TestMain(t *testing.T) {
	if !wait4Server(t) {
		t.Fatalf("Server didn't became ready in time.")
	}

	resp, err := http.Get(baseUrl + "/")
	if err != nil {
		t.Fatalf("Failed to request path /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Request / got status %d, expected 200", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	expectedContentType := "text/plain; charset=UTF-8"
	if contentType != expectedContentType {
		t.Fatalf("Request / got Content-Type %s, expected %s", contentType, expectedContentType)
	}
}

func wait4Server(t *testing.T) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseUrl + "/")
		if err == nil {
			resp.Body.Close()
			return true
		}
		<-ticker.C
	}

	return false
}
