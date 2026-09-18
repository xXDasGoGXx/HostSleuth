package core

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	proxyStoryTimeout       = 12 * time.Second
	proxyHTTPRedirectLimit  = 5
	proxyHTTPHeaderMaxBytes = 1024
)

type ProxyStoryStage struct {
	ID       string   `json:"id"`
	Group    string   `json:"group"`
	Title    string   `json:"title"`
	Status   string   `json:"status"`
	Summary  string   `json:"summary"`
	Evidence []string `json:"evidence,omitempty"`
}

type ProxyHTTPHop struct {
	URL         string `json:"url"`
	StatusCode  int    `json:"status_code"`
	Status      string `json:"status"`
	Location    string `json:"location,omitempty"`
	Server      string `json:"server,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type ProxyHTTPProbe struct {
	URL                      string         `json:"url"`
	DialTarget               string         `json:"dial_target"`
	HostHeader               string         `json:"host_header"`
	TLSServerName            string         `json:"tls_server_name,omitempty"`
	Status                   string         `json:"status"`
	DurationMS               int64          `json:"duration_ms"`
	StatusCode               int            `json:"status_code,omitempty"`
	HTTPStatus               string         `json:"http_status,omitempty"`
	Location                 string         `json:"location,omitempty"`
	Server                   string         `json:"server,omitempty"`
	ContentType              string         `json:"content_type,omitempty"`
	Redirects                []ProxyHTTPHop `json:"redirects,omitempty"`
	CrossHostRedirectStopped bool           `json:"cross_host_redirect_stopped,omitempty"`
	RedirectLimitReached     bool           `json:"redirect_limit_reached,omitempty"`
	Error                    string         `json:"error,omitempty"`
}

type ProxyStory struct {
	PublicURL              string            `json:"public_url"`
	UpstreamURL            string            `json:"upstream_url"`
	PublicTarget           string            `json:"public_target"`
	UpstreamTarget         string            `json:"upstream_target"`
	StartedAt              time.Time         `json:"started_at"`
	Status                 string            `json:"status"`
	Conclusion             string            `json:"conclusion"`
	FirstProblem           string            `json:"first_problem,omitempty"`
	PublicDiagnosis        Diagnosis         `json:"public_diagnosis"`
	UpstreamDiagnosis      Diagnosis         `json:"upstream_diagnosis"`
	PublicHTTP             *ProxyHTTPProbe   `json:"public_http,omitempty"`
	UpstreamNativeHTTP     *ProxyHTTPProbe   `json:"upstream_native_http,omitempty"`
	UpstreamPublicHostHTTP *ProxyHTTPProbe   `json:"upstream_public_host_http,omitempty"`
	Stages                 []ProxyStoryStage `json:"stages"`
}

var proxyStoryDiagnose = Diagnose
var proxyPublicHTTPProbe = inspectProxyPublicHTTP
var proxyUpstreamHTTPProbe = inspectProxyUpstreamHTTP

func BuildProxyStory(ctx context.Context, publicURL, upstreamURL string, snap Snapshot) (*ProxyStory, error) {
	publicParsed, err := normalizeProxyStoryURL(publicURL)
	if err != nil {
		return nil, fmt.Errorf("public URL: %w", err)
	}
	upstreamParsed, err := normalizeProxyStoryURL(upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("upstream URL: %w", err)
	}

	publicTarget, err := proxyURLTarget(publicParsed)
	if err != nil {
		return nil, fmt.Errorf("public URL: %w", err)
	}
	upstreamTarget, err := proxyURLTarget(upstreamParsed)
	if err != nil {
		return nil, fmt.Errorf("upstream URL: %w", err)
	}

	probeCtx, cancel := context.WithTimeout(ctx, proxyStoryTimeout)
	defer cancel()

	story := &ProxyStory{
		PublicURL:      displayProxyStoryURL(publicParsed),
		UpstreamURL:    displayProxyStoryURL(upstreamParsed),
		PublicTarget:   publicTarget,
		UpstreamTarget: upstreamTarget,
		StartedAt:      time.Now().UTC(),
	}

	story.PublicDiagnosis = proxyStoryDiagnose(probeCtx, publicTarget, snap)
	story.UpstreamDiagnosis = proxyStoryDiagnose(probeCtx, upstreamTarget, snap)

	story.PublicHTTP = proxyPublicHTTPProbe(probeCtx, publicParsed)
	story.UpstreamNativeHTTP = proxyUpstreamHTTPProbe(probeCtx, upstreamParsed, "", "")

	if !sameDNSName(publicParsed.Hostname(), upstreamParsed.Hostname()) {
		story.UpstreamPublicHostHTTP = proxyUpstreamHTTPProbe(
			probeCtx,
			upstreamParsed,
			proxyPublicHostHeader(publicParsed),
			publicParsed.Hostname(),
		)
	}

	story.Stages = buildProxyStoryStages(story, publicParsed, upstreamParsed)
	story.Status, story.FirstProblem, story.Conclusion = proxyStoryOutcome(story.Stages)
	return story, nil
}

func normalizeProxyStoryURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("a complete http:// or https:// URL is required")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("only http:// and https:// URLs are supported")
	}
	if parsed.User != nil {
		return nil, errors.New("URLs containing credentials are not accepted")
	}
	if strings.TrimSpace(parsed.Hostname()) == "" {
		return nil, errors.New("URL hostname is required")
	}
	if port := parsed.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return nil, errors.New("URL port must be between 1 and 65535")
		}
	}
	parsed.Fragment = ""
	return parsed, nil
}

func proxyURLTarget(parsed *url.URL) (string, error) {
	if parsed == nil || parsed.Hostname() == "" {
		return "", errors.New("URL hostname is required")
	}
	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	if _, err := strconv.Atoi(port); err != nil {
		return "", errors.New("URL port must be numeric")
	}
	return net.JoinHostPort(parsed.Hostname(), port), nil
}

func displayProxyStoryURL(parsed *url.URL) string {
	if parsed == nil {
		return ""
	}
	copyURL := *parsed
	hadQuery := copyURL.RawQuery != ""
	copyURL.RawQuery = ""
	copyURL.ForceQuery = false
	copyURL.Fragment = ""
	value := copyURL.String()
	if hadQuery {
		value += "?[redacted]"
	}
	return value
}

func proxyPublicHostHeader(parsed *url.URL) string {
	if parsed == nil {
		return ""
	}
	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		return host
	}
	if (parsed.Scheme == "https" && port == "443") || (parsed.Scheme == "http" && port == "80") {
		return host
	}
	return net.JoinHostPort(host, port)
}

func inspectProxyPublicHTTP(ctx context.Context, parsed *url.URL) *ProxyHTTPProbe {
	if parsed == nil {
		return &ProxyHTTPProbe{Status: "fail", Error: "public URL is unavailable"}
	}
	started := time.Now()
	result := &ProxyHTTPProbe{
		URL:        displayProxyStoryURL(parsed),
		DialTarget: parsed.Host,
		HostHeader: parsed.Host,
		Status:     "fail",
	}
	if parsed.Scheme == "https" {
		result.TLSServerName = parsed.Hostname()
	}

	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: 4 * time.Second}).DialContext,
		TLSHandshakeTimeout:   4 * time.Second,
		ResponseHeaderTimeout: 6 * time.Second,
		DisableCompression:    true,
	}
	defer transport.CloseIdleConnections()

	originHost := normalizeDNSName(parsed.Hostname())
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.Response != nil {
				appendProxyHTTPHop(result, req.Response)
			}
			if normalizeDNSName(req.URL.Hostname()) != originHost {
				result.CrossHostRedirectStopped = true
				return http.ErrUseLastResponse
			}
			if len(via) >= proxyHTTPRedirectLimit {
				result.RedirectLimitReached = true
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, parsed.String(), nil)
	if err != nil {
		result.Error = boundedEvidence(err.Error(), 512)
		result.DurationMS = time.Since(started).Milliseconds()
		return result
	}
	req.Header.Set("User-Agent", "HostSleuth-ProxyStory/1")
	resp, err := client.Do(req)
	result.DurationMS = time.Since(started).Milliseconds()
	if err != nil {
		result.Error = boundedEvidence(err.Error(), 512)
		return result
	}
	defer resp.Body.Close()
	appendProxyHTTPHop(result, resp)
	populateProxyHTTPFinal(result, resp)
	result.Status = "pass"
	return result
}

func inspectProxyUpstreamHTTP(ctx context.Context, parsed *url.URL, hostHeader, tlsServerName string) *ProxyHTTPProbe {
	if parsed == nil {
		return &ProxyHTTPProbe{Status: "fail", Error: "upstream URL is unavailable"}
	}
	target, err := proxyURLTarget(parsed)
	if err != nil {
		return &ProxyHTTPProbe{Status: "fail", Error: boundedEvidence(err.Error(), 512)}
	}

	started := time.Now()
	if strings.TrimSpace(hostHeader) == "" {
		hostHeader = parsed.Host
	}
	if strings.TrimSpace(tlsServerName) == "" {
		tlsServerName = parsed.Hostname()
	}
	result := &ProxyHTTPProbe{
		URL:        displayProxyStoryURL(parsed),
		DialTarget: target,
		HostHeader: hostHeader,
		Status:     "fail",
	}
	if parsed.Scheme == "https" {
		result.TLSServerName = tlsServerName
	}

	dialTarget := target
	transport := &http.Transport{
		Proxy: nil,
		DialContext: func(dialCtx context.Context, network, _ string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: 4 * time.Second}
			return dialer.DialContext(dialCtx, network, dialTarget)
		},
		TLSHandshakeTimeout:   4 * time.Second,
		ResponseHeaderTimeout: 6 * time.Second,
		DisableCompression:    true,
	}
	if parsed.Scheme == "https" {
		transport.TLSClientConfig = &tls.Config{ServerName: tlsServerName, MinVersion: tls.VersionTLS12}
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.Response != nil {
				appendProxyHTTPHop(result, req.Response)
			}
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, parsed.String(), nil)
	if err != nil {
		result.Error = boundedEvidence(err.Error(), 512)
		result.DurationMS = time.Since(started).Milliseconds()
		return result
	}
	req.Host = hostHeader
	req.Header.Set("User-Agent", "HostSleuth-ProxyStory/1")
	resp, err := client.Do(req)
	result.DurationMS = time.Since(started).Milliseconds()
	if err != nil {
		result.Error = boundedEvidence(err.Error(), 512)
		return result
	}
	defer resp.Body.Close()
	appendProxyHTTPHop(result, resp)
	populateProxyHTTPFinal(result, resp)
	result.Status = "pass"
	return result
}

func appendProxyHTTPHop(result *ProxyHTTPProbe, resp *http.Response) {
	if result == nil || resp == nil || resp.Request == nil {
		return
	}
	hop := ProxyHTTPHop{
		URL:         displayProxyStoryURL(resp.Request.URL),
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		Location:    boundedEvidence(resp.Header.Get("Location"), proxyHTTPHeaderMaxBytes),
		Server:      boundedEvidence(resp.Header.Get("Server"), 256),
		ContentType: boundedEvidence(resp.Header.Get("Content-Type"), 256),
	}
	if len(result.Redirects) > 0 {
		last := result.Redirects[len(result.Redirects)-1]
		if last.URL == hop.URL && last.StatusCode == hop.StatusCode && last.Location == hop.Location {
			return
		}
	}
	if len(result.Redirects) < proxyHTTPRedirectLimit+1 {
		result.Redirects = append(result.Redirects, hop)
	}
}

func populateProxyHTTPFinal(result *ProxyHTTPProbe, resp *http.Response) {
	if result == nil || resp == nil {
		return
	}
	result.StatusCode = resp.StatusCode
	result.HTTPStatus = resp.Status
	result.Location = boundedEvidence(resp.Header.Get("Location"), proxyHTTPHeaderMaxBytes)
	result.Server = boundedEvidence(resp.Header.Get("Server"), 256)
	result.ContentType = boundedEvidence(resp.Header.Get("Content-Type"), 256)
}

func buildProxyStoryStages(story *ProxyStory, publicParsed, upstreamParsed *url.URL) []ProxyStoryStage {
	if story == nil {
		return nil
	}
	stages := make([]ProxyStoryStage, 0, 11)
	stages = append(stages,
		proxyDiagnosisCheckStage("public-dns", "front", "Public DNS", story.PublicDiagnosis, "dns", true),
		proxyRouteStage("public-route", "front", "Public route", story.PublicDiagnosis),
		proxyDiagnosisCheckStage("public-tcp", "front", "Public TCP", story.PublicDiagnosis, "tcp", true),
	)
	if publicParsed != nil && publicParsed.Scheme == "https" {
		stages = append(stages, proxyTLSStage("public-tls", "front", "Public TLS", story.PublicDiagnosis, story.PublicHTTP, nil))
	}
	stages = append(stages, proxyHTTPStage("public-http", "front", "Public HTTP", story.PublicHTTP))
	stages = append(stages, proxyLocalContextStage(story.PublicDiagnosis))

	stages = append(stages,
		proxyDiagnosisCheckStage("upstream-dns", "upstream", "Upstream DNS", story.UpstreamDiagnosis, "dns", true),
		proxyRouteStage("upstream-route", "upstream", "Upstream route", story.UpstreamDiagnosis),
		proxyDiagnosisCheckStage("upstream-tcp", "upstream", "Upstream TCP", story.UpstreamDiagnosis, "tcp", true),
	)
	if upstreamParsed != nil && upstreamParsed.Scheme == "https" {
		stages = append(stages, proxyTLSStage("upstream-tls", "upstream", "Upstream TLS", story.UpstreamDiagnosis, story.UpstreamNativeHTTP, story.UpstreamPublicHostHTTP))
	}
	stages = append(stages, proxyUpstreamHTTPStage(story.UpstreamNativeHTTP, story.UpstreamPublicHostHTTP))
	return stages
}

func proxyDiagnosisCheckStage(id, group, title string, diagnosis Diagnosis, checkName string, critical bool) ProxyStoryStage {
	check, ok := proxyDiagnosisCheck(diagnosis, checkName)
	if !ok {
		status := "unknown"
		if !critical {
			status = "info"
		}
		return ProxyStoryStage{ID: id, Group: group, Title: title, Status: status, Summary: "no " + checkName + " evidence was produced"}
	}
	status := check.Status
	if status != "pass" && status != "fail" && status != "unknown" {
		status = "unknown"
	}
	if !critical && status == "unknown" {
		status = "info"
	}
	return ProxyStoryStage{
		ID:       id,
		Group:    group,
		Title:    title,
		Status:   status,
		Summary:  proxyCheckSummary(title, status),
		Evidence: []string{boundedEvidence(check.Evidence, 1200)},
	}
}

func proxyRouteStage(id, group, title string, diagnosis Diagnosis) ProxyStoryStage {
	stage := proxyDiagnosisCheckStage(id, group, title, diagnosis, "route", false)
	if stage.Status == "fail" {
		return stage
	}
	if stage.Status == "unknown" {
		stage.Status = "info"
	}
	return stage
}

func proxyTLSStage(id, group, title string, diagnosis Diagnosis, nativeHTTP, forwardedHTTP *ProxyHTTPProbe) ProxyStoryStage {
	evidence := make([]string, 0, 4)
	tlsEvidence := diagnosis.TLS
	nativeVerified := false
	if nativeHTTP != nil && nativeHTTP.Status == "pass" && nativeHTTP.StatusCode > 0 {
		nativeVerified = true
		evidence = append(evidence, "validated HTTPS request completed with "+nativeHTTP.HTTPStatus)
	}
	forwardedVerified := false
	if forwardedHTTP != nil && forwardedHTTP.Status == "pass" && forwardedHTTP.StatusCode > 0 {
		forwardedVerified = true
		evidence = append(evidence, "public Host/SNI HTTPS request completed with "+forwardedHTTP.HTTPStatus)
	}

	if forwardedHTTP != nil && nativeVerified != forwardedVerified {
		return ProxyStoryStage{
			ID:       id,
			Group:    group,
			Title:    title,
			Status:   "warn",
			Summary:  "TLS behavior changes when the public Host/SNI is used against the upstream address",
			Evidence: append(evidence, proxyTLSDiagnosticEvidence(tlsEvidence)...),
		}
	}
	if nativeVerified || forwardedVerified {
		return ProxyStoryStage{
			ID:       id,
			Group:    group,
			Title:    title,
			Status:   "pass",
			Summary:  "TLS completed with hostname/trust validation for the tested request path",
			Evidence: append(evidence, proxyTLSDiagnosticEvidence(tlsEvidence)...),
		}
	}
	if tlsEvidence == nil {
		return ProxyStoryStage{ID: id, Group: group, Title: title, Status: "unknown", Summary: "TLS evidence is unavailable"}
	}
	evidence = append(evidence, proxyTLSDiagnosticEvidence(tlsEvidence)...)
	if tlsEvidence.HandshakeStatus != "pass" || tlsEvidence.HostnameStatus == "fail" || tlsEvidence.TrustStatus == "fail" {
		return ProxyStoryStage{ID: id, Group: group, Title: title, Status: "fail", Summary: "TLS validation failed for the tested endpoint identity", Evidence: evidence}
	}
	return ProxyStoryStage{ID: id, Group: group, Title: title, Status: "unknown", Summary: "TLS connected but validation evidence is incomplete", Evidence: evidence}
}

func proxyTLSDiagnosticEvidence(tlsEvidence *TLSEvidence) []string {
	if tlsEvidence == nil {
		return nil
	}
	parts := []string{
		"handshake=" + valueOrFallback(tlsEvidence.HandshakeStatus, "unknown"),
		"hostname=" + valueOrFallback(tlsEvidence.HostnameStatus, "unknown"),
		"trust=" + valueOrFallback(tlsEvidence.TrustStatus, "unknown"),
	}
	if tlsEvidence.ServerName != "" {
		parts = append(parts, "server_name="+tlsEvidence.ServerName)
	}
	if tlsEvidence.Protocol != "" {
		parts = append(parts, "protocol="+tlsEvidence.Protocol)
	}
	if tlsEvidence.HandshakeError != "" {
		parts = append(parts, "error="+boundedEvidence(tlsEvidence.HandshakeError, 384))
	}
	return parts
}

func proxyHTTPStage(id, group, title string, probe *ProxyHTTPProbe) ProxyStoryStage {
	if probe == nil {
		return ProxyStoryStage{ID: id, Group: group, Title: title, Status: "unknown", Summary: "HTTP probe evidence is unavailable"}
	}
	if probe.Error != "" || probe.Status != "pass" {
		return ProxyStoryStage{ID: id, Group: group, Title: title, Status: "fail", Summary: "HTTP request did not complete", Evidence: []string{boundedEvidence(probe.Error, 900)}}
	}
	status, summary := proxyHTTPStatusClass(probe.StatusCode)
	if probe.CrossHostRedirectStopped {
		status = "warn"
		summary = "HTTP redirected to a different hostname; HostSleuth recorded the redirect but did not follow it"
	}
	if probe.RedirectLimitReached {
		status = "warn"
		summary = "HTTP redirect limit was reached before the path settled"
	}
	return ProxyStoryStage{
		ID:       id,
		Group:    group,
		Title:    title,
		Status:   status,
		Summary:  summary,
		Evidence: proxyHTTPProbeEvidence(probe),
	}
}

func proxyUpstreamHTTPStage(native, forwarded *ProxyHTTPProbe) ProxyStoryStage {
	if native == nil {
		return ProxyStoryStage{ID: "upstream-http", Group: "upstream", Title: "Upstream HTTP", Status: "unknown", Summary: "upstream HTTP evidence is unavailable"}
	}
	nativeStatus, nativeSummary := proxyHTTPProbeClass(native)
	if forwarded == nil {
		return ProxyStoryStage{
			ID:       "upstream-http",
			Group:    "upstream",
			Title:    "Upstream HTTP",
			Status:   nativeStatus,
			Summary:  nativeSummary,
			Evidence: proxyHTTPProbeEvidence(native),
		}
	}

	forwardedStatus, forwardedSummary := proxyHTTPProbeClass(forwarded)
	evidence := append([]string{"native Host/SNI: " + nativeSummary}, proxyHTTPProbeEvidence(native)...)
	evidence = append(evidence, "public Host/SNI: "+forwardedSummary)
	evidence = append(evidence, proxyHTTPProbeEvidence(forwarded)...)

	if nativeStatus == "fail" && forwardedStatus == "fail" {
		return ProxyStoryStage{ID: "upstream-http", Group: "upstream", Title: "Upstream HTTP", Status: "fail", Summary: "upstream HTTP failed with both native and public Host/SNI variants", Evidence: evidence}
	}
	if nativeStatus != forwardedStatus || native.StatusCode != forwarded.StatusCode {
		return ProxyStoryStage{ID: "upstream-http", Group: "upstream", Title: "Upstream HTTP", Status: "warn", Summary: "upstream behavior changes with Host/SNI; virtual-host or proxy Host handling may matter", Evidence: evidence}
	}
	if nativeStatus == "warn" {
		return ProxyStoryStage{ID: "upstream-http", Group: "upstream", Title: "Upstream HTTP", Status: "warn", Summary: "upstream is reachable but both Host/SNI variants return application-level rejection", Evidence: evidence}
	}
	return ProxyStoryStage{ID: "upstream-http", Group: "upstream", Title: "Upstream HTTP", Status: nativeStatus, Summary: "upstream HTTP behavior is consistent across native and public Host/SNI variants", Evidence: evidence}
}

func proxyHTTPProbeClass(probe *ProxyHTTPProbe) (string, string) {
	if probe == nil {
		return "unknown", "HTTP evidence is unavailable"
	}
	if probe.Error != "" || probe.Status != "pass" {
		return "fail", "request failed: " + boundedEvidence(probe.Error, 384)
	}
	return proxyHTTPStatusClass(probe.StatusCode)
}

func proxyHTTPStatusClass(code int) (string, string) {
	switch {
	case code >= 100 && code < 400:
		return "pass", fmt.Sprintf("HTTP %d confirms the request path answered", code)
	case code >= 400 && code < 500:
		return "warn", fmt.Sprintf("HTTP %d reached the application/proxy but the request was rejected", code)
	case code >= 500 && code < 600:
		return "fail", fmt.Sprintf("HTTP %d is a server-side/proxy/upstream failure response", code)
	default:
		return "unknown", "no recognized HTTP response status was captured"
	}
}

func proxyHTTPProbeEvidence(probe *ProxyHTTPProbe) []string {
	if probe == nil {
		return nil
	}
	evidence := []string{
		"dial_target=" + valueOrFallback(probe.DialTarget, "unknown"),
		"host=" + valueOrFallback(probe.HostHeader, "unknown"),
	}
	if probe.TLSServerName != "" {
		evidence = append(evidence, "sni="+probe.TLSServerName)
	}
	if probe.HTTPStatus != "" {
		evidence = append(evidence, "status="+probe.HTTPStatus)
	}
	if probe.Location != "" {
		evidence = append(evidence, "location="+probe.Location)
	}
	if probe.Server != "" {
		evidence = append(evidence, "server="+probe.Server)
	}
	if probe.ContentType != "" {
		evidence = append(evidence, "content_type="+probe.ContentType)
	}
	evidence = append(evidence, fmt.Sprintf("duration_ms=%d", probe.DurationMS))
	if probe.Error != "" {
		evidence = append(evidence, "error="+boundedEvidence(probe.Error, 512))
	}
	return evidence
}

func proxyLocalContextStage(diagnosis Diagnosis) ProxyStoryStage {
	evidence := make([]string, 0, 2)
	for _, name := range []string{"local-listener", "docker-port"} {
		if check, ok := proxyDiagnosisCheck(diagnosis, name); ok {
			evidence = append(evidence, name+": "+boundedEvidence(check.Evidence, 1000))
		}
	}
	if len(evidence) == 0 {
		return ProxyStoryStage{
			ID:      "proxy-local-context",
			Group:   "proxy",
			Title:   "Local proxy evidence",
			Status:  "info",
			Summary: "the public endpoint is not proven local to this HostSleuth instance, so local listener/container ownership is not asserted",
		}
	}
	return ProxyStoryStage{
		ID:       "proxy-local-context",
		Group:    "proxy",
		Title:    "Local proxy evidence",
		Status:   "info",
		Summary:  "current snapshot contains local listener/container context for the public endpoint",
		Evidence: evidence,
	}
}

func proxyDiagnosisCheck(diagnosis Diagnosis, name string) (Check, bool) {
	for _, check := range diagnosis.Checks {
		if check.Name == name {
			return check, true
		}
	}
	return Check{}, false
}

func proxyCheckSummary(title, status string) string {
	switch status {
	case "pass":
		return title + " evidence passed"
	case "fail":
		return title + " evidence failed"
	case "unknown":
		return title + " evidence is unresolved"
	default:
		return title + " evidence is informational"
	}
}

func proxyStoryOutcome(stages []ProxyStoryStage) (status, first, conclusion string) {
	for _, wanted := range []string{"fail", "warn", "unknown"} {
		for _, stage := range stages {
			if stage.Status != wanted {
				continue
			}
			switch wanted {
			case "fail":
				return "fail", stage.ID, "first proven request-path failure: " + stage.Title + " — " + stage.Summary
			case "warn":
				return "warn", stage.ID, "request path is reachable but needs attention at: " + stage.Title + " — " + stage.Summary
			case "unknown":
				return "unknown", stage.ID, "no proven failure was found, but evidence is unresolved at: " + stage.Title
			}
		}
	}
	return "pass", "", "public endpoint and expected upstream path passed the tested DNS/TCP/TLS/HTTP evidence"
}
