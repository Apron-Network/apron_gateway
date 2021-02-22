package internal

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	"apron.network/gateway/internal/models"
)

type ProxyHandler struct{}

// ForwardHandler receives request and forward to configured services, which contains those actions
// - Identify request user
// - Authenticate user
// - Find request related service (based on passed in user credentials)
// - Transparent proxy
func (h *ProxyHandler) ForwardHandler(ctx *fasthttp.RequestCtx) {
	h.validateRequest(ctx)

	debugResponse, _ := BuildProxyDebugResponseFromCtx(ctx)

	ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))
	ctx.SetStatusCode(200)
	jsonStr, _ := json.Marshal(debugResponse)
	ctx.Write(jsonStr)
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
