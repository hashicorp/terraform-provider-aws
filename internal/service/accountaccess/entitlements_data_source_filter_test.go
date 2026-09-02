// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const entitlementsDataSourceTestAccountID = "123456789012" // nosemgrep:ci.literal-12Digit-string-test-constant

var entitlementsDataSourceTestRoleARN = awsarn.ARN{
	Partition: endpoints.AwsPartitionID,
	Service:   names.IAMServiceID,
	AccountID: entitlementsDataSourceTestAccountID,
	Resource:  "role/example",
}.String()

func TestEntitlementsDataSourcePrincipalRoleFilter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		principal identityCenterPrincipalModel
		wantType  string
		wantID    string
	}{
		"group": {
			principal: identityCenterPrincipalModel{GroupID: types.StringValue("group-id")},
			wantType:  "group",
			wantID:    "group-id",
		},
		"user": {
			principal: identityCenterPrincipalModel{UserID: types.StringValue("user-id")},
			wantType:  "user",
			wantID:    "user-id",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data := entitlementsDataSourceModel{
				AccountID: types.StringValue(entitlementsDataSourceTestAccountID),
				Principal: testEntitlementsDataSourcePrincipal(ctx, t, tc.principal),
				RoleARN:   fwtypes.ARNValue(entitlementsDataSourceTestRoleARN),
			}

			filter, diags := entitlementsDataSourcePrincipalRoleFilter(ctx, &data)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if got, want := aws.ToString(filter.Account), entitlementsDataSourceTestAccountID; got != want {
				t.Errorf("Account = %q, want %q", got, want)
			}
			if got, want := aws.ToString(filter.RoleArn), entitlementsDataSourceTestRoleARN; got != want {
				t.Errorf("RoleArn = %q, want %q", got, want)
			}

			principal, ok := filter.Principal.(*awstypes.PrincipalFilterMemberIdentityCenter)
			if !ok {
				t.Fatalf("Principal = %T, want *PrincipalFilterMemberIdentityCenter", filter.Principal)
			}
			switch value := principal.Value.(type) {
			case *awstypes.IdentityCenterPrincipalFilterMemberGroupId:
				if tc.wantType != "group" || value.Value != tc.wantID {
					t.Errorf("group filter = %#v, want %q", value, tc.wantID)
				}
			case *awstypes.IdentityCenterPrincipalFilterMemberUserId:
				if tc.wantType != "user" || value.Value != tc.wantID {
					t.Errorf("user filter = %#v, want %q", value, tc.wantID)
				}
			default:
				t.Errorf("Identity Center filter = %T, want %s", value, tc.wantType)
			}
		})
	}
}

func testEntitlementsDataSourcePrincipal(ctx context.Context, t *testing.T, identityCenter identityCenterPrincipalModel) fwtypes.ListNestedObjectValueOf[principalModel] {
	t.Helper()

	identityCenters, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []identityCenterPrincipalModel{identityCenter})
	if diags.HasError() {
		t.Fatalf("creating identity center value: %v", diags)
	}
	principals, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []principalModel{{IdentityCenter: identityCenters}})
	if diags.HasError() {
		t.Fatalf("creating principal value: %v", diags)
	}

	return principals
}
