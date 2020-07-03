package goreq

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestWithFilterLimiter(t *testing.T) {
	c := NewClient(WithFilterLimiter(true, &FilterLimiterOpinion{
		LimiterMatcher: LimiterMatcher{
			Glob: "*.taobao.com",
		},
		Allow: false,
	}))
	assert.Equal(t, ReqRejectedErr, c.Do(Get("https://www.taobao.com/")).Err)
	assert.NoError(t, c.Do(Get("https://www.baidu.com/")).Err)
	c = NewClient(WithFilterLimiter(false, &FilterLimiterOpinion{
		LimiterMatcher: LimiterMatcher{
			Glob: "*.taobao.com",
		},
		Allow: false,
	}))
	assert.Equal(t, ReqRejectedErr, c.Do(Get("https://www.taobao.com/")).Err)
	assert.Equal(t, ReqRejectedErr, c.Do(Get("https://www.baidu.com/")).Err)
}

func TestWithDelayLimiter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	c := NewClient(WithDelayLimiter(false, &DelayLimiterOpinion{
		LimiterMatcher: LimiterMatcher{
			Glob: "*",
		},
		Delay: 5 * time.Second,
	}))
	var wg sync.WaitGroup
	wg.Add(5)
	t1 := time.Now()
	go func() { c.Do(Get(ts.URL)); wg.Done() }()
	go func() { c.Do(Get(ts.URL)); wg.Done() }()
	go func() { c.Do(Get(ts.URL)); wg.Done() }()
	go func() { c.Do(Get(ts.URL)); wg.Done() }()
	go func() { c.Do(Get(ts.URL)); wg.Done() }()
	wg.Wait()
	assert.True(t, time.Since(t1) >= 20*time.Second)
}

func TestWithDelayLimiterEach(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts2.Close()
	c := NewClient(WithDelayLimiter(true, &DelayLimiterOpinion{
		LimiterMatcher: LimiterMatcher{
			Glob: "*",
		},
		Delay: 5 * time.Second,
	}))
	var wg sync.WaitGroup
	wg.Add(10)
	t1 := time.Now()
	go func() { c.Do(Get(ts.URL)); wg.Done() }()
	go func() { c.Do(Get(ts2.URL)); wg.Done() }()
	go func() { c.Do(Get(ts.URL)); wg.Done() }()
	go func() { c.Do(Get(ts2.URL)); wg.Done() }()
	go func() { c.Do(Get(ts.URL)); wg.Done() }()
	go func() { c.Do(Get(ts2.URL)); wg.Done() }()
	go func() { c.Do(Get(ts.URL)); wg.Done() }()
	go func() { c.Do(Get(ts2.URL)); wg.Done() }()
	go func() { c.Do(Get(ts.URL)); wg.Done() }()
	go func() { c.Do(Get(ts2.URL)); wg.Done() }()
	wg.Wait()
	assert.True(t, time.Since(t1) >= 20*time.Second)
}

func TestWithRateLimiter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	c := NewClient(WithRateLimiter(false, &RateLimiterOpinion{
		LimiterMatcher: LimiterMatcher{
			Glob: "*",
		},
		Rate: 2,
	}))
	var wg sync.WaitGroup
	a := 32
	wg.Add(a)
	start := time.Now()
	for a > 0 {
		aa := a
		go func() { Get(ts.URL).SetClient(c).Do(); fmt.Println("finish", aa); wg.Done() }()
		a -= 1
	}
	wg.Wait()
	assert.True(t, time.Since(start) >= 15*time.Second)
}

func TestWithRateLimiterEach(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts2.Close()
	c := NewClient(WithRateLimiter(true, &RateLimiterOpinion{
		LimiterMatcher: LimiterMatcher{
			Glob: "*",
		},
		Rate: 2,
	}))
	var wg sync.WaitGroup
	start := time.Now()
	a := 32
	wg.Add(a)
	for a > 0 {
		aa := a
		go func() { Get(ts.URL).SetClient(c).Do(); fmt.Println("finish a", aa); wg.Done() }()
		a -= 1
	}
	b := 32
	wg.Add(b)
	for b > 0 {
		bb := b
		go func() { Get(ts2.URL).SetClient(c).Do(); fmt.Println("finish b", bb); wg.Done() }()
		b -= 1
	}
	wg.Wait()
	assert.True(t, time.Since(start) >= 15*time.Second)
}

func TestWithParallelismLimiter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	c := NewClient(WithParallelismLimiter(false, &ParallelismLimiterOpinion{
		LimiterMatcher: LimiterMatcher{
			Glob: "*",
		},
		Parallelism: 2,
	}))
	var wg sync.WaitGroup
	start := time.Now()
	a := 10
	wg.Add(a)
	for a > 0 {
		aa := a
		go func() { Get(ts.URL).SetClient(c).Do(); fmt.Println("finish", aa); wg.Done() }()
		a -= 1
	}
	wg.Wait()
	assert.True(t, time.Since(start) >= 25*time.Second)
}

func TestWithParallelismLimiterEach(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts2.Close()
	c := NewClient(WithParallelismLimiter(true, &ParallelismLimiterOpinion{
		LimiterMatcher: LimiterMatcher{
			Glob: "*",
		},
		Parallelism: 2,
	}))
	var wg sync.WaitGroup
	start := time.Now()
	a := 10
	wg.Add(a)
	for a > 0 {
		aa := a
		go func() { Get(ts.URL).SetClient(c).Do(); fmt.Println("finish a", aa); wg.Done() }()
		a -= 1
	}
	b := 10
	wg.Add(b)
	for b > 0 {
		bb := b
		go func() { Get(ts2.URL).SetClient(c).Do(); fmt.Println("finish b", bb); wg.Done() }()
		b -= 1
	}
	wg.Wait()
	assert.True(t, time.Since(start) >= 25*time.Second)
}
