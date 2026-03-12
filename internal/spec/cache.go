// internal/spec/cache.go
package spec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheTTLHours    = 24
	parsedBinFile    = "parsed.bin"
	rawSpecFile      = "spec.yml"
	metaJSONFile     = "meta.json"
)

// CacheMeta holds metadata about a cached spec.
type CacheMeta struct {
	Version   string    `json:"version"`
	FetchedAt time.Time `json:"fetched_at"`
	Source    string    `json:"source"`
	TTLHours  int       `json:"ttl_hours"`
}

// Cache manages on-disk caching of a parsed APISpec using gob encoding.
type Cache struct {
	dir string
}

// NewCache returns a Cache that stores files in dir.
func NewCache(dir string) *Cache {
	return &Cache{dir: dir}
}

// DefaultCacheDir returns the default cache directory (~/.apideck-cli/cache/).
func DefaultCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".apideck-cli", "cache")
	}
	return filepath.Join(home, ".apideck-cli", "cache")
}

// Save writes the parsed APISpec (gob), the raw spec bytes, and metadata to disk.
// All writes are atomic.
func (c *Cache) Save(apiSpec *APISpec, rawSpec []byte) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	// Encode APISpec with gob.
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(apiSpec); err != nil {
		return fmt.Errorf("gob encode spec: %w", err)
	}
	if err := c.atomicWrite(filepath.Join(c.dir, parsedBinFile), buf.Bytes()); err != nil {
		return fmt.Errorf("write parsed.bin: %w", err)
	}

	// Write raw spec.
	if err := c.atomicWrite(filepath.Join(c.dir, rawSpecFile), rawSpec); err != nil {
		return fmt.Errorf("write spec.yml: %w", err)
	}

	// Write metadata.
	meta := &CacheMeta{
		Version:   apiSpec.Version,
		FetchedAt: time.Now().UTC(),
		TTLHours:  cacheTTLHours,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}
	if err := c.atomicWrite(filepath.Join(c.dir, metaJSONFile), metaBytes); err != nil {
		return fmt.Errorf("write meta.json: %w", err)
	}

	return nil
}

// Load reads and decodes the cached APISpec from parsed.bin.
func (c *Cache) Load() (*APISpec, error) {
	data, err := os.ReadFile(filepath.Join(c.dir, parsedBinFile))
	if err != nil {
		return nil, fmt.Errorf("read parsed.bin: %w", err)
	}

	var apiSpec APISpec
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&apiSpec); err != nil {
		return nil, fmt.Errorf("gob decode spec: %w", err)
	}

	return &apiSpec, nil
}

// LoadMeta reads and returns the cache metadata from meta.json.
func (c *Cache) LoadMeta() (*CacheMeta, error) {
	data, err := os.ReadFile(filepath.Join(c.dir, metaJSONFile))
	if err != nil {
		return nil, fmt.Errorf("read meta.json: %w", err)
	}

	var meta CacheMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal meta: %w", err)
	}

	return &meta, nil
}

// IsFresh returns true if the cache exists and was saved within the TTL window.
func (c *Cache) IsFresh() bool {
	meta, err := c.LoadMeta()
	if err != nil {
		return false
	}
	ttl := time.Duration(meta.TTLHours) * time.Hour
	return time.Since(meta.FetchedAt) < ttl
}

// LoadRawSpec reads and returns the raw spec bytes from spec.yml.
func (c *Cache) LoadRawSpec() ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(c.dir, rawSpecFile))
	if err != nil {
		return nil, fmt.Errorf("read spec.yml: %w", err)
	}
	return data, nil
}

// atomicWrite writes data to a temporary file in the same directory and then
// renames it to path, ensuring an atomic replacement.
func (c *Cache) atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	// Ensure the temp file is removed on failure.
	success := false
	defer func() {
		if !success {
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	success = true
	return nil
}
