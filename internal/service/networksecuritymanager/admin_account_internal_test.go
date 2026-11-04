// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
)

func TestIsAccountNotOnboardedError(t *testing.T) {
	t.Parallel()

	notOnboarded := &awstypes.ValidationException{
		Message: aws.String("The account is not onboarded to Network Security Manager. Create a resource to onboard the account, or contact your organization administrator."),
	}

	testCases := map[string]struct {
		err      error
		expected bool
	}{
		"nil": {
			err: nil,
		},
		"not onboarded": {
			err:      notOnboarded,
			expected: true,
		},
		"not onboarded, wrapped": {
			err:      smarterr.NewError(fmt.Errorf("listing: %w", notOnboarded)),
			expected: true,
		},
		"other validation error": {
			err: &awstypes.ValidationException{Message: aws.String("1 validation error detected: Value at 'priority' failed to satisfy constraint")},
		},
		"same message, other error type": {
			err: &awstypes.ConflictException{Message: notOnboarded.Message},
		},
		"unrelated error": {
			err: errors.New("is not onboarded to Network Security Manager"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := isAccountNotOnboardedError(tc.err); got != tc.expected {
				t.Errorf("isAccountNotOnboardedError() = %t, want %t", got, tc.expected)
			}
		})
	}
}
