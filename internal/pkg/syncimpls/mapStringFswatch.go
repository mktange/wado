package syncimpls

import (
	"sync"

	"github.com/fsnotify/fsnotify"
)

// MapStringFsWatch is a concurrent imlementation of map[string]*fsnotify.Watcher
type MapStringFsWatch struct {
	sync.RWMutex
	internal map[string]*fsnotify.Watcher
}

// NewMapStringFsWatch is a concurrent imlementation of map[string]*fsnotify.Watcher
func NewMapStringFsWatch() *MapStringFsWatch {
	return &MapStringFsWatch{
		internal: make(map[string]*fsnotify.Watcher),
	}
}

// Load loads the value based on the given key
func (rm *MapStringFsWatch) Load(key string) (value *fsnotify.Watcher, ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}

// Delete removes the given key from the map
func (rm *MapStringFsWatch) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

// Store saves the key/value pair in the map
func (rm *MapStringFsWatch) Store(key string, value *fsnotify.Watcher) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}

// ClearAndClose closes all file watches and clears the map
func (rm *MapStringFsWatch) ClearAndClose() error {
	rm.Lock()
	for _, fsWatch := range rm.internal {
		err := fsWatch.Close()
		if err != nil {
			return err
		}
	}
	rm.internal = make(map[string]*fsnotify.Watcher)
	rm.Unlock()
	return nil
}

// Size returns the size of the map
func (rm *MapStringFsWatch) Size() int {
	return len(rm.internal)
}
