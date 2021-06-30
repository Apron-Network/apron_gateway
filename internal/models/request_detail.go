package models

import (
	"apron.network/gateway/internal"
	"bytes"
	"fmt"
	"github.com/valyala/fasthttp"
	"regexp"
	"strconv"
)

var ProxyRequestPathPattern = regexp.MustCompile(`(?m)/v([1-9]\d{0,3})/([\w-]+)/(.*)`)

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

	pathWithKey := detail.Path

	if bytes.HasPrefix(detail.Path, []byte("/ws")) {
		// Remove /ws prefix to match regexp for extracting version and key
		pathWithKey = pathWithKey[3:]
	}

	detail.ServiceName = internal.ServiceHostnameToIdByte(detail.Host)

	pathMatchResult := ProxyRequestPathPattern.FindAllSubmatch(pathWithKey, -1)
	fmt.Printf("Path with k: %+q\n", pathWithKey)
	fmt.Printf("Match rslt: %+q\n", pathMatchResult)

	if len(pathMatchResult) == 1 && len(pathMatchResult[0]) == 4 {
		detail.Version, _ = strconv.Atoi(string(pathMatchResult[0][1]))
		detail.ApiKey = pathMatchResult[0][2]
		detail.ProxyRequestPath = pathMatchResult[0][3]

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
