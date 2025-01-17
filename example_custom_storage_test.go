package httpcache_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aibarbetta/httpcache"
	"github.com/aibarbetta/httpcache/cache"
	patrickCache "github.com/patrickmn/go-cache"
)

type customInMemStorage struct {
	cacheHandler *patrickCache.Cache
}

// NewCustomInMemStorage will return a custom in memory cache
func NewCustomInMemStorage() cache.ICacheInteractor {
	return &customInMemStorage{
		cacheHandler: patrickCache.New(patrickCache.DefaultExpiration, time.Second*10),
	}
}

func (c customInMemStorage) Set(key string, value cache.CachedResponse) error { //nolint
	c.cacheHandler.Set(key, value, patrickCache.DefaultExpiration)
	return nil
}

func (c customInMemStorage) Get(key string) (res cache.CachedResponse, err error) {
	cachedRes, ok := c.cacheHandler.Get(key)
	if !ok {
		err = cache.ErrCacheMissed
		return
	}
	res, ok = cachedRes.(cache.CachedResponse)
	if !ok {
		err = cache.ErrInvalidCachedResponse
		return
	}
	return
}
func (c customInMemStorage) Delete(key string) error {
	c.cacheHandler.Delete(key)
	return nil
}
func (c customInMemStorage) Flush() error {
	c.cacheHandler.Flush()
	return nil
}
func (c customInMemStorage) Origin() string {
	return "MY-OWN-CUSTOM-INMEMORY-CACHED"
}

func Example_withCustomStorage() {
	client := &http.Client{}
	handler, err := httpcache.NewWithCustomStorageCache(client, true, false, NewCustomInMemStorage())
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		startTime := time.Now()
		req, err := http.NewRequestWithContext(context.TODO(), "GET", "https://imantumorang.com", http.NoBody)
		if err != nil {
			log.Fatal((err))
		}
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Response time: %v micro-second\n", time.Since(startTime).Microseconds())
		fmt.Println("Status Code", res.StatusCode)
		time.Sleep(time.Second * 1)
		fmt.Println("Sequence >>> ", i)
		if i%5 == 0 {
			err := handler.CacheInteractor.Flush()
			if err != nil {
				log.Fatal(err)
			}
		}
		res.Body.Close()
	}
	// Example Output:
	/*
		2020/06/21 13:14:51 Cache item's missing failed to retrieve from cache, trying with a live version
		Response time: 940086 micro-second
		Status Code 200
		Sequence >>>  0
		2020/06/21 13:14:53 Cache item's missing failed to retrieve from cache, trying with a live version
		Response time: 73679 micro-second
		Status Code 200
		Sequence >>>  1
		Response time: 126 micro-second
		Status Code 200
		Sequence >>>  2
		Response time: 96 micro-second
		Status Code 200
		Sequence >>>  3
		Response time: 102 micro-second
		Status Code 200
		Sequence >>>  4
		Response time: 94 micro-second
		Status Code 200
		Sequence >>>  5
	*/
}
