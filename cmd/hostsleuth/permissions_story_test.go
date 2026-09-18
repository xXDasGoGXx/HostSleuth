package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

func TestPermissionStoryAPIRejectsMissingInputs(t *testing.T) {
	mux := http.NewServeMux()
	registerPermissionStoryAPI(mux, core.Store{Dir: t.TempDir()})

	for _, path := range []string{
		"/api/permissions-story",
		"/api/permissions-story?service=nginx.service",
		"/api/permissions-story?path=%2Fsrv%2Fapp",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", path, res.Code)
		}
	}
}
