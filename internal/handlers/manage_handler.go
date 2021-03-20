package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"apron.network/gateway/internal"
	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"

	"apron.network/gateway/internal/models"
)

// TODO: Add database client to fetch registered service and api keys
type ManagerHandler struct {
	AggrAccessRecordManager models.AggregatedAccessRecordManager

	storageManager   *models.StorageManager
	r                *router.Router
	AccessLogChannel chan string
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

	h.r.GET("/detailed_logs", h.detailedUserReportHandler)

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
	apiKeyRouter := serviceRouter.Group("/{service_id}/keys")
	apiKeyRouter.GET("/", h.listApiKeysHandler)
	apiKeyRouter.POST("/", h.newApiKeyHandler)
	apiKeyRouter.GET("/{key_id}", h.apiKeyDetailHandler)
	apiKeyRouter.PUT("/{key_id}", h.updateApiKeyHandler)
	apiKeyRouter.DELETE("/{key_id}", h.deleteApiKeyHandler)

	// User mgmt related
	userRouter := h.r.Group("/users")
	userRouter.GET("/", h.listAllUsersHandler)
	userRouter.PUT("/", h.updateUserProfileHandler)
	userRouter.GET("/keys", h.listAllUserKeysHandler)
}

func (h *ManagerHandler) indexHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "It Works!")
}

func (h *ManagerHandler) detailedUserReportHandler(ctx *fasthttp.RequestCtx) {
	if websocket.FastHTTPIsWebSocketUpgrade(ctx) {
		upgrader := websocket.FastHTTPUpgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
				return true
			},
		}

		var logMsg string

		err := upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
			for {
				logMsg = <-h.AccessLogChannel
				if err := ws.WriteMessage(websocket.TextMessage, []byte(logMsg)); err != nil {
					log.Println(err)
					return
				}
			}
		})
		internal.CheckError(err)
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}

func (h *ManagerHandler) allUsageReportHandler(ctx *fasthttp.RequestCtx) {
	if rslt, err := h.AggrAccessRecordManager.ExportAllUsage(); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(err.Error())
	} else {
		usageRecordsJsonByte, err := json.Marshal(rslt)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.SetBodyString(err.Error())
		}
		ctx.SetBody(usageRecordsJsonByte)
	}
}
