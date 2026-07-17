package store

import (
	"fmt"
	"sync"
	"testing"
)

func mustInsert(t *testing.T,b *Btree,key,value string){
	t.Helper()
	b.Insert(key,[]byte(value))
}

func mustSearch(t *testing.T,b *Btree,key,want string){
	t.Helper()
	val,ok:=b.Search(key)
	if(!ok){
		t.Fatalf("key %q not found",key)
	}
	if(string(val)!=want){
		t.Fatalf("key %q: want %q, got %q",key,want,string(val))
	}
}

func mustNotFound(t *testing.T,b *Btree,key string){
	t.Helper()
	_,ok:=b.Search(key)
	if(ok){
		t.Fatalf("key %q should not exist",key)
	}
}

func TestBtree_SingleInsert(t *testing.T){
	b:=NewBtree(3)
	mustInsert(t,b,"a","1")
	mustSearch(t,b,"a","1")
}

func TestBtree_MultipleInsert(t *testing.T){
	b:=NewBtree(3)
	keys:="fedcba"
	for i:=0;i<len(keys);i++{
		mustInsert(t,b,string(keys[i]),fmt.Sprintf("%d",i))
	}

	for i:=0;i<len(keys);i++{
		mustSearch(t,b,string(keys[i]),fmt.Sprintf("%d",i))
	}
}

func TestBtree_InsertTriggersSplit(t *testing.T){
	b:=NewBtree(2)
	for i:=0;i<10;i++{
		mustInsert(t,b,fmt.Sprintf("key%02d",i),fmt.Sprintf("val%02d",i))
	}

	for i:=0;i<10;i++{
		mustSearch(t,b,fmt.Sprintf("key%02d",i),fmt.Sprintf("val%02d",i))
	}
}

func TestBtree_SearchMissing(t *testing.T){
	b:=NewBtree(3)
	mustInsert(t,b,"a","1")
	mustNotFound(t,b,"z")
}

func TestBtree_SearchEmpty(t *testing.T){
	b:=NewBtree(3)
	mustNotFound(t,b,"anything")
}

func TestBtree_DeleteLeaf(t *testing.T){
	b:=NewBtree(3)
	mustInsert(t,b,"a","1")
	mustInsert(t,b,"b","2")
	mustInsert(t,b,"c","3")

	b.Delete("b")
	mustNotFound(t,b,"b")
	mustSearch(t,b,"a","1")
	mustSearch(t,b,"c","3")
}

func TestBtree_DeleteNonLeaf(t *testing.T){
	b:=NewBtree(2)
	for i:=0;i<10;i++{
		mustInsert(t,b,fmt.Sprintf("key%02d",i),fmt.Sprintf("val%02d",i))
	}

	b.Delete("key05")
	mustNotFound(t,b,"key05")

	for i:=0;i<10;i++{
		if(i==5){
			continue
		}
		mustSearch(t,b,fmt.Sprintf("key%02d",i),fmt.Sprintf("val%02d",i))
	}
}

func TestBtree_DeleteAll(t *testing.T){
	b:=NewBtree(2)
	for i:=0;i<10;i++{
		mustInsert(t,b,fmt.Sprintf("key%02d",i),fmt.Sprintf("val%02d",i))
	}

	for i:=0;i<10;i++{
		b.Delete(fmt.Sprintf("key%02d",i))
	}

	for i:=0;i<10;i++{
		mustNotFound(t,b,fmt.Sprintf("key%02d",i))
	}
}

func TestBtree_DeleteMissing(t *testing.T){
	b:=NewBtree(3)
	mustInsert(t,b,"a","1")
	b.Delete("z")
	mustSearch(t,b,"a","1")
}

func TestBtree_DeleteEmpty(t *testing.T){
	b:=NewBtree(3)
	b.Delete("a")
}

func TestBtree_DeleteFromEmpty(t *testing.T){
	b:=NewBtree(3)
	b.Delete("x")
}

func TestBtree_InsertEmptyKey(t *testing.T){
	b:=NewBtree(3)
	b.Insert("",[]byte("value"))
	mustNotFound(t,b,"")
}

func TestBtree_DeleteEmptyKey(t *testing.T){
	b:=NewBtree(3)
	mustInsert(t,b,"a","1")
	b.Delete("")
	mustSearch(t,b,"a","1")
}

