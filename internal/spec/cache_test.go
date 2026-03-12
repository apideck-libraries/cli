package spec

import (
	"testing"
)

func TestCacheSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	apiSpec := &APISpec{
		Name:    "test",
		Version: "1.0.0",
		BaseURL: "https://example.com",
		APIGroups: map[string]*APIGroup{
			"accounting": {
				Name:      "accounting",
				Resources: map[string]*Resource{},
			},
		},
	}

	err := cache.Save(apiSpec, []byte("raw spec data"))
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := cache.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Name != "test" {
		t.Errorf("Name = %q, want %q", loaded.Name, "test")
	}
	if loaded.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", loaded.Version, "1.0.0")
	}
}

func TestCacheIsFresh(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	if cache.IsFresh() {
		t.Error("empty cache should not be fresh")
	}

	apiSpec := &APISpec{Name: "test", Version: "1.0.0", APIGroups: map[string]*APIGroup{}}
	cache.Save(apiSpec, []byte("data"))

	if !cache.IsFresh() {
		t.Error("just-saved cache should be fresh")
	}
}

func TestCacheMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewCache(tmpDir)

	apiSpec := &APISpec{Name: "test", Version: "2.0.0", APIGroups: map[string]*APIGroup{}}
	cache.Save(apiSpec, []byte("data"))

	meta, err := cache.LoadMeta()
	if err != nil {
		t.Fatalf("LoadMeta failed: %v", err)
	}
	if meta.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", meta.Version, "2.0.0")
	}
}
