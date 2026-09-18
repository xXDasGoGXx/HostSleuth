package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

func runProxyStory(args []string) {
	fs := flag.NewFlagSet("proxy", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	upstream := fs.String("upstream", "", "expected upstream HTTP/HTTPS URL")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		log.Fatal("proxy requires one public HTTP/HTTPS URL")
	}
	if strings.TrimSpace(*upstream) == "" {
		log.Fatal("proxy requires --upstream with the expected HTTP/HTTPS upstream URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	store := core.Store{Dir: *state}
	snap, err := store.LoadSnapshot()
	if err != nil {
		snap = core.Collect(ctx)
	}
	story, err := core.BuildProxyStory(ctx, fs.Arg(0), *upstream, snap)
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(story, "", "  ")
	fmt.Println(string(b))
}

func registerProxyStoryAPI(mux *http.ServeMux, store core.Store) {
	mux.HandleFunc("/api/proxy-story", func(w http.ResponseWriter, r *http.Request) {
		publicURL := strings.TrimSpace(r.URL.Query().Get("public"))
		upstreamURL := strings.TrimSpace(r.URL.Query().Get("upstream"))
		if publicURL == "" {
			http.Error(w, "public URL is required", http.StatusBadRequest)
			return
		}
		if upstreamURL == "" {
			http.Error(w, "upstream URL is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		snap, err := store.LoadSnapshot()
		if err != nil {
			snap = core.Collect(ctx)
		}
		story, err := core.BuildProxyStory(ctx, publicURL, upstreamURL, snap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(story)
	})
}