func TestBtree_LargeDataset(t *testing.T){
	b:=NewBtree(5)
	n:=500

	for i:=0;i<n;i++{
		mustInsert(t,b,fmt.Sprintf("key%04d",i),fmt.Sprintf("val%04d",i))
	}

	for i:=0;i<n;i++{
		mustSearch(t,b,fmt.Sprintf("key%04d",i),fmt.Sprintf("val%04d",i))
	}

	for i:=0;i<n;i+=2{
		b.Delete(fmt.Sprintf("key%04d",i))
	}

	for i:=0;i<n;i++{
		if(i%2==0){
			mustNotFound(t,b,fmt.Sprintf("key%04d",i))
		}else{
			mustSearch(t,b,fmt.Sprintf("key%04d",i),fmt.Sprintf("val%04d",i))
		}
	}
}

func TestBtree_Update(t *testing.T){
	b:=NewBtree(3)
	mustInsert(t,b,"a","1")
	mustSearch(t,b,"a","1")

	mustInsert(t,b,"a","999")
	mustSearch(t,b,"a","999")
}

func TestBtree_ConcurrentInsert(t *testing.T){
	b:=NewBtree(3)
	var wg sync.WaitGroup
	n:=100

	for i:=0;i<n;i++{
		wg.Add(1)
		go func(i int){
			defer wg.Done()
			key:=fmt.Sprintf("key%03d",i)
			b.Insert(key,[]byte(fmt.Sprintf("val%03d",i)))
		}(i)
	}
	wg.Wait()

	for i:=0;i<n;i++{
		mustSearch(t,b,fmt.Sprintf("key%03d",i),fmt.Sprintf("val%03d",i))
	}
}

func TestBtree_ConcurrentReadWrite(t *testing.T){
	b:=NewBtree(3)
	var wg sync.WaitGroup

	for i:=0;i<50;i++{
		mustInsert(t,b,fmt.Sprintf("key%03d",i),fmt.Sprintf("val%03d",i))
	}

	for i:=0;i<50;i++{
		wg.Add(1)
		go func(i int){
			defer wg.Done()
			key:=fmt.Sprintf("key%03d",i)
			b.Search(key)
			b.Insert(fmt.Sprintf("new%03d",i),[]byte(fmt.Sprintf("newval%03d",i)))
		}(i)
	}
	wg.Wait()

	for i:=0;i<50;i++{
		mustSearch(t,b,fmt.Sprintf("key%03d",i),fmt.Sprintf("val%03d",i))
		mustSearch(t,b,fmt.Sprintf("new%03d",i),fmt.Sprintf("newval%03d",i))
	}
}

func TestBtree_ConcurrentDelete(t *testing.T){
	b:=NewBtree(3)

	for i:=0;i<100;i++{
		mustInsert(t,b,fmt.Sprintf("key%03d",i),fmt.Sprintf("val%03d",i))
	}

	var wg sync.WaitGroup
	for i:=0;i<50;i++{
		wg.Add(1)
		go func(i int){
			defer wg.Done()
			b.Delete(fmt.Sprintf("key%03d",i))
		}(i)
	}
	wg.Wait()

	for i:=0;i<50;i++{
		mustNotFound(t,b,fmt.Sprintf("key%03d",i))
	}
	for i:=50;i<100;i++{
		mustSearch(t,b,fmt.Sprintf("key%03d",i),fmt.Sprintf("val%03d",i))
	}
}

func TestBtree_SortedKeys(t *testing.T){
	b:=NewBtree(3)
	// insert in reverse order
	for i:=99;i>=0;i--{
		mustInsert(t,b,fmt.Sprintf("key%03d",i),fmt.Sprintf("val%03d",i))
	}

	for i:=0;i<100;i++{
		mustSearch(t,b,fmt.Sprintf("key%03d",i),fmt.Sprintf("val%03d",i))
	}
}

func BenchmarkBtree_Insert(b *testing.B){
	tree:=NewBtree(16)
	b.ResetTimer()
	for i:=0;i<b.N;i++{
		tree.Insert(fmt.Sprintf("key%08d",i),[]byte("value"))
	}
}

func BenchmarkBtree_Search(b *testing.B){
	tree:=NewBtree(16)
	n:=10000
	for i:=0;i<n;i++{
		tree.Insert(fmt.Sprintf("key%08d",i),[]byte("value"))
	}
	b.ResetTimer()
	for i:=0;i<b.N;i++{
		tree.Search(fmt.Sprintf("key%08d",i%n))
	}
}

func BenchmarkBtree_Delete(b *testing.B){
	for i:=0;i<b.N;i++{
		b.StopTimer()
		tree:=NewBtree(16)
		for j:=0;j<1000;j++{
			tree.Insert(fmt.Sprintf("key%08d",j),[]byte("value"))
		}
		b.StartTimer()
		for j:=0;j<1000;j++{
			tree.Delete(fmt.Sprintf("key%08d",j))
		}
	}
}
