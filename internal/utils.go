package internal

import (
	"context"
	"fmt"

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

const ServiceBucketName = "ApronService"
