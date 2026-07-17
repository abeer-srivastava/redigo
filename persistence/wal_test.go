package persistence

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/abeer-srivastava/redigo/store"
)

func mustSet(t *testing.T, s *WalStore, key, value string) {
	t.Helper()

	if err := s.Set(key, []byte(value)); err != nil {
		t.Fatalf("Set(%q,%q) failed: %v", key, value, err)
	}
}

func mustGet(t *testing.T, s *WalStore, key, want string) {
	t.Helper()

	val, err := s.Get(key)
	if err != nil {
		t.Fatalf("Get(%q) failed: %v", key, err)
	}

	if string(val) != want {
		t.Fatalf("Get(%q)=%q want=%q", key, string(val), want)
	}
}

func TestWal_BasicOperation(t *testing.T){
	path:=filepath.Join(t.TempDir(),"test_wal")
	s,err:=NewWalStore(path,store.NewMemoryStore())
	if err!=nil{
		t.Fatalf("failed creating wal store %v",err)
	}
	mustSet(t,s,"name","test_name")
	mustGet(t,s,"name","test_name")

	mustSet(t,s,"name","test_name_overwrite")
	mustGet(t,s,"name","test_name_overwrite")

	if err:=s.Delete("name");err!=nil{
		t.Fatalf("Delete failed %v",err)
	}
	_,err=s.Get("name")
	if(!errors.Is(err,store.ErrKeyNotFound)){
		t.Fatalf("expected ErrKeyNotFound,got %v",err)
	}
	_,err=s.Get("name")
	if(!errors.Is(err,store.ErrKeyNotFound)){
		t.Fatalf("expected ErrKeyNotFound,got %v",err)
	}
	err=s.Set("",[]byte("testingEmptyKey"))
	if(!errors.Is(err,store.ErrEmptyKey)){
		t.Fatalf("expected ErrEmptyKey , got %v",err)
	}
	
}

