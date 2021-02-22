package internal

import (
	"fmt"
	"net"
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

var h = &UserHandler{}

func TestMain(m *testing.M) {
	h.InitRouters()
}

func serve(handler fasthttp.RequestHandler, req *fasthttp.Request, res *fasthttp.Response) error {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go func() {
		err := fasthttp.Serve(ln, handler)
		if err != nil {
			panic(fmt.Errorf("failed to serve: %v", err))
		}
	}()

	client := fasthttp.HostClient{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}

	return client.Do(req, res)
}

func TestUserHandlerNewService(t *testing.T) {
	req := fasthttp.AcquireRequest()
	req.Header.SetHost("test.com")
	req.SetRequestURI("/services") // task URI
	req.Header.SetMethod("GET")

	resp := fasthttp.AcquireResponse()
	err := serve(h.Handler(), req, resp)

	if err != nil {
		t.Errorf("New service API error: %+v\n", err)
	}
}
