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
	SERVIC_URI_STR = "http://localhost:2345/anything"
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
	serviceUrl, _ := url.Parse(SERVIC_URI_STR)
	serviceUrl.Path += requestDetail.Path
	fmt.Printf("ServiceURL: %+v\n", requestDetail.URI)
	for k, values := range requestDetail.QueryParams {
		for _, v := range values {
			serviceUrl.Query().Set(k, v)
		}
	}

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

func (h *ProxyHandler) respondToClient(ctx *fasthttp.RequestCtx, resp *http.Response) {
	ctx.SetStatusCode(resp.StatusCode)
	for k, values := range resp.Header {
		for _, v:= range values {
			ctx.Response.Header.SetCanonical([]byte(k), []byte(v))
		}
	}

	// TODO: Check cookies

	ctx.SetBodyStream(resp.Body, int(resp.ContentLength))
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
