package main

import "embed"

//go:embed web/actions.js web/actions.css
var actionAssets embed.FS

func appendActionAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := actionAssets.ReadFile("web/actions.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := actionAssets.ReadFile("web/actions.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
