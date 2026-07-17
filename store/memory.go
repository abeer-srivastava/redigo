package store

import (
	"sort"
	"sync"
)

type MemoryStore struct {
	data map[string][]byte
	mu   sync.RWMutex
	closed bool
}

func (m *MemoryStore) Set(key string, value []byte) error {
	if key == "" {
		return ErrEmptyKey
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if(m.closed){
		return ErrStoreShutDown
	}
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
	if(m.closed){
		return nil,ErrStoreShutDown
	}
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
	if(m.closed){
		return ErrStoreShutDown
	}
	delete(m.data,key)
	return nil	
}

func (m *MemoryStore)Exists(key string) bool{
	if(key==""){
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if(m.closed){
		return false
	}
	_,ok:=m.data[key]
	return ok
}

func (m *MemoryStore)Scan(start,end string)([]KeyValue,error){
	m.mu.RLock()
	defer m.mu.RUnlock()
	if(m.closed){
		return nil,ErrStoreShutDown
	}
	keys:=make([]string,0,len(m.data))
	for k:=range m.data{
		if(k>=start && k<=end){
			keys=append(keys,k)
		}
	}
	sort.Strings(keys)
	result:=make([]KeyValue,len(keys))
	for i,k:=range keys{
		val:=make([]byte,len(m.data[k]))
		copy(val,m.data[k])
		result[i]=KeyValue{Key:k,Value:val}
	}
	return result,nil
}

func (m *MemoryStore)Close() error{
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed=true
	return nil
}

func NewMemoryStore() Store{
	return &MemoryStore{
		data: make(map[string][]byte),
	}
}
