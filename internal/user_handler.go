package internal

import (
	"fmt"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// TODO: Add database client to fetch regisited service and api keys
type UserHandler struct {
	r *router.Router
}

func (h *UserHandler) Handler() func(ctx *fasthttp.RequestCtx) {
	return h.r.Handler
}

func (h *UserHandler) InitRouters() {
	h.r = router.New()

	h.r.GET("/", h.indexHandler)

	// Service related
	serviceRouter := h.r.Group("/service")
	serviceRouter.GET("/", h.listServiceHandler)
	serviceRouter.POST("/", h.newServiceHandler)
	serviceRouter.POST("/{service_id}", h.serviceDetailHandler)
	serviceRouter.PUT("/{service_id}", h.updateServiceHandler)
	serviceRouter.DELETE("/{service_id}", h.deleteServiceHandler)

	// API key related
	apiKeyRouter := h.r.Group("/keys")
	apiKeyRouter.GET("/", h.listApiKeysHandler)
	apiKeyRouter.POST("/", h.newApiKeyHandler)
	apiKeyRouter.POST("/{service_id}", h.apiKeyDetailHandler)
	apiKeyRouter.PUT("/{service_id}", h.updateApiKeyHandler)
	apiKeyRouter.DELETE("/{service_id}", h.deleteApiKeyHandler)
}

func (h *UserHandler) indexHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Admin index")
}

func (h *UserHandler) listServiceHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "List Service")
}
func (h *UserHandler) newServiceHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "New Service")
}
func (h *UserHandler) serviceDetailHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Service Detail")
}
func (h *UserHandler) updateServiceHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Update Service")
}

func (h *UserHandler) deleteServiceHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Delete Service")
}

func (h *UserHandler) listApiKeysHandler(ctx *fasthttp.RequestCtx) {
}
func (h *UserHandler) newApiKeyHandler(ctx *fasthttp.RequestCtx) {

}
func (h *UserHandler) apiKeyDetailHandler(ctx *fasthttp.RequestCtx) {

}
func (h *UserHandler) updateApiKeyHandler(ctx *fasthttp.RequestCtx) {

}
func (h *UserHandler) deleteApiKeyHandler(ctx *fasthttp.RequestCtx) {

}
