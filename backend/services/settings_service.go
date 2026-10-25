package services

import (
	"log"

	"calc-app/backend/config"
)

// Settings is the persisted user preference state.
type Settings struct {
	Theme     string `json:"theme"`     // "dark" or "light"
	Mode      string `json:"mode"`      // last active calculator mode
	AngleUnit string `json:"angleUnit"` // "deg", "rad" or "grad"
	DevBits   int    `json:"devBits"`   // 8, 16, 32 or 64
	DevSigned bool   `json:"devSigned"`
	Currency  string `json:"currency"` // ISO code used by finance formatting
}

// DefaultSettings returns the first-run preferences.
func DefaultSettings() Settings {
	return Settings{
		Theme:     "dark",
		Mode:      "normal",
		AngleUnit: "deg",
		DevBits:   32,
		DevSigned: false,
		Currency:  "USD",
	}
}

// SettingsService loads and saves user preferences. It is pure Go; the host
// wires OnChanged to forward updates to the frontend.
type SettingsService struct {
	// OnChanged, when set, is invoked after every save.
	OnChanged func()

	settings Settings
	path     string
}

// Load reads persisted settings, filling gaps with defaults. Call once
// before serving.
func (s *SettingsService) Load() error {
	path, err := config.Filepath("settings.json")
	if err != nil {
		return err
	}
	s.path = path
	s.settings = DefaultSettings()
	var loaded Settings
	if _, err := config.LoadJSON(path, &loaded); err != nil {
		log.Printf("settings: could not load %s: %v", path, err)
		return nil
	}
	s.merge(loaded)
	return nil
}

// Get returns the current settings.
func (s *SettingsService) Get() Settings { return s.settings }

// Save replaces the settings, persists them and fires OnChanged.
func (s *SettingsService) Save(next Settings) error {
	s.settings = next
	if s.path == "" {
		return nil
	}
	if err := config.SaveJSON(s.path, s.settings); err != nil {
		return err
	}
	if s.OnChanged != nil {
		s.OnChanged()
	}
	return nil
}

// merge overlays non-zero fields of loaded onto the defaults so new fields
// introduced in later versions keep sane values.
func (s *SettingsService) merge(loaded Settings) {
	cur := &s.settings
	if loaded.Theme != "" {
		cur.Theme = loaded.Theme
	}
	if loaded.Mode != "" {
		cur.Mode = loaded.Mode
	}
	if loaded.AngleUnit != "" {
		cur.AngleUnit = loaded.AngleUnit
	}
	if loaded.DevBits != 0 {
		cur.DevBits = loaded.DevBits
	}
	cur.DevSigned = loaded.DevSigned
	if loaded.Currency != "" {
		cur.Currency = loaded.Currency
	}
}
