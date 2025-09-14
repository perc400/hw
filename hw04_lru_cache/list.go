package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
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
	front *ListItem
	back  *ListItem
	len   int
}

func NewList() List {
	return new(list)
}

func (list *list) Len() int {
	return list.len
}

func (list *list) Front() *ListItem {
	return list.front
}

func (list *list) Back() *ListItem {
	return list.back
}

func (list *list) PushFront(v interface{}) *ListItem {
	newItem := &ListItem{Value: v}
	if list.len == 0 {
		list.front = newItem
		list.back = newItem
	} else {
		newItem.Prev = nil
		newItem.Next = list.front
		list.front.Prev = newItem
		list.front = newItem
	}
	list.len++
	return newItem
}

func (list *list) PushBack(v interface{}) *ListItem {
	newItem := &ListItem{Value: v}
	if list.len == 0 {
		list.front = newItem
		list.back = newItem
	} else {
		newItem.Prev = list.back
		newItem.Next = nil
		list.back.Next = newItem
		list.back = newItem
	}
	list.len++
	return newItem
}

func (list *list) Remove(i *ListItem) {
	if i == nil {
		return
	}

	if i.Prev == nil {
		list.front = i.Next
	} else {
		i.Prev.Next = i.Next
	}

	if i.Next == nil {
		list.back = i.Prev
	} else {
		i.Next.Prev = i.Prev
	}
	list.len--
}

func (list *list) MoveToFront(i *ListItem) {
	if i == nil || i == list.front {
		return
	}

	if i.Prev != nil {
		i.Prev.Next = i.Next
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		list.back = i.Prev
	}

	i.Prev = nil
	i.Next = list.front
	list.front.Prev = i
	list.front = i
}
