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

func runPermissionStory(args []string) {
	fs := flag.NewFlagSet("permissions", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	service := fs.String("service", "", "native systemd service unit")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		log.Fatal("permissions requires one absolute path")
	}
	if strings.TrimSpace(*service) == "" {
		log.Fatal("permissions requires --service with a native systemd unit")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	store := core.Store{Dir: *state}
	snap, err := store.LoadSnapshot()
	if err != nil {
		snap = core.Collect(ctx)
	}
	story := core.BuildPermissionStory(ctx, *service, fs.Arg(0), snap)
	b, _ := json.MarshalIndent(story, "", "  ")
	fmt.Println(string(b))
}

func registerPermissionStoryAPI(mux *http.ServeMux, store core.Store) {
	mux.HandleFunc("/api/permissions-story", func(w http.ResponseWriter, r *http.Request) {
		service := strings.TrimSpace(r.URL.Query().Get("service"))
		path := strings.TrimSpace(r.URL.Query().Get("path"))
		if service == "" {
			http.Error(w, "service is required", http.StatusBadRequest)
			return
		}
		if path == "" {
			http.Error(w, "path is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()
		snap, err := store.LoadSnapshot()
		if err != nil {
			snap = core.Collect(ctx)
		}
		story := core.BuildPermissionStory(ctx, service, path, snap)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(story)
	})
}
