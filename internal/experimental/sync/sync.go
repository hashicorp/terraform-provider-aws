// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sync

import (
	"os"
	"strconv"
	"sync"

	testing "github.com/mitchellh/go-testing-interface"
)

// Semaphore can be used to limit concurrent executions.
// This can be used to work with resources with low quotas.
type Semaphore chan struct{}

var semaphoreKV = &struct {
	lock  sync.Locker
	store map[string]Semaphore
}{
	lock:  &sync.Mutex{},
	store: make(map[string]Semaphore),
}

// GetSemaphore returns a named semaphore with a default capacity or overrides it using an environment variable
// NOTE: this is currently an experimental feature and is likely to change. DO NOT USE.
func GetSemaphore(key, envvar string, defaultLimit int) Semaphore {
	semaphoreKV.lock.Lock()
	defer semaphoreKV.lock.Unlock()

	semaphore, ok := semaphoreKV.store[key]
	if !ok {
		limit := defaultLimit
		if v := os.Getenv(envvar); v != "" {
			if v, err := strconv.Atoi(v); err == nil {
				limit = v
			}
		}

		semaphore = make(Semaphore, limit)
		semaphoreKV.store[key] = semaphore
	}

	return semaphore
}

// Wait waits for a semaphore before continuing
// NOTE: this is currently an experimental feature and is likely to change. DO NOT USE.
func (s Semaphore) Wait() {
	s <- struct{}{}
}

// Notify releases a semaphore
// NOTE: this is currently an experimental feature and is likely to change. DO NOT USE.
func (s Semaphore) Notify() {
	// Make the Notify non-blocking. This can happen if a Wait was never issued
	select {
	case <-s:
	default:
		// log.Println("[WARN] Notifying semaphore without Wait")
	}
}

// TestAccPreCheckSyncronized waits for a semaphore and skips the test if there is no capacity
// NOTE: this is currently an experimental feature and is likely to change. DO NOT USE.
func TestAccPreCheckSyncronize(t testing.T, semaphore Semaphore, resource string) {
	if cap(semaphore) == 0 {
		t.Skipf("concurrency for %s testing set to 0", resource)
	}

	semaphore.Wait()
}
