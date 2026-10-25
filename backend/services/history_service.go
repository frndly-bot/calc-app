package services

import (
	"log"
	"time"

	"calc-app/backend/config"
)

// HistoryEntry is one calculation record.
type HistoryEntry struct {
	ID        int64  `json:"id"`
	Mode      string `json:"mode"` // normal | scientific | developer | finance
	Expr      string `json:"expr"`
	Result    string `json:"result"`
	CreatedAt string `json:"createdAt"`
}

// maxHistory caps the persisted list so the file stays small.
const maxHistory = 200

// HistoryService records calculations and persists them to disk. It is pure
// Go; the host wires OnChanged to forward updates to the frontend.
type HistoryService struct {
	// OnChanged, when set, is invoked after every mutation.
	OnChanged func()

	entries []HistoryEntry
	nextID  int64
	path    string
}

// Load reads the persisted history. Call once before serving.
func (s *HistoryService) Load() error {
	path, err := config.Filepath("history.json")
	if err != nil {
		return err
	}
	s.path = path
	if _, err := config.LoadJSON(path, &s.entries); err != nil {
		log.Printf("history: could not load %s: %v", path, err)
		s.entries = nil
	}
	for _, e := range s.entries {
		if e.ID >= s.nextID {
			s.nextID = e.ID + 1
		}
	}
	return nil
}

// List returns the most recent entries first.
func (s *HistoryService) List() []HistoryEntry {
	out := make([]HistoryEntry, len(s.entries))
	copy(out, s.entries)
	// Reverse: newest first.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// Add appends an entry, trims to the cap, persists and fires OnChanged.
func (s *HistoryService) Add(mode, expr, result string) (HistoryEntry, error) {
	entry := HistoryEntry{
		ID:        s.nextID,
		Mode:      mode,
		Expr:      expr,
		Result:    result,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	s.nextID++
	s.entries = append(s.entries, entry)
	if excess := len(s.entries) - maxHistory; excess > 0 {
		s.entries = s.entries[excess:]
	}
	if err := s.persist(); err != nil {
		return entry, err
	}
	s.changed()
	return entry, nil
}

// Clear removes every entry, persists and fires OnChanged.
func (s *HistoryService) Clear() error {
	s.entries = nil
	if err := s.persist(); err != nil {
		return err
	}
	s.changed()
	return nil
}

// Delete removes a single entry by id.
func (s *HistoryService) Delete(id int64) error {
	for i, e := range s.entries {
		if e.ID == id {
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			if err := s.persist(); err != nil {
				return err
			}
			s.changed()
			return nil
		}
	}
	return nil
}

func (s *HistoryService) persist() error {
	if s.path == "" {
		return nil
	}
	return config.SaveJSON(s.path, s.entries)
}

func (s *HistoryService) changed() {
	if s.OnChanged != nil {
		s.OnChanged()
	}
}
