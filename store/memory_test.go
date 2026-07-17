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

func TestScan_Basic(t *testing.T){
	s:=NewMemoryStore()
	s.Set("a",[]byte("1"))
	s.Set("b",[]byte("2"))
	s.Set("c",[]byte("3"))
	s.Set("d",[]byte("4"))
	s.Set("e",[]byte("5"))

	result,err:=s.Scan("b","d")
	if(err!=nil){
		t.Fatalf("Scan failed: %v",err)
	}
	if(len(result)!=3){
		t.Fatalf("expected 3 results, got %d",len(result))
	}
	if(result[0].Key!="b" || result[1].Key!="c" || result[2].Key!="d"){
		t.Fatalf("unexpected keys: %v",result)
	}
}

func TestScan_Empty(t *testing.T){
	s:=NewMemoryStore()
	s.Set("a",[]byte("1"))

	result,err:=s.Scan("x","z")
	if(err!=nil){
		t.Fatalf("Scan failed: %v",err)
	}
	if(len(result)!=0){
		t.Fatalf("expected 0 results, got %d",len(result))
	}
}

func TestScan_SingleKey(t *testing.T){
	s:=NewMemoryStore()
	s.Set("a",[]byte("1"))
	s.Set("b",[]byte("2"))
	s.Set("c",[]byte("3"))

	result,err:=s.Scan("b","b")
	if(err!=nil){
		t.Fatalf("Scan failed: %v",err)
	}
	if(len(result)!=1){
		t.Fatalf("expected 1 result, got %d",len(result))
	}
	if(result[0].Key!="b"){
		t.Fatalf("expected key b, got %s",result[0].Key)
	}
}

func TestClosed_SetAfterClose(t *testing.T){
	s:=NewMemoryStore()
	s.Close()
	err:=s.Set("a",[]byte("1"))
	if(err!=ErrStoreShutDown){
		t.Fatalf("expected ErrStoreShutDown, got %v",err)
	}
}

func TestClosed_GetAfterClose(t *testing.T){
	s:=NewMemoryStore()
	s.Set("a",[]byte("1"))
	s.Close()
	_,err:=s.Get("a")
	if(err!=ErrStoreShutDown){
		t.Fatalf("expected ErrStoreShutDown, got %v",err)
	}
}

func TestClosed_DeleteAfterClose(t *testing.T){
	s:=NewMemoryStore()
	s.Close()
	err:=s.Delete("a")
	if(err!=ErrStoreShutDown){
		t.Fatalf("expected ErrStoreShutDown, got %v",err)
	}
}

func TestClosed_ScanAfterClose(t *testing.T){
	s:=NewMemoryStore()
	s.Close()
	_,err:=s.Scan("a","z")
	if(err!=ErrStoreShutDown){
		t.Fatalf("expected ErrStoreShutDown, got %v",err)
	}
}