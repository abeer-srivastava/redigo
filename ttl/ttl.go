package ttl

import (
	"fmt"
	"sync"
	"time"

	"github.com/abeer-srivastava/redigo/store"
)

type TTLStore struct{
	inner store.Store
	expires map[string]time.Time
	mu sync.RWMutex
	done chan struct{}
	wg sync.WaitGroup
	sweepInterval time.Duration
}

func NewTTLStore(inner store.Store,sweepInterval time.Duration) *TTLStore{
	t:=&TTLStore{
		inner:inner,
		expires: make(map[string]time.Time),
		done:make(chan struct{}),
		sweepInterval: sweepInterval,
	}
	t.wg.Add(1)
	go func(){
		defer t.wg.Done()
		for {
			select{
				case <-t.done:
					return
			 	case <-time.After(t.sweepInterval):
					t.sweep()
			}
		}
	}()
	return t
}

func (t *TTLStore) sweep(){
// collecting the expired keys 
	t.mu.RLock()
	expiredKeys:=make([]string,0,len(t.expires))
	for key,deadline:=range t.expires{
		if(time.Now().After(deadline)){
			expiredKeys=append(expiredKeys, key)
		}
	}
	t.mu.RUnlock()
	for _,key:=range expiredKeys{
		t.mu.Lock()
		var shouldDelete bool
		deadline,ok:=t.expires[key]
		if(ok && time.Now().After(deadline)){
			delete(t.expires,key)
			shouldDelete=true
		}
		t.mu.Unlock()
		if(shouldDelete){
			t.inner.Delete(key)
		}
	}
}

func (t *TTLStore) SetWithTtl(key string,value []byte,ttl time.Duration) error{
	if(key==""){
		return  store.ErrEmptyKey
	}
	if err:=t.inner.Set(key,value);err!=nil{
		return fmt.Errorf("ttl set failed : %w",err)
	}
	t.mu.Lock()
	t.expires[key]=time.Now().Add(ttl)
	t.mu.Unlock()
	return nil
}

func (t *TTLStore) Set(key string ,value []byte) error{
	if(key==""){
		return store.ErrEmptyKey
	}
	t.mu.Lock()
	delete(t.expires,key)
	t.mu.Unlock()
	return t.inner.Set(key,value)
}

func (t *TTLStore) Get(key string) ([]byte,error){
	if(key==""){
		return nil,store.ErrEmptyKey
	}
	t.mu.RLock()
	deadline,ok:=t.expires[key]
	t.mu.RUnlock()
	if ok && time.Now().After(deadline){
		t.mu.Lock()
		deadline,ok=t.expires[key]
		if(ok && time.Now().After(deadline)){
			delete(t.expires,key)
			t.mu.Unlock()
			_=t.inner.Delete(key)
			return nil,store.ErrKeyNotFound
		}
		t.mu.Unlock()
	}
	return t.inner.Get(key)
}
func (t *TTLStore) Delete(key string) error{
	if(key==""){
		return store.ErrEmptyKey
	}
	t.mu.Lock()
	delete(t.expires,key)
	t.mu.Unlock()
	return t.inner.Delete(key)
}

func (t *TTLStore) Exists(key string)bool{
	if(key==""){
		return false
	}
	t.mu.RLock()
	deadline,ok:=t.expires[key]
	t.mu.RUnlock()
	if(ok && time.Now().After(deadline)){
		t.mu.Lock()
		deadline,ok=t.expires[key]
		if(ok && time.Now().After(deadline)){
			delete(t.expires,key)
			t.mu.Unlock()
			return false
		}
		t.mu.Unlock()
	}
	return t.inner.Exists(key)
}

func (t *TTLStore) Scan(start,end string)([]store.KeyValue,error){
	return t.inner.Scan(start,end)
}

func (t *TTLStore)Close()error{
	close(t.done)
	t.wg.Wait()
	return t.inner.Close()
}
