package main

import (
	"fmt"
	"testing"
)

func TestInsertValue(t *testing.T) {
	a := [2]int64{1, 4}
	b := [2]int64{4, 7}
	c := [2]int64{2, 3}
	d := [2]int64{0, 1}

	l := NewSortList()
	l.InsertValue(a)
	l.InsertValue(b)
	l.InsertValue(c)
	l.InsertValue(d)

	for i := l.Next(); i != nil; i = i.next {
		fmt.Println(i.Offset)
	}

	o := NewOffset()
	o.List = l
	s := o.SlideCur([2]int64{7, 8}, true)
	fmt.Println(o.Cur, s)
	for i := o.List.Next(); i != nil; i = i.next {
		fmt.Println(i.Offset)
	}

}
