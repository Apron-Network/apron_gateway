package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"apron.network/gateway/internal/handlers/ratelimiter"

	"github.com/fasthttp/websocket"
	"github.com/golang/protobuf/proto"
	"github.com/valyala/fasthttp"

	"apron.network/gateway/internal"
	"apron.network/gateway/internal/models"
)

type ProxyHandler struct {
	HttpClient              *http.Client
	StorageManager          *models.StorageManager
	RateLimiter             *ratelimiter.Limiter
	Logger                  *internal.GatewayLogger
	AggrAccessRecordManager models.AggregatedAccessRecordManager
	AccessLogChannel        chan string

	upgrader         *websocket.FastHTTPUpgrader
	requestDetail    *models.RequestDetail
	serviceAggrCount map[string]uint32 // Simple aggr count for detail logs
}

// InternalHandler ...
func (h *ProxyHandler) InternalHandler(ctx *fasthttp.RequestCtx) {
	requestDetail, err := models.ExtractCtxRequestDetail(ctx)
	internal.CheckError(err)
	h.requestDetail = requestDetail
	h.validateRequest(ctx)

	key := string(ctx.Path()) // TODO: need process path before handle rate limit
	res, err := h.RateLimiter.Get(key)
	if err != nil {
		return
	}

	fmt.Printf("X-Ratelimit-Limit: %s\n", strconv.FormatInt(int64(res.Total), 10))
	fmt.Printf("X-Ratelimit-Remaining: %s\n", strconv.FormatInt(int64(res.Remaining), 10))
	fmt.Printf("X-Ratelimit-Reset: %s\n", strconv.FormatInt(res.Reset.Unix(), 10))

	// TODO: handle the case of no remain resource left, the key point is how to notify clients in response
	if res.Remaining < 0 {
		after := int64(res.Reset.Sub(time.Now())) / 1e9
		fmt.Printf("Retry-After: %s\n", strconv.FormatInt(after, 10))

		h.Logger.Log(fmt.Sprintf("%s|429 error|%s: from %s, service: %s, api_key: %s\n",
			time.Now().UTC().Format("2006-01-02 15:04:05"),
			requestDetail.URI.String(),
			ctx.RemoteIP().String(),
			requestDetail.ServiceNameStr,
			requestDetail.ApiKeyStr,
		))

		return
	}

	h.Logger.Log(fmt.Sprintf("%s|%s: from %s, service: %s, api_key: %s\n",
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		requestDetail.URI.String(),
		ctx.RemoteIP().String(),
		requestDetail.ServiceNameStr,
		requestDetail.ApiKeyStr,
	))
	h.AggrAccessRecordManager.IncUsage(requestDetail.ServiceNameStr, requestDetail.ApiKeyStr)
	access_log := models.AccessLog{
		Ts:          int64(int(time.Now().UnixNano() / 1e6)),
		ServiceName: requestDetail.ServiceNameStr,
		UserKey:     requestDetail.ApiKeyStr,
		RequestIp:   string(ctx.RemoteIP().String()),
		RequestPath: string(requestDetail.ProxyRequestPath),
	}
	access_log_bytes, err := json.Marshal(&access_log)
	internal.CheckError(err)
	h.AccessLogChannel <- string(access_log_bytes)

	h.ForwardHandler(ctx)
}

// ForwardHandler receives request and forward to configured services, which contains those actions
// - Identify request user
// - Authenticate user
// - Find request related service (based on passed in user credentials)
// - Transparent proxy
func (h *ProxyHandler) ForwardHandler(ctx *fasthttp.RequestCtx) {
	if err := h.validateRequest(ctx); err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetBodyString(err.Error())
		return
	}

	service := h.loadService(h.requestDetail.ServiceNameStr)

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
	h.upgrader = &websocket.FastHTTPUpgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
			return true
		},
	}

	var err error

	serviceUrlStr := fmt.Sprintf("%s://%s", service.Schema, service.BaseUrl)
	serviceUrl, _ := url.Parse(serviceUrlStr)
	fmt.Printf("Service url: %+v\n", serviceUrl)

	dialer := websocket.Dialer{
		HandshakeTimeout: 15 * time.Second,
	}

	// TODO: Check whether header information are required for service ws
	proxyServerWsConn, _, err := dialer.Dial(serviceUrl.String(), nil)
	internal.CheckError(err)

	err = h.upgrader.Upgrade(ctx, func(clientWsConn *websocket.Conn) {
		defer clientWsConn.Close()

		var (
			errClient      = make(chan error, 1)
			errProxyServer = make(chan error, 1)
		)

		go forwardWsMessage(clientWsConn, proxyServerWsConn, errClient)
		go forwardWsMessage(proxyServerWsConn, clientWsConn, errProxyServer)

		for {
			select {
			case err := <-errClient:
				fmt.Sprintf("Error while forwarding response: %+v\n", err.Error())
			case err := <-errProxyServer:
				fmt.Sprintf("Error while forwarding request: %+v\n", err.Error())
			}
		}
	})
	internal.CheckError(err)
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

func forwardWsMessage(src, dest *websocket.Conn, errCh chan error) {
	for {
		msgType, msgBytes, err := src.ReadMessage()

		if err != nil {
			fmt.Printf("src.ReadMessage failed, msgType=%d, msg=%s, err=%v\n", msgType, msgBytes, err)
			if ce, ok := err.(*websocket.CloseError); ok {
				msgBytes = websocket.FormatCloseMessage(ce.Code, ce.Text)
			} else {
				msgBytes = websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, err.Error())
			}

			errCh <- err

			if err = dest.WriteMessage(websocket.CloseMessage, msgBytes); err != nil {
				fmt.Printf("write close message failed, err=%v", err)
			}

			break
		}

		dest.WriteMessage(msgType, msgBytes)
		if err != nil {
			fmt.Printf("dest.WriteMessage error: %v\n", err)
			errCh <- err
			break
		}
	}
}
