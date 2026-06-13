package ttl

import (
	"errors"
	"testing"
	"time"

	"github.com/abeer-srivastava/redigo/store"
)
func mustSetTTL(
	t *testing.T,
	s *TTLStore,
	key, value string,
	ttl time.Duration,
) {
	t.Helper()

	if err := s.SetWithTtl(key, []byte(value), ttl); err != nil {
		t.Fatalf(
			"SetWithTTL(%q,%q) failed: %v",
			key,
			value,
			err,
		)
	}
}

func mustGet(
	t *testing.T,
	s *TTLStore,
	key,
	want string,
) {
	t.Helper()

	val, err := s.Get(key)
	if err != nil {
		t.Fatalf("Get(%q) failed: %v", key, err)
	}

	if string(val) != want {
		t.Fatalf(
			"Get(%q)=%q want=%q",
			key,
			string(val),
			want,
		)
	}
}

func mustExpired(
	t *testing.T,
	s *TTLStore,
	key string,
) {
	t.Helper()

	_, err := s.Get(key)

	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf(
			"expected %q to be expired, got %v",
			key,
			err,
		)
	}
}
func TestTTL_ExpiresAfterTTL(t *testing.T) {
	s := NewTTLStore(
		store.NewMemoryStore(),
		10*time.Millisecond,
	)
	defer s.Close()

	mustSetTTL(
		t,
		s,
		"otp",
		"889923",
		50*time.Millisecond,
	)
	time.Sleep(100 * time.Millisecond)
	mustExpired(t, s, "otp")
}
func TestTTL_AccessibleBeforeExpiry(t *testing.T) {
	s := NewTTLStore(
		store.NewMemoryStore(),
		10*time.Millisecond,
	)
	defer s.Close()

	mustSetTTL(
		t,
		s,
		"token",
		"abc123",
		500*time.Millisecond,
	)

	mustGet(
		t,
		s,
		"token",
		"abc123",
	)
}
func TestTTL_PlainSetClearsTTL(t *testing.T) {
	s := NewTTLStore(
		store.NewMemoryStore(),
		10*time.Millisecond,
	)
	defer s.Close()

	mustSetTTL(
		t,
		s,
		"x",
		"first",
		50*time.Millisecond,
	)

	if err := s.Set(
		"x",
		[]byte("second"),
	); err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	mustGet(
		t,
		s,
		"x",
		"second",
	)
}
func TestTTL_DeleteClearsTTL(t *testing.T) {
	s := NewTTLStore(
		store.NewMemoryStore(),
		10*time.Millisecond,
	)
	defer s.Close()

	mustSetTTL(
		t,
		s,
		"temp",
		"data",
		500*time.Millisecond,
	)

	if err := s.Delete("temp"); err != nil {
		t.Fatal(err)
	}

	_, err := s.Get("temp")

	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf(
			"expected ErrKeyNotFound, got %v",
			err,
		)
	}
}
func TestTTL_SweepCleansExpiredKeys(t *testing.T) {
	s := NewTTLStore(
		store.NewMemoryStore(),
		10*time.Millisecond,
	)
	defer s.Close()

	mustSetTTL(
		t,
		s,
		"ghost",
		"boo",
		20*time.Millisecond,
	)

	time.Sleep(100 * time.Millisecond)

	mustExpired(
		t,
		s,
		"ghost",
	)
}
func TestTTL_CloseStopsGoroutine(t *testing.T) {
	s := NewTTLStore(
		store.NewMemoryStore(),
		10*time.Millisecond,
	)

	done := make(chan struct{})

	go func() {
		defer close(done)

		_ = s.Close()
	}()

	select {
	case <-done:
		// success

	case <-time.After(1 * time.Second):
		t.Fatal(
			"Close() timed out; goroutine likely leaked",
		)
	}
}
func TestTTL_OverwriteTTL(t *testing.T) {
	s := NewTTLStore(
		store.NewMemoryStore(),
		10*time.Millisecond,
	)
	defer s.Close()

	mustSetTTL(
		t,
		s,
		"x",
		"v1",
		50*time.Millisecond,
	)

	time.Sleep(30 * time.Millisecond)

	mustSetTTL(
		t,
		s,
		"x",
		"v2",
		500*time.Millisecond,
	)

	time.Sleep(50 * time.Millisecond)

	mustGet(
		t,
		s,
		"x",
		"v2",
	)
}