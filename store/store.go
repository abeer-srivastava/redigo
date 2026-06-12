package store

import "errors"



var (
	ErrKeyNotFound=errors.New("Key Not Found")
	ErrEmptyKey=errors.New("Key Cannot Be Empty")
	ErrStoreShutDown=errors.New("Store Shutdown")
)


type Store interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	Exist(key string) bool
	Close() error
}


