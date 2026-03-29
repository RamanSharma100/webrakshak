package config

import (
	"encoding/json"
	"io"
	"os"
)

// BlockedWebsites holds the in-memory list of blocked host substrings.
var BlockedWebsites []string

// LoadBlocked loads a JSON array of strings from the given file into
// BlockedWebsites. If the file does not exist, BlockedWebsites will be empty.
func LoadBlocked(path string) error {
	f, err := os.Open(path)
	if err != nil {
		// file missing -> empty list
		BlockedWebsites = []string{}
		return nil
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	BlockedWebsites = arr
	return nil
}

// GetBlocked returns the currently loaded blocklist.
func GetBlocked() []string {
	return BlockedWebsites
}
