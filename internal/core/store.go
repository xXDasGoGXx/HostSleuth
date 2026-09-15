package core

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Store struct { Dir string }

func (s Store) ensure() error {
	if s.Dir == "" { return errors.New("store directory is empty") }
	return os.MkdirAll(s.Dir, 0o700)
}

func (s Store) SnapshotPath() string { return filepath.Join(s.Dir, "snapshot.json") }
func (s Store) EventsPath() string { return filepath.Join(s.Dir, "events.jsonl") }

func (s Store) LoadSnapshot() (Snapshot, error) {
	var snap Snapshot
	b, err := os.ReadFile(s.SnapshotPath())
	if err != nil { return snap, err }
	err = json.Unmarshal(b, &snap)
	return snap, err
}

func (s Store) SaveSnapshot(snap Snapshot) error {
	if err := s.ensure(); err != nil { return err }
	b, err := json.MarshalIndent(snap, "", "  ")
	if err != nil { return err }
	tmp := s.SnapshotPath() + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil { return err }
	return os.Rename(tmp, s.SnapshotPath())
}

func (s Store) AppendEvents(events []Event) error {
	if len(events) == 0 { return nil }
	if err := s.ensure(); err != nil { return err }
	f, err := os.OpenFile(s.EventsPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil { return err }
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range events {
		if e.At.IsZero() { e.At = time.Now().UTC() }
		if err := enc.Encode(e); err != nil { return err }
	}
	return nil
}

func (s Store) ReadEvents(limit int) ([]Event, error) {
	f, err := os.Open(s.EventsPath())
	if errors.Is(err, os.ErrNotExist) { return nil, nil }
	if err != nil { return nil, err }
	defer f.Close()
	var events []Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e Event
		if json.Unmarshal(scanner.Bytes(), &e) == nil { events = append(events, e) }
	}
	if err := scanner.Err(); err != nil { return nil, err }
	if limit > 0 && len(events) > limit { events = events[len(events)-limit:] }
	return events, nil
}
