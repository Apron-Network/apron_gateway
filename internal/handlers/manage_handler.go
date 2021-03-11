package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"apron.network/gateway/internal"
	"apron.network/gateway/internal/models"
)

// TODO: Add database client to fetch registered service and api keys
type ManagerHandler struct {
	AggrAccessRecordManager models.AggregatedAccessRecordManager

	storageManager *models.StorageManager
	r              *router.Router
}

func (h *ManagerHandler) InitStore(storeMgr *models.StorageManager) {
	h.storageManager = storeMgr
}

func (h *ManagerHandler) Handler() func(ctx *fasthttp.RequestCtx) {
	return h.r.Handler
}

func (h *ManagerHandler) InitRouters() {
	h.r = router.New()

	h.r.GET("/", h.indexHandler)

	// Service related
	serviceRouter := h.r.Group("/service")
	serviceRouter.GET("/", h.listServiceHandler)
	serviceRouter.GET("/{service_name}/report/{key_id}", h.serviceUsageReportHandler)
	serviceRouter.GET("/report/", h.allUsageReportHandler)
	serviceRouter.POST("/", h.newServiceHandler)
	serviceRouter.POST("/{service_name}", h.serviceDetailHandler)
	serviceRouter.PUT("/{service_name}", h.updateServiceHandler)
	serviceRouter.DELETE("/{service_name}", h.deleteServiceHandler)

	// API key related
	apiKeyRouter := serviceRouter.Group("/{service_name}/keys")
	apiKeyRouter.GET("/", h.listApiKeysHandler)
	apiKeyRouter.POST("/", h.newApiKeyHandler)
	apiKeyRouter.GET("/{key_id}", h.apiKeyDetailHandler)
	apiKeyRouter.PUT("/{key_id}", h.updateApiKeyHandler)
	apiKeyRouter.DELETE("/{key_id}", h.deleteApiKeyHandler)

	// User mgmt related
	userRouter := h.r.Group("/users")
	userRouter.GET("/", h.userProfileHandler)
	userRouter.PUT("/", h.updateUserProfileHandler)
}

func (h *ManagerHandler) indexHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "It Works!")
}

func (h *ManagerHandler) allUsageReportHandler(ctx *fasthttp.RequestCtx) {
	if rslt, err := h.AggrAccessRecordManager.ExportAllUsage(); err != nil {
		internal.GenerateServerErrorResponse(ctx, err)
	} else {
		usageRecordsJsonByte, err := json.Marshal(rslt)
		if err != nil {
			internal.GenerateServerErrorResponse(ctx, err)
			return
		}
		ctx.SetBody(usageRecordsJsonByte)
	}
}
