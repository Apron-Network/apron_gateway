package internal

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
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

func ServiceHostnameToId(hostname string) string {
	return strings.Replace(hostname, ".", "_", -1)
}

func ServiceHostnameToIdByte(hostname []byte) []byte {
	return bytes.Replace(hostname, []byte("."), []byte("_"), -1)
}

func ServiceApiKeyStorageBucketName(serviceId string) string {
	return fmt.Sprintf("ApronApiKey:%s", serviceId)
}

// GenTimestamp ...
func GenTimestamp() string {
	time := time.Now().UnixNano() / 1e6
	return strconv.FormatInt(time, 10)
}

const ServiceBucketName = "ApronService"
const UserBucketName = "ApronUser"
