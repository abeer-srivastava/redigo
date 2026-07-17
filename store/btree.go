package store

import "sync"

type Btree struct{
	root *BtreeNode
	t int
	mu sync.RWMutex
}

func NewBtree(t int)*Btree{
	return &Btree{
		root: newBtreeNode(t,true),
		t: t,
	}
}

func (b *Btree) Search(key string)([]byte,bool){
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.root.search(key)
}

func (b *Btree) Insert(key string,value []byte){
	if key==""{
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	r:=b.root
	if(r.n==2*b.t-1){
		s:=newBtreeNode(b.t,false)
		s.children=append(s.children,r)
		s.splitChild(b.t,0)
		s.insertNonFull(b.t,key,value)
		b.root=s
		return
	}
	r.insertNonFull(b.t,key,value)
}

func (b *Btree) Delete(key string){
	if key==""{
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	if(b.root.n==0){
		return
	}
	b.root.deleteFromNode(b.t,key)
	if(b.root.n==0 && !b.root.isLeaf){
		b.root=b.root.children[0]
	}
}

func (b *Btree) Scan(start,end string)([]KeyValue){
	b.mu.RLock()
	defer b.mu.RUnlock()
	var result []KeyValue
	b.root.scanRange(start,end,&result)
	return result
}
