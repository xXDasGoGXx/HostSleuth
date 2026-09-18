package main

import (
	"strings"
	"testing"
)

func assembledWebAssetsForTest(t *testing.T) ([]byte, []byte, []byte) {
	t.Helper()

	indexHTML, err := webAssets.ReadFile("web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	appCSS, err := webAssets.ReadFile("web/app.css")
	if err != nil {
		t.Fatal(err)
	}
	appJS, err := webAssets.ReadFile("web/app.js")
	if err != nil {
		t.Fatal(err)
	}

	appCSS, appJS = appendServiceStoryAssets(appCSS, appJS)
	appCSS, appJS = appendIncidentLensAssets(appCSS, appJS)
	appCSS, appJS = appendRebootStoryAssets(appCSS, appJS)
	appCSS, appJS = appendWorkbenchAssets(appCSS, appJS)
	appCSS, appJS = appendEndpointContractAssets(appCSS, appJS)
	appCSS, appJS = appendDNSDetectiveAssets(appCSS, appJS)
	appCSS, appJS = appendProxyStoryAssets(appCSS, appJS)
	appCSS, appJS = appendPermissionStoryAssets(appCSS, appJS)
	appCSS, appJS = appendStartTLSStoryAssets(appCSS, appJS)
	appCSS, appJS = appendCertificateRolloutAssets(appCSS, appJS)
	appCSS, appJS = appendActionAssets(appCSS, appJS)
	appCSS, appJS = appendAdminConsoleAssets(appCSS, appJS)
	appCSS, appJS = appendAdminConsoleV2Assets(appCSS, appJS)
	return indexHTML, appCSS, appJS
}

func TestWebV1ReadinessAccessibilityContract(t *testing.T) {
	indexHTML, appCSS, appJS := assembledWebAssetsForTest(t)
	html := string(indexHTML)
	css := string(appCSS)
	js := string(appJS)

	for _, marker := range []string{
		`lang="en"`,
		`aria-label="Primary navigation"`,
		`aria-live="polite"`,
		`aria-atomic="true"`,
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("index is missing accessibility marker %q", marker)
		}
	}

	for _, marker := range []string{
		`setAttribute("role", "tab")`,
		`setAttribute("aria-selected", String(active))`,
		`setAttribute("role", "tabpanel")`,
		`setAttribute("aria-hidden", String(!active))`,
		`"ArrowLeft"`,
		`"ArrowRight"`,
		`"Home"`,
		`"End"`,
		`new MutationObserver`,
		`prefers-reduced-motion: reduce`,
	} {
		if !strings.Contains(js, marker) {
			t.Fatalf("served JavaScript is missing navigation/accessibility marker %q", marker)
		}
	}

	for _, marker := range []string{
		`:focus-visible`,
		`[role="tabpanel"][hidden]`,
		`@media (prefers-reduced-motion: reduce)`,
		`@media (max-width: 620px)`,
	} {
		if !strings.Contains(css, marker) {
			t.Fatalf("served CSS is missing accessibility/responsive marker %q", marker)
		}
	}
}

func TestWebV1ReadinessAssetBudget(t *testing.T) {
	indexHTML, appCSS, appJS := assembledWebAssetsForTest(t)

	const (
		maxHTML = 64 * 1024
		maxCSS  = 128 * 1024
		maxJS   = 256 * 1024
	)
	if len(indexHTML) > maxHTML {
		t.Fatalf("served HTML exceeded v1 budget: %d > %d bytes", len(indexHTML), maxHTML)
	}
	if len(appCSS) > maxCSS {
		t.Fatalf("served CSS exceeded v1 budget: %d > %d bytes", len(appCSS), maxCSS)
	}
	if len(appJS) > maxJS {
		t.Fatalf("served JavaScript exceeded v1 budget: %d > %d bytes", len(appJS), maxJS)
	}
}
