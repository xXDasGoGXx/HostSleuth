package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

func runRebootStory(args []string) {
	fs := flag.NewFlagSet("reboot", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	_ = fs.Parse(args)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	store := core.Store{Dir: *state}
	snap, err := store.LoadSnapshot()
	if err != nil {
		snap = core.Collect(ctx)
	}
	events, err := store.ReadEvents(500)
	if err != nil {
		events = nil
	}
	story := core.BuildRebootStory(ctx, snap, events)
	b, _ := json.MarshalIndent(story, "", "  ")
	fmt.Println(string(b))
}

func registerRebootStoryAPI(mux *http.ServeMux, store core.Store) {
	mux.HandleFunc("/api/reboot-story", func(w http.ResponseWriter, r *http.Request) {
		snap, err := store.LoadSnapshot()
		if err != nil {
			snap = core.Collect(r.Context())
		}
		events, err := store.ReadEvents(500)
		if err != nil {
			events = nil
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(core.BuildRebootStory(r.Context(), snap, events))
	})
}
