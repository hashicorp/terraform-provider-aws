// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sync

import (
	"errors"
	"sync"
)

// Copied from github.com/hashicorp/go-multierror and adapted to use `errors.Join`.

type Group struct {
	mutex sync.Mutex
	errs  []error
	wg    sync.WaitGroup
}

// Go calls the given function in a new goroutine.
//
// If the function returns an error it is added to the group's errors.
func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.mutex.Lock()
			g.errs = append(g.errs, err)
			g.mutex.Unlock()
		}
	}()
}

// Wait blocks until all function calls from the Go method have returned,
// then returns the group's errors wrapped via `errors.Join`.
func (g *Group) Wait() error {
	g.wg.Wait()
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return errors.Join(g.errs...)
}
