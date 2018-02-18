package syncimpls

import "sync"

// MapStringString is a concurrent imlementation of map[string]string
type MapStringString struct {
	sync.RWMutex
	internal map[string]string
}

// NewMapStringString is a concurrent imlementation of map[string]string
func NewMapStringString() *MapStringString {
	return &MapStringString{
		internal: make(map[string]string),
	}
}

// Load loads the value based on the given key
func (rm *MapStringString) Load(key string) (value string, ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}

// Delete removes the given key from the map
func (rm *MapStringString) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

// Store saves the key/value pair in the map
func (rm *MapStringString) Store(key, value string) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}

// Clear clears the map
func (rm *MapStringString) Clear() {
	rm.Lock()
	rm.internal = make(map[string]string)
	rm.Unlock()
}

// Size returns the size of the map
func (rm *MapStringString) Size() int {
	return len(rm.internal)
}
