package store


type BtreeNode struct {
	keys     []string
	values   [][]byte
	children []*BtreeNode
	isLeaf   bool
	n        int
}

func newBtreeNode(t int,isLeaf bool) *BtreeNode{
	return &BtreeNode{
		keys:make([]string,0,2*t-1),
		values: make([][]byte,0,2*t-1),
		children: make([]*BtreeNode,0,2*t),
		isLeaf: isLeaf,
		n: 0,
	}
}

func (node *BtreeNode) FindKeyIndex(key string) int{
	low:=0;
	high:=node.n;
	for low<high{
		mid:=low+high/2;
		if(node.keys[mid]==key){
			return mid;
		}
		if node.keys[mid]<key {
			low=mid+1;
		}else{
			high=mid;
		}
	}
	return low
}

func (node *BtreeNode) splitChild(t int,i int){
	y:=node.children[i]
	z:=newBtreeNode(t,y.isLeaf)
	z.n=t-1

	for j:=0;j<t-1;j++{
		z.keys=append(z.keys,y.keys[j+t])
		z.values=append(z.values,y.values[j+t])
	}

	if(!y.isLeaf){
		for j:=0;j<t;j++{
			z.children=append(z.children,y.children[j+t])
		}
	}

	node.keys=append(node.keys,"")
	copy(node.keys[i+1:],node.keys[i:])
	node.keys[i]=y.keys[t-1]

	node.values=append(node.values,nil)
	copy(node.values[i+1:],node.values[i:])
	node.values[i]=y.values[t-1]

	node.children=append(node.children,nil)
	copy(node.children[i+2:],node.children[i+1:])
	node.children[i+1]=z

	y.keys=y.keys[:t-1]
	y.values=y.values[:t-1]
	if(!y.isLeaf){
		y.children=y.children[:t]
	}
	y.n=t-1

	node.n++
}

func (node *BtreeNode) insertNonFull(t int,key string,value []byte){
	i:=node.n-1

	if(node.isLeaf){
		for j:=0;j<node.n;j++{
			if(node.keys[j]==key){
				node.values[j]=value
				return
			}
		}
		node.keys=append(node.keys,"")
		node.values=append(node.values,nil)
		for i>=0 && key<node.keys[i]{
			node.keys[i+1]=node.keys[i]
			node.values[i+1]=node.values[i]
			i--
		}
		node.keys[i+1]=key
		node.values[i+1]=value
		node.n++
		return
	}

	for i>=0 && key<node.keys[i]{
		i--
	}
	i++

	if(node.children[i].n==2*t-1){
		node.splitChild(t,i)
		if(key==node.keys[i]){
			node.values[i]=value
			return
		}
		if(key>node.keys[i]){
			i++
		}
	}
	node.children[i].insertNonFull(t,key,value)
}

func (node *BtreeNode) search(key string)([]byte,bool){
	i:=0
	for i<node.n && key>node.keys[i]{
		i++
	}

	if(i<node.n && key==node.keys[i]){
		return node.values[i],true
	}

	if(node.isLeaf){
		return nil,false
	}

	return node.children[i].search(key)
}

func (node *BtreeNode) scanRange(start,end string,result *[]KeyValue){
	if(node.isLeaf){
		for i:=0;i<node.n;i++{
			if(node.keys[i]>=start && node.keys[i]<=end){
				val:=make([]byte,len(node.values[i]))
				copy(val,node.values[i])
				*result=append(*result,KeyValue{Key:node.keys[i],Value:val})
			}
		}
		return
	}

	i:=0
	for i<node.n && node.keys[i]<start{
		i++
	}

	if(i<=len(node.children)){
		node.children[i].scanRange(start,end,result)
	}

	for i<node.n && node.keys[i]<=end{
		val:=make([]byte,len(node.values[i]))
		copy(val,node.values[i])
		*result=append(*result,KeyValue{Key:node.keys[i],Value:val})

		if(i+1<=len(node.children)){
			node.children[i+1].scanRange(start,end,result)
		}
		i++
	}
}

