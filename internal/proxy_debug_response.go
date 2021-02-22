package internal

import "github.com/valyala/fasthttp"

type ProxyDebugResponse struct {
	Host string
	Method string
	Headers map[string][]string
	Cookies map[string][]string
	QueryParams map[string][]string
	FormParams map[string][]string
	RequestBody string
}

func BuildProxyDebugResponseFromCtx(ctx *fasthttp.RequestCtx) (ProxyDebugResponse, error){
	debugResponse := ProxyDebugResponse{}
	debugResponse.Host = string(ctx.Host())
	debugResponse.Method = string(ctx.Method())

	debugResponse.Headers = make(map[string][]string)
	debugResponse.QueryParams = make(map[string][]string)
	debugResponse.FormParams = make(map[string][]string)

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		debugResponse.Headers[string(key)] = append(debugResponse.Headers[string(key)], string(value))
	})

	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		debugResponse.QueryParams[string(key)] = append(debugResponse.QueryParams[string(key)], string(value))
	})

	debugResponse.RequestBody = string(ctx.PostBody())

	requestContentTypeStr := string(ctx.Request.Header.Peek("Content-Type"))
	if requestContentTypeStr != "application/json" {
		// Form request
		ctx.PostArgs().VisitAll(func(key, value []byte) {
			debugResponse.FormParams[string(key)] = append(debugResponse.FormParams[string(key)], string(value))
		})
	}


	return debugResponse, nil
}
