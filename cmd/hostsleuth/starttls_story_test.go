package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStartTLSStoryAPIRejectsMissingOrInvalidInputs(t *testing.T) {
	mux := http.NewServeMux()
	registerStartTLSStoryAPI(mux)

	tests := []string{
		"/api/starttls-story",
		"/api/starttls-story?protocol=smtp",
		"/api/starttls-story?target=localhost%3A25",
		"/api/starttls-story?protocol=ftp&target=localhost%3A21",
		"/api/starttls-story?protocol=smtp&target=missing-port",
	}
	for _, path := range tests {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", path, res.Code)
		}
	}
}
