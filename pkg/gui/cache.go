package gui

import (
	"container/list"
	"image"
	"sync"
)

const thumbCacheCapacity = 200

type thumbCacheEntry struct {
	key string
	img image.Image
}

// ThumbCache is a small in-memory LRU cache of decoded thumbnail images,
// keyed by thumbnail URL. It lets paginating back through search results
// repaint instantly instead of re-fetching already-seen thumbnails.
type ThumbCache struct {
	mu    sync.Mutex
	items map[string]*list.Element
	order *list.List // front = most recently used
}

// NewThumbCache returns an empty cache with a fixed capacity.
func NewThumbCache() *ThumbCache {
	return &ThumbCache{
		items: make(map[string]*list.Element),
		order: list.New(),
	}
}

// Get returns the cached image for key, if present, and marks it as
// recently used.
func (c *ThumbCache) Get(key string) (image.Image, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[key]
	if !ok {
		return nil, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*thumbCacheEntry).img, true
}

// Put stores img under key, evicting the least recently used entry if the
// cache is at capacity.
func (c *ThumbCache) Put(key string, img image.Image) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[key]; ok {
		el.Value.(*thumbCacheEntry).img = img
		c.order.MoveToFront(el)
		return
	}

	el := c.order.PushFront(&thumbCacheEntry{key: key, img: img})
	c.items[key] = el

	if c.order.Len() > thumbCacheCapacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			delete(c.items, oldest.Value.(*thumbCacheEntry).key)
		}
	}
}
