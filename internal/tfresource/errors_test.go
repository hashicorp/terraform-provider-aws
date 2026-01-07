// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package tfresource_test

import (
	"errors"
	"strings"
	"testing"

	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestSetLastError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name     string
		Err      error
		LastErr  error
		Expected bool
	}{
		{
			Name: "nil error",
		},
		{
			Name:    "other error",
			Err:     errors.New("test"),
			LastErr: errors.New("last"),
		},
		{
			Name: "timeout error lastErr is nil",
			Err:  &sdkretry.TimeoutError{},
		},
		{
			Name:     "timeout error",
			Err:      &sdkretry.TimeoutError{},
			LastErr:  errors.New("lasttest"),
			Expected: true,
		},
		{
			Name: "timeout error non-nil last error lastErr is nil",
			Err:  &sdkretry.TimeoutError{LastError: errors.New("test")},
		},
		{
			Name:    "timeout error non-nil last error no overwrite",
			Err:     &sdkretry.TimeoutError{LastError: errors.New("test")},
			LastErr: errors.New("lasttest"),
		},
		{
			Name: "unexpected state error lastErr is nil",
			Err:  &sdkretry.UnexpectedStateError{},
		},
		{
			Name:     "unexpected state error",
			Err:      &sdkretry.UnexpectedStateError{},
			LastErr:  errors.New("lasttest"),
			Expected: true,
		},
		{
			Name: "unexpected state error non-nil last error lastErr is nil",
			Err:  &sdkretry.UnexpectedStateError{LastError: errors.New("test")},
		},
		{
			Name:    "unexpected state error non-nil last error no overwrite",
			Err:     &sdkretry.UnexpectedStateError{LastError: errors.New("test")},
			LastErr: errors.New("lasttest"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			tfresource.SetLastError(testCase.Err, testCase.LastErr)

			if testCase.Err != nil {
				got := testCase.Err.Error()
				contains := strings.Contains(got, "lasttest")

				if (testCase.Expected && !contains) || (!testCase.Expected && contains) {
					t.Errorf("got %s", got)
				}
			}
		})
	}
}
