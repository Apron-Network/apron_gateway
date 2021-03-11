package internal

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

// CheckError is a helper function to simplify error checking
func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func Ctx() context.Context {
	return context.Background()
}

// ExtractQueryIntValue tries to extract int value from query string, and return default value if not presents
func ExtractQueryIntValue(ctx *fasthttp.RequestCtx, argName string, defaultValue int) int {
	val, err := ctx.QueryArgs().GetUint(argName)
	if err == fasthttp.ErrNoArgValue {
		return defaultValue
	} else {
		return val
	}
}

func ServiceApiKeyStorageBucketName(service_id string) string {
	return fmt.Sprintf("ApronApiKey:%s", service_id)
}

// GenerateServerErrorResponse set respond status to 500 and fill detail message if error occurred,
// this function will replace CheckError in web request related code
// TODO: Update the function to involve all HTTP errors occurred in project, such as 404, 429
func GenerateServerErrorResponse(ctx *fasthttp.RequestCtx, err error) {
	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	ctx.SetBodyString(err.Error())
}

// GenTimestamp ...
func GenTimestamp() string {
	time := time.Now().UnixNano() / 1e6
	return strconv.FormatInt(time, 10)
}

const ServiceBucketName = "ApronService"
