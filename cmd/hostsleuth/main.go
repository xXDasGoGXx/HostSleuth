package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

const version = "0.1.0-dev"

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
		fmt.Println(version)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("HostSleuth - local-first Linux change recorder and diagnostics")
	fmt.Println("usage: hostsleuth <snapshot|diagnose|events|serve|version> [options]")
}

func defaultStateDir() string {
	if v := os.Getenv("HOSTSLEUTH_STATE_DIR"); v != "" { return v }
	if os.Geteuid() == 0 { return "/var/lib/hostsleuth" }
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
		if err := store.AppendEvents(events); err != nil { log.Fatal(err) }
		fmt.Printf("recorded %d change event(s)\n", len(events))
	}
	if err := store.SaveSnapshot(cur); err != nil { log.Fatal(err) }
	fmt.Println(core.SnapshotSummary(cur))
}

func runDiagnose(args []string) {
	fs := flag.NewFlagSet("diagnose", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	_ = fs.Parse(args)
	if fs.NArg() != 1 { log.Fatal("diagnose requires target in host:port form") }
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	store := core.Store{Dir: *state}
	snap, err := store.LoadSnapshot()
	if err != nil { snap = core.Collect(ctx) }
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
	if err != nil { log.Fatal(err) }
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
		if old, err := store.LoadSnapshot(); err == nil { _ = store.AppendEvents(core.DiffSnapshots(old, cur)) }
		_ = store.SaveSnapshot(cur)
	}
	capture()
	go func() {
		t := time.NewTicker(*interval)
		defer t.Stop()
		for { select { case <-ctx.Done(): return; case <-t.C: capture() } }
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		s, err := store.LoadSnapshot(); if err != nil { http.Error(w, err.Error(), 500); return }
		_ = json.NewEncoder(w).Encode(s)
	})
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		e, err := store.ReadEvents(100); if err != nil { http.Error(w, err.Error(), 500); return }
		_ = json.NewEncoder(w).Encode(e)
	})
	mux.HandleFunc("/api/diagnose", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" { http.Error(w, "target is required", 400); return }
		s, err := store.LoadSnapshot(); if err != nil { s = core.Collect(r.Context()) }
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(core.Diagnose(r.Context(), target, s))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s, _ := store.LoadSnapshot(); e, _ := store.ReadEvents(20)
		_ = dashboard.Execute(w, struct { Snapshot core.Snapshot; Events []core.Event; Version string }{s, e, version})
	})
	log.Printf("HostSleuth %s listening on http://%s", version, *listen)
	log.Fatal(http.ListenAndServe(*listen, mux))
}

var dashboard = template.Must(template.New("dashboard").Parse(`<!doctype html><html><head><meta charset="utf-8"><title>HostSleuth</title><style>body{font-family:system-ui;margin:2rem;max-width:1100px;background:#0b1220;color:#e5edf7}a{color:#8ab4ff}.card{background:#121c2d;padding:1rem 1.2rem;border-radius:12px;margin:1rem 0}.muted{color:#95a3b8}input,button{padding:.6rem;font:inherit}code{color:#a7f3d0}.warn{color:#fbbf24}</style></head><body><h1>HostSleuth</h1><p class="muted">What changed, what broke, and why?</p><div class="card"><h2>{{.Snapshot.Host.Hostname}}</h2><p>{{.Snapshot.Host.OS}} · kernel {{.Snapshot.Host.Kernel}}</p><p>{{len .Snapshot.Services}} services · {{len .Snapshot.Listeners}} listeners · {{len .Snapshot.Containers}} containers</p></div><div class="card"><h2>Diagnose</h2><form action="/api/diagnose"><input name="target" placeholder="host:port" required><button>Run diagnosis</button></form></div><div class="card"><h2>Recent changes</h2>{{if .Events}}{{range .Events}}<p><code>{{.At.Format "2006-01-02 15:04:05"}}</code> <span class="{{if eq .Severity "warning"}}warn{{end}}">{{.Summary}}</span></p>{{end}}{{else}}<p class="muted">No recorded changes yet.</p>{{end}}</div><p class="muted">HostSleuth {{.Version}}</p></body></html>`))
