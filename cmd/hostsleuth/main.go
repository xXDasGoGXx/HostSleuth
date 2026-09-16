package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

var version = "0.1.0-dev"

//go:embed web/index.html web/app.css web/app.js
var webAssets embed.FS

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "snapshot":
		runSnapshot(os.Args[2:])
	case "diagnose":
		runDiagnose(os.Args[2:])
	case "events":
		runEvents(os.Args[2:])
	case "serve":
		runServe(os.Args[2:])
	case "version":
		fmt.Println(versionDisplay())
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("HostSleuth - local-first Linux change recorder and diagnostics")
	fmt.Println("usage: hostsleuth <snapshot|diagnose|events|serve|version> [options]")
}

func buildRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key != "vcs.revision" || setting.Value == "" {
			continue
		}
		if len(setting.Value) > 12 {
			return setting.Value[:12]
		}
		return setting.Value
	}
	return ""
}

func versionDisplay() string {
	if revision := buildRevision(); revision != "" {
		return fmt.Sprintf("%s (%s)", version, revision)
	}
	return version
}

func defaultStateDir() string {
	if v := os.Getenv("HOSTSLEUTH_STATE_DIR"); v != "" {
		return v
	}
	if os.Geteuid() == 0 {
		return "/var/lib/hostsleuth"
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "hostsleuth")
}

func runSnapshot(args []string) {
	fs := flag.NewFlagSet("snapshot", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	_ = fs.Parse(args)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	store := core.Store{Dir: *state}
	old, oldErr := store.LoadSnapshot()
	cur := core.Collect(ctx)
	if oldErr == nil {
		events := core.DiffSnapshots(old, cur)
		if err := store.AppendEvents(events); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("recorded %d change event(s)\n", len(events))
	}
	if err := store.SaveSnapshot(cur); err != nil {
		log.Fatal(err)
	}
	fmt.Println(core.SnapshotSummary(cur))
}

func runDiagnose(args []string) {
	fs := flag.NewFlagSet("diagnose", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		log.Fatal("diagnose requires target in host:port form")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	store := core.Store{Dir: *state}
	snap, err := store.LoadSnapshot()
	if err != nil {
		snap = core.Collect(ctx)
	}
	d := core.Diagnose(ctx, fs.Arg(0), snap)
	b, _ := json.MarshalIndent(d, "", "  ")
	fmt.Println(string(b))
}

func runEvents(args []string) {
	fs := flag.NewFlagSet("events", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	limit := fs.Int("limit", 50, "maximum recent events")
	_ = fs.Parse(args)
	events, err := (core.Store{Dir: *state}).ReadEvents(*limit)
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(events, "", "  ")
	fmt.Println(string(b))
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	listen := fs.String("listen", "127.0.0.1:8787", "HTTP listen address")
	interval := fs.Duration("interval", 60*time.Second, "snapshot interval")
	_ = fs.Parse(args)
	store := core.Store{Dir: *state}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	capture := func() {
		cctx, ccancel := context.WithTimeout(ctx, 20*time.Second)
		defer ccancel()
		cur := core.Collect(cctx)
		if old, err := store.LoadSnapshot(); err == nil {
			_ = store.AppendEvents(core.DiffSnapshots(old, cur))
		}
		_ = store.SaveSnapshot(cur)
	}
	capture()
	go func() {
		t := time.NewTicker(*interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				capture()
			}
		}
	}()

	indexHTML, err := webAssets.ReadFile("web/index.html")
	if err != nil {
		log.Fatal(err)
	}
	appCSS, err := webAssets.ReadFile("web/app.css")
	if err != nil {
		log.Fatal(err)
	}
	appJS, err := webAssets.ReadFile("web/app.js")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/about", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"version":  version,
			"revision": buildRevision(),
		})
	})
	mux.HandleFunc("/api/snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		s, err := store.LoadSnapshot()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(s)
	})
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		e, err := store.ReadEvents(100)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(e)
	})
	mux.HandleFunc("/api/diagnose", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "target is required", http.StatusBadRequest)
			return
		}
		s, err := store.LoadSnapshot()
		if err != nil {
			s = core.Collect(r.Context())
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(core.Diagnose(r.Context(), target, s))
	})
	mux.HandleFunc("/assets/app.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(appCSS)
	})
	mux.HandleFunc("/assets/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(appJS)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(indexHTML)
	})

	log.Printf("HostSleuth %s listening on http://%s", versionDisplay(), *listen)
	log.Fatal(http.ListenAndServe(*listen, mux))
}
