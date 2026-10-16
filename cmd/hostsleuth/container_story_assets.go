package main

import "embed"

//go:embed web/share.js web/container_story.js web/container_story.css
var containerStoryAssets embed.FS

// appendContainerStoryAssets adds the share dialog (used by every evidence
// view) and the Containers view.
func appendContainerStoryAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := containerStoryAssets.ReadFile("web/container_story.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	for _, name := range []string{"web/share.js", "web/container_story.js"} {
		if js, err := containerStoryAssets.ReadFile(name); err == nil {
			appJS = append(appJS, '\n')
			appJS = append(appJS, js...)
		}
	}
	return appCSS, appJS
}
