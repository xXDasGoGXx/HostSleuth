package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

// containerStoryEventLimit covers a week of history on a busy host; the
// timeline file itself is compacted past 4 MiB.
const containerStoryEventLimit = 5000

func registerContainerStoryAPI(mux *http.ServeMux, store core.Store) {
	mux.HandleFunc("/api/containers", func(w http.ResponseWriter, r *http.Request) {
		snap, err := store.LoadSnapshot()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		events, err := store.ReadDetailedEvents(containerStoryEventLimit)
		if err != nil {
			events = nil
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(core.BuildContainerList(snap, events))
	})
	mux.HandleFunc("/api/container-story", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" || len(name) > 255 {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		snap, err := store.LoadSnapshot()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		events, err := store.ReadDetailedEvents(containerStoryEventLimit)
		if err != nil {
			events = nil
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(core.BuildContainerStory(snap, events, name))
	})
}

// runContainers prints every container's summary, or one container's
// story when a name is given.
func runContainers(args []string) {
	fs := flag.NewFlagSet("containers", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	_ = fs.Parse(args)
	store := core.Store{Dir: *state}
	snap, err := store.LoadSnapshot()
	if err != nil {
		log.Fatal("no snapshot yet; run hostsleuth snapshot or serve first: ", err)
	}
	events, err := store.ReadDetailedEvents(containerStoryEventLimit)
	if err != nil {
		events = nil
	}
	var out any
	switch fs.NArg() {
	case 0:
		out = core.BuildContainerList(snap, events)
	case 1:
		out = core.BuildContainerStory(snap, events, fs.Arg(0))
	default:
		log.Fatal("containers takes at most one container name")
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}