func (node *BtreeNode) getPredecessor(i int)(string,[]byte){
	cur:=node.children[i]
	for !cur.isLeaf{
		cur=cur.children[cur.n]
	}
	return cur.keys[cur.n-1],cur.values[cur.n-1]
}

func (node *BtreeNode) getSuccessor(i int)(string,[]byte){
	cur:=node.children[i+1]
	for !cur.isLeaf{
		cur=cur.children[0]
	}
	return cur.keys[0],cur.values[0]
}

func (node *BtreeNode) merge(t int,i int){
	child:=node.children[i]
	sibling:=node.children[i+1]

	child.keys=append(child.keys,node.keys[i])
	child.values=append(child.values,node.values[i])

	for j:=0;j<sibling.n;j++{
		child.keys=append(child.keys,sibling.keys[j])
		child.values=append(child.values,sibling.values[j])
	}

	if(!child.isLeaf){
		for j:=0;j<=sibling.n;j++{
			child.children=append(child.children,sibling.children[j])
		}
	}

	node.keys=append(node.keys[:i],node.keys[i+1:]...)
	node.values=append(node.values[:i],node.values[i+1:]...)
	node.children=append(node.children[:i+1],node.children[i+2:]...)
	node.n--

	child.n=sibling.n+t
}

func (node *BtreeNode) deleteFromNode(t int,key string){
	i:=0
	for i<node.n && key>node.keys[i]{
		i++
	}

	if(i<node.n && key==node.keys[i]){
		if(node.isLeaf){
			node.keys=append(node.keys[:i],node.keys[i+1:]...)
			node.values=append(node.values[:i],node.values[i+1:]...)
			node.n--
			return
		}

		if(node.children[i].n>=t){
			pred,predVal:=node.getPredecessor(i)
			node.keys[i]=pred
			node.values[i]=predVal
			node.children[i].deleteFromNode(t,pred)
			return
		}

		if(node.children[i+1].n>=t){
			succ,succVal:=node.getSuccessor(i)
			node.keys[i]=succ
			node.values[i]=succVal
			node.children[i+1].deleteFromNode(t,succ)
			return
		}

		node.merge(t,i)
		node.children[i].deleteFromNode(t,key)
		return
	}

	if(node.isLeaf){
		return
	}

	childIdx:=i

	if(node.children[childIdx].n==t-1){
		if(childIdx>0 && node.children[childIdx-1].n>=t){
			sibling:=node.children[childIdx-1]
			child:=node.children[childIdx]

			child.keys=append([]string{node.keys[childIdx-1]},child.keys...)
			child.values=append([][]byte{node.values[childIdx-1]},child.values...)

			node.keys[childIdx-1]=sibling.keys[sibling.n-1]
			node.values[childIdx-1]=sibling.values[sibling.n-1]

			if(!child.isLeaf){
				child.children=append([]*BtreeNode{sibling.children[sibling.n]},child.children...)
			}

			sibling.keys=sibling.keys[:sibling.n-1]
			sibling.values=sibling.values[:sibling.n-1]
			if(!sibling.isLeaf){
				sibling.children=sibling.children[:sibling.n]
			}
			sibling.n--
			child.n++
		}else if(childIdx<node.n && node.children[childIdx+1].n>=t){
			sibling:=node.children[childIdx+1]
			child:=node.children[childIdx]

			child.keys=append(child.keys,node.keys[childIdx])
			child.values=append(child.values,node.values[childIdx])

			node.keys[childIdx]=sibling.keys[0]
			node.values[childIdx]=sibling.values[0]

			sibling.keys=sibling.keys[1:]
			sibling.values=sibling.values[1:]

			if(!child.isLeaf){
				child.children=append(child.children,sibling.children[0])
				sibling.children=sibling.children[1:]
			}

			sibling.n--
			child.n++
		}else{
			if(childIdx<node.n){
				node.merge(t,childIdx)
			}else{
				node.merge(t,childIdx-1)
				childIdx=childIdx-1
			}
		}
	}

	node.children[childIdx].deleteFromNode(t,key)
}
