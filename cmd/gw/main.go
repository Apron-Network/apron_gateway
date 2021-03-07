package main

import (
	"apron.network/gateway/internal"
	"encoding/json"
	"go.uber.org/zap"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/go-redis/redis/v8"

	"apron.network/gateway/internal/handlers"
	"apron.network/gateway/internal/handlers/ratelimiter"
	"apron.network/gateway/internal/models"
)

func startAdminService(addr string, wg *sync.WaitGroup, redisClient *redis.Client) {
	h := handlers.ManagerHandler{}
	h.InitStore(&models.StorageManager{
		RedisClient: redisClient,
	})
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

	zapConfigByte := []byte(`{
	  "level": "info",
	  "encoding": "json",
	  "outputPaths": ["stdout", "logs/proxy.log"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var zapCfg zap.Config
	err := json.Unmarshal(zapConfigByte, &zapCfg)
	internal.CheckError(err)

	proxyLogger, err := zapCfg.Build()
	internal.CheckError(err)
	defer proxyLogger.Sync()

	h := handlers.ProxyHandler{
		HttpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: t,
		},
		StorageManager: &models.StorageManager{
			RedisClient: redisClient,
		},
		RateLimiter: ratelimiter.New(ratelimiter.Options{
			Max:      60,
			Duration: time.Minute,
		}),
		Logger: proxyLogger,
	}

	if err := fasthttp.ListenAndServe(addr, h.InternalHandler); err != nil {
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
