// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_iam_role_policy")
func newRolePolicyResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceRolePolicy{}
	l.SetResourceSchema(resourceRolePolicy())
	return &l
}

var _ list.ListResource = &listResourceRolePolicy{}

type listResourceRolePolicy struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceRolePolicy) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"role_name": listschema.StringAttribute{
				Required:    true,
				Description: "Name of the IAM role to list policies from.",
			},
		},
	}
}

func (l *listResourceRolePolicy) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().IAMClient(ctx)
	awsClient := l.Meta()

	var query listRolePolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	roleName := query.RoleName.ValueString()

	tflog.Info(ctx, "Listing IAM (Identity & Access Management) Role Policy")
	stream.Results = func(yield func(list.ListResult) bool) {
		input := &iam.ListRolePoliciesInput{
			RoleName: aws.String(roleName),
		}
		for policyName, err := range listRolePolicies(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := createRolePolicyImportID(roleName, policyName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)

			tflog.Info(ctx, "Reading IAM (Identity & Access Management) Role Policy")
			diags := resourceRolePolicyRead(ctx, rd, awsClient)
			if diags.HasError() {
				tflog.Error(ctx, "Reading IAM (Identity & Access Management) Role Policy", map[string]any{
					names.AttrID: id,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}
			if rd.Id() == "" {
				// Resource is logically deleted
				continue
			}

			result.DisplayName = policyName

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listRolePolicyModel struct {
	RoleName types.String `tfsdk:"role_name"`
}

func listRolePolicies(ctx context.Context, conn *iam.Client, input *iam.ListRolePoliciesInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := iam.NewListRolePoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield("", fmt.Errorf("listing IAM (Identity & Access Management) Role Policy resources: %w", err))
				return
			}

			for _, policyName := range page.PolicyNames {
				if !yield(policyName, nil) {
					return
				}
			}
		}
	}
}
