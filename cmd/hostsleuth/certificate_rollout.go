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

func runCertificateRollout(args []string) {
	fs := flag.NewFlagSet("cert-rollout", flag.ExitOnError)
	fingerprint := fs.String("fingerprint", "", "expected SHA-256 certificate fingerprint")
	reference := fs.String("reference", "", "reference direct-TLS endpoint in host:port form")
	endpoints := &stringListFlag{}
	fs.Var(endpoints, "endpoint", "direct-TLS endpoint in host:port form; repeat up to 16 times")
	_ = fs.Parse(args)
	if fs.NArg() != 0 {
		log.Fatal("cert-rollout accepts endpoints only through repeated --endpoint flags")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Second)
	defer cancel()
	story, err := core.BuildCertificateRolloutStory(ctx, *fingerprint, *reference, *endpoints)
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(story, "", "  ")
	fmt.Println(string(b))
}

func registerCertificateRolloutAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/certificate-rollout", func(w http.ResponseWriter, r *http.Request) {
		fingerprint := strings.TrimSpace(r.URL.Query().Get("fingerprint"))
		reference := strings.TrimSpace(r.URL.Query().Get("reference"))
		endpoints := r.URL.Query()["endpoint"]

		ctx, cancel := context.WithTimeout(r.Context(), 24*time.Second)
		defer cancel()
		story, err := core.BuildCertificateRolloutStory(ctx, fingerprint, reference, endpoints)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(story)
	})
}
