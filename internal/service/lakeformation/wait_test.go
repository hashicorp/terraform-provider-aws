// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
)

const testWaitPermissionsReadyTimeout = 5 * time.Second

func testPermissionsRefresh(results ...func() (any, string, error)) retry.StateRefreshFunc {
	var n int

	return func(ctx context.Context) (any, string, error) {
		// Mirror statusPermissions: a request made with an expired context fails.
		if err := ctx.Err(); err != nil {
			return nil, "", fmt.Errorf("listing permissions: %w", err)
		}

		result := results[min(n, len(results)-1)]
		n++

		return result()
	}
}

func testPermissionsFound() (any, string, error) {
	return []awstypes.PrincipalResourcePermissions{
		{
			Permissions: []awstypes.Permission{awstypes.PermissionAlter},
			Principal: &awstypes.DataLakePrincipal{
				DataLakePrincipalIdentifier: aws.String("arn:aws:iam::123456789012:role/example"),
			},
		},
	}, "AVAILABLE", nil
}

func testPermissionsEmpty() (any, string, error) {
	return nil, "", nil
}

func testPermissionsIAMDelay() (any, string, error) {
	return nil, "IAM DELAY", nil
}

func TestWaitPermissionsReadyWithRefresh(t *testing.T) {
	t.Parallel()

	errThrottled := errors.New("listing permissions: ThrottlingException: Rate exceeded")

	testCases := map[string]struct {
		refresh             retry.StateRefreshFunc
		expectedPermissions int
		expectNotFound      bool
		expectErr           bool
	}{
		"found": {
			refresh:             testPermissionsRefresh(testPermissionsFound),
			expectedPermissions: 1,
		},
		"found after empty results": {
			refresh:             testPermissionsRefresh(testPermissionsEmpty, testPermissionsEmpty, testPermissionsFound),
			expectedPermissions: 1,
		},
		"consistently empty": {
			refresh:        testPermissionsRefresh(testPermissionsEmpty),
			expectNotFound: true,
			expectErr:      true,
		},
		"consistently invalid principal": {
			refresh:        testPermissionsRefresh(testPermissionsIAMDelay),
			expectNotFound: true,
			expectErr:      true,
		},
		"no request completes before timeout": {
			refresh: func(ctx context.Context) (any, string, error) {
				<-ctx.Done()
				return nil, "", fmt.Errorf("listing permissions: %w", ctx.Err())
			},
			expectErr: true,
		},
		"empty then no request completes before timeout": {
			refresh: func() retry.StateRefreshFunc {
				var n int

				return func(ctx context.Context) (any, string, error) {
					n++
					if n == 1 {
						return testPermissionsEmpty()
					}

					<-ctx.Done()
					return nil, "", fmt.Errorf("listing permissions: %w", ctx.Err())
				}
			}(),
			expectErr: true,
		},
		"request error": {
			refresh: testPermissionsRefresh(func() (any, string, error) {
				return nil, "", errThrottled
			}),
			expectErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			permissions, err := tflakeformation.WaitPermissionsReadyWithRefresh(t.Context(), testCase.refresh, testWaitPermissionsReadyTimeout)

			if got, want := len(permissions), testCase.expectedPermissions; got != want {
				t.Errorf("len(permissions) = %d, want %d", got, want)
			}

			if got, want := err != nil, testCase.expectErr; got != want {
				t.Errorf("err = %v, want error: %t", err, want)
			}

			if got, want := retry.NotFound(err), testCase.expectNotFound; got != want {
				t.Errorf("NotFound(%v) = %t, want %t", err, got, want)
			}

			if got, want := tflakeformation.PermissionsNotFound(err), testCase.expectNotFound; got != want {
				t.Errorf("PermissionsNotFound(%v) = %t, want %t", err, got, want)
			}
		})
	}
}

func TestPermissionsNotFound(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		err      error
		expected bool
	}{
		"nil": {
			err:      nil,
			expected: false,
		},
		"not found": {
			err:      &retry.NotFoundError{},
			expected: true,
		},
		"entity not found": {
			err:      fmt.Errorf("listing permissions: %w", &awstypes.EntityNotFoundException{Message: aws.String("not found")}),
			expected: true,
		},
		"access denied resource does not exist": {
			err:      fmt.Errorf("listing permissions: %w", &awstypes.AccessDeniedException{Message: aws.String("Resource does not exist or requester is not authorized to access requested permissions.")}),
			expected: true,
		},
		"access denied": {
			err:      fmt.Errorf("listing permissions: %w", &awstypes.AccessDeniedException{Message: aws.String("Insufficient Lake Formation permission(s)")}),
			expected: false,
		},
		"timeout": {
			err:      &retry.TimeoutError{LastError: context.DeadlineExceeded},
			expected: false,
		},
		"internal service error": {
			err:      fmt.Errorf("listing permissions: %w", &awstypes.InternalServiceException{Message: aws.String("internal error")}),
			expected: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got, want := tflakeformation.PermissionsNotFound(testCase.err), testCase.expected; got != want {
				t.Errorf("PermissionsNotFound(%v) = %t, want %t", testCase.err, got, want)
			}
		})
	}
}
