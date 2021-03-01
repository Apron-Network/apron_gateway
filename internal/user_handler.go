package internal

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"apron.network/gateway/internal/models"
	"apron.network/gateway/internal/network_struct"

	"github.com/fasthttp/router"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

// TODO: Add database client to fetch registered service and api keys
type UserHandler struct {
	RedisClient *redis.Client
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
	apiKeyRouter :=serviceRouter.Group("/{service_id}/keys")
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
	// Parse service data from request body
	listApiKeysRequest := network_struct.ListApiKeysRequest{
		ServiceId: ctx.UserValue("service_id").(string),
	}

	rcds, err := h.RedisClient.HGetAll(Ctx(), models.ServiceApiKeyStorageBucketName(listApiKeysRequest.ServiceId)).Result()
	CheckError(err)

	fmt.Printf("%+v\n", rcds)
}
func (h *UserHandler) newApiKeyHandler(ctx *fasthttp.RequestCtx) {
	// Parse service data from request body
	newApiRequest := network_struct.NewApiKeyRequest{
		ServiceId: ctx.UserValue("service_id").(string),
	}

	newApiKeyMessage := models.ApronApiKey{
		Name: uuid.NewString(),
		Val: uuid.NewString(),
		ServiceId: newApiRequest.ServiceId,
		IssuedAt: time.Now().Unix(),
	}

	binaryNewApiKey, err := proto.Marshal(&newApiKeyMessage)
	CheckError(err)

	_, err = h.RedisClient.HSet(Ctx(), newApiKeyMessage.StoreBucketName(), newApiKeyMessage.Val, binaryNewApiKey).Result()
	CheckError(err)

	m := jsonpb.Marshaler{}

	respBody, _ := m.MarshalToString(&newApiKeyMessage)
	ctx.WriteString(respBody)
}
func (h *UserHandler) apiKeyDetailHandler(ctx *fasthttp.RequestCtx) {

}
func (h *UserHandler) updateApiKeyHandler(ctx *fasthttp.RequestCtx) {

}
func (h *UserHandler) deleteApiKeyHandler(ctx *fasthttp.RequestCtx) {

}

func (h *UserHandler) userProfileHandler(ctx *fasthttp.RequestCtx) {}
func (h *UserHandler) updateUserProfileHandler(ctx *fasthttp.RequestCtx) {}
