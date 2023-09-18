// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func TestAddErrorPredicateRetrier(t *testing.T) {
	t.Parallel()

	f := func(err error) bool {
		return errs.Contains(err, "testing")
	}
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "no error",
		},
		{
			name: "non-retryable",
			err:  errors.New(`this is not retryable`),
		},
		{
			name:     "retryable",
			err:      errors.New(`this is testing`),
			expected: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := AddErrorPredicateRetrier(retry.NewStandard(), f).IsErrorRetryable(testCase.err)
			if got, want := got, testCase.expected; got != want {
				t.Errorf("IsErrorRetryable(%q) = %v, want %v", testCase.err, got, want)
			}
		})
	}
}
