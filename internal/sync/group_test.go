// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sync

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// Copied from github.com/hashicorp/go-multierror.

func TestGroup(t *testing.T) {
	t.Parallel()

	err1 := errors.New("group_test: 1")
	err2 := errors.New("group_test: 2")
	cases := []struct {
		errs      []error
		nilResult bool
	}{
		{errs: []error{}, nilResult: true},
		{errs: []error{nil}, nilResult: true},
		{errs: []error{err1}},
		{errs: []error{err1, nil}},
		{errs: []error{err1, nil, err2}},
	}
	ctx := t.Context()

	for _, tc := range cases {
		var g Group

		for _, err := range tc.errs {
			g.Go(ctx, func(context.Context) error { return err })
		}

		gErr := g.Wait(ctx)
		if gErr != nil {
			for i := range tc.errs {
				if tc.errs[i] != nil && !strings.Contains(gErr.Error(), tc.errs[i].Error()) {
					t.Fatalf("expected error to contain %q, actual: %v", tc.errs[i].Error(), gErr)
				}
			}
		} else if !tc.nilResult {
			t.Fatalf("Group.Wait() should not have returned nil for errs: %v", tc.errs)
		}
	}
}
