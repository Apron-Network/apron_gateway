package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/go-redis/redis/v8"

	"apron.network/gateway/internal"
)

func startAdminService(addr string, wg *sync.WaitGroup, redisClient *redis.Client) {
	h := internal.UserHandler{
		RedisClient: redisClient,
	}
	h.InitRouters()

	if err := fasthttp.ListenAndServe(addr, h.Handler()); err != nil {
		log.Fatalf("Error in Admin service: %s", err)
		wg.Done()
	}
	wg.Done()
}

func startProxyService(addr string, wg *sync.WaitGroup, redisClient *redis.Client) {
	// TODO: Load from configurations
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	h := internal.ProxyHandler{HttpClient: &http.Client{
		Timeout:   10 * time.Second,
		Transport: t,
	},
		RedisClient: redisClient}

	if err := fasthttp.ListenAndServe(addr, h.ForwardHandler); err != nil {
		log.Fatalf("Error in Proxy service: %s", err)
		wg.Done()
	}
	wg.Done()
}

func main() {
	// TODO: Add cli params support, to pass in config file path
	// TODO: Define config file format - After logic finalized
	wg := new(sync.WaitGroup)
	wg.Add(2)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	go startProxyService(":8080", wg, rdb)
	go startAdminService("127.0.0.1:8082", wg, rdb)

	wg.Wait()
}
