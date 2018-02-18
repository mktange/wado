package syncimpls

import (
	"sync"

	"github.com/mktange/wado/internal/pkg/util"
)

// MapStringFileStats is a concurrent imlementation of map[string]string
type MapStringFileStats struct {
	sync.RWMutex
	internal map[string]*util.FileStats
}

// NewMapStringFileStats is a concurrent imlementation of map[string]string
func NewMapStringFileStats() *MapStringFileStats {
	return &MapStringFileStats{
		internal: make(map[string]*util.FileStats),
	}
}

// Load loads the value based on the given key
func (rm *MapStringFileStats) Load(key string) (value *util.FileStats, ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}

// Delete removes the given key from the map
func (rm *MapStringFileStats) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

// Store saves the key/value pair in the map
func (rm *MapStringFileStats) Store(key string, value *util.FileStats) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}

// Clear clears the map
func (rm *MapStringFileStats) Clear() {
	rm.Lock()
	rm.internal = make(map[string]*util.FileStats)
	rm.Unlock()
}

// Size returns the size of the map
func (rm *MapStringFileStats) Size() int {
	return len(rm.internal)
}

// Range returns the size of the map
func (rm *MapStringFileStats) Range() map[string]*util.FileStats {
	return rm.internal
}
