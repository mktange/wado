package syncimpls

import (
	"bytes"
	"sync"
)

// SyncBuffer is a thread-safe bytes.Buffer
type SyncBuffer struct {
	buffer bytes.Buffer
	mutex  *sync.Mutex
}

// NewSyncBuffer creates a new thread-safe bytes.Buffer
func NewSyncBuffer() *SyncBuffer {
	return &SyncBuffer{
		buffer: *bytes.NewBufferString(""),
		mutex:  &sync.Mutex{},
	}
}

func (s *SyncBuffer) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.buffer.Write(p)
}

func (s *SyncBuffer) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.buffer.String()
}

func (s *SyncBuffer) Len() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.buffer.Len()
}
