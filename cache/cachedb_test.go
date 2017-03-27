package cache

import (
	"testing"
)

func TestInterface(t *testing.T) {
	var cache CacheDB
	// make sure we satisfy the interface
	cache = NewBoltCache()
	if cache == nil {
		t.Errorf("Cache is nil.")
	}
}
