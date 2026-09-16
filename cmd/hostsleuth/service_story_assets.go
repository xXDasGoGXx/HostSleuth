package main

import "embed"

//go:embed web/service_story.js web/service_story.css
var serviceStoryAssets embed.FS

func appendServiceStoryAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := serviceStoryAssets.ReadFile("web/service_story.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := serviceStoryAssets.ReadFile("web/service_story.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
