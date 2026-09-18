package main

import "embed"

//go:embed web/starttls_story.js web/starttls_story.css
var startTLSStoryAssets embed.FS

func appendStartTLSStoryAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := startTLSStoryAssets.ReadFile("web/starttls_story.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := startTLSStoryAssets.ReadFile("web/starttls_story.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
