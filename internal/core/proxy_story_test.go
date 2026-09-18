package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNormalizeProxyStoryURLRejectsUnsafeInput(t *testing.T) {
	tests := []string{
		"example.com",
		"ftp://example.com/file",
		"https://user:pass@example.com/",
		"http://example.com:70000/",
	}
	for _, raw := range tests {
		if _, err := normalizeProxyStoryURL(raw); err == nil {
			t.Fatalf("expected validation failure for %q", raw)
		}
	}
}

func TestProxyStoryOutcomeProvenFailureWinsOverEarlierUnknown(t *testing.T) {
	stages := []ProxyStoryStage{
		{ID: "route", Title: "Route", Status: "unknown"},
		{ID: "tcp", Title: "TCP", Status: "fail", Summary: "connection refused"},
	}
	status, first, _ := proxyStoryOutcome(stages)
	if status != "fail" || first != "tcp" {
		t.Fatalf("unexpected outcome: status=%s first=%s", status, first)
	}
}

func TestInspectProxyPublicHTTPFollowsSameHostRedirect(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, server.URL+"/final", http.StatusFound)
			return
		}
		w.Header().Set("Server", "fixture")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	parsed, err := normalizeProxyStoryURL(server.URL + "/start")
	if err != nil {
		t.Fatal(err)
	}
	probe := inspectProxyPublicHTTP(context.Background(), parsed)
	if probe.Status != "pass" || probe.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected probe: %#v", probe)
	}
	if probe.CrossHostRedirectStopped {
		t.Fatalf("same-host redirect should not stop: %#v", probe)
	}
	if len(probe.Redirects) < 2 {
		t.Fatalf("expected redirect chain, got %#v", probe.Redirects)
	}
}

func TestInspectProxyPublicHTTPStopsCrossHostRedirect(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/other", http.StatusFound)
	}))
	defer source.Close()

	sourceURL := strings.Replace(source.URL, "127.0.0.1", "localhost", 1)
	parsed, err := normalizeProxyStoryURL(sourceURL + "/start")
	if err != nil {
		t.Fatal(err)
	}
	probe := inspectProxyPublicHTTP(context.Background(), parsed)
	if probe.Status != "pass" || probe.StatusCode != http.StatusFound || !probe.CrossHostRedirectStopped {
		t.Fatalf("expected bounded cross-host redirect stop, got %#v", probe)
	}
}

func TestInspectProxyUpstreamHTTPUsesPublicHostHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "public.example" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	parsed, err := normalizeProxyStoryURL(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	native := inspectProxyUpstreamHTTP(context.Background(), parsed, "", "")
	forwarded := inspectProxyUpstreamHTTP(context.Background(), parsed, "public.example", "public.example")
	if native.StatusCode != http.StatusNotFound {
		t.Fatalf("expected native Host to return 404, got %#v", native)
	}
	if forwarded.StatusCode != http.StatusNoContent || forwarded.HostHeader != "public.example" {
		t.Fatalf("expected public Host to return 204, got %#v", forwarded)
	}
}

func TestProxyUpstreamHTTPStageWarnsWhenHostBehaviorChanges(t *testing.T) {
	native := &ProxyHTTPProbe{Status: "pass", StatusCode: 404, HTTPStatus: "404 Not Found", HostHeader: "127.0.0.1"}
	forwarded := &ProxyHTTPProbe{Status: "pass", StatusCode: 204, HTTPStatus: "204 No Content", HostHeader: "public.example"}
	stage := proxyUpstreamHTTPStage(native, forwarded)
	if stage.Status != "warn" || !strings.Contains(stage.Summary, "Host/SNI") {
		t.Fatalf("expected Host/SNI warning, got %#v", stage)
	}
}

func TestBuildProxyStoryUsesForwardedPublicHostVariant(t *testing.T) {
	oldDiagnose := proxyStoryDiagnose
	oldPublic := proxyPublicHTTPProbe
	oldUpstream := proxyUpstreamHTTPProbe
	t.Cleanup(func() {
		proxyStoryDiagnose = oldDiagnose
		proxyPublicHTTPProbe = oldPublic
		proxyUpstreamHTTPProbe = oldUpstream
	})

	proxyStoryDiagnose = func(_ context.Context, target string, _ Snapshot) Diagnosis {
		return Diagnosis{
			Target: target,
			Checks: []Check{
				{Name: "dns", Status: "pass", Evidence: "127.0.0.1"},
				{Name: "route", Status: "pass", Evidence: "local"},
				{Name: "tcp", Status: "pass", Evidence: "accepted"},
			},
		}
	}
	proxyPublicHTTPProbe = func(_ context.Context, parsed *url.URL) *ProxyHTTPProbe {
		return &ProxyHTTPProbe{URL: parsed.String(), Status: "pass", StatusCode: 200, HTTPStatus: "200 OK"}
	}
	var forwardedHost string
	proxyUpstreamHTTPProbe = func(_ context.Context, _ *url.URL, host, sni string) *ProxyHTTPProbe {
		if host != "" {
			forwardedHost = host
			return &ProxyHTTPProbe{Status: "pass", StatusCode: 200, HTTPStatus: "200 OK", HostHeader: host, TLSServerName: sni}
		}
		return &ProxyHTTPProbe{Status: "pass", StatusCode: 200, HTTPStatus: "200 OK", HostHeader: "127.0.0.1"}
	}

	story, err := BuildProxyStory(context.Background(), "http://public.example/", "http://127.0.0.1:8080/", Snapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if forwardedHost != "public.example" || story.UpstreamPublicHostHTTP == nil {
		t.Fatalf("expected public Host variant, got %#v", story.UpstreamPublicHostHTTP)
	}
	if story.Status != "pass" {
		t.Fatalf("expected passing story, got %#v", story)
	}
}

func TestProxyHTTPStatusClasses(t *testing.T) {
	cases := map[int]string{204: "pass", 302: "pass", 401: "warn", 404: "warn", 502: "fail", 503: "fail"}
	for code, want := range cases {
		got, _ := proxyHTTPStatusClass(code)
		if got != want {
			t.Fatalf("HTTP %d: got %s want %s", code, got, want)
		}
	}
}


func TestProxyPublicHostHeaderIPv6(t *testing.T) {
	parsed, err := normalizeProxyStoryURL("https://[2001:db8::10]/")
	if err != nil {
		t.Fatal(err)
	}
	if got := proxyPublicHostHeader(parsed); got != "[2001:db8::10]" {
		t.Fatalf("unexpected IPv6 Host header: %q", got)
	}
}

func TestSanitizeProxyLocationRedactsQuery(t *testing.T) {
	base, _ := normalizeProxyStoryURL("https://example.com/start")
	got := sanitizeProxyLocation("/login?token=secret", base)
	if got != "https://example.com/login?[redacted]" {
		t.Fatalf("unexpected sanitized Location: %q", got)
	}
}

func TestProxyHTTPMethodNotAllowedExplainsHEADBoundary(t *testing.T) {
	status, summary := proxyHTTPStatusClass(http.StatusMethodNotAllowed)
	if status != "warn" || !strings.Contains(summary, "HEAD") {
		t.Fatalf("unexpected 405 classification: %s %s", status, summary)
	}
}
