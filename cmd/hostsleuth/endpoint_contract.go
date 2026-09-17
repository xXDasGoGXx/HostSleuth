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

func runEndpointContract(args []string) {
	fs := flag.NewFlagSet("contract", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	name := fs.String("name", "", "optional contract name")
	tlsMode := fs.String("tls", core.EndpointTLSIgnore, "TLS expectation: ignore, present, verified, or forbidden")
	service := fs.String("service", "", "optional systemd service expected to be active")
	container := fs.String("container", "", "optional container expected to be running")
	expectedAddresses := &stringListFlag{}
	fs.Var(expectedAddresses, "expect-ip", "expected DNS address; repeat to require an exact address set")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		log.Fatal("contract requires one target in host:port form")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	store := core.Store{Dir: *state}
	snap, err := store.LoadSnapshot()
	if err != nil {
		snap = core.Collect(ctx)
	}
	evaluation, err := core.EvaluateEndpointContract(ctx, core.EndpointContract{
		Name:              *name,
		Target:            fs.Arg(0),
		ExpectedAddresses: []string(*expectedAddresses),
		TLS:               *tlsMode,
		Service:           *service,
		Container:         *container,
	}, snap)
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(evaluation, "", "  ")
	fmt.Println(string(b))
}

func registerEndpointContractAPI(mux *http.ServeMux, store core.Store) {
	mux.HandleFunc("/api/contract", func(w http.ResponseWriter, r *http.Request) {
		target := strings.TrimSpace(r.URL.Query().Get("target"))
		if target == "" {
			http.Error(w, "target is required", http.StatusBadRequest)
			return
		}
		snap, err := store.LoadSnapshot()
		if err != nil {
			snap = core.Collect(r.Context())
		}
		evaluation, err := core.EvaluateEndpointContract(r.Context(), core.EndpointContract{
			Name:              r.URL.Query().Get("name"),
			Target:            target,
			ExpectedAddresses: contractAddressesFromQuery(r),
			TLS:               r.URL.Query().Get("tls"),
			Service:           r.URL.Query().Get("service"),
			Container:         r.URL.Query().Get("container"),
		}, snap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(evaluation)
	})
}

func contractAddressesFromQuery(r *http.Request) []string {
	if r == nil {
		return nil
	}
	var out []string
	for _, value := range r.URL.Query()["ip"] {
		for _, part := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	return out
}
