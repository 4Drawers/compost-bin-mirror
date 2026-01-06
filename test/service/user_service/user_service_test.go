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

	// user data -> expected status code
	mockData := map[struct {
		username string
		password string
	}]int{
		{
			username: "abc12c_07",
			password: "adasfeiafnav393fanfe9affabgbh8fnbbc0",
		}: http.StatusOK,
		{
			username: "你好",
			// NOT a test case of sql injection, just want to test if api still works when the
			// input includes some sensitive words.
			password: "-- SELECT *",
		}: http.StatusOK,
		{
			username: "一二三四五六七八九十一二三四五六七八九十",
			password: "123.com",
		}: http.StatusOK,
		{
			username: "一二三四五六七八九十一二三四五六七八九十一",
			password: "123.com",
		}: http.StatusBadRequest,
	}

	apiVersion := 1
	for md, exp := range mockData {
		resp, err := http.PostForm(fmt.Sprintf("%s/v%d/user/register", baseUrl, apiVersion), url.Values{
			"username": []string{md.username},
			"password": []string{md.password},
		})
		if err != nil {
			t.Fatalf("Failed to reach /user/register: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != exp {
			t.Fatalf("Request /user/register with %v got status %d, expected %d", md, resp.StatusCode, exp)
		}

		contentType := resp.Header.Get("Content-Type")
		expectedContentType := "application/json"
		if contentType != expectedContentType {
			t.Fatalf("Request /user/register got Content-Type %s, expected %s", contentType, expectedContentType)
		}
	}
}

func TestLogin(t *testing.T) {
	once.Do(func() {
		if !wait4Server(t) {
			t.Fatalf("Server didn't became ready in time.")
		}
	})

	mockDataSuccess := map[struct {
		username string
		password string
	}]struct{}{
		{
			username: "abab",
			password: "123.com",
		}: {},
	}

	mockDataFail := map[struct {
		username string
		password string
	}]int{
		{
			username: "一二三四五六七八九十一二三四五六七八九十一",
			password: "123.com",
		}: http.StatusBadRequest,
		{
			username: "'or''='",
			password: "",
		}: http.StatusBadRequest,
		{
			username: "abab",
			password: "124.com",
		}: http.StatusBadRequest,
	}

	apiVersion := 1
	urlPrefix := fmt.Sprintf("%s/v%d/user", baseUrl, apiVersion)

	for md := range mockDataSuccess {
		regResp, err := http.PostForm(fmt.Sprintf("%s/%s", urlPrefix, "register"), url.Values{
			"username": []string{md.username},
			"password": []string{md.password},
		})
		if err != nil {
			t.Fatalf("Failed to send %v to /user/register: %v", md, err)
		}
		if regResp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected status code %d received while request /user/register with %v, expected 200",
				regResp.StatusCode, md)
		}
		defer regResp.Body.Close()

		loginResp, err := http.PostForm(fmt.Sprintf("%s/%s", urlPrefix, "login"), url.Values{
			"user_info": []string{md.username},
			"password":  []string{md.password},
		})
		if err != nil {
			t.Fatalf("Failed to send %v to /user/login: %v", md, err)
		}
		defer loginResp.Body.Close()

		_, _, result, err := unmarshalRespBody(loginResp.Body)
		if err != nil {
			t.Fatalf("Unexpected response from /user/login with %v: %v", md, err)
		}
		if result == nil {
			t.Fatalf("Got null user id from /user/login with %v", md)
		}
	}

	for md, exp := range mockDataFail {
		loginResp, err := http.PostForm(fmt.Sprintf("%s/%s", urlPrefix, "login"), url.Values{
			"user_info": []string{md.username},
			"password":  []string{md.password},
		})
		if err != nil {
			t.Fatalf("Failed to send %v to /user/login: %v", md, err)
		}
		defer loginResp.Body.Close()

		if loginResp.StatusCode != exp {
			t.Fatalf("Request /user/login with %v got status %d, expected %d", md, loginResp.StatusCode, exp)
		}
	}
}

func TestProfile(t *testing.T) {
	once.Do(func() {
		if !wait4Server(t) {
			t.Fatalf("Server didn't became ready in time.")
		}
	})

	apiVersion := 1
	urlPrefix := fmt.Sprintf("%s/v%d/user", baseUrl, apiVersion)

	mustExistUsername := "user-4-profile-test"
	mustExistPassword := "pswd-4-profile-test"

	_, err := http.PostForm(fmt.Sprintf("%s/%s", urlPrefix, "register"), url.Values{
		"username": []string{mustExistUsername},
		"password": []string{mustExistPassword},
	})
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	loginResp, err := http.PostForm(fmt.Sprintf("%s/%s", urlPrefix, "login"), url.Values{
		"user_info": []string{mustExistUsername},
		"password":  []string{mustExistPassword},
	})
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer loginResp.Body.Close()

	_, _, loginRes, err := unmarshalRespBody(loginResp.Body)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	var id int64
	if userId, ok := loginRes.(float64); !ok {
		t.Fatalf("Api /user/login didn't return user's id correctly, got data type: %T", loginRes)
	} else {
		id = int64(userId)
	}

	auth := loginResp.Header.Get("X-Authorization")
	ref := loginResp.Header.Get("X-Refresh")

	profReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/%d", urlPrefix, "profile", id), nil)
	if err != nil {
		t.Fatalf("Failed to build profile request: %v", err)
	}
	profReq.Header.Add("Authorization", fmt.Sprintf("%s %s", "Bearer", auth))
	profReq.Header.Add("Refresh", fmt.Sprintf("%s %s", "Bearer", ref))

	client := &http.Client{Timeout: 5 * time.Second}
	profResp, err := client.Do(profReq)
	if err != nil {
		t.Fatalf("Failed to get user %d's profile: %v", id, profResp)
	}
	defer profResp.Body.Close()

	_, _, profRes, err := unmarshalRespBody(profResp.Body)
	if err != nil {
		t.Fatalf("Failed to get profile: %v", err)
	}

	profile, ok := profRes.(map[string]any)
	if !ok {
		t.Fatalf("Illegal profile %v", profile)
	}

	if username, ok := profile["username"].(string); !ok || username != mustExistUsername {
		t.Fatalf("Unexpected username <%v, %T>, expected <%s, string>", username, username, mustExistUsername)
	}
	if userId, ok := profile["id"].(float64); !ok || int64(userId) != id {
		t.Fatalf("Unexpected user id <%v, %T>, expected <%d, float64>", userId, userId, id)
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

func unmarshalRespBody(body io.ReadCloser) (code int, msg string, result any, err error) {
	data, err := io.ReadAll(body)
	if err != nil && err != io.EOF {
		return 0, "", nil, fmt.Errorf("illegal response body: %v", err)
	}

	wrapped := make(map[string]any)
	if err = json.Unmarshal(data, &wrapped); err != nil {
		return 0, "", nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	cf, ok := wrapped["code"].(float64)
	if !ok {
		return 0, "", nil, fmt.Errorf("%v isn't a stardard response body, code not exist", wrapped)
	}
	code = int(cf)

	msg, ok = wrapped["msg"].(string)
	if !ok {
		return 0, "", nil, fmt.Errorf("%v isn't a stardard response body, msg not exist", wrapped)
	}

	result = wrapped["result"]

	return code, msg, result, nil
}
