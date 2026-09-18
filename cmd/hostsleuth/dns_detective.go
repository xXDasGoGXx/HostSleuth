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

func runDNSDetective(args []string) {
	fs := flag.NewFlagSet("dns", flag.ExitOnError)
	resolvers := &stringListFlag{}
	fs.Var(resolvers, "resolver", "custom DNS resolver as LABEL=IP or IP; repeat up to four times")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		log.Fatal("dns requires one DNS name or IP address")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	result, err := core.InspectDNSDetective(ctx, fs.Arg(0), []string(*resolvers))
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
}

func registerDNSDetectiveAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/dns-detective", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSpace(r.URL.Query().Get("name"))
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		resolvers := r.URL.Query()["resolver"]
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()
		result, err := core.InspectDNSDetective(ctx, name, resolvers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	})
}
