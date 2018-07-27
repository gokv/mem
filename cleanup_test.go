package mem

import (
	"context"
	"errors"
	"testing"
	"time"
)

type value string

func (s *value) UnmarshalJSON(data []byte) error {
	*s = value(data)
	return nil
}

func (s value) MarshalJSON() (data []byte, err error) {
	return []byte(s), nil
}

func TestCleanup(t *testing.T) {
	s := New()
	defer s.Close()

	key := "key"

	d := time.Nanosecond
	s.SetWithTimeout(context.Background(), key, value("wazzup"), d)
	time.Sleep(d)
	if _, ok := s.m[key]; !ok {
		panic(errors.New("expected the value to still be present after short delay"))
	}

	time.Sleep(time.Millisecond * 1001)
	if _, ok := s.m[key]; ok {
		t.Error("expected the value to be garbage collected")
	}
}
