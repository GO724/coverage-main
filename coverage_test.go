package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// тут писать код тестов

type TestSearchServerAuthCase struct {
	token  string
	status int
}

func TestSearchServerAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	testCases := make([]TestSearchServerAuthCase, 3, 3)
	testCases[0] = TestSearchServerAuthCase{"valid_token", http.StatusOK}
	testCases[1] = TestSearchServerAuthCase{"invalid_token", http.StatusUnauthorized}
	testCases[2] = TestSearchServerAuthCase{"", http.StatusUnauthorized}

	for i, testCase := range testCases {
		// Test request
		req, err := http.NewRequest("GET", ts.URL+"?query=John&limit=5", nil)
		if err != nil {
			t.Errorf("TestSearchServerAuth failed: http.NewRequest():%v", err)
		}
		req.Header.Add("AccessToken", testCase.token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("TestSearchServerAuth failed: http.DefaultClient.Do():%v", err)
		}
		// Check answer
		if resp.StatusCode != testCase.status {
			t.Errorf("TestSearchServerAuth [%d] failed: token(%s) expected %d, got %d", i, testCase.token, testCase.status, resp.StatusCode)
		}
	}
}
