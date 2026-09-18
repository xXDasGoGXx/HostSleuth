package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

func TestProxyStoryAPIRejectsMissingInputs(t *testing.T) {
	mux := http.NewServeMux()
	registerProxyStoryAPI(mux, core.Store{Dir: t.TempDir()})

	for _, path := range []string{
		"/api/proxy-story",
		"/api/proxy-story?public=https%3A%2F%2Fexample.com%2F",
		"/api/proxy-story?upstream=http%3A%2F%2F127.0.0.1%3A8080%2F",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", path, res.Code)
		}
	}
}
