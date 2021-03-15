package main

import (
	"apron.network/gateway/internal"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/go-redis/redis/v8"

	"apron.network/gateway/internal/handlers"
	"apron.network/gateway/internal/handlers/ratelimiter"
	"apron.network/gateway/internal/models"
)

var (
	corsAllowHeaders     = "authorization"
	corsAllowMethods     = "HEAD,GET,POST,PUT,DELETE,OPTIONS"
	corsAllowOrigin      = "*"
	corsAllowCredentials = "true"
)

func CORS(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", corsAllowCredentials)
		ctx.Response.Header.Set("Access-Control-Allow-Headers", corsAllowHeaders)
		ctx.Response.Header.Set("Access-Control-Allow-Methods", corsAllowMethods)
		ctx.Response.Header.Set("Access-Control-Allow-Origin", corsAllowOrigin)

		next(ctx)
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	} else {
		return defaultVal
	}
}

func startAdminService(addr string, wg *sync.WaitGroup, redisClient *redis.Client, manager models.AggregatedAccessRecordManager, accessLogChannel chan string) {
	h := handlers.ManagerHandler{
		AggrAccessRecordManager: manager,
		AccessLogChannel:        accessLogChannel,
	}
	h.InitStore(&models.StorageManager{
		RedisClient: redisClient,
	})
	h.InitRouters()

	if err := fasthttp.ListenAndServe(addr, CORS(h.Handler())); err != nil {
		log.Fatalf("Error in Admin service: %s", err)
		wg.Done()
	}
	wg.Done()
}

func startProxyService(addr string, wg *sync.WaitGroup, redisClient *redis.Client, manager models.AggregatedAccessRecordManager, accessLogChannel chan string) {
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
		AccessLogChannel:        accessLogChannel,
	}

	if err := fasthttp.ListenAndServe(addr, CORS(h.InternalHandler)); err != nil {
		log.Fatalf("Error in Proxy service: %s", err)
		wg.Done()
	}
	wg.Done()
}

func main() {
	// TODO: Define config file format - After logic finalized
	wg := new(sync.WaitGroup)
	wg.Add(2)

	proxyPort, err := strconv.ParseInt(getEnv("PROXY_PORT", "8080"), 10, 32)
	internal.CheckError(err)
	adminAddrStr := getEnv("ADMIN_ADDR", "127.0.0.1:8082")
	redisServer := getEnv("REDIS_SERVER", "localhost:6379")

	proxyServerAddr := fmt.Sprintf(":%d", proxyPort)

	fmt.Println("Service info:")
	fmt.Printf("\tProxy addr: %s\n", proxyServerAddr)
	fmt.Printf("\tAdmin service addr: %s\n", adminAddrStr)
	fmt.Printf("\tRedis server: %s\n", redisServer)

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisServer,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	aggrAccessRecordManager := models.AggregatedAccessRecordManager{}
	aggrAccessRecordManager.Init()

	accessLogChannel := make(chan string, 4096)
	defer close(accessLogChannel)

	go startProxyService(proxyServerAddr, wg, rdb, aggrAccessRecordManager, accessLogChannel)
	go startAdminService(adminAddrStr, wg, rdb, aggrAccessRecordManager, accessLogChannel)

	wg.Wait()
}
