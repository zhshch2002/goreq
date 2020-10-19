package main

import (
	"fmt"
	"github.com/zhshch2002/goreq"
	"time"
)

func main() {
	c := goreq.NewClient(goreq.WithFilterLimiter(true, &goreq.FilterLimiterOpinion{
		LimiterMatcher: goreq.LimiterMatcher{
			Glob: "*.taobao.com",
		},
		Allow: false,
	}))
	fmt.Println(c.Do(goreq.Get("https://www.taobao.com/")).Err) // ReqRejectedErr

	c = goreq.NewClient(goreq.WithDelayLimiter(false, &goreq.DelayLimiterOpinion{
		LimiterMatcher: goreq.LimiterMatcher{
			Glob: "*",
		},
		Delay: 5 * time.Second,
		// RandomDelay: 5 * time.Second,
	}))

	c = goreq.NewClient(goreq.WithRateLimiter(false, &goreq.RateLimiterOpinion{
		LimiterMatcher: goreq.LimiterMatcher{
			Glob: "*",
		},
		Rate: 2,
	}))
}
