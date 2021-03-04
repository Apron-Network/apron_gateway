package models

import (
	"regexp"
	"strconv"

	"github.com/valyala/fasthttp"
)

var ProxyRequestPathPattern = regexp.MustCompile(`(?m)/v([1-9]\d{0,3})/([\w-]+)/([\w-]+)/(.*)`)

type RequestDetail struct {
	URI              *fasthttp.URI
	Host             []byte
	Path             []byte
	Method           string
	Headers          map[string][]string
	Cookies          map[string][]string
	QueryParams      map[string][]string
	FormParams       map[string][]string
	RequestBody      []byte
	Version          int
	ServiceName      []byte
	ServiceNameStr   string
	ApiKey           []byte
	ApiKeyStr        string
	ProxyRequestPath []byte
}

func ExtractCtxRequestDetail(ctx *fasthttp.RequestCtx) (*RequestDetail, error) {
	detail := RequestDetail{}

	detail.URI = ctx.URI()
	detail.Host = ctx.Host()
	detail.Path = ctx.URI().Path()

	pathMatchResult := ProxyRequestPathPattern.FindAllSubmatch(detail.Path, -1)
	if len(pathMatchResult) == 1 && len(pathMatchResult[0]) == 5 {
		detail.Version, _ = strconv.Atoi(string(pathMatchResult[0][1]))
		detail.ServiceName = pathMatchResult[0][2]
		detail.ApiKey = pathMatchResult[0][3]
		detail.ProxyRequestPath = pathMatchResult[0][4]

		detail.ServiceNameStr = string(detail.ServiceName)
		detail.ApiKeyStr = string(detail.ApiKey)
	}

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

	detail.RequestBody = ctx.PostBody()

	requestContentTypeStr := string(ctx.Request.Header.Peek("Content-Type"))
	if requestContentTypeStr != "application/json" {
		// Form request
		ctx.PostArgs().VisitAll(func(key, value []byte) {
			detail.FormParams[string(key)] = append(detail.FormParams[string(key)], string(value))
		})
	}

	return &detail, nil
}
