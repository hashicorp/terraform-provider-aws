// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/aws/smithy-go"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// Verifies the Create retry predicate against the real error shape observed from
// CreateRule on a not-yet-onboarded account, and confirms an unrelated
// ConflictException is NOT retried so a genuine conflict still surfaces.
func TestRuleCreateOnboardingRetryPredicate(t *testing.T) {
	t.Parallel()

	opErr := func(msg string) error {
		return &smithy.OperationError{
			ServiceID:     "Network Security Manager",
			OperationName: "CreateRule",
			Err:           &awstypes.ConflictException{Message: aws.String(msg)},
		}
	}

	t.Run("onboarding conflict is retried then succeeds", func(t *testing.T) {
		t.Parallel()
		calls := 0
		_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ConflictException](
			context.Background(), 30*time.Second,
			func(context.Context) (any, error) {
				calls++
				if calls == 1 {
					return nil, opErr("Account or org is already onboarded")
				}
				return struct{}{}, nil
			},
			"already onboarded",
		)
		if err != nil {
			t.Fatalf("expected retry to succeed, got %v", err)
		}
		if calls < 2 {
			t.Fatalf("expected a retry, got %d call(s)", calls)
		}
	})

	t.Run("unrelated conflict is not retried", func(t *testing.T) {
		t.Parallel()
		calls := 0
		_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ConflictException](
			context.Background(), 30*time.Second,
			func(context.Context) (any, error) {
				calls++
				return nil, opErr("Rule with name foo already exists")
			},
			"already onboarded",
		)
		if err == nil {
			t.Fatal("expected the unrelated conflict to surface")
		}
		if calls != 1 {
			t.Fatalf("expected no retry, got %d calls", calls)
		}
	})
}
