package handlers

import (
	"encoding/json"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"apron.network/gateway/internal"
	"apron.network/gateway/internal/models"
)

// listApiKeysHandler loads specified size keys from service api key hash bucket and return
func (h *ManagerHandler) listApiKeysHandler(ctx *fasthttp.RequestCtx) {
	serviceName := ctx.UserValue("service_name").(string)
	scanResultMap, cursor, resultCount, err := h.storageManager.FetchRecords(
		internal.ServiceApiKeyStorageBucketName(serviceName),
		internal.ExtractQueryIntValue(ctx, "start", 0),
		"",
		internal.ExtractQueryIntValue(ctx, "count", 10),
	)
	internal.CheckError(err)

	// Build response
	rslt := make([]models.ApronApiKey, resultCount)
	idx := 0
	for _, v := range scanResultMap {
		tmpRcd := models.ApronApiKey{}
		err := proto.Unmarshal([]byte(v), &tmpRcd)
		internal.CheckError(err)
		rslt[idx] = tmpRcd
		idx++
	}
	internal.CheckError(err)

	resp := ListApiKeysResponse{
		ServiceId:  serviceName,
		Records:    rslt,
		Count:      resultCount,
		NextCursor: cursor,
	}

	respBody, err := json.Marshal(resp)
	internal.CheckError(err)
	ctx.Write(respBody)
}

// newApiKeyHandler create a new key and relationship between key and service.
// The new key will be saved in table/bucket ApronApiKey:<service_name>,
// and uses its Key value as store key,
// while a protobuf serialized ApronApiKey object will be saved as its content.
func (h *ManagerHandler) newApiKeyHandler(ctx *fasthttp.RequestCtx) {
	// Build key object and save to redis
	newApiKeyMessage := models.ApronApiKey{
		Key:         uuid.NewString(),
		ServiceName: ctx.UserValue("service_name").(string),
		IssuedAt:    time.Now().Unix(),
	}

	binaryNewApiKey, err := proto.Marshal(&newApiKeyMessage)
	internal.CheckError(err)
	err = h.storageManager.SaveBinaryKeyData(
		internal.ServiceApiKeyStorageBucketName(newApiKeyMessage.ServiceName),
		newApiKeyMessage.Key,
		binaryNewApiKey,
	)
	internal.CheckError(err)

	// Build response
	m := jsonpb.Marshaler{}
	respBody, _ := m.MarshalToString(&newApiKeyMessage)
	ctx.WriteString(respBody)
}

func (h *ManagerHandler) apiKeyDetailHandler(ctx *fasthttp.RequestCtx) {
	serviceId := ctx.UserValue("service_name").(string)
	key := ctx.UserValue("key_id").(string)
	storageBucketName := internal.ServiceApiKeyStorageBucketName(serviceId)

	if h.storageManager.IsKeyExistingInBucket(storageBucketName, key) {
		binaryKeyData, err := h.RedisClient.HGet(internal.Ctx(), storageBucketName, key).Result()
		internal.CheckError(err)

		keyDetail := models.ApronApiKey{}
		err = proto.Unmarshal([]byte(binaryKeyData), &keyDetail)
		internal.CheckError(err)

		// Build response
		m := jsonpb.Marshaler{}
		respBody, _ := m.MarshalToString(&keyDetail)
		ctx.WriteString(respBody)
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}

func (h *ManagerHandler) updateApiKeyHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNotImplemented)
}

func (h *ManagerHandler) deleteApiKeyHandler(ctx *fasthttp.RequestCtx) {
	serviceId := ctx.UserValue("service_name").(string)
	key := ctx.UserValue("key_id").(string)
	storageBucketName := internal.ServiceApiKeyStorageBucketName(serviceId)

	if h.storageManager.IsKeyExistingInBucket(storageBucketName, key) {
		err := h.storageManager.DeleteKey(storageBucketName, key)
		internal.CheckError(err)
		ctx.SetStatusCode(fasthttp.StatusOK)
	} else {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}
