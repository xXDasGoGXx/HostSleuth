package main

import "embed"

//go:embed web/certificate_rollout.js web/certificate_rollout.css
var certificateRolloutAssets embed.FS

func appendCertificateRolloutAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := certificateRolloutAssets.ReadFile("web/certificate_rollout.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := certificateRolloutAssets.ReadFile("web/certificate_rollout.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
