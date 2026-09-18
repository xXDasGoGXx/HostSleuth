package main

import "embed"

//go:embed web/proxy_story.js web/proxy_story.css
var proxyStoryAssets embed.FS

func appendProxyStoryAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := proxyStoryAssets.ReadFile("web/proxy_story.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := proxyStoryAssets.ReadFile("web/proxy_story.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
