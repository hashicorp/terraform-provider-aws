// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"math"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/retry"
)

// AWS SDK for Go v1 compatible Backoff.
// See https://github.com/aws/aws-sdk-go/blob/e7dfa8a81550571e247af1ed63a698f9f43a4d51/aws/client/default_retryer.go#L78.

type v1CompatibleBackoff struct {
	maxRetryDelay time.Duration
}

// AWS SDK for Go v1 compatible Backoff.
// See https://github.com/aws/aws-sdk-go/blob/e7dfa8a81550571e247af1ed63a698f9f43a4d51/aws/client/default_retryer.go#L78.
func (c *v1CompatibleBackoff) BackoffDelay(attempt int, err error) (time.Duration, error) {
	const (
		defaultMinRetryDelay    = 30 * time.Millisecond
		defaultMinThrottleDelay = 500 * time.Millisecond
	)
	minDelay := defaultMinRetryDelay

	if retry.IsErrorThrottles(retry.DefaultThrottles).IsErrorThrottle(err).Bool() {
		minDelay = defaultMinThrottleDelay
	}

	maxDelay := c.maxRetryDelay
	var delay time.Duration

	// Logic to cap the retry count based on the minDelay provided.
	actualRetryCount := int(math.Log2(float64(minDelay))) + 1
	if actualRetryCount < 63-attempt {
		delay = time.Duration(1<<uint64(attempt)) * getJitterDelay(minDelay)
		if delay > maxDelay {
			delay = getJitterDelay(maxDelay / 2) //nolint:mnd // copied verbatim
		}
	} else {
		delay = getJitterDelay(maxDelay / 2) //nolint:mnd // copied verbatim
	}

	return delay, nil
}

func getJitterDelay(duration time.Duration) time.Duration {
	return time.Duration(seededRand.Int63n(int64(duration)) + int64(duration))
}

// lockedSource is a thread-safe implementation of rand.Source.
type lockedSource struct {
	src rand.Source
}

func (r *lockedSource) Int63() (n int64) {
	r.lock()
	n = r.src.Int63()
	r.unlock()
	return
}

func (r *lockedSource) Seed(seed int64) {
	r.lock()
	r.src.Seed(seed)
	r.unlock()
}

func (r *lockedSource) key() string {
	return "backoff-rand-source"
}

func (r *lockedSource) lock() {
	GlobalMutexKV.Lock(r.key())
}

func (r *lockedSource) unlock() {
	GlobalMutexKV.Unlock(r.key())
}

// seededRand is a new RNG using a thread safe implementation of rand.Source.
var seededRand = rand.New(&lockedSource{src: rand.NewSource(time.Now().UnixNano())})
