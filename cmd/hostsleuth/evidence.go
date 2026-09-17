package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/xXDasGoGXx/HostSleuth/internal/core"
)

func runEvidence(args []string) {
	if len(args) == 0 {
		log.Fatal("evidence requires one of: preview, export")
	}
	switch args[0] {
	case "preview":
		runEvidencePreview(args[1:])
	case "export":
		runEvidenceExport(args[1:])
	default:
		log.Fatal("unknown evidence operation; use preview or export")
	}
}

func runEvidencePreview(args []string) {
	fs := flag.NewFlagSet("evidence preview", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	_ = fs.Parse(args)
	if fs.NArg() != 0 {
		log.Fatal("evidence preview accepts no positional arguments")
	}
	preview, err := core.BuildEvidencePreview(core.Store{Dir: *state}, versionDisplay(), time.Now().UTC())
	if err != nil {
		log.Fatal(err)
	}
	writeEvidenceJSON(preview)
}

func runEvidenceExport(args []string) {
	fs := flag.NewFlagSet("evidence export", flag.ExitOnError)
	state := fs.String("state-dir", defaultStateDir(), "state directory")
	output := fs.String("output", "", "output ZIP path; defaults to a timestamped file in the current directory")
	_ = fs.Parse(args)
	if fs.NArg() != 0 {
		log.Fatal("evidence export accepts no positional arguments")
	}
	now := time.Now().UTC()
	path := *output
	if path == "" {
		path = core.DefaultEvidenceFilename(now)
	}
	result, err := core.ExportEvidence(core.Store{Dir: *state}, versionDisplay(), path, now)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s\nsha256 %s\nbytes %d\n", result.Path, result.SHA256, result.Bytes)
}

func writeEvidenceJSON(value any) {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
