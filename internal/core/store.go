package core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Event timeline retention. The timeline is append-only JSON lines and is
// read in full whenever the UI or a story needs recent history, so it must
// not grow without bound. When the file passes eventFileCompactBytes it is
// rewritten atomically with routine listener churn removed and only the
// newest eventFileKeep events retained.
const (
	eventFileCompactBytes = 4 << 20
	eventFileKeep         = 5000
)

// eventMu serializes timeline writes within one process (scheduled
// snapshots and compaction).
var eventMu sync.Mutex

type Store struct{ Dir string }

func (s Store) ensure() error {
	if s.Dir == "" {
		return errors.New("store directory is empty")
	}
	return os.MkdirAll(s.Dir, 0o700)
}

func (s Store) SnapshotPath() string { return filepath.Join(s.Dir, "snapshot.json") }
func (s Store) EventsPath() string   { return filepath.Join(s.Dir, "events.jsonl") }

func (s Store) LoadSnapshot() (Snapshot, error) {
	var snap Snapshot
	b, err := os.ReadFile(s.SnapshotPath())
	if err != nil {
		return snap, err
	}
	err = json.Unmarshal(b, &snap)
	return snap, err
}

func (s Store) SaveSnapshot(snap Snapshot) error {
	if snap.SchemaVersion == 0 {
		snap.SchemaVersion = 1
	}
	if err := s.ensure(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.SnapshotPath() + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.SnapshotPath())
}

func (s Store) AppendEvents(events []Event) error {
	if len(events) == 0 {
		return nil
	}
	if err := s.ensure(); err != nil {
		return err
	}
	eventMu.Lock()
	defer eventMu.Unlock()
	f, err := os.OpenFile(s.EventsPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	for _, e := range events {
		if e.SchemaVersion == 0 {
			e.SchemaVersion = 1
		}
		if e.At.IsZero() {
			e.At = time.Now().UTC()
		}
		if err := enc.Encode(e); err != nil {
			_ = f.Close()
			return err
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	return s.compactEventsIfLarge()
}

// compactEventsIfLarge must be called with eventMu held.
func (s Store) compactEventsIfLarge() error {
	info, err := os.Stat(s.EventsPath())
	if err != nil || info.Size() <= eventFileCompactBytes {
		return nil
	}
	events, err := s.readEventFile()
	if err != nil {
		return err
	}
	events = MeaningfulEvents(events)
	if len(events) > eventFileKeep {
		events = events[len(events)-eventFileKeep:]
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, e := range events {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	tmp := s.EventsPath() + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.EventsPath())
}

// ReadEvents returns the newest `limit` meaningful events, oldest first.
// Routine listener churn recorded by earlier versions is skipped and
// listeners that went away and came back are collapsed into one summary
// per snapshot, both before the limit is applied, so neither can crowd out
// real history such as a reboot.
func (s Store) ReadEvents(limit int) ([]Event, error) {
	events, err := s.readEventFile()
	if err != nil {
		return nil, err
	}
	return newestEvents(CollapseListenerBlips(MeaningfulEvents(events)), limit), nil
}

// ReadDetailedEvents is ReadEvents without blip collapsing, for stories
// that correlate individual listener events with one service or port.
func (s Store) ReadDetailedEvents(limit int) ([]Event, error) {
	events, err := s.readEventFile()
	if err != nil {
		return nil, err
	}
	return newestEvents(MeaningfulEvents(events), limit), nil
}

// ReadAllEvents returns the newest `limit` recorded events including
// routine listener churn, for operators who want the raw timeline.
func (s Store) ReadAllEvents(limit int) ([]Event, error) {
	events, err := s.readEventFile()
	if err != nil {
		return nil, err
	}
	return newestEvents(events, limit), nil
}

func newestEvents(events []Event, limit int) []Event {
	if limit > 0 && len(events) > limit {
		events = events[len(events)-limit:]
	}
	return events
}

func (s Store) readEventFile() ([]Event, error) {
	f, err := os.Open(s.EventsPath())
	if errors.Is(err, os.ErrNotExist) {
		return []Event{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	events := []Event{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scanner.Scan() {
		var e Event
		if json.Unmarshal(scanner.Bytes(), &e) == nil {
			events = append(events, e)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
