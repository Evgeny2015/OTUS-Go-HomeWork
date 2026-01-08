package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	Clear()
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	front *ListItem
	back  *ListItem
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) Clear() {
	for i := l.front; i != nil; i = i.Next {
		i.Prev = nil
	}
	l.front = nil
	l.back = nil
	l.len = 0
}

func (l *list) PushFront(v interface{}) *ListItem {
	// creates new ListItem
	item := &ListItem{Value: v}

	// link items
	l.LinkItems(item, l.front)

	// set top of the list
	l.front = item

	// if list is empty then set back pointer too
	if l.len == 0 {
		l.back = l.front
	}

	// increment item counter
	l.len++

	return item
}

func (l *list) PushBack(v interface{}) *ListItem {
	// creates new ListItem
	item := &ListItem{Value: v}

	// link items
	l.LinkItems(l.back, item)

	// set bottom of the list
	l.back = item

	// if list is empty then set front pointer too
	if l.len == 0 {
		l.front = l.back
	}

	// increment item counter
	l.len++

	return item
}

func (l *list) Remove(i *ListItem) {
	l.LinkItems(i.Prev, i.Next)

	if l.front == i {
		l.front = i.Next
	}

	if l.back == i {
		l.back = i.Prev
	}

	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	l.Remove(i)
	l.PushFront(i.Value)
}

func (l *list) LinkItems(from, to *ListItem) {
	if from != nil {
		from.Next = to
	}

	if to != nil {
		to.Prev = from
	}
}

func NewList() List {
	return new(list)
}
