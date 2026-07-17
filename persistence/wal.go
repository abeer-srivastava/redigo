package persistence

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/abeer-srivastava/redigo/store"
)


const (
	OpSET byte=1
	OpDelete byte=2
)

type WalStore struct{
	store store.Store
	file *os.File
	path string
	mu sync.Mutex
	closed bool
}

func encodeEntry(op byte,key string,value []byte)[]byte{
	// 0p + uint32(len(key))+uint32(len(value))+byte(len(key))+byte(len(value))
	size:=9+len(key)+len(value)
	buf:=make([]byte,size)
	buf[0]=op
	binary.BigEndian.PutUint32(
		buf[1:5],
		uint32(len(key)),
	)
	binary.BigEndian.PutUint32(
		buf[5:9],
		uint32(len(value)),
	)
	copy(buf[9:],key)
	copy(buf[9+len(key):],value)
	return buf
}

func (w *WalStore)Set(key string,value []byte) error{
	if(key==""){
		return store.ErrEmptyKey
	}
	entry:=encodeEntry(OpSET,key,value)
	w.mu.Lock()
	defer w.mu.Unlock()
	if(w.closed){
		return store.ErrStoreShutDown
	}
	if _,err:=w.file.Write(entry);err!=nil{
		return fmt.Errorf("wal write %w",err)
	}
	if err:=w.file.Sync();err!=nil{
		return fmt.Errorf("wal sync %w",err)
	}
	return w.store.Set(key,value)
}

func (w *WalStore)Get(key string)([]byte ,error){
	return w.store.Get(key)
}

func (w *WalStore)Exists(key string) bool{
	return w.store.Exists(key)
}

func (w *WalStore)Scan(start,end string)([]store.KeyValue,error){
	return w.store.Scan(start,end)
}

func (w *WalStore)Delete(key string) error{
	if(key==""){
		return store.ErrEmptyKey
	}
	entry:=encodeEntry(OpDelete,key,nil)
	w.mu.Lock()
	defer w.mu.Unlock()
	if(w.closed){
		return store.ErrStoreShutDown
	}
	if _,err:=w.file.Write(entry);err!=nil{
		return fmt.Errorf("wal write %w",err)
	}
	if err:=w.file.Sync();err!=nil{
		return fmt.Errorf("wal sync %w",err)
	}
	return w.store.Delete(key)
}

func (w *WalStore)Close()error{
	w.mu.Lock()
	defer w.mu.Unlock()
	if(w.closed){
		return nil //already closed
	}
	w.closed=true
	if err:=w.file.Sync();err!=nil{
		return fmt.Errorf("wal closing sync %w",err)
	}
	if err:=w.file.Close();err!=nil{
		return fmt.Errorf("wal closing %w",err)
	}
	return nil
}


func NewWalStore(path string , inner store.Store) (*WalStore,error){
	f,err:=os.OpenFile(path,os.O_RDWR | os.O_CREATE | os.O_APPEND,0644)
	if(err!=nil){
		return nil,fmt.Errorf("failed to open file %w",err)
	}
	defer func(){
		if(err!=nil){
			f.Close()
		}
	}()
	if _,err=f.Seek(0,io.SeekStart);err!=nil{ 
		return nil,fmt.Errorf("wal seek to start %w",err)
	}
	var offset int64=0
	reader:=bufio.NewReader(f)
	for {
		header:=make([]byte,9)
		_,err=io.ReadFull(reader,header)
		if(err==io.EOF){
			err=nil
			break
		}
		if(err!=nil){
			// partial header truncate to last good offset
			if tErr:=f.Truncate(offset);tErr!=nil{
				err=tErr
				return nil,fmt.Errorf("wal truncate %w",tErr)
			}
			err=nil
			break
		}
		// parsing header
		op:=header[0]
		keyLen:=binary.BigEndian.Uint32(header[1:5])
		valLen:=binary.BigEndian.Uint32(header[5:9])
		// reading key 
		keyBuf:=make([]byte,keyLen)
		if _,err=io.ReadFull(reader,keyBuf);err!=nil{
			if tErr:=f.Truncate(offset);tErr!=nil{
				err=tErr
				return nil,fmt.Errorf("wal truncate %w",tErr)
			}
			err=nil
			break
		}
		// reader value
		valBuf:=make([]byte,valLen)
		if(valLen>0){
			if _,err=io.ReadFull(reader,valBuf);err!=nil{
				if tErr:=f.Truncate(offset);tErr!=nil{
					err=tErr
					return nil,fmt.Errorf("wal truncate %w",tErr)
				}
				err=nil
				break
			}
		}
		switch op{
			case OpSET:
				if err=inner.Set(string(keyBuf),valBuf);err!=nil{
					return nil,fmt.Errorf("replay set failed %w",err)
				}
			case OpDelete:
				inner.Delete(string(keyBuf))
		}
		offset+=9+int64(keyLen)+int64(valLen)
		
	}
	
	return &WalStore{
		store:inner,
		file:f,
		path:path,
	},nil
}