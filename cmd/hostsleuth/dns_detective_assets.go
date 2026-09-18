package main

import "embed"

//go:embed web/dns_detective.js web/dns_detective.css
var dnsDetectiveAssets embed.FS

func appendDNSDetectiveAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := dnsDetectiveAssets.ReadFile("web/dns_detective.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := dnsDetectiveAssets.ReadFile("web/dns_detective.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}
