package handlers

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	"apron.network/gateway/internal"
	"apron.network/gateway/internal/models"
)

func (h *ManagerHandler) listAllUsersHandler(ctx *fasthttp.RequestCtx) {
	if !h.storageManager.IsKeyExisting(internal.UserBucketName) {
		ctx.SetBodyString("[]")
	} else {
		var cursor uint64
		rslt := make([]string, 0, 100)

		// TODO: Refactor this scanall with function
		for {
			scanResultMap, nextCursor, _, err := h.storageManager.FetchRecords(
				internal.UserBucketName,
				int(cursor),
				"",
				100,
			)
			internal.CheckError(err)

			for userId, _ := range scanResultMap {
				rslt = append(rslt, userId)
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
}

func (h *ManagerHandler) updateUserProfileHandler(ctx *fasthttp.RequestCtx) {}
func (h *ManagerHandler) listAllUserKeysHandler(ctx *fasthttp.RequestCtx) {
	detail, err := models.ExtractCtxRequestDetail(ctx)
	internal.CheckError(err)

	accountId := detail.QueryParams["account_id"][0]

	if !h.storageManager.IsKeyExistingInBucket(internal.UserBucketName, accountId) {
		ctx.SetBodyString("[]")
	} else {
		userKeys, err := h.storageManager.GetRecord(internal.UserBucketName, accountId)
		internal.CheckError(err)
		ctx.SetBodyString(userKeys)
	}
}
