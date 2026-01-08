package hw04lrucache

import "sync"

var mu sync.Mutex

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

type KeyValue struct {
	key   Key
	value interface{}
}

func (cache *lruCache) Set(key Key, value interface{}) bool {
	mu.Lock()
	defer mu.Unlock()

	v, ok := cache.items[key]

	if ok {
		cache.queue.Remove(v)
	} else {
		if cache.queue.Len() == cache.capacity {
			item := cache.queue.Back().Value.(KeyValue)

			delete(cache.items, item.key)
			cache.queue.Remove(cache.queue.Back())
		}
	}

	cache.queue.PushFront(KeyValue{key, value})
	cache.items[key] = cache.queue.Front()

	return ok
}
func (cache *lruCache) Get(key Key) (interface{}, bool) {
	mu.Lock()
	defer mu.Unlock()

	v, ok := cache.items[key]

	if ok {
		cache.queue.MoveToFront(v)
		cache.items[key] = cache.queue.Front()

		return v.Value.(KeyValue).value, true
	}

	return v, ok
}
func (cache *lruCache) Clear() {
	mu.Lock()
	defer mu.Unlock()

	cache.queue.Clear()
	clear(cache.items)
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
