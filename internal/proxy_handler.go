package internal

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

type ProxyHandler struct{}

// ForwardHandler receives request and forward to configured services, which contains those actions
// - Identify request user
// - Authenticate user
// - Find request related service (based on passed in user credentials)
// - Transparent proxy
func (h ProxyHandler) ForwardHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Proxy index")
}
