package store

import (
	"sync"

	"github.com/abeer-srivastava/redigo/store"
)

type MemoryStore struct{
	data map[string][]byte
	mu sync.RWMutex
}
func (m *MemoryStore) Set(key string,value []byte) error{
	if(key==""){
		return store.ErrEmptyKey
	}
}