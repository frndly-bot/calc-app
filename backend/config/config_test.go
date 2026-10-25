package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirCreatesAndPersists(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("expected directory at %s: %v", dir, err)
	}
	// Second call returns the same path.
	dir2, _ := Dir()
	if dir2 != dir {
		t.Errorf("Dir() not stable: %s vs %s", dir, dir2)
	}
}

func TestLoadSaveJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	type state struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	var s state
	found, err := LoadJSON(path, &s)
	if err != nil || found {
		t.Fatalf("missing file: found=%v err=%v", found, err)
	}

	s = state{Name: "calc", Count: 3}
	if err := SaveJSON(path, &s); err != nil {
		t.Fatal(err)
	}

	var loaded state
	found, err = LoadJSON(path, &loaded)
	if err != nil || !found {
		t.Fatalf("load: found=%v err=%v", found, err)
	}
	if loaded != s {
		t.Errorf("roundtrip mismatch: %+v vs %+v", loaded, s)
	}
}
