package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCertificateRolloutAPIRejectsInvalidInputs(t *testing.T) {
	mux := http.NewServeMux()
	registerCertificateRolloutAPI(mux)

	tests := []string{
		"/api/certificate-rollout",
		"/api/certificate-rollout?fingerprint=abcd&endpoint=example.test%3A443",
		"/api/certificate-rollout?fingerprint=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"/api/certificate-rollout?reference=example.test%3A443&fingerprint=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&endpoint=example.test%3A443",
		"/api/certificate-rollout?fingerprint=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&endpoint=missing-port",
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
