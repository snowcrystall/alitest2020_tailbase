package main

import "sync"

type Element struct {
	next  *Element
	Value interface{}
}

type Queue struct {
	first *Element
	last  *Element
	len   int
	m     sync.Mutex
}

func (q *Queue) Init() *Queue {
	q.first = nil
	q.last = nil
	q.len = 0
	q.m = sync.Mutex{}
	return q
}

func NewQueue() *Queue { return new(Queue).Init() }

func (q *Queue) Len() int { return q.len }

func (q *Queue) Pop() interface{} {
	q.m.Lock()
	defer q.m.Unlock()
	if q.len == 0 {
		return nil
	}
	e := q.first
	q.first = q.first.next
	q.len--
	e.next = nil
	return e.Value
}

func (q *Queue) Push(v interface{}) {
	q.m.Lock()
	defer q.m.Unlock()
	e := Element{nil, v}
	if q.len == 0 {
		q.first = &e
	} else {
		q.last.next = &e
	}
	q.last = &e
	q.len++
}
