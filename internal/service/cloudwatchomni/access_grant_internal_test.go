// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

func TestFlattenExpandAccessGrantRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testCases := map[string]struct {
		apiObject awstypes.AccessGrant
		want      cloudwatchomni.CreateAccessGrantInput
	}{
		"minimal": {
			apiObject: awstypes.AccessGrant{
				DomainId:   aws.String("d-0123456789"),
				Name:       aws.String("test"),
				Permission: awstypes.AccessGrantPermissionRead,
				Principal: &awstypes.AccessGrantPrincipal{
					PrincipalId:         aws.String("a1b2c3d4-5678-90ab-cdef-EXAMPLE11111"),
					PrincipalType:       awstypes.AccessGrantPrincipalTypeIdcGroup,
					PrincipalAttributes: []awstypes.AccessGrantPrincipalAttribute{},
				},
				ScopedActions: []awstypes.ScopedActions{},
				SpaceId:       aws.String("11111111-2222-4333-8444-555555555555"),
			},
			want: cloudwatchomni.CreateAccessGrantInput{
				DomainId:   aws.String("d-0123456789"),
				Name:       aws.String("test"),
				Permission: awstypes.AccessGrantPermissionRead,
				Principal: &awstypes.AccessGrantPrincipal{
					PrincipalId:   aws.String("a1b2c3d4-5678-90ab-cdef-EXAMPLE11111"),
					PrincipalType: awstypes.AccessGrantPrincipalTypeIdcGroup,
				},
				SpaceId: aws.String("11111111-2222-4333-8444-555555555555"),
			},
		},
		"scoped actions": {
			apiObject: awstypes.AccessGrant{
				DomainId:   aws.String("d-0123456789"),
				Name:       aws.String("test"),
				Permission: awstypes.AccessGrantPermissionCustom,
				Principal: &awstypes.AccessGrantPrincipal{
					PrincipalType: awstypes.AccessGrantPrincipalTypeIdcUser,
					PrincipalAttributes: []awstypes.AccessGrantPrincipalAttribute{
						{Key: aws.String("enterprise.department"), Value: aws.String("engineering")},
					},
				},
				ScopedActions: []awstypes.ScopedActions{
					{
						Actions: []string{"cloudwatch:GetOmniDashboard"},
					},
					{
						Actions:           []string{"cloudwatch:GetRecords"},
						ContextConditions: map[string][]string{"channelNames": {"a", "b"}},
						Resources: []awstypes.ResourceScope{
							{
								ResourceType: aws.String("DataSet"),
								ResourceArns: []string{},
								RowScopeGroups: [][]awstypes.RowScope{
									{
										{Field: aws.String("serviceName"), Operator: awstypes.RowScopeOperatorIn, Values: []string{"y", "x"}},
										{Field: aws.String("accountId"), Operator: awstypes.RowScopeOperatorIn, Values: []string{"example"}},
									},
									{
										{Field: aws.String("serviceName"), Operator: awstypes.RowScopeOperatorIn, Values: []string{"z"}},
									},
								},
								SignalTypes: []awstypes.SignalType{awstypes.SignalTypeLogs, awstypes.SignalTypeTraces},
								Tags:        map[string]string{"env": "test"},
							},
						},
					},
				},
				SpaceId: aws.String("11111111-2222-4333-8444-555555555555"),
			},
			want: cloudwatchomni.CreateAccessGrantInput{
				DomainId:   aws.String("d-0123456789"),
				Name:       aws.String("test"),
				Permission: awstypes.AccessGrantPermissionCustom,
				Principal: &awstypes.AccessGrantPrincipal{
					PrincipalType: awstypes.AccessGrantPrincipalTypeIdcUser,
					PrincipalAttributes: []awstypes.AccessGrantPrincipalAttribute{
						{Key: aws.String("enterprise.department"), Value: aws.String("engineering")},
					},
				},
				ScopedActions: []awstypes.ScopedActions{
					{
						Actions: []string{"cloudwatch:GetOmniDashboard"},
					},
					{
						Actions: []string{"cloudwatch:GetRecords"},
						Resources: []awstypes.ResourceScope{
							{
								ResourceType: aws.String("DataSet"),
								RowScopeGroups: [][]awstypes.RowScope{
									{
										{Field: aws.String("serviceName"), Operator: awstypes.RowScopeOperatorIn, Values: []string{"x", "y"}},
										{Field: aws.String("accountId"), Operator: awstypes.RowScopeOperatorIn, Values: []string{"example"}},
									},
									{
										{Field: aws.String("serviceName"), Operator: awstypes.RowScopeOperatorIn, Values: []string{"z"}},
									},
								},
								SignalTypes: []awstypes.SignalType{awstypes.SignalTypeLogs, awstypes.SignalTypeTraces},
								Tags:        map[string]string{"env": "test"},
							},
						},
					},
				},
				SpaceId: aws.String("11111111-2222-4333-8444-555555555555"),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var r accessGrantResource
			var data accessGrantResourceModel
			if diags := r.flatten(ctx, &testCase.apiObject, &data); diags.HasError() {
				t.Fatalf("flatten: %v", diags)
			}

			var got cloudwatchomni.CreateAccessGrantInput
			if diags := fwflex.Expand(ctx, data, &got); diags.HasError() {
				t.Fatalf("expand: %v", diags)
			}

			// Row scope values are a set, so their expanded order is not significant.
			for _, scopedAction := range got.ScopedActions {
				for _, scope := range scopedAction.Resources {
					for _, group := range scope.RowScopeGroups {
						for _, rowScope := range group {
							slices.Sort(rowScope.Values)
						}
					}
				}
			}

			if diff := cmp.Diff(testCase.want, got, cmpopts.IgnoreUnexported(
				cloudwatchomni.CreateAccessGrantInput{},
				awstypes.AccessGrantPrincipal{},
				awstypes.AccessGrantPrincipalAttribute{},
				awstypes.ScopedActions{},
				awstypes.ResourceScope{},
				awstypes.RowScope{},
			)); diff != "" {
				t.Errorf("unexpected diff (+got, -want): %s", diff)
			}
		})
	}
}
