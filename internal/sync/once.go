// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sync

import (
	"context"
	"sync"
)

// Adapted from sync.OnceValue in the Go standard library

func OnceValueCtx[T any](f func(context.Context) T) func(context.Context) T {
	d := struct {
		f      func(context.Context) T
		once   sync.Once
		valid  bool
		p      any
		result T
	}{
		f: f,
	}
	return func(ctx context.Context) T {
		d.once.Do(func() {
			defer func() {
				d.f = nil // release reference to f after first call
				d.p = recover()
				if !d.valid {
					panic(d.p)
				}
			}()
			d.result = d.f(ctx)
			d.valid = true
		})
		if !d.valid {
			panic(d.p)
		}
		return d.result
	}
}