func TestWAL_PersistsAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wal")

	// ---------- write phase ----------
	s1, err := NewWalStore(path, store.NewMemoryStore())
	if err != nil {
		t.Fatalf("failed creating store: %v", err)
	}

	mustSet(t, s1, "lang", "go")
	mustSet(t, s1, "version", "1.21")
	mustSet(t, s1, "author", "abeer")

	if err := s1.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// ---------- restart ----------
	s2, err := NewWalStore(path, store.NewMemoryStore())
	if err != nil {
		t.Fatalf("failed reopening store: %v", err)
	}

	mustGet(t, s2, "lang", "go")
	mustGet(t, s2, "version", "1.21")
	mustGet(t, s2, "author", "abeer")
}
func TestWAL_OverwritePersistsAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wal")

	s1, err := NewWalStore(path, store.NewMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	mustSet(t, s1, "x", "first")
	mustSet(t, s1, "x", "second")

	if err := s1.Close(); err != nil {
		t.Fatal(err)
	}

	s2, err := NewWalStore(path, store.NewMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	mustGet(t, s2, "x", "second")
}
func TestWAL_DeletePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wal")

	s1, err := NewWalStore(path, store.NewMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	mustSet(t, s1, "temp", "data")

	if err := s1.Delete("temp"); err != nil {
		t.Fatal(err)
	}

	if err := s1.Close(); err != nil {
		t.Fatal(err)
	}

	s2, err := NewWalStore(path, store.NewMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	_, err = s2.Get("temp")

	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}
func TestWAL_CorruptedTailRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wal")

	// write valid data
	s1, err := NewWalStore(path, store.NewMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	mustSet(t, s1, "good", "data")

	if err := s1.Close(); err != nil {
		t.Fatal(err)
	}

	// manually append corruption
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Write([]byte{
		0x01,
		0x00,
		0x00,
	})
	if err != nil {
		t.Fatal(err)
	}

	f.Close()

	// recovery should succeed
	s2, err := NewWalStore(path, store.NewMemoryStore())
	if err != nil {
		t.Fatalf("expected successful recovery, got %v", err)
	}

	mustGet(t, s2, "good", "data")
}

func TestWAL_Compact_ShrinksFile(t *testing.T){
	path:=filepath.Join(t.TempDir(),"compact.wal")
	s,err:=NewWalStore(path,store.NewMemoryStore())
	if(err!=nil){
		t.Fatal(err)
	}

	// write 100 overwrites to key "x"
	for i:=0;i<100;i++{
		mustSet(t,s,"x",string(rune('a'+i%26)))
	}

	before,_:=os.Stat(path)
	if err:=s.Compact();err!=nil{
		t.Fatalf("Compact failed: %v",err)
	}
	after,_:=os.Stat(path)

	if(after.Size()>=before.Size()){
		t.Fatalf("expected compaction to shrink file: before=%d after=%d",before.Size(),after.Size())
	}
}

func TestWAL_Compact_PreservesData(t *testing.T){
	path:=filepath.Join(t.TempDir(),"compact.wal")
	s,err:=NewWalStore(path,store.NewMemoryStore())
	if(err!=nil){
		t.Fatal(err)
	}

	mustSet(t,s,"a","1")
	mustSet(t,s,"b","2")
	mustSet(t,s,"c","3")

	if err:=s.Compact();err!=nil{
		t.Fatalf("Compact failed: %v",err)
	}

	mustGet(t,s,"a","1")
	mustGet(t,s,"b","2")
	mustGet(t,s,"c","3")
}

func TestWAL_Compact_WorksAfterRestart(t *testing.T){
	path:=filepath.Join(t.TempDir(),"compact.wal")

	// phase 1: write data
	s1,err:=NewWalStore(path,store.NewMemoryStore())
	if(err!=nil){
		t.Fatal(err)
	}
	for i:=0;i<50;i++{
		mustSet(t,s1,"key",string(rune('a'+i%26)))
	}

	// phase 2: compact
	if err:=s1.Compact();err!=nil{
		t.Fatalf("Compact failed: %v",err)
	}

	// phase 3: write more after compact
	mustSet(t,s1,"key","final")
	mustSet(t,s1,"after","compact")

	if err:=s1.Close();err!=nil{
		t.Fatal(err)
	}

	// phase 4: restart and verify
	s2,err:=NewWalStore(path,store.NewMemoryStore())
	if(err!=nil){
		t.Fatal(err)
	}
	mustGet(t,s2,"key","final")
	mustGet(t,s2,"after","compact")
}

func TestWAL_Compact_DeletePreserved(t *testing.T){
	path:=filepath.Join(t.TempDir(),"compact.wal")
	s,err:=NewWalStore(path,store.NewMemoryStore())
	if(err!=nil){
		t.Fatal(err)
	}

	mustSet(t,s,"temp","data")
	s.Delete("temp")

	if err:=s.Compact();err!=nil{
		t.Fatalf("Compact failed: %v",err)
	}

	_,err=s.Get("temp")
	if(!errors.Is(err,store.ErrKeyNotFound)){
		t.Fatalf("expected ErrKeyNotFound after compact, got %v",err)
	}
}

func TestWAL_Compact_ClosedStore(t *testing.T){
	path:=filepath.Join(t.TempDir(),"compact.wal")
	s,err:=NewWalStore(path,store.NewMemoryStore())
	if(err!=nil){
		t.Fatal(err)
	}
	s.Close()

	err=s.Compact()
	if(!errors.Is(err,store.ErrStoreShutDown)){
		t.Fatalf("expected ErrStoreShutDown, got %v",err)
	}
}

func TestWAL_Compact_ContinuesWorking(t *testing.T){
	path:=filepath.Join(t.TempDir(),"compact.wal")
	s,err:=NewWalStore(path,store.NewMemoryStore())
	if(err!=nil){
		t.Fatal(err)
	}

	// compact multiple times
	for i:=0;i<3;i++{
		mustSet(t,s,"x","iter"+string(rune('0'+i)))
		if err:=s.Compact();err!=nil{
			t.Fatalf("Compact iter %d failed: %v",i,err)
		}
	}

	mustGet(t,s,"x","iter2")
	mustSet(t,s,"y","after")
	mustGet(t,s,"y","after")
}