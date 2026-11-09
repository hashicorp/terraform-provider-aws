// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

const testAdminAccountID = "123456789012" // nosemgrep:ci.literal-12Digit-string-test-constant -- same-package resource tests cannot import acctest without a provider import cycle

func TestAdminAccountAccountIDMapping(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := adminAccountResourceModel{
		AccountID:  types.StringValue(testAdminAccountID),
		AdminScope: fwtypes.NewListNestedObjectValueOfNull[adminScopeModel](ctx),
		Priority:   types.Int32Value(1),
	}

	var input networksecuritymanager.PutAdminAccountInput
	if diags := flex.Expand(ctx, model, &input); diags.HasError() {
		t.Fatalf("expanding admin account: %v", diags)
	}
	if got := aws.ToString(input.AccountId); got != model.AccountID.ValueString() {
		t.Fatalf("AccountId = %q, want %q", got, model.AccountID.ValueString())
	}

	// GetAdminAccount can omit the account ID; configuration or import retains it.
	output := awstypes.AdminAccountDetails{
		Priority: aws.Int32(1),
		Status:   awstypes.AdminAccountStatusOnboarded,
	}
	r := &adminAccountResource{}
	if diags := r.flatten(ctx, &output, &model); diags.HasError() {
		t.Fatalf("flattening admin account: %v", diags)
	}
	if got := model.AccountID.ValueString(); got != testAdminAccountID {
		t.Fatalf("account_id = %q, want %q", got, testAdminAccountID)
	}
}

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

func TestScopeFilterModelExpandExcludeOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	selection := &scopeSelectionModel{
		Accounts:            fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{types.StringValue(testAdminAccountID)}),
		OrganizationalUnits: fwtypes.NewSetValueOfNull[types.String](ctx),
	}

	for name, includeOnly := range map[string]fwtypes.ListNestedObjectValueOf[scopeSelectionModel]{
		"include_only null":  fwtypes.NewListNestedObjectValueOfNull[scopeSelectionModel](ctx),
		"include_only empty": fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []scopeSelectionModel{}),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := scopeFilterModel{
				ExcludeOnly: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, selection),
				IncludeAll:  types.BoolNull(),
				IncludeOnly: includeOnly,
			}

			output, diags := model.Expand(ctx)
			if diags.HasError() {
				t.Fatalf("expanding scope filter: %v", diags)
			}

			apiObject, ok := output.(*awstypes.AdminScopeFilterInputMemberExcludeOnly)
			if !ok {
				t.Fatalf("output = %T, want *AdminScopeFilterInputMemberExcludeOnly", output)
			}
			if got := apiObject.Value.Accounts; len(got) != 1 || got[0] != testAdminAccountID {
				t.Fatalf("exclude_only accounts = %v, want [%s]", got, testAdminAccountID)
			}
		})
	}
}
