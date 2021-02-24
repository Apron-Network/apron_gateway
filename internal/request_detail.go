package internal

import "github.com/valyala/fasthttp"

type RequestDetail struct {
	URI string
	Host string
	Path string
	Method string
	Headers map[string][]string
	Cookies map[string][]string
	QueryParams map[string][]string
	FormParams map[string][]string
	RequestBody string
}

func ExtractCtxRequestDetail(ctx *fasthttp.RequestCtx) (RequestDetail, error){
	detail := RequestDetail{}

	detail.URI = ctx.URI().String()
	detail.Host = string(ctx.Host())
	detail.Path = string(ctx.URI().Path())

	detail.Method = string(ctx.Method())

	detail.Headers = make(map[string][]string)
	detail.QueryParams = make(map[string][]string)
	detail.FormParams = make(map[string][]string)

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		detail.Headers[string(key)] = append(detail.Headers[string(key)], string(value))

		// If a request is websocket request, it shouldn't change back to restful one, so this check should only be applied on potential websocket request
		if !detail.IsWebsocket {
			if (bytes.Compare(detail.URI.Scheme(), []byte("wss")) == 0) ||
				(bytes.Compare(detail.URI.Scheme(), []byte("ws")) == 0) ||
				(bytes.Compare(key, []byte("Upgrade")) == 0 && bytes.Compare(value, []byte("websocket")) == 0) ||
				(bytes.Compare(key, []byte("Connection")) == 0 && bytes.Compare(value, []byte("Upgrade")) == 0) {
				detail.IsWebsocket = true
			}
		}
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
