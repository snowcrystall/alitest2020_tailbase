package main

import (
	"sync"
)

type Offset struct {
	Cur     int64
	List    *LNode
	l       *sync.Mutex
	Waiting int
}

type LNode struct {
	next   *LNode
	Offset [2]int64
}

func NewOffset() *Offset {
	return &Offset{0, NewSortList(), &sync.Mutex{}, 0}
}
func NewSortList() *LNode {
	return &LNode{nil, [2]int64{}}
}

func (l *LNode) InsertValue(v [2]int64) {
	node := &LNode{nil, v}
	i := l
	for i.next != nil {
		if v[0] > i.next.Offset[0] {
			i = i.next
		} else {
			break
		}
	}
	node.next = i.next
	i.next = node
}

func (l *LNode) Next() *LNode {
	if l != nil {
		return l.next
	}
	return nil
}

func (l *LNode) DeleteFirst() *LNode {
	i := l.next
	if i != nil {
		l.next = i.next
		i.next = nil
	}
	return l
}
func (o *Offset) SlideCur(offset [2]int64, withSlice bool) [][2]int64 {
	o.l.Lock()
	defer o.l.Unlock()
	o.List.InsertValue(offset)
	offsetSlice := [][2]int64{}
	i := o.List.Next()
	if o.Cur == 0 && i.Offset[0] == 0 {
		o.Cur = i.Offset[1]
		if withSlice {
			offsetSlice = append(offsetSlice, i.Offset)
		}
	}
	if o.Cur == i.Offset[1] {
		for i.Next() != nil {
			if i.Next().Offset[0]-i.Offset[1] < 300 {
				if withSlice {
					offsetSlice = append(offsetSlice, i.Next().Offset)
				}
				o.Cur = i.Next().Offset[1]
				o.List.DeleteFirst()
				i = o.List.Next()
			} else {
				break
			}
		}
	}
	return offsetSlice
}
