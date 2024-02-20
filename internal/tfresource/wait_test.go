// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfresource_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestWaitUntil(t *testing.T) { //nolint:tparallel
	ctx := acctest.Context(t)
	t.Parallel()

	var retryCount int32

	testCases := []struct {
		Name        string
		F           func() (bool, error)
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func() (bool, error) {
				return true, nil
			},
		},
		{
			Name: "immediate error",
			F: func() (bool, error) {
				return false, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "never reaches state",
			F: func() (bool, error) {
				return false, nil
			},
			ExpectError: true,
		},
		{
			Name: "retry then success",
			F: func() (bool, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return true, nil
				}

				return false, nil
			},
		},
	}

	for _, testCase := range testCases { //nolint:paralleltest
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			err := tfresource.WaitUntil(ctx, 5*time.Second, testCase.F, tfresource.WaitOpts{})

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}
