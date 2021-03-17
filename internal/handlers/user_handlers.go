package handlers

import (
	"apron.network/gateway/internal"
	"apron.network/gateway/internal/models"
	"github.com/valyala/fasthttp"
)

func (h *ManagerHandler) userProfileHandler(ctx *fasthttp.RequestCtx)       {}
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
