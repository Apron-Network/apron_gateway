package handlers

import (
	"apron.network/gateway/ratelimiter"
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/golang/protobuf/proto"
	"github.com/valyala/fasthttp"

	"apron.network/gateway/internal"
	"apron.network/gateway/internal/models"
)

type ProxyHandler struct {
	HttpClient     *http.Client
	StorageManager *models.StorageManager
	RateLimiter    *ratelimiter.Limiter

	requestDetail *models.RequestDetail
}

// InternalHandler ...
func (h *ProxyHandler) InternalHandler(ctx *fasthttp.RequestCtx) {
	h.validateRequest(ctx)

	key := string(ctx.Path()) // TODO: need process path before handle rate limit
	res, err := h.RateLimiter.Get(key)
	if err != nil {
		return
	}

	fmt.Printf("X-Ratelimit-Limit: %s\n", strconv.FormatInt(int64(res.Total), 10))
	fmt.Printf("X-Ratelimit-Remaining: %s\n", strconv.FormatInt(int64(res.Remaining), 10))
	fmt.Printf("X-Ratelimit-Reset: %s\n", strconv.FormatInt(res.Reset.Unix(), 10))

	//TODO: handle the case of no remain resource left, the key point is how to notify clients in response
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
	requestDetail, err := models.ExtractCtxRequestDetail(ctx)
	internal.CheckError(err)
	h.requestDetail = requestDetail

	if err := h.validateRequest(ctx); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetBodyString(err.Error())
		return
	}

	service := h.loadService(requestDetail.ServiceNameStr)

	if websocket.FastHTTPIsWebSocketUpgrade(ctx) && (service.Schema == "ws" || service.Schema == "wss") {
		h.forwardWebsocketRequest(ctx, service)
	} else if service.Schema == "http" || service.Schema == "https" {
		h.forwardHttpRequest(ctx, service)
	} else {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("regisited service has different schema with request")
	}
}
func (h *ProxyHandler) forwardWebsocketRequest(ctx *fasthttp.RequestCtx, service models.ApronService) {
	upgrader := websocket.FastHTTPUpgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
		serviceUrlStr := fmt.Sprintf("%s://%s", service.Schema, service.BaseUrl)
		serviceUrl, _ := url.Parse(serviceUrlStr)
		fmt.Printf("Service url: %+v\n", serviceUrl)

		dialer := websocket.Dialer{
			HandshakeTimeout: 15 * time.Second,
		}

		// TODO: Check whether header information are required for service ws
		serviceConnection, _, err := dialer.Dial(serviceUrl.String(), nil)
		internal.CheckError(err)

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

func (h *ProxyHandler) forwardHttpRequest(ctx *fasthttp.RequestCtx, service models.ApronService) {
	// Build URI, the forward URL is local httpbin URL
	serviceUrlStr := fmt.Sprintf("%s://%s", service.Schema, service.BaseUrl)
	serviceUrl, _ := url.Parse(serviceUrlStr)
	if bytes.Compare(h.requestDetail.Path, []byte("/")) != 0 {
		serviceUrl.Path += string(h.requestDetail.ProxyRequestPath)
	}

	query := serviceUrl.Query()
	for k, values := range h.requestDetail.QueryParams {
		for _, v := range values {
			query.Add(k, v)
		}
	}
	serviceUrl.RawQuery = query.Encode()

	// fmt.Printf("host: %+v, path: %+v, queries: %+v\n", serviceUrl.Host, serviceUrl.Path, serviceUrl.RawQuery)

	// Build request, query params are included in URI
	proxyReq := fasthttp.AcquireRequest()
	proxyResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(proxyReq)
	defer fasthttp.ReleaseResponse(proxyResp)

	proxyReq.SetRequestURI(serviceUrl.String())
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		proxyReq.Header.SetCanonical(k, v)
	})
	proxyReq.Header.SetMethod(h.requestDetail.Method)
	proxyReq.SetBody(h.requestDetail.RequestBody)
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

	ctx.SetBody(respBody)
}

// TODO: Validator related, perhaps can move to a new middleware

// validateRequest checks whether the request can be forwarded to backend services.
// It will check whether the key is existing in ApronApiKey:<service_name> bucket/table
func (h *ProxyHandler) validateRequest(ctx *fasthttp.RequestCtx) error {
	// Check whether API key and service has related record
	serviceBucketName := internal.ServiceApiKeyStorageBucketName(h.requestDetail.ServiceNameStr)
	if h.StorageManager.IsKeyExisting(serviceBucketName) && h.StorageManager.IsKeyExistingInBucket(serviceBucketName, h.requestDetail.ApiKeyStr) {
		return nil
	}
	// Key not found in service bucket, return forbidden
	ctx.SetStatusCode(fasthttp.StatusForbidden)
	return errors.New("unauthorized")
}

func (h *ProxyHandler) loadService(serviceName string) models.ApronService {
	r, err := h.StorageManager.GetRecord(internal.ServiceBucketName, serviceName)
	internal.CheckError(err)

	service := models.ApronService{}
	err = proto.Unmarshal([]byte(r), &service)
	internal.CheckError(err)

	return service
}
