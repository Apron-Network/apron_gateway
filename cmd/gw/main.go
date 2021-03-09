package main

import (
	"apron.network/gateway/internal"
	"flag"
	"fmt"
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

func startAdminService(addr string, wg *sync.WaitGroup, redisClient *redis.Client, manager models.AggregatedAccessRecordManager) {
	h := handlers.ManagerHandler{
		AggrAccessRecordManager: manager,
	}
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

func startProxyService(addr string, wg *sync.WaitGroup, redisClient *redis.Client, manager models.AggregatedAccessRecordManager) {
	// TODO: Load from configurations
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	proxyLogger := internal.GatewayLogger{
		LogFile: "logs/proxy_log.txt",
	}
	proxyLogger.Init()

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
		Logger:                  &proxyLogger,
		AggrAccessRecordManager: manager,
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

	// CLI parser
	redisServerPtr := flag.String("redis", "localhost:6379", "redis server used for saving service and user key")
	proxyPortPtr := flag.Int("proxy-port", 8080, "proxy server port")
	adminAddrPtr := flag.String("admin-addr", "127.0.0.1:8082", "Admin service address")
	flag.Parse()

	proxyServerAddr := fmt.Sprintf(":%d", *proxyPortPtr)

	fmt.Println("Service info:")
	fmt.Printf("\tProxy addr: %s\n", proxyServerAddr)
	fmt.Printf("\tAdmin service addr: %s\n", *adminAddrPtr)
	fmt.Printf("\tRedis server: %s\n", *redisServerPtr)

	rdb := redis.NewClient(&redis.Options{
		Addr:     *redisServerPtr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	aggrAccessRecordManager := models.AggregatedAccessRecordManager{}
	aggrAccessRecordManager.Init()

	go startProxyService(proxyServerAddr, wg, rdb, aggrAccessRecordManager)
	go startAdminService(*adminAddrPtr, wg, rdb, aggrAccessRecordManager)

	wg.Wait()
}
