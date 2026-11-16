// Package identity keeps the agent's device ID: a UUID made on first run and kept in a file, so BoxServer
// recognises the machine across restarts and hostname changes.
package identity

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// LoadOrCreate returns the ID saved at path, creating it on first run. A file that holds something other than
// a UUID is an error rather than replaced: a new ID would make the server see a new device.
func LoadOrCreate(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err == nil {
		id := strings.TrimSpace(string(b))
		if !uuidPattern.MatchString(id) {
			return "", fmt.Errorf("%s does not hold a device ID", path)
		}
		return id, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	id := newUUID()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(id+"\n"), 0o600); err != nil {
		return "", err
	}
	return id, nil
}

// newUUID makes a random (version 4) UUID.
func newUUID() string {
	var b [16]byte
	rand.Read(b[:]) // never fails since Go 1.24
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
