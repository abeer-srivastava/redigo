package store

import (
	"sync"
)

type MemoryStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func (m *MemoryStore) Set(key string, value []byte) error {
	if key == "" {
		return ErrEmptyKey
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	data:=make([]byte,len(value))
	copy(data,value)
	m.data[key]=data
	return nil
}

func (m *MemoryStore) Get(key string)([]byte,error){
	if(key==""){
		return nil,ErrEmptyKey
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	val,ok:=m.data[key]
	if(!ok){
		return nil,ErrKeyNotFound
	}
	copiedData:=make([]byte,len(val))
	copy(copiedData,val)
	return copiedData,nil
}

func (m *MemoryStore)Delete(key string) error{
	if(key==""){
		return ErrEmptyKey
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data,key)
	return nil	
}

func (m *MemoryStore)Exists(key string) bool{
	if(key==""){
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	_,ok:=m.data[key]
	return ok
}

func (m *MemoryStore)Close() error{
	// TODO
	return nil
}

func NewMemoryStore() Store{
	return &MemoryStore{
		data: make(map[string][]byte),
	}
}