package mem

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gokv/store"
	"github.com/google/uuid"
)

const (
	cleanupInterval = time.Second
	cleanupTimeout  = time.Millisecond
)

// ErrKeyExists is returned when the Add method generates a non-unique ID.
var ErrKeyExists = errors.New("the key already exists")

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

	close func()
}

// New initialises the map underlying Store.
func New() *Store {
	s := &Store{
		m: make(map[interface{}]entry),
	}
	s.close = start(s.Cleanup, cleanupTimeout, cleanupInterval)
	return s
}

// Get returns the value corresponding the key, and a nil error.
// If no match is found, returns (false, nil).
func (s *Store) Get(ctx context.Context, k string, v json.Unmarshaler) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.m[k]
	if !ok || !e.validAt(time.Now()) {
		return false, nil
	}

	return true, v.UnmarshalJSON(e.data)
}

// GetAll returns all values. Error is non-nil if the context is Done.
func (s *Store) GetAll(ctx context.Context, k string, c store.Collection) error {
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

// Add persists a new object and returns its unique UUIDv4 key.
// Err is non-nil in case of failure.
func (s *Store) Add(ctx context.Context, v json.Marshaler) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	b, err := v.MarshalJSON()
	if err != nil {
		return "", err
	}

	k := uuid.New().String()

	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if _, ok := s.m[k]; ok {
		return "", ErrKeyExists
	}

	s.m[k] = entry{data: b}
	return k, nil
}

// Set assigns the given value to the given key, possibly overwriting.
// The returned error is not nil if the context is Done.
func (s *Store) Set(ctx context.Context, k string, v json.Marshaler) error {
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

	s.m[k] = entry{data: b}
	return nil
}

// SetWithTimeout assigns the given value to the given key, possibly
// overwriting.
// The assigned key will clear after timeout. The lifespan starts when this
// function is called.
func (s *Store) SetWithTimeout(ctx context.Context, k string, v json.Marshaler, timeout time.Duration) error {
	return s.SetWithDeadline(ctx, k, v, time.Now().Add(timeout))
}

// SetWithDeadline assigns the given value to the given key, possibly
// overwriting.
// The assigned key will clear after deadline.
func (s *Store) SetWithDeadline(ctx context.Context, k string, v json.Marshaler, deadline time.Time) error {
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

	s.m[k] = entry{data: b, validTo: deadline.UnixNano()}
	return nil
}

// Delete removes the corresponding entry if present.
// Returns a non-nil error if the key is not known or if the context is Done.
func (s *Store) Delete(ctx context.Context, k string) error {
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

	if _, ok := s.m[k]; !ok {
		return store.ErrNoRows
	}

	delete(s.m, k)
	return nil
}

// Close releases the resources associated with the Store.
func (s *Store) Close() error {
	s.close()
	return nil
}
