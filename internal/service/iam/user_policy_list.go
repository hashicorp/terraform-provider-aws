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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_iam_user_policy")
func newUserPolicyResourceAsListResource() inttypes.ListResourceForSDK {
	l := userPolicyListResource{}
	l.SetResourceSchema(resourceUserPolicy())
	return &l
}

var _ list.ListResource = &userPolicyListResource{}

type userPolicyListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *userPolicyListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrUserName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the IAM user to list policies from.",
			},
		},
	}
}

func (l *userPolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.IAMClient(ctx)

	var query listUserPolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	userName := query.UserName.ValueString()

	tflog.Info(ctx, "Listing IAM User Policy")
	stream.Results = func(yield func(list.ListResult) bool) {
		input := &iam.ListUserPoliciesInput{
			UserName: aws.String(userName),
		}
		for policyName, err := range listUserPolicies(ctx, conn, input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := userPolicyCreateResourceID(userName, policyName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)
			rd.Set("user", userName)
			rd.Set(names.AttrName, policyName)

			if request.IncludeResource {
				policyDocument, err := findUserPolicyByTwoPartKey(ctx, conn, userName, policyName)
				if err != nil {
					tflog.Error(ctx, "Reading IAM User Policy", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if err := resourceUserPolicyFlatten(rd, userName, policyName, policyDocument); err != nil {
					tflog.Error(ctx, "Reading IAM User Policy", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = policyName

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

type listUserPolicyModel struct {
	UserName types.String `tfsdk:"user_name"`
}

func listUserPolicies(ctx context.Context, conn *iam.Client, input *iam.ListUserPoliciesInput) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		pages := iam.NewListUserPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield("", fmt.Errorf("listing IAM User Policy resources: %w", err))
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
