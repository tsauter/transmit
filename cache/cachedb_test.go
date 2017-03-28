package cache

import (
	"testing"
)

// TestInterface makes sure that all cache backends satisfy the interfaces
func TestInterface(t *testing.T) {
	var cache CacheDB
	// make sure we satisfy the interface
	cache = NewBoltCache()
	if cache == nil {
		t.Errorf("Cache is nil.")
	}
}
