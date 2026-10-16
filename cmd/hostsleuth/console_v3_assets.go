package main

import "embed"

//go:embed web/console_v3.css
var consoleV3Assets embed.FS

// appendConsoleV3Assets adds the 1.2 layout layer last so it can refine
// every earlier view. Its behavior changes live in app.js.
func appendConsoleV3Assets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := consoleV3Assets.ReadFile("web/console_v3.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	return appCSS, appJS
}
