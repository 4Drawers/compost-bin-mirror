package user_service_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"
)

const (
	baseUrl = "http://localhost:17890"
	timeout = 10 * time.Second
)

var once sync.Once

func TestRegister(t *testing.T) {
	once.Do(func() {
		if !wait4Server(t) {
			t.Fatalf("Server didn't became ready in time.")
		}
	})

	apiVersion := 1
	resp, err := http.PostForm(fmt.Sprintf("%s/v%d/user/register", baseUrl, apiVersion), url.Values{
		"username": []string{"user1"},
		"password": []string{"passwd1"},
	})
	if err != nil {
		t.Fatalf("Failed to request /user/register: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Request /user/register got status %d, expected 200", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Fatalf("Request /user/register got Content-Type %s, expected %s", contentType, expectedContentType)
	}
}

func TestLogin(t *testing.T) {
	once.Do(func() {
		if !wait4Server(t) {
			t.Fatalf("Server didn't became ready in time.")
		}
	})

	apiVersion := 1
	urlPrefix := fmt.Sprintf("%s/v%d/user", baseUrl, apiVersion)
	regResp, err := http.PostForm(fmt.Sprintf("%s/%s", urlPrefix, "register"), url.Values{
		"username": []string{"Login Tester"},
		"password": []string{"Login Tester Password"},
	})
	if err != nil {
		t.Fatalf("Failed to request /user/register: %v", err)
	}
	defer regResp.Body.Close()

	loginResp, err := http.PostForm(fmt.Sprintf("%s/%s", urlPrefix, "login"), url.Values{
		"user_info": []string{"Login Tester"},
		"password":  []string{"Login Tester Password"},
	})
	if err != nil {
		t.Fatalf("Failed to request /user/login: %v", err)
	}
	defer loginResp.Body.Close()

	var data []byte
	if data, err = io.ReadAll(loginResp.Body); err != nil && err != io.EOF {
		t.Fatalf("Failed to request /user/login (illegal response body): %v", err)
	}
	result := map[string]any{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	userId, ok := result["result"].(float64)
	if !ok {
		t.Fatalf("Got illegal response body from /user/login request: %v", result)
	}
	jwtToken := loginResp.Header.Get("X-Authorization")
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/%d", urlPrefix, "info", int64(userId)), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to request /user/info: %v", err)
	}
	defer resp.Body.Close()

	if data, err = io.ReadAll(resp.Body); err != nil && err != io.EOF {
		t.Fatalf("Failed to request /user/info (illegal response body): %v", err)
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body %s: %v", string(data), err)
	}

	res, ok := result["result"].(map[string]any)
	if !ok {
		t.Fatalf("Got illegal response body from /user/info request: %v", result)
	}
	username := res["username"].(string)
	if username != "Login Tester" {
		t.Fatalf("Unexpected username from /user/info %s, expected: %s", username, "Login Tester")
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
