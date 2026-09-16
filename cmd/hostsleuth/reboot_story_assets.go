package main

import "embed"

//go:embed web/reboot_story.js web/reboot_story.css
var rebootStoryAssets embed.FS

func appendRebootStoryAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := rebootStoryAssets.ReadFile("web/reboot_story.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := rebootStoryAssets.ReadFile("web/reboot_story.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
