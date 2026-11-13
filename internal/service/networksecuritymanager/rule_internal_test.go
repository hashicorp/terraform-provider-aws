// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/aws/smithy-go"
)

func TestRetryWhenOnboarding(t *testing.T) {
	t.Parallel()

	// The error as CreateRule returns it.
	conflict := func(message string) error {
		return &smithy.OperationError{
			ServiceID:     "Network Security Manager",
			OperationName: "CreateRule",
			Err:           &awstypes.ConflictException{Message: aws.String(message)},
		}
	}

	t.Run("onboarding conflict is retried", func(t *testing.T) {
		t.Parallel()

		calls := 0
		got, err := retryWhenOnboarding(t.Context(), func(context.Context) (string, error) {
			calls++
			if calls < 3 {
				return "", conflict("Account or org is already onboarded")
			}
			return "created", nil
		})

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if calls != 3 {
			t.Errorf("calls = %d, want 3", calls)
		}
		// The result of the call that succeeded is returned.
		if got != "created" {
			t.Errorf("result = %q, want %q", got, "created")
		}
	})

	t.Run("other conflict is not retried", func(t *testing.T) {
		t.Parallel()

		calls := 0
		_, err := retryWhenOnboarding(t.Context(), func(context.Context) (string, error) {
			calls++
			return "", conflict("Invalid update token - entity was modified by another process")
		})

		if err == nil {
			t.Fatal("expected an error")
		}
		if calls != 1 {
			t.Errorf("calls = %d, want 1", calls)
		}
	})

	t.Run("same message on another error type is not retried", func(t *testing.T) {
		t.Parallel()

		calls := 0
		_, err := retryWhenOnboarding(t.Context(), func(context.Context) (string, error) {
			calls++
			return "", &awstypes.ValidationException{Message: aws.String("Account or org is already onboarded")}
		})

		if err == nil {
			t.Fatal("expected an error")
		}
		if calls != 1 {
			t.Errorf("calls = %d, want 1", calls)
		}
	})
}
