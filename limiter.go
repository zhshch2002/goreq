package goreq

import (
	"github.com/gobwas/glob"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LimitRuleAllow uint8

const (
	NotSet LimitRuleAllow = iota
	Allow
	Disallow
)

type LimiterMatcher struct {
	Regexp, Glob   string
	compiledRegexp *regexp.Regexp
	compiledGlob   glob.Glob
}

func (s *LimiterMatcher) Match(u *url.URL) bool {
	match := false
	if s.compiledGlob != nil {
		match = s.compiledGlob.Match(strings.ToLower(u.Host))
	} else {
		match = s.compiledRegexp.MatchString(strings.ToLower(u.Host))
	}
	return match
}

func (s *LimiterMatcher) Compile() {
	if s.Glob != "" {
		s.compiledGlob = glob.MustCompile(s.Glob)
	} else if s.Regexp != "" {
		s.compiledRegexp = regexp.MustCompile(s.Regexp)
	}
}

type FilterLimiterOpinion struct {
	LimiterMatcher
	Allow bool
}

func WithFilterLimiter(noneMatchAllow bool, opts ...*FilterLimiterOpinion) Middleware {
	for i := range opts {
		opts[i].Compile()
	}
	return func(c *Client, h Handler) Handler {
		return func(req *Request) *Response {
			for i := range opts {
				if opts[i].Match(req.URL) {
					if opts[i].Allow {
						return h(req)
					} else {
						return nil
					}
				}
			}
			if noneMatchAllow {
				return h(req)
			} else {
				return nil
			}
		}
	}
}

type delayLimiterVal struct {
	Delay       time.Duration
	RandomDelay time.Duration
	lastReqTime time.Time
	lock        sync.Mutex
}

type DelayLimiterOpinion struct {
	LimiterMatcher
	Delay       time.Duration
	RandomDelay time.Duration
	lastReqTime time.Time
	lock        sync.Mutex
}

func WithDelayLimiter(eachSite bool, opts ...*DelayLimiterOpinion) Middleware {
	for i := range opts {
		opts[i].Compile()
		opts[i].lock = sync.Mutex{}
	}
	sites := sync.Map{}
	return func(c *Client, h Handler) Handler {
		return func(req *Request) *Response {
			if !eachSite {
				for i := range opts {
					if opts[i].Match(req.URL) {
						opts[i].lock.Lock()
						since := time.Since(opts[i].lastReqTime)
						if since < opts[i].Delay {
							time.Sleep(opts[i].Delay - since)
						}
						if opts[i].RandomDelay > 0 {
							ra := rand.New(rand.NewSource(time.Now().Unix()))
							time.Sleep(time.Duration(ra.Int63n(int64(opts[i].RandomDelay))))
						}
						res := h(req)
						opts[i].lastReqTime = time.Now()
						opts[i].lock.Unlock()
						return res
					}
				}
			}
			for i := range opts {
				if opts[i].Match(req.URL) {
					v, _ := sites.LoadOrStore(req.URL.Host, &delayLimiterVal{
						Delay:       opts[i].Delay,
						RandomDelay: opts[i].Delay,
						lastReqTime: time.Time{},
						lock:        sync.Mutex{},
					})
					val := v.(*delayLimiterVal)
					val.lock.Lock()
					since := time.Since(val.lastReqTime)
					if since < val.Delay {
						time.Sleep(val.Delay - since)
					}
					if val.RandomDelay > 0 {
						ra := rand.New(rand.NewSource(time.Now().Unix()))
						time.Sleep(time.Duration(ra.Int63n(int64(val.RandomDelay))))
					}
					res := h(req)
					val.lastReqTime = time.Now()
					val.lock.Unlock()
					return res
				}
			}
			return h(req)
		}
	}
}

type rateLimiterVal struct {
	Rate     int64
	rateLeft int64
}

type RateLimiterOpinion struct {
	LimiterMatcher
	Rate     int64
	rateLeft int64
}

func WithRateLimiter(eachSite bool, opts ...*RateLimiterOpinion) Middleware {
	for i := range opts {
		opts[i].Compile()
	}
	sites := sync.Map{}
	go func() {
		for {
			if eachSite {
				sites.Range(func(k, v interface{}) bool {
					data := v.(*rateLimiterVal)
					data.rateLeft = data.Rate
					sites.Store(k, data)
					return true
				})
			} else {
				for i := range opts {
					opts[i].rateLeft = opts[i].Rate
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return func(c *Client, h Handler) Handler {
		return func(req *Request) *Response {
			if !eachSite {
				for i := range opts {
					if opts[i].Match(req.URL) {
						wait := true
						for wait {
							if atomic.LoadInt64(&opts[i].rateLeft) > 0 {
								atomic.AddInt64(&opts[i].rateLeft, -1)
								wait = false
							} else {
								time.Sleep(100 * time.Microsecond)
							}
						}
						return h(req)
					}
				}
			}
			for i := range opts {
				if opts[i].Match(req.URL) {
					v, _ := sites.LoadOrStore(req.URL.Host, &rateLimiterVal{
						Rate:     opts[i].Rate,
						rateLeft: opts[i].rateLeft,
					})
					val := v.(*rateLimiterVal)
					wait := true
					for wait {
						if atomic.LoadInt64(&val.rateLeft) > 0 {
							atomic.AddInt64(&val.rateLeft, -1)
							wait = false
						} else {
							time.Sleep(100 * time.Microsecond)
						}
					}
					return h(req)
				}
			}
			return h(req)
		}
	}
}

type parallelismLimiterVal struct {
	Parallelism        int64
	workingParallelism int64
}

type ParallelismLimiterOpinion struct {
	LimiterMatcher
	Parallelism        int64
	workingParallelism int64
}

func WithParallelismLimiter(eachSite bool, opts ...*ParallelismLimiterOpinion) Middleware {
	for i := range opts {
		opts[i].Compile()
	}
	sites := sync.Map{}
	return func(c *Client, h Handler) Handler {
		return func(req *Request) *Response {
			if !eachSite {
				for i := range opts {
					if opts[i].Match(req.URL) {
						wait := true
						for wait {
							if atomic.LoadInt64(&opts[i].workingParallelism) < opts[i].Parallelism {
								atomic.AddInt64(&opts[i].workingParallelism, 1)
								wait = false
							} else {
								time.Sleep(100 * time.Microsecond)
							}
						}
						resp := h(req)
						atomic.AddInt64(&opts[i].workingParallelism, -1)
						return resp
					}
				}
			}
			for i := range opts {
				if opts[i].Match(req.URL) {
					v, _ := sites.LoadOrStore(req.URL.Host, &parallelismLimiterVal{
						Parallelism:        opts[i].Parallelism,
						workingParallelism: opts[i].workingParallelism,
					})
					val := v.(*parallelismLimiterVal)
					wait := true
					for wait {
						if atomic.LoadInt64(&val.workingParallelism) < val.Parallelism {
							atomic.AddInt64(&val.workingParallelism, 1)
							wait = false
						} else {
							time.Sleep(100 * time.Microsecond)
						}
					}
					resp := h(req)
					atomic.AddInt64(&val.workingParallelism, -1)
					return resp
				}
			}
			return h(req)
		}
	}
}
