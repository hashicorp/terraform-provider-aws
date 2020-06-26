package tfawsresource

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

// Semaphore can be used to limit concurrent executions. This can be used to work with resources with low quotas
type Semaphore chan struct{}

// InitializeSemaphore initializes a semaphore with a default capacity or overrides it using an environment variable
func InitializeSemaphore(envvar string, defaultLimit int) Semaphore {
	limit := defaultLimit
	x := os.Getenv(envvar)
	if x != "" {
		var err error
		limit, err = strconv.Atoi(x)
		if err != nil {
			panic(fmt.Errorf("could not parse %q: expected integer, got %q", envvar, x))
		}
	}
	return make(Semaphore, limit)
}

// Wait waits for a semaphore before continuing
func (s Semaphore) Wait() {
	s <- struct{}{}
}

// Notify releases a semaphore
func (s Semaphore) Notify() {
	<-s
}

// TestAccPreCheckSyncronized waits for a semaphore and skips the test if there is no capacity
func TestAccPreCheckSyncronize(t *testing.T, semaphore Semaphore, resource string) {
	if cap(semaphore) == 0 {
		t.Skipf("concurrency for %s testing set to 0", resource)
	}

	semaphore.Wait()
}
