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

func runStartTLSStory(args []string) {
	fs := flag.NewFlagSet("starttls", flag.ExitOnError)
	protocol := fs.String("protocol", "", "mail protocol: smtp, imap, or pop3")
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		log.Fatal("starttls requires one target in host:port form")
	}
	if strings.TrimSpace(*protocol) == "" {
		log.Fatal("starttls requires --protocol with smtp, imap, or pop3")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	story, err := core.BuildStartTLSStory(ctx, *protocol, fs.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(story, "", "  ")
	fmt.Println(string(b))
}

func registerStartTLSStoryAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/starttls-story", func(w http.ResponseWriter, r *http.Request) {
		protocol := strings.TrimSpace(r.URL.Query().Get("protocol"))
		target := strings.TrimSpace(r.URL.Query().Get("target"))
		if protocol == "" {
			http.Error(w, "protocol is required", http.StatusBadRequest)
			return
		}
		if target == "" {
			http.Error(w, "target is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
		defer cancel()
		story, err := core.BuildStartTLSStory(ctx, protocol, target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(story)
	})
}
