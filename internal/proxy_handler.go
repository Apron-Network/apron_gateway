package internal

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/go-redis/redis/v8"
	"github.com/valyala/fasthttp"

	"apron.network/gateway/internal/models"
	"apron.network/gateway/ratelimiter"
)

var (
	RestServiceUrlStr = "http://localhost:2345/anything"
	WsServiceUrlStr   = "wss://stream.binance.com:9443/ws/bnbbtc@depth"
)

type ProxyHandler struct {
	HttpClient  *http.Client
	RedisClient *redis.Client
	RateLimiter *ratelimiter.Limiter
}

// InternalHandler ...
func (h *ProxyHandler) InternalHandler(ctx *fasthttp.RequestCtx) {
	h.validateRequest(ctx)

	res, err := h.RateLimiter.Get(string(ctx.Path()))
	if err != nil {
		return
	}

	fmt.Printf("X-Ratelimit-Limit: %s\n", strconv.FormatInt(int64(res.Total), 10))
	fmt.Printf("X-Ratelimit-Remaining: %s\n", strconv.FormatInt(int64(res.Remaining), 10))
	fmt.Printf("X-Ratelimit-Reset: %s\n", strconv.FormatInt(res.Reset.Unix(), 10))

	//TODO: handle the case of no remain resource left, the key point is how to notify clients
	if res.Remaining < 0 {
		after := int64(res.Reset.Sub(time.Now())) / 1e9
		fmt.Printf("Retry-After: %s\n", strconv.FormatInt(after, 10))
		return
	}

	h.ForwardHandler(ctx)
}

// ForwardHandler receives request and forward to configured services, which contains those actions
// - Identify request user
// - Authenticate user
// - Find request related service (based on passed in user credentials)
// - Transparent proxy
func (h *ProxyHandler) ForwardHandler(ctx *fasthttp.RequestCtx) {
	h.validateRequest(ctx)

	requestDetail, _ := ExtractCtxRequestDetail(ctx)

	if websocket.FastHTTPIsWebSocketUpgrade(ctx) {
		h.forwardWebsocketRequest(ctx, &requestDetail)
	} else {
		h.forwardHttpRequest(ctx, &requestDetail)
	}
}

func (h *ProxyHandler) forwardWebsocketRequest(ctx *fasthttp.RequestCtx, requestDetail *RequestDetail) {
	upgrader := websocket.FastHTTPUpgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
		serviceUrl, _ := url.Parse(WsServiceUrlStr)
		fmt.Printf("Service url: %+v\n", serviceUrl)

		dialer := websocket.Dialer{
			HandshakeTimeout: 15 * time.Second,
		}

		// TODO: Check whether header information are required for service ws
		serviceConnection, _, err := dialer.Dial(serviceUrl.String(), nil)
		CheckError(err)

		for {
			messageType, p, err := serviceConnection.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			if err := ws.WriteMessage(messageType, p); err != nil {
				log.Println(err)
				return
			}
		}
	})
}

func (h *ProxyHandler) forwardHttpRequest(ctx *fasthttp.RequestCtx, requestDetail *RequestDetail) {
	// Build URI, the forward URL is local httpbin URL
	serviceUrl, _ := url.Parse(RestServiceUrlStr)
	if bytes.Compare(requestDetail.Path, []byte("/")) != 0 {
		serviceUrl.Path += string(requestDetail.Path)
	}

	query := serviceUrl.Query()
	for k, values := range requestDetail.QueryParams {
		for _, v := range values {
			query.Add(k, v)
		}
	}
	serviceUrl.RawQuery = query.Encode()

	fmt.Printf("host: %+v, path: %+v, queries: %+v\n", serviceUrl.Host, serviceUrl.Path, serviceUrl.RawQuery)

	// Build request, query params are included in URI
	proxyReq := fasthttp.AcquireRequest()
	proxyResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(proxyReq)
	defer fasthttp.ReleaseResponse(proxyResp)

	proxyReq.SetRequestURI(serviceUrl.String())
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		proxyReq.Header.SetCanonical(k, v)
	})
	proxyReq.Header.SetMethod(requestDetail.Method)
	proxyReq.SetBodyString(requestDetail.RequestBody)
	if err := fasthttp.Do(proxyReq, proxyResp); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	}

	respBody := proxyResp.Body()

	ctx.SetStatusCode(proxyResp.StatusCode())

	// TODO: Only set fields should be visible for client
	proxyResp.Header.VisitAll(func(k, v []byte) {
		ctx.Response.Header.SetCanonical(k, v)
	})

	// TODO: Check cookies

	ctx.SetBody(respBody)
}

// TODO: Validator related, perhaps can move to a new middleware

//validateRequest checks whether the request can be forwarded to backend services
func (h *ProxyHandler) validateRequest(ctx *fasthttp.RequestCtx) {
	user, err := h.identifyUser(ctx)
	CheckError(err)

	service, err := h.identifyService(ctx)
	CheckError(err)

	if !h.canUserAccessService(&user, &service) {
		panic("Not allowed")
	}
}

func (h *ProxyHandler) identifyUser(ctx *fasthttp.RequestCtx) (models.User, error) {
	// Parse API key from request
	// Find user via API key

	// Scaffold code
	return models.User{}, nil
}

func (h *ProxyHandler) identifyService(ctx *fasthttp.RequestCtx) (models.ApronService, error) {
	// Scaffold code
	return models.ApronService{}, nil
}

func (h *ProxyHandler) canUserAccessService(user *models.User, service *models.ApronService) bool {
	return true
}
