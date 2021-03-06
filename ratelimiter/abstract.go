package ratelimiter

type abstractLimiter interface {
	getLimit(key string, policy ...int) ([]interface{}, error)
	removeLimit(key string) error
}
