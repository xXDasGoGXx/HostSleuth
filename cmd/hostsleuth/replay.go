package main

import (
	"embed"
	"encoding/json"
	"net/http"
	"path"
	"strings"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

// The B612 typeface (SIL Open Font License, web/fonts/OFL.txt) is served by
// HostSleuth itself so the console never loads anything from the internet.
//
//go:embed web/replay.js web/replay.css web/fonts/*.woff web/fonts/OFL.txt
var replayAssets embed.FS

func appendReplayAssets(appCSS, appJS []byte) ([]byte, []byte) {
	if css, err := replayAssets.ReadFile("web/replay.css"); err == nil {
		appCSS = append(appCSS, '\n')
		appCSS = append(appCSS, css...)
	}
	if js, err := replayAssets.ReadFile("web/replay.js"); err == nil {
		appJS = append(appJS, '\n')
		appJS = append(appJS, js...)
	}
	return appCSS, appJS
}

func registerReplayAPI(mux *http.ServeMux, store core.Store) {
	mux.HandleFunc("/api/replay", func(w http.ResponseWriter, r *http.Request) {
		snap, err := store.LoadSnapshot()
		if err != nil {
			http.Error(w, "no snapshot yet; HostSleuth takes its first one at start", http.StatusServiceUnavailable)
			return
		}
		events, err := store.ReadDetailedEvents(containerStoryEventLimit)
		if err != nil {
			events = nil
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(core.BuildReplay(snap, events))
	})
	mux.HandleFunc("/assets/fonts/", func(w http.ResponseWriter, r *http.Request) {
		name := path.Base(r.URL.Path)
		if name != path.Clean(name) || strings.ContainsAny(name, "/\\") {
			http.NotFound(w, r)
			return
		}
		b, err := replayAssets.ReadFile("web/fonts/" + name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		switch {
		case strings.HasSuffix(name, ".woff"):
			w.Header().Set("Content-Type", "font/woff")
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		case strings.HasSuffix(name, ".txt"):
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		default:
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(b)
	})
}
