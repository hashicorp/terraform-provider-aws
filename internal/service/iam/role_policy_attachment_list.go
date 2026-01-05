// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_iam_role_policy_attachment")
func newRolePolicyAttachmentResourceAsListResource() inttypes.ListResourceForSDK {
	l := rolePolicyAttachmentListResource{}
	l.SetResourceSchema(resourceRolePolicyAttachment())

	return &l
}

var _ list.ListResourceWithRawV5Schemas = &rolePolicyAttachmentListResource{}

type rolePolicyAttachmentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type rolePolicyAttachmentListResourceModel struct {
}

func (l *rolePolicyAttachmentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.IAMClient(ctx)

	var query rolePolicyAttachmentListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input iam.ListRolesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for role, err := range listNonServiceLinkedRoles(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			tflog.Info(ctx, "Listing attached policies for role", map[string]any{
				logging.ResourceAttributeKey(names.AttrRole): aws.ToString(role.RoleName),
			})

			input := iam.ListAttachedRolePoliciesInput{
				RoleName: role.RoleName,
			}
			pages := iam.NewListAttachedRolePoliciesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if errs.IsA[*awstypes.NoSuchEntityException](err) {
					tflog.Warn(ctx, "Resource disappeared during listing, skipping", map[string]any{
						logging.ResourceAttributeKey(names.AttrRole): aws.ToString(role.RoleName),
					})
					continue
				}
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}

				for _, attachedPolicy := range page.AttachedPolicies {
					ctx := resourceRolePolicyAttachmentListItemLoggingContext(ctx, role, attachedPolicy)

					result := request.NewListResult(ctx)

					rd := l.ResourceData()
					rd.SetId(resourceRolePolicyAttachmentImportIDFromData(role, attachedPolicy))

					tflog.Info(ctx, "Reading resource")
					rd.Set(names.AttrRole, aws.ToString(role.RoleName))
					rd.Set("policy_arn", aws.ToString(attachedPolicy.PolicyArn))

					result.DisplayName = resourceRolePolicyAttachmentDisplayName(role, attachedPolicy)

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
	}
}

func resourceRolePolicyAttachmentListItemLoggingContext(ctx context.Context, role awstypes.Role, attachedPolicy awstypes.AttachedPolicy) context.Context {
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrRole), aws.ToString(role.RoleName))
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("policy_arn"), aws.ToString(attachedPolicy.PolicyArn))
	return ctx
}

func resourceRolePolicyAttachmentImportIDFromData(role awstypes.Role, attachedPolicy awstypes.AttachedPolicy) string {
	return (rolePolicyAttachmentImportID{}).create(aws.ToString(role.RoleName), aws.ToString(attachedPolicy.PolicyArn))
}

func resourceRolePolicyAttachmentDisplayName(role awstypes.Role, attachedPolicy awstypes.AttachedPolicy) string {
	foo := aws.ToString(attachedPolicy.PolicyArn)
	policyARN, err := arn.Parse(foo)
	if err == nil {
		foo = strings.TrimPrefix(policyARN.Resource, "policy/")
	}

	return fmt.Sprintf("Role: %s - Policy: %s", resourceRoleDisplayName(role), foo)
}
