package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/valyala/fasthttp"

	"apron.network/gateway/internal"
	"apron.network/gateway/internal/models"
)

func (h *ManagerHandler) listServiceHandler(ctx *fasthttp.RequestCtx) {
	var cursor uint64
	rslt := make([]*models.ApronService, 0, 100)

	// TODO: Refactor this scanall with function
	for {
		scanResultMap, nextCursor, _, err := h.storageManager.FetchRecords(
			internal.ServiceBucketName,
			int(cursor),
			"",
			100,
		)
		internal.CheckError(err)

		for _, v := range scanResultMap {
			tmpRcd := &models.ApronService{}
			err := proto.Unmarshal([]byte(v), tmpRcd)
			internal.CheckError(err)
			rslt = append(rslt, tmpRcd)
		}

		if nextCursor == 0 {
			break
		} else {
			cursor = nextCursor
		}
	}

	respBody, err := json.Marshal(rslt)
	internal.CheckError(err)
	ctx.Write(respBody)
}

// newServiceHandler parse request and create service in store.
// The table/bucket `ApronService` will be created if not existing,
// and a new record reflect this service will be inserted,
// the key is service name while content is ApronService object serialized by protobuf.
// An error will be respond if service with same name already existing.
func (h *ManagerHandler) newServiceHandler(ctx *fasthttp.RequestCtx) {
	detail, err := models.ExtractCtxRequestDetail(ctx)

	service := models.ApronService{}
	err = json.Unmarshal(detail.RequestBody, &service)
	internal.CheckError(err)

	if h.storageManager.IsKeyExistingInBucket(internal.ServiceBucketName, service.Id) {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.WriteString("duplicated service name")
	} else {
		binaryService, err := proto.Marshal(&service)
		internal.CheckError(err)
		err = h.storageManager.SaveBinaryKeyData(internal.ServiceBucketName, service.Id, binaryService)
		internal.CheckError(err)

		ctx.SetStatusCode(fasthttp.StatusCreated)
	}
}
func (h *ManagerHandler) serviceDetailHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Service Detail")
}
func (h *ManagerHandler) updateServiceHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Update Service")
}

func (h *ManagerHandler) deleteServiceHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Delete Service")
}
func (h *ManagerHandler) serviceUsageReportHandler(ctx *fasthttp.RequestCtx) {
	serviceId := ctx.UserValue("service_name").(string)
	keyId := ctx.UserValue("key_id").(string)
	rslt, err := h.AggrAccessRecordManager.ExportUsage(serviceId, keyId)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(err.Error())
	} else {
		ctx.SetBodyString(rslt)
	}
}
