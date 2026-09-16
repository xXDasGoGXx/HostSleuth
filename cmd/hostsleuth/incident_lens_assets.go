package main

import "embed"

//go:embed web/incident_lens.js web/incident_lens.css
var incidentLensAssets embed.FS

func appendIncidentLensAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := incidentLensAssets.ReadFile("web/incident_lens.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := incidentLensAssets.ReadFile("web/incident_lens.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
