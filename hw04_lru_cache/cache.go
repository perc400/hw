package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (lruCache *lruCache) Set(key Key, value interface{}) bool {
	if element, exists := lruCache.items[key]; exists {
		lruCache.queue.MoveToFront(element)
		element.Value = value
		return true
	}

	newElement := lruCache.queue.PushFront(value)
	lruCache.items[key] = newElement

	if lruCache.queue.Len() > lruCache.capacity {
		lruCache.Clear()
	}

	return false
}

func (lruCache *lruCache) Get(key Key) (interface{}, bool) {
	if element, exists := lruCache.items[key]; exists {
		lruCache.queue.MoveToFront(element)
		return element.Value, true
	}
	return nil, false
}

func (lruCache *lruCache) Clear() {
	valueFromQueue := lruCache.queue.Back().Value
	lruCache.queue.Remove(lruCache.queue.Back())
	for key, element := range lruCache.items {
		if element.Value == valueFromQueue {
			delete(lruCache.items, key)
		}
	}
}
