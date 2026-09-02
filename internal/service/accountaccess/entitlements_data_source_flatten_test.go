// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
)

func TestFlattenEntitlementSummary(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var model entitlementsDataSourceItemModel

	diags := flattenEntitlementSummary(ctx, awstypes.EntitlementsListMember{
		CreatedAt:     aws.Time(time.Date(2026, time.September, 2, 1, 0, 0, 0, time.UTC)),
		EntitlementId: aws.String("entitlement-id"),
		Entitlement: &awstypes.EntitlementSummaryMemberPrincipalRole{Value: awstypes.PrincipalRoleEntitlementSummary{
			Account:     aws.String(entitlementsDataSourceTestAccountID),
			AccountName: aws.String("example-account"),
			Principal: &awstypes.PrincipalMemberIdentityCenter{Value: &awstypes.IdentityCenterPrincipalMemberUserId{
				Value: "user-id",
			}},
			RoleArn: aws.String(entitlementsDataSourceTestRoleARN),
		}},
	}, &model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got, want := model.EntitlementID.ValueString(), "entitlement-id"; got != want {
		t.Errorf("EntitlementID = %q, want %q", got, want)
	}

	entitlement, diags := model.Entitlement.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading entitlement: %v", diags)
	}
	principalRole, diags := entitlement.PrincipalRole.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading principal role: %v", diags)
	}
	principal, diags := principalRole.Principal.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading principal: %v", diags)
	}
	identityCenter, diags := principal.IdentityCenter.ToPtr(ctx)
	if diags.HasError() {
		t.Fatalf("reading identity center: %v", diags)
	}

	if got, want := principalRole.Account.ValueString(), entitlementsDataSourceTestAccountID; got != want {
		t.Errorf("Account = %q, want %q", got, want)
	}
	if got, want := principalRole.AccountName.ValueString(), "example-account"; got != want {
		t.Errorf("AccountName = %q, want %q", got, want)
	}
	if got, want := principalRole.RoleARN.ValueString(), entitlementsDataSourceTestRoleARN; got != want {
		t.Errorf("RoleARN = %q, want %q", got, want)
	}
	if got, want := identityCenter.UserID.ValueString(), "user-id"; got != want {
		t.Errorf("UserID = %q, want %q", got, want)
	}
	if !identityCenter.GroupID.IsNull() {
		t.Errorf("GroupID = %q, want null", identityCenter.GroupID.ValueString())
	}
}
