package main

import (
	"net/http/httptest"
	"testing"
)

func TestDNSDetectiveAPIRejectsMissingName(t *testing.T) {
	mux := httpTestMuxDNSDetective()
	req := httptest.NewRequest("GET", "/api/dns-detective", nil)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != 400 {
		t.Fatalf("expected 400, got %d", res.Code)
	}
}

func httpTestMuxDNSDetective() *http.ServeMux {
	mux := http.NewServeMux()
	registerDNSDetectiveAPI(mux)
	return mux
}
