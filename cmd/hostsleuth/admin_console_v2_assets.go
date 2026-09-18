package main

import "embed"

//go:embed web/admin_console_v2.js web/admin_console_v2.css
var adminConsoleV2Assets embed.FS

func appendAdminConsoleV2Assets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := adminConsoleV2Assets.ReadFile("web/admin_console_v2.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := adminConsoleV2Assets.ReadFile("web/admin_console_v2.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
