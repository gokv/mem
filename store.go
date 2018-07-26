package mem

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gokv/store"
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
	m  map[interface{}]entry
}

// New initialises the map underlying Store.
func New() *Store {
	return &Store{
		m: make(map[interface{}]entry),
	}
}

// Get returns the value corresponding the key, and a nil error.
// If no match is found, returns (false, nil).
func (s *Store) Get(ctx context.Context, key interface{}, v json.Unmarshaler) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.m[key]
	if !ok || !e.validAt(time.Now()) {
		return false, nil
	}

	return true, v.UnmarshalJSON(e.data)
}

// GetAll returns all values. Error is non-nil if the context is Done.
func (s *Store) GetAll(ctx context.Context, key interface{}, c store.Collection) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()

	for _, e := range s.m {
		if e.validAt(now) {
			if err := c.New().UnmarshalJSON(e.data); err != nil {
				return err
			}
		}
	}

	return nil
}

// Add assigns the given value to the given key if it doesn't exist already.
// Err is non-nil if key was already present, or in case of failure.
func (s *Store) Add(ctx context.Context, key interface{}, v json.Marshaler) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	b, err := v.MarshalJSON()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, ok := s.m[key]; ok {
		return store.ErrKeyExists
	}

	s.m[key] = entry{data: b}
	return nil
}

// Set assigns the given value to the given key, possibly overwriting.
// The returned error is not nil if the context is Done.
func (s *Store) Set(ctx context.Context, key interface{}, v json.Marshaler) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	b, err := v.MarshalJSON()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.m[key] = entry{data: b}
	return nil
}

// SetWithTimeout assigns the given value to the given key, possibly
// overwriting.
// The assigned key will clear after timeout. The lifespan starts when this
// function is called.
func (s *Store) SetWithTimeout(ctx context.Context, key interface{}, v json.Marshaler, timeout time.Duration) error {
	return s.SetWithDeadline(ctx, key, v, time.Now().Add(timeout))
}

// SetWithDeadline assigns the given value to the given key, possibly
// overwriting.
// The assigned key will clear after deadline.
func (s *Store) SetWithDeadline(ctx context.Context, key interface{}, v json.Marshaler, deadline time.Time) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	b, err := v.MarshalJSON()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.m[key] = entry{data: b, validTo: deadline.UnixNano()}
	return nil
}

// Delete removes the corresponding entry if present.
// Returns a non-nil error if the key is not known or if the context is Done.
func (s *Store) Delete(ctx context.Context, key interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, ok := s.m[key]; !ok {
		return store.ErrNoRows
	}

	delete(s.m, key)
	return nil
}
