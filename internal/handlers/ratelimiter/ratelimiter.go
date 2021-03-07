package ratelimiter

import (
	"errors"
	"time"
)

// Limiter struct.
type Limiter struct {
	abstractLimiter
}

// Options for Limiter
type Options struct {
	Max      int           // The max count in duration for no policy, default is 100.
	Duration time.Duration // Count duration for no policy, default is 1 Minute.
}

// Result of limiter.Get
type Result struct {
	Total     int           // It Equals Options.Max, or policy max
	Remaining int           // It will always >= -1
	Duration  time.Duration // It Equals Options.Duration, or policy duration
	Reset     time.Time     // The limit record reset time
}

// New returns a Limiter instance with given options.
// If options.Client omit, the limiter is a memory limiter
func New(opts Options) *Limiter {
	if opts.Max <= 0 {
		opts.Max = 100
	}
	if opts.Duration <= 0 {
		opts.Duration = time.Minute
	}
	return newMemoryLimiter(&opts)
}

/*
Get get a limiter result:

    key := "user-123456"
    res, err := limiter.Get(key)
    if err == nil {
        fmt.Println(res.Reset)     // 2016-10-11 21:17:53.362 +0800 CST
        fmt.Println(res.Total)     // 100
        fmt.Println(res.Remaining) // 100
        fmt.Println(res.Duration)  // 1m
    }


    key := "id-123456"
    policy := []int{100, 60000, 50, 60000, 50, 120000}
    res, err := limiter.Get(key, policy...)
	if err == nil {
        fmt.Println(res.Reset)     // 2016-10-11 21:17:53.362 +0800 CST
        fmt.Println(res.Total)     // 100
        fmt.Println(res.Remaining) // 100
        fmt.Println(res.Duration)  // 1m
    }
*/
/// Get limiter
func (l *Limiter) Get(id string, policy ...int) (Result, error) {
	var result Result
	key := id

	if odd := len(policy) % 2; odd == 1 {
		return result, errors.New("ratelimiter: must be paired values")
	}

	res, err := l.getLimit(key, policy...)
	if err != nil {
		return result, err
	}

	result = Result{}
	result.Remaining = res[0].(int)
	result.Total = res[1].(int)
	result.Duration = res[2].(time.Duration)
	result.Reset = res[3].(time.Time)

	return result, nil
}

// Remove remove limiter record for id
func (l *Limiter) Remove(id string) error {
	return l.removeLimit(id)
}
