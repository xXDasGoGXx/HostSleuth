package main

import "embed"

//go:embed web/permissions_story.js web/permissions_story.css
var permissionStoryAssets embed.FS

func appendPermissionStoryAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := permissionStoryAssets.ReadFile("web/permissions_story.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := permissionStoryAssets.ReadFile("web/permissions_story.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
