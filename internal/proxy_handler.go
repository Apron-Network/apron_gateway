package internal

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/valyala/fasthttp"

	"apron.network/gateway/internal/models"
)

var (
	SERVICE_URI_STR = "http://localhost:2345/anything"
)

type ProxyHandler struct {
	HttpClient *http.Client
}

// ForwardHandler receives request and forward to configured services, which contains those actions
// - Identify request user
// - Authenticate user
// - Find request related service (based on passed in user credentials)
// - Transparent proxy
func (h *ProxyHandler) ForwardHandler(ctx *fasthttp.RequestCtx) {
	h.validateRequest(ctx)

	resp, err := h.sendRequestToService(ctx)
	CheckError(err)

	h.respondToClient(ctx, resp)
}

func (h *ProxyHandler) sendRequestToService(ctx *fasthttp.RequestCtx) (*http.Response, error) {
	requestDetail, _ := ExtractCtxRequestDetail(ctx)

	// Build URI, the forward URL is local httpbin URL
	serviceUrl, _ := url.Parse(SERVICE_URI_STR)
	if requestDetail.Path != "/" {
		serviceUrl.Path += requestDetail.Path
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
	req, _ := http.NewRequest(requestDetail.Method, serviceUrl.String(), bytes.NewBuffer([]byte(requestDetail.RequestBody)))

	// Fill header data
	for k, values := range requestDetail.Headers {
		for _, v := range values {
			req.Header.Set(string(k), string(v))
		}

	}

	return h.HttpClient.Do(req)
}

func (h *ProxyHandler) respondToClient(ctx *fasthttp.RequestCtx, proxyResponse *http.Response) {
	ctx.SetStatusCode(proxyResponse.StatusCode)
	fmt.Printf("Proxy response header: %+v\n", proxyResponse.Header)

	// TODO: Only set fields should be visible for client
	for k, values := range proxyResponse.Header {
		for _, v := range values {
			fmt.Printf("proxy Resp header: %s\n", k)
				ctx.Response.Header.Set(k, v)
		}
	}

	// TODO: Check cookies

	ctx.SetBodyStream(proxyResponse.Body, int(proxyResponse.ContentLength))
}

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

func (h *ProxyHandler) identifyService(ctx *fasthttp.RequestCtx) (models.Service, error) {
	// Scaffold code
	return models.Service{}, nil
}

func (h *ProxyHandler) canUserAccessService(user *models.User, service *models.Service) bool {
	return true
}
