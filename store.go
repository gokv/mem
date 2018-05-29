package mem

import (
	"encoding"
	"sync"
	"time"
)

type entry struct {
	data    []byte
	validTo int64
}

func (e *entry) validAt(t time.Time) bool {
	if e.validTo != 0 && t.UnixNano() > e.validTo {
		return false
	}
	return true
}

// Store implements an in-memory key-value store.
// It is implemented as a Go map and protected by a mutex.
// The zero value is not ready to use: initialise with New.
//
// Store is safe for concurrent use.
type Store struct {
	mu sync.RWMutex
	m  map[string]entry
}

// New initialises the map inderlying Store.
func New() *Store {
	return &Store{
		m: make(map[string]entry),
	}
}

// Length returns the current size of the store, including the expired values.
func (s *Store) Length() int {
	s.mu.RLock()
	s.mu.RUnlock()

	return len(s.m)
}

// Del deletes the corresponding entry if present, and returns nil.
func (s *Store) Del(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.m, key)
	return nil
}

// Get returns the value corresponding the key, and a nil error.
// If no match is found, returns (false, nil).
func (s *Store) Get(key string, v encoding.BinaryUnmarshaler) (bool, error) {
	s.mu.RLock()
	s.mu.RUnlock()

	e, ok := s.m[key]
	if !ok || !e.validAt(time.Now()) {
		return false, nil
	}

	return true, v.UnmarshalBinary(e.data)
}

// Set assigns the given value to the given key, possibly overwriting.
func (s *Store) Set(key string, v encoding.BinaryMarshaler) error {
	b, err := v.MarshalBinary()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[key] = entry{data: b}
	return nil
}

// SetWithTimeout assigns the given value to the given key, possibly
// overwriting.
// The assigned key will clear after timeout. The lifespan starts when this
// function is called.
func (s *Store) SetWithTimeout(key string, v encoding.BinaryMarshaler, timeout time.Duration) error {
	return s.SetWithDeadline(key, v, time.Now().Add(timeout))
}

// SetWithDeadline assigns the given value to the given key, possibly
// overwriting.
// The assigned key will clear after deadline.
func (s *Store) SetWithDeadline(key string, v encoding.BinaryMarshaler, deadline time.Time) error {
	b, err := v.MarshalBinary()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[key] = entry{data: b, validTo: deadline.UnixNano()}
	return nil
}
