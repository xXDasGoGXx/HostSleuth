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

func runWorkbench(args []string) {
	if len(args) == 0 {
		log.Fatal("workbench requires one of: file, compare, dns, http, cert")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var result any
	var err error
	switch args[0] {
	case "file":
		fs := flag.NewFlagSet("workbench file", flag.ExitOnError)
		expected := fs.String("expect", "", "optional expected SHA-256 or SHA-512 checksum")
		_ = fs.Parse(args[1:])
		if fs.NArg() != 1 {
			log.Fatal("workbench file requires one local file path")
		}
		result, err = core.InspectFile(ctx, fs.Arg(0), *expected)
	case "compare":
		fs := flag.NewFlagSet("workbench compare", flag.ExitOnError)
		_ = fs.Parse(args[1:])
		if fs.NArg() != 2 {
			log.Fatal("workbench compare requires two local file paths")
		}
		result, err = core.CompareFiles(ctx, fs.Arg(0), fs.Arg(1))
	case "dns":
		fs := flag.NewFlagSet("workbench dns", flag.ExitOnError)
		_ = fs.Parse(args[1:])
		if fs.NArg() != 1 {
			log.Fatal("workbench dns requires one name or IP")
		}
		result, err = core.InspectDNS(ctx, fs.Arg(0))
	case "http":
		fs := flag.NewFlagSet("workbench http", flag.ExitOnError)
		_ = fs.Parse(args[1:])
		if fs.NArg() != 1 {
			log.Fatal("workbench http requires one http:// or https:// URL")
		}
		result, err = core.InspectHTTP(ctx, fs.Arg(0))
	case "cert":
		fs := flag.NewFlagSet("workbench cert", flag.ExitOnError)
		target := fs.String("target", "", "optional host:port whose served certificate should be compared")
		_ = fs.Parse(args[1:])
		if fs.NArg() != 1 {
			log.Fatal("workbench cert requires one certificate PEM path")
		}
		if strings.TrimSpace(*target) == "" {
			result, err = core.InspectCertificateFile(fs.Arg(0))
		} else {
			result, err = core.CompareCertificateFileToServed(ctx, fs.Arg(0), *target)
		}
	default:
		log.Fatal("unknown workbench tool; use file, compare, dns, http, or cert")
	}
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(b))
}

func registerWorkbenchAPI(mux *http.ServeMux) {
	write := func(w http.ResponseWriter, value any, err error) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(value)
	}

	mux.HandleFunc("/api/workbench/file", func(w http.ResponseWriter, r *http.Request) {
		write(w, core.InspectFile(r.Context(), r.URL.Query().Get("path"), r.URL.Query().Get("expected")))
	})
	mux.HandleFunc("/api/workbench/compare", func(w http.ResponseWriter, r *http.Request) {
		write(w, core.CompareFiles(r.Context(), r.URL.Query().Get("left"), r.URL.Query().Get("right")))
	})
	mux.HandleFunc("/api/workbench/dns", func(w http.ResponseWriter, r *http.Request) {
		write(w, core.InspectDNS(r.Context(), r.URL.Query().Get("name")))
	})
	mux.HandleFunc("/api/workbench/http", func(w http.ResponseWriter, r *http.Request) {
		write(w, core.InspectHTTP(r.Context(), r.URL.Query().Get("url")))
	})
	mux.HandleFunc("/api/workbench/cert", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		target := strings.TrimSpace(r.URL.Query().Get("target"))
		if target == "" {
			write(w, core.InspectCertificateFile(path))
			return
		}
		write(w, core.CompareCertificateFileToServed(r.Context(), path, target))
	})
}
