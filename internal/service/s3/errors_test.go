// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func TestErrBucketRegionMismatch(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		err         error
		wantWrapped bool
	}{
		"nil": {
			err: nil,
		},
		"unrelated": {
			err: errs.APIError(errCodeAccessDenied, "Access Denied"),
		},
		"permanent redirect": {
			err:         errs.APIError(errCodePermanentRedirect, "The bucket is in this region"),
			wantWrapped: true,
		},
		"authorization header malformed": {
			err:         errs.APIError(errCodeAuthorizationHeaderMalformed, "The region is wrong"),
			wantWrapped: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := errBucketRegionMismatch("example-bucket", testCase.err)

			if testCase.err == nil {
				if got != nil {
					t.Fatalf("expected no error, got %s", got)
				}
				return
			}

			if !errors.Is(got, testCase.err) {
				t.Fatalf("expected original error to be preserved, got %s", got)
			}

			wantMessage := "S3 Bucket (example-bucket) was redirected to another Region; verify that the bucket name is the actual S3 bucket name and that the resource's `region` argument matches the bucket Region"
			if gotWrapped := strings.Contains(got.Error(), wantMessage); gotWrapped != testCase.wantWrapped {
				t.Fatalf("expected wrapped = %t, got %q", testCase.wantWrapped, got)
			}
		})
	}
}
