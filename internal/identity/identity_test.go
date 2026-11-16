package identity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateKeepsTheSameID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "id")
	first, err := LoadOrCreate(path)
	if err != nil {
		t.Fatal(err)
	}
	if !uuidPattern.MatchString(first) {
		t.Fatalf("%q is not a UUID", first)
	}
	again, err := LoadOrCreate(path)
	if err != nil || again != first {
		t.Fatalf("second load = %q, %v; want %q", again, err, first)
	}
}

func TestLoadOrCreateRefusesAGarbledFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "id")
	if err := os.WriteFile(path, []byte("not-an-id"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadOrCreate(path); err == nil {
		t.Fatal("want an error for a file that holds no UUID")
	}
}

func TestNewUUIDIsVersion4(t *testing.T) {
	id := newUUID()
	if id[14] != '4' || !uuidPattern.MatchString(id) {
		t.Fatalf("%q is not a version 4 UUID", id)
	}
}
