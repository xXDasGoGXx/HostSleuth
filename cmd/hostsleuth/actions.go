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

type stringListFlag []string

func (v *stringListFlag) String() string { return strings.Join(*v, ",") }
func (v *stringListFlag) Set(value string) error {
	*v = append(*v, value)
	return nil
}

func runAction(args []string) {
	if len(args) == 0 {
		log.Fatal("action requires one of: list, preview, run, audit")
	}
	switch args[0] {
	case "list":
		fs, state, enabled, allowed := actionFlagSet("action list")
		_ = fs.Parse(args[1:])
		manager := actionManagerFromFlags(*state, *enabled, *allowed)
		writeActionJSON(manager.Capabilities(actionMode()))
	case "preview":
		fs, state, enabled, allowed := actionFlagSet("action preview")
		id := fs.String("id", "service.restart", "fixed action ID")
		target := fs.String("target", "", "action target")
		_ = fs.Parse(args[1:])
		if strings.TrimSpace(*target) == "" {
			log.Fatal("action preview requires --target")
		}
		manager := actionManagerFromFlags(*state, *enabled, *allowed)
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		writeActionJSON(manager.Preview(ctx, *id, *target, actionMode()))
	case "run":
		fs, state, enabled, allowed := actionFlagSet("action run")
		id := fs.String("id", "service.restart", "fixed action ID")
		target := fs.String("target", "", "action target")
		confirm := fs.String("confirm", "", "exact confirmation value returned by preview")
		_ = fs.Parse(args[1:])
		if strings.TrimSpace(*target) == "" {
			log.Fatal("action run requires --target")
		}
		manager := actionManagerFromFlags(*state, *enabled, *allowed)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		result, err := manager.Run(ctx, *id, *target, *confirm, actionMode())
		writeActionJSON(result)
		if err != nil {
			log.Fatal(err)
		}
		if result.Status != "success" {
			log.Fatalf("action finished with status %s", result.Status)
		}
	case "audit":
		fs := flag.NewFlagSet("action audit", flag.ExitOnError)
		state := fs.String("state-dir", defaultStateDir(), "state directory")
		limit := fs.Int("limit", 50, "maximum recent action audit records")
		_ = fs.Parse(args[1:])
		audits, err := (core.Store{Dir: *state}).ReadActionAudits(*limit)
		if err != nil {
			log.Fatal(err)
		}
		writeActionJSON(audits)
	default:
		log.Fatal("unknown action operation; use list, preview, run, or audit")
	}
}

func actionFlagSet(name string) (*flag.FlagSet, *string, *bool, *stringListFlag) {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	enabled := fs.Bool("enable-actions", false, "explicitly enable optional safe actions")
	allowed := &stringListFlag{}
	fs.Var(allowed, "allow-restart-service", "allow one systemd service for service.restart; repeat for additional services")
	return fs, state, enabled, allowed
}

func actionManagerFromFlags(state string, enabled bool, allowed []string) *core.ActionManager {
	policy, err := core.NewActionPolicy(enabled, allowed)
	if err != nil {
		log.Fatal(err)
	}
	return &core.ActionManager{Policy: policy, Store: core.Store{Dir: state}}
}

func actionMode() string {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return core.Collect(ctx).Mode
}

func writeActionJSON(value any) {
	b, _ := json.MarshalIndent(value, "", "  ")
	fmt.Println(string(b))
}

type actionAPIRequest struct {
	ActionID     string `json:"action_id"`
	Target       string `json:"target"`
	Confirmation string `json:"confirmation,omitempty"`
}

func registerActionAPI(mux *http.ServeMux, manager *core.ActionManager, mode string) {
	guard := func(w http.ResponseWriter, r *http.Request) bool {
		if workbenchRequestAllowed(r) {
			return true
		}
		http.Error(w, "Action Web/API operations are loopback-only; use the local UI, an SSH tunnel, or the HostSleuth CLI", http.StatusForbidden)
		return false
	}
	write := func(w http.ResponseWriter, status int, value any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(value)
	}
	decode := func(w http.ResponseWriter, r *http.Request) (actionAPIRequest, bool) {
		var request actionAPIRequest
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type"))), "application/json") {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return request, false
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "invalid action request: "+err.Error(), http.StatusBadRequest)
			return request, false
		}
		return request, true
	}

	mux.HandleFunc("/api/actions", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r) {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		write(w, http.StatusOK, manager.Capabilities(mode))
	})
	mux.HandleFunc("/api/actions/preview", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		request, ok := decode(w, r)
		if !ok {
			return
		}
		preview := manager.Preview(r.Context(), request.ActionID, request.Target, mode)
		write(w, http.StatusOK, preview)
	})
	mux.HandleFunc("/api/actions/run", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-HostSleuth-Action") != "confirm" {
			http.Error(w, "X-HostSleuth-Action: confirm is required", http.StatusPreconditionRequired)
			return
		}
		request, ok := decode(w, r)
		if !ok {
			return
		}
		result, err := manager.Run(r.Context(), request.ActionID, request.Target, request.Confirmation, mode)
		if err != nil {
			write(w, http.StatusInternalServerError, map[string]any{"result": result, "error": err.Error()})
			return
		}
		status := http.StatusOK
		switch result.Status {
		case "denied", "blocked":
			status = http.StatusForbidden
		case "failed":
			status = http.StatusConflict
		}
		write(w, status, result)
	})
	mux.HandleFunc("/api/actions/audit", func(w http.ResponseWriter, r *http.Request) {
		if !guard(w, r) {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		audits, err := manager.Store.ReadActionAudits(100)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		write(w, http.StatusOK, audits)
	})
}
