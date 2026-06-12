package store

import (
	"bytes"
	"sync"
	"testing"
)

func TestSetGet(t *testing.T) {
	store := NewMemoryStore()

	err := store.Set("key", []byte("test"))
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := store.Get("key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !bytes.Equal(val, []byte("test")) {
		t.Fatalf("expected test, got %s", string(val))
	}
}
func TestGetMissingKey(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.Get("missing")

	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}
func TestDelete(t *testing.T) {
	store := NewMemoryStore()

	err := store.Set("key", []byte("test"))
	if err != nil {
		t.Fatal(err)
	}

	err = store.Delete("key")
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Get("key")

	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound after delete, got %v", err)
	}
}
func TestOverwrite(t *testing.T) {
	store := NewMemoryStore()

	err := store.Set("key", []byte("first"))
	if err != nil {
		t.Fatal(err)
	}

	err = store.Set("key", []byte("second"))
	if err != nil {
		t.Fatal(err)
	}

	val, err := store.Get("key")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(val, []byte("second")) {
		t.Fatalf("expected second, got %s", string(val))
	}
}
func TestConcurrentSet(t *testing.T) {
	store := NewMemoryStore()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			key := "key"
			value := []byte("value")

			if err := store.Set(key, value); err != nil {
				t.Errorf("set failed: %v", err)
			}
		}(i)
	}

	wg.Wait()
}
func TestConcurrentReadWrite(t *testing.T) {
	store := NewMemoryStore()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(2)

		go func(i int) {
			defer wg.Done()
			store.Set("key", []byte("value"))
		}(i)

		go func() {
			defer wg.Done()
			store.Get("key")
		}()
	}

	wg.Wait()
}