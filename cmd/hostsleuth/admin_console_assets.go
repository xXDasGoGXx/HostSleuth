package main

import "embed"

//go:embed web/admin_console.js web/admin_console.css
var adminConsoleAssets embed.FS

func appendAdminConsoleAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := adminConsoleAssets.ReadFile("web/admin_console.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := adminConsoleAssets.ReadFile("web/admin_console.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
