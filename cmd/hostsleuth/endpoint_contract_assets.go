package main

import "embed"

//go:embed web/endpoint_contract.js web/endpoint_contract.css
var endpointContractAssets embed.FS

func appendEndpointContractAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := endpointContractAssets.ReadFile("web/endpoint_contract.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := endpointContractAssets.ReadFile("web/endpoint_contract.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
