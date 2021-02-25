package internal

import (
	"github.com/valyala/fasthttp"
)

type RequestDetail struct {
	URI         *fasthttp.URI
	Host        []byte
	Path        []byte
	Method      string
	Headers     map[string][]string
	Cookies     map[string][]string
	QueryParams map[string][]string
	FormParams  map[string][]string
	RequestBody string
}

func ExtractCtxRequestDetail(ctx *fasthttp.RequestCtx) (RequestDetail, error) {
	detail := RequestDetail{}

	detail.URI = ctx.URI()
	detail.Host = ctx.Host()
	detail.Path = ctx.URI().Path()

	detail.Method = string(ctx.Method())

	detail.Headers = make(map[string][]string)
	detail.QueryParams = make(map[string][]string)
	detail.FormParams = make(map[string][]string)

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		detail.Headers[string(key)] = append(detail.Headers[string(key)], string(value))
	})

	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		detail.QueryParams[string(key)] = append(detail.QueryParams[string(key)], string(value))
	})

	detail.RequestBody = string(ctx.PostBody())

	requestContentTypeStr := string(ctx.Request.Header.Peek("Content-Type"))
	if requestContentTypeStr != "application/json" {
		// Form request
		ctx.PostArgs().VisitAll(func(key, value []byte) {
			detail.FormParams[string(key)] = append(detail.FormParams[string(key)], string(value))
		})
	}

	return detail, nil
}
