package main

import (
	"container/list"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type CacheItem struct {
	key        string
	value      string
	expiration int64
}

type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List
	lock     sync.Mutex
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (c *LRUCache) Get(key string) (string, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if elem, found := c.items[key]; found {
		item := elem.Value.(*CacheItem)
		if time.Now().Unix() > item.expiration {
			c.order.Remove(elem)
			delete(c.items, key)
			return "", false
		}
		c.order.MoveToFront(elem)
		return item.value, true
	}
	return "", false
}

func (c *LRUCache) Set(key, value string, expiration int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if elem, found := c.items[key]; found {
		c.order.MoveToFront(elem)
		elem.Value.(*CacheItem).value = value
		elem.Value.(*CacheItem).expiration = time.Now().Unix() + expiration
		return
	}

	if c.order.Len() >= c.capacity {
		backElem := c.order.Back()
		if backElem != nil {
			backItem := backElem.Value.(*CacheItem)
			delete(c.items, backItem.key)
			c.order.Remove(backElem)
		}
	}

	item := &CacheItem{
		key:        key,
		value:      value,
		expiration: time.Now().Unix() + expiration,
	}
	elem := c.order.PushFront(item)
	c.items[key] = elem
}

func (c *LRUCache) DeleteExpired() {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now().Unix()
	for elem := c.order.Back(); elem != nil; elem = elem.Prev() {
		item := elem.Value.(*CacheItem)
		if item.expiration > now {
			break
		}
		c.order.Remove(elem)
		delete(c.items, item.key)
	}
}

func main() {
	r := mux.NewRouter()
	cache := NewLRUCache(1024)

	r.HandleFunc("/get/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["key"]
		if value, found := cache.Get(key); found {
			json.NewEncoder(w).Encode(map[string]string{"key": key, "value": value})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}).Methods("GET")

	r.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		var data map[string]string
		_ = json.NewDecoder(r.Body).Decode(&data)
		key := data["key"]
		value := data["value"]
		expiration, _ := strconv.ParseInt(data["expiration"], 10, 64)
		cache.Set(key, value, expiration)
		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	go func() {
		for {
			time.Sleep(1 * time.Second)
			cache.DeleteExpired()
		}
	}()

	handler := cors.Default().Handler(r)
	http.ListenAndServe(":8080", handler)
}
