package internal

import (
	"context"
	"errors"

	"github.com/valyala/fasthttp"
)

//CheckError is a helper function to simplify error checking
func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func Ctx() context.Context {
	return context.Background()
}

//ExtractQueryIntValue tries to extract int value from query string, and return default value if not presents
func ExtractQueryIntValue(ctx *fasthttp.RequestCtx, argName string, defaultValue int) int {
	val, err := ctx.QueryArgs().GetUint(argName)
	if err == fasthttp.ErrNoArgValue {
		return defaultValue
	} else {
		return val
	}
}

func ParseHscanResultToObjectMap(rcds []string) (map[string]string, uint, error) {
	if len(rcds) % 2 != 0 {
		return nil, 0, errors.New("record length should be even number")
	}

	resultCount := uint(len(rcds) / 2)
	rslt := make(map[string]string)
	for idx := 0; idx <len(rcds); idx+=2 {
		rslt[rcds[idx]] = rcds[idx+1]
	}

	return rslt, resultCount, nil
}
