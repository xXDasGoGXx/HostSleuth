package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

func actionTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	policy, err := core.NewActionPolicy(false, []string{"demo"})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	registerActionAPI(mux, &core.ActionManager{Policy: policy, Store: core.Store{Dir: t.TempDir()}}, "native")
	return mux
}

func TestActionAPIIsLoopbackOnly(t *testing.T) {
	mux := actionTestMux(t)
	req := httptest.NewRequest(http.MethodGet, "http://192.168.2.181/api/actions", nil)
	req.RemoteAddr = "192.168.2.50:54321"
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("remote status=%d body=%q", res.Code, res.Body.String())
	}
}

func TestActionRunRequiresExplicitCustomHeader(t *testing.T) {
	mux := actionTestMux(t)
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/api/actions/run", strings.NewReader(`{"action_id":"service.restart","target":"demo.service","confirmation":"RESTART demo.service"}`))
	req.RemoteAddr = "127.0.0.1:54321"
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusPreconditionRequired {
		t.Fatalf("status=%d body=%q", res.Code, res.Body.String())
	}
}

func TestActionAPIRequiresJSONAndRejectsUnknownFields(t *testing.T) {
	mux := actionTestMux(t)

	plain := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/api/actions/preview", strings.NewReader(`{"action_id":"service.restart","target":"demo.service"}`))
	plain.RemoteAddr = "127.0.0.1:12345"
	plain.Header.Set("Content-Type", "text/plain")
	plainRes := httptest.NewRecorder()
	mux.ServeHTTP(plainRes, plain)
	if plainRes.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("plain content-type status=%d", plainRes.Code)
	}

	body := bytes.NewBufferString(`{"action_id":"service.restart","target":"demo.service","command":"reboot"}`)
	unknown := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/api/actions/preview", body)
	unknown.RemoteAddr = "127.0.0.1:12345"
	unknown.Header.Set("Content-Type", "application/json")
	unknownRes := httptest.NewRecorder()
	mux.ServeHTTP(unknownRes, unknown)
	if unknownRes.Code != http.StatusBadRequest {
		t.Fatalf("unknown field status=%d body=%q", unknownRes.Code, unknownRes.Body.String())
	}
}

func TestActionCapabilitiesRemainDisabledWithoutOptIn(t *testing.T) {
	mux := actionTestMux(t)
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/actions", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"enabled":false`) || !strings.Contains(res.Body.String(), `"available":false`) {
		t.Fatalf("disabled boundary missing from response: %s", res.Body.String())
	}
}

func TestActionFlagSetKeepsRestartAndReloadAllowlistsSeparate(t *testing.T) {
	fs, _, enabled, restart, reload := actionFlagSet("test")
	if err := fs.Parse([]string{
		"--enable-actions",
		"--allow-restart-service", "restart-only",
		"--allow-reload-service", "reload-only",
	}); err != nil {
		t.Fatal(err)
	}
	if !*enabled {
		t.Fatal("actions were not enabled")
	}
	if len(*restart) != 1 || (*restart)[0] != "restart-only" {
		t.Fatalf("restart flags=%v", *restart)
	}
	if len(*reload) != 1 || (*reload)[0] != "reload-only" {
		t.Fatalf("reload flags=%v", *reload)
	}
}

func TestActionAPICapabilitiesIncludeIndependentReloadAction(t *testing.T) {
	policy, err := core.NewActionPolicyWithReload(false, []string{"restart-only"}, []string{"reload-only"})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	registerActionAPI(mux, &core.ActionManager{Policy: policy, Store: core.Store{Dir: t.TempDir()}}, "native")

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/actions", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", res.Code, res.Body.String())
	}
	body := res.Body.String()
	if !strings.Contains(body, `"action_id":"service.restart"`) || !strings.Contains(body, `"action_id":"service.reload"`) {
		t.Fatalf("both fixed actions were not exposed: %s", body)
	}
	if !strings.Contains(body, "restart-only.service") || !strings.Contains(body, "reload-only.service") {
		t.Fatalf("independent action targets missing: %s", body)
	}
}
