package persistence

import (
	"bufio"
	"fmt"
	"os"

	"github.com/abeer-srivastava/redigo/store"
)

func (w *WalStore) Compact() error{
	w.mu.Lock()
	defer w.mu.Unlock()

	if(w.closed){
		return store.ErrStoreShutDown
	}

	snapshot,ok:=w.store.(*store.MemoryStore)
	if(!ok){
		return fmt.Errorf("compact: inner store must be *MemoryStore")
	}

	allKeys,err:=snapshot.Scan("\x00","\xff\xff\xff\xff\xff\xff\xff\xff")
	if(err!=nil){
		return fmt.Errorf("compact scan %w",err)
	}

	compactPath:=w.path+".compact"
	cf,err:=os.OpenFile(compactPath,os.O_RDWR|os.O_CREATE|os.O_TRUNC,0644)
	if(err!=nil){
		return fmt.Errorf("compact create %w",err)
	}

	writer:=bufio.NewWriter(cf)
	for _,kv:=range allKeys{
		entry:=encodeEntry(OpSET,kv.Key,kv.Value)
		if _,err:=writer.Write(entry);err!=nil{
			cf.Close()
			os.Remove(compactPath)
			return fmt.Errorf("compact write %w",err)
		}
	}
	if err:=writer.Flush();err!=nil{
		cf.Close()
		os.Remove(compactPath)
		return fmt.Errorf("compact flush %w",err)
	}
	if err:=cf.Sync();err!=nil{
		cf.Close()
		os.Remove(compactPath)
		return fmt.Errorf("compact sync %w",err)
	}
	if err:=cf.Close();err!=nil{
		os.Remove(compactPath)
		return fmt.Errorf("compact close %w",err)
	}

	if err:=w.file.Sync();err!=nil{
		os.Remove(compactPath)
		return fmt.Errorf("compact wal sync %w",err)
	}
	if err:=w.file.Close();err!=nil{
		os.Remove(compactPath)
		return fmt.Errorf("compact wal close %w",err)
	}

	if err:=os.Rename(compactPath,w.path);err!=nil{
		os.Remove(compactPath)
		return fmt.Errorf("compact rename %w",err)
	}

	nf,err:=os.OpenFile(w.path,os.O_RDWR|os.O_APPEND,0644)
	if(err!=nil){
		return fmt.Errorf("compact reopen %w",err)
	}
	w.file=nf

	return nil
}
