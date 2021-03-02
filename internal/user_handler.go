package internal

import (
	"encoding/json"
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
	r           *router.Router
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
	apiKeyRouter := serviceRouter.Group("/{service_id}/keys")
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

//listApiKeysHandler loads specified size keys from service api key hash bucket and return
func (h *UserHandler) listApiKeysHandler(ctx *fasthttp.RequestCtx) {
	// Parse service data from request body
	listApiKeysRequest := network_struct.ListApiKeysRequest{
		ServiceId: ctx.UserValue("service_id").(string),
		Start:     ExtractQueryIntValue(ctx, "start", 0),
		Count:     ExtractQueryIntValue(ctx, "count", 10),
	}

	// Invoke hscan command to fetch keys
	// Note: Using hscan instead of hgetall here is to avoid performance loss of redis
	fmt.Printf("%+v\n", listApiKeysRequest)
	rcds, cursor, err := h.RedisClient.HScan(Ctx(),
		ServiceApiKeyStorageBucketName(listApiKeysRequest.ServiceId),
		uint64(listApiKeysRequest.Start),
		"",
		int64(listApiKeysRequest.Count)).Result()
	CheckError(err)

	// Rebuilt hscan result to map[string]string
	scanResultMap, resultCount, err := ParseHscanResultToObjectMap(rcds)
	CheckError(err)

	// Build response
	rslt := make([]models.ApronApiKey, resultCount)
	idx := 0
	for _, v := range scanResultMap {
		tmpRcd := models.ApronApiKey{}
		err := proto.Unmarshal([]byte(v), &tmpRcd)
		CheckError(err)
		rslt[idx] = tmpRcd
		idx++
	}
	CheckError(err)

	resp := network_struct.ListApiKeysResponse{
		ServiceId:  listApiKeysRequest.ServiceId,
		Records:    rslt,
		Count:      resultCount,
		NextCursor: cursor,
	}

	respBody, err := json.Marshal(resp)
	CheckError(err)
	ctx.Write(respBody)
}

func (h *UserHandler) newApiKeyHandler(ctx *fasthttp.RequestCtx) {
	// Build key object and save to redis
	newApiKeyMessage := models.ApronApiKey{
		Name:      uuid.NewString(),
		Key:       uuid.NewString(),
		ServiceId: ctx.UserValue("service_id").(string),
		IssuedAt:  time.Now().Unix(),
	}

	binaryNewApiKey, err := proto.Marshal(&newApiKeyMessage)
	CheckError(err)

	_, err = h.RedisClient.HSet(Ctx(),
		ServiceApiKeyStorageBucketName(newApiKeyMessage.ServiceId),
		newApiKeyMessage.Key,
		binaryNewApiKey).Result()
	CheckError(err)

	// Build response
	m := jsonpb.Marshaler{}
	respBody, _ := m.MarshalToString(&newApiKeyMessage)
	ctx.WriteString(respBody)
}
func (h *UserHandler) apiKeyDetailHandler(ctx *fasthttp.RequestCtx) {
	serviceId := ctx.UserValue("service_id").(string)
	key := ctx.UserValue("key_id").(string)
	storageBucketName := ServiceApiKeyStorageBucketName(serviceId)

	existing, err := h.RedisClient.HExists(Ctx(), storageBucketName, key).Result()
	CheckError(err)
	if existing {
		binaryKeyData, err := h.RedisClient.HGet(Ctx(), storageBucketName, key).Result()
		CheckError(err)

		keyDetail := models.ApronApiKey{}
		err = proto.Unmarshal([]byte(binaryKeyData), &keyDetail)
		CheckError(err)

		// Build response
		m := jsonpb.Marshaler{}
		respBody, _ := m.MarshalToString(&keyDetail)
		ctx.WriteString(respBody)
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}
func (h *UserHandler) updateApiKeyHandler(ctx *fasthttp.RequestCtx) {

}
func (h *UserHandler) deleteApiKeyHandler(ctx *fasthttp.RequestCtx) {

}

func (h *UserHandler) userProfileHandler(ctx *fasthttp.RequestCtx)       {}
func (h *UserHandler) updateUserProfileHandler(ctx *fasthttp.RequestCtx) {}
