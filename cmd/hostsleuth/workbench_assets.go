package main

import "embed"

//go:embed web/workbench.js web/workbench.css
var workbenchAssets embed.FS

func appendWorkbenchAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := workbenchAssets.ReadFile("web/workbench.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := workbenchAssets.ReadFile("web/workbench.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
