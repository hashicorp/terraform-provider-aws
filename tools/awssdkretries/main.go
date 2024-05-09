// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
)

func main() {
	maxBackoff := 300 * time.Second
	maxRetries := 25
	v2 := retry.NewExponentialJitterBackoff(maxBackoff)
	v1 := client.DefaultRetryer{
		NumMaxRetries:    maxRetries,
		MaxRetryDelay:    maxBackoff,
		MaxThrottleDelay: maxBackoff,
	}
	v2compat := &compat{
		MaxRetryDelay: maxBackoff,
	}

	err := awserr.New("ThrottlingException", "Rate exceeded", nil)
	req := request.Request{
		Error: err,
		HTTPResponse: &http.Response{
			StatusCode: 400,
		},
	}
	for i := 0; i < maxRetries; i++ {
		d1 := v1.RetryRules(&req)
		req.RetryCount++
		d2, _ := v2.BackoffDelay(i, err)
		d2compat, _ := v2compat.BackoffDelay(i, err)

		fmt.Printf("%d v1: %s, v2: %s v2compat: %s\n", i, d1.String(), d2.String(), d2compat.String())
	}
}

type compat struct {
	MaxRetryDelay time.Duration
}

// AWS SDK for Go v1 compatible Backoff.
// See https://github.com/aws/aws-sdk-go/blob/e7dfa8a81550571e247af1ed63a698f9f43a4d51/aws/client/default_retryer.go#L78.
func (c *compat) BackoffDelay(attempt int, err error) (time.Duration, error) {
	minDelay := 30 * time.Millisecond
	maxDelay := c.MaxRetryDelay
	var delay time.Duration

	// Logic to cap the retry count based on the minDelay provided
	actualRetryCount := int(math.Log2(float64(minDelay))) + 1
	if actualRetryCount < 63-attempt {
		delay = time.Duration(1<<uint64(attempt)) * getJitterDelay(minDelay)
		if delay > maxDelay {
			delay = getJitterDelay(maxDelay / 2)
		}
	} else {
		delay = getJitterDelay(maxDelay / 2)
	}

	return delay, nil
}

func getJitterDelay(duration time.Duration) time.Duration {
	return time.Duration(SeededRand.Int63n(int64(duration)) + int64(duration))
}

// lockedSource is a thread-safe implementation of rand.Source
type lockedSource struct {
	lk  sync.Mutex
	src rand.Source
}

func (r *lockedSource) Int63() (n int64) {
	r.lk.Lock()
	n = r.src.Int63()
	r.lk.Unlock()
	return
}

func (r *lockedSource) Seed(seed int64) {
	r.lk.Lock()
	r.src.Seed(seed)
	r.lk.Unlock()
}

// SeededRand is a new RNG using a thread safe implementation of rand.Source
var SeededRand = rand.New(&lockedSource{src: rand.NewSource(time.Now().UnixNano())})
