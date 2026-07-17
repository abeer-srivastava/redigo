package store

import (
	"sync"
)

type BTreeStore struct{
	tree *Btree
	mu sync.RWMutex
	closed bool
}

func NewBTreeStore(t int) Store{
	return &BTreeStore{
		tree: NewBtree(t),
	}
}

func (b *BTreeStore) Set(key string,value []byte) error{
	if(key==""){
		return ErrEmptyKey
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if(b.closed){
		return ErrStoreShutDown
	}
	copied:=make([]byte,len(value))
	copy(copied,value)
	b.tree.Insert(key,copied)
	return nil
}

func (b *BTreeStore) Get(key string)([]byte,error){
	if(key==""){
		return nil,ErrEmptyKey
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if(b.closed){
		return nil,ErrStoreShutDown
	}
	val,ok:=b.tree.Search(key)
	if(!ok){
		return nil,ErrKeyNotFound
	}
	copied:=make([]byte,len(val))
	copy(copied,val)
	return copied,nil
}

func (b *BTreeStore) Delete(key string) error{
	if(key==""){
		return ErrEmptyKey
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if(b.closed){
		return ErrStoreShutDown
	}
	b.tree.Delete(key)
	return nil
}

func (b *BTreeStore) Exists(key string) bool{
	if(key==""){
		return false
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if(b.closed){
		return false
	}
	_,ok:=b.tree.Search(key)
	return ok
}

func (b *BTreeStore) Scan(start,end string)([]KeyValue,error){
	b.mu.RLock()
	defer b.mu.RUnlock()
	if(b.closed){
		return nil,ErrStoreShutDown
	}
	return b.tree.Scan(start,end),nil
}

func (b *BTreeStore) Close() error{
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed=true
	return nil
}
