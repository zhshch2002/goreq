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

func TestLimitDelay(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	c := NewClient()
	c.Use(WithLimiter(true, &LimitRule{
		Glob: "*",
		//Allow: Allow,
		Delay: 5 * time.Second,
	}))
	start := time.Now()
	Get(ts.URL).SetClient(c).Do()
	Get(ts.URL).SetClient(c).Do()
	Get(ts.URL).SetClient(c).Do()
	Get(ts.URL).SetClient(c).Do()
	Get(ts.URL).SetClient(c).Do()
	assert.True(t, time.Since(start) >= 20*time.Second)
}

func TestLimiterRate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	c := NewClient()
	c.Use(WithLimiter(true, &LimitRule{
		Glob: "*",
		Rate: 2,
	}))
	var wg sync.WaitGroup
	start := time.Now()
	a := 30
	wg.Add(a)
	for a > 0 {
		aa := a
		go func() { Get(ts.URL).SetClient(c).Do(); fmt.Println("finish", aa); wg.Done() }()
		a -= 1
	}
	wg.Wait()
	assert.True(t, time.Since(start) >= 10*time.Second)
}

func TestLimiterParallelism(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	c := NewClient()
	c.Use(WithLimiter(true, &LimitRule{
		Glob:        "*",
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
