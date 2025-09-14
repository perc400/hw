package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type cacheItem struct {
	key   Key
	value interface{}
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
		element.Value.(*cacheItem).value = value
		return true
	}

	newElement := lruCache.queue.PushFront(&cacheItem{key: key, value: value})
	lruCache.items[key] = newElement

	if lruCache.queue.Len() > lruCache.capacity {
		item := lruCache.queue.Back()
		lruCache.queue.Remove(item)
		delete(lruCache.items, item.Value.(*cacheItem).key)
	}

	return false
}

func (lruCache *lruCache) Get(key Key) (interface{}, bool) {
	if element, exists := lruCache.items[key]; exists {
		lruCache.queue.MoveToFront(element)
		return element.Value.(*cacheItem).value, true
	}
	return nil, false
}

func (lruCache *lruCache) Clear() {
	newList := NewList()
	newMap := make(map[Key]*ListItem, lruCache.capacity)
	lruCache.queue = newList
	lruCache.items = newMap
}
