package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

// registerShareAPI serves POST /api/share: the browser sends the evidence
// summary it already shows, and gets back a redacted copy to paste or
// download. Nothing is stored or sent anywhere else.
func registerShareAPI(mux *http.ServeMux, store core.Store) {
	mux.HandleFunc("/api/share", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "POST the text to redact", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, core.ShareTextLimit+1))
		if err != nil {
			http.Error(w, "could not read the text", http.StatusBadRequest)
			return
		}
		if len(body) > core.ShareTextLimit {
			http.Error(w, "text is too long to redact", http.StatusRequestEntityTooLarge)
			return
		}
		snap, err := store.LoadSnapshot()
		if err != nil {
			http.Error(w, "no snapshot is available to seed redaction", http.StatusServiceUnavailable)
			return
		}
		events, err := store.ReadDetailedEvents(500)
		if err != nil {
			events = nil
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{"text": core.RedactForSharing(snap, events, string(body))})
	})
}
