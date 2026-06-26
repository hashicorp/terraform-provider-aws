// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_iam_group_policy_attachment")
func newGroupPolicyAttachmentResourceAsListResource() inttypes.ListResourceForSDK {
	l := groupPolicyAttachmentListResource{}
	l.SetResourceSchema(resourceGroupPolicyAttachment())
	return &l
}

var _ list.ListResource = &groupPolicyAttachmentListResource{}

type groupPolicyAttachmentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type listGroupPolicyAttachmentModel struct {
}

func (l *groupPolicyAttachmentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().IAMClient(ctx)

	var query listGroupPolicyAttachmentModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input iam.ListGroupsInput
	tflog.Info(ctx, "Listing Resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for group, err := range listGroups(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			groupName := aws.ToString(group.GroupName)
			attachInput := iam.ListAttachedGroupPoliciesInput{
				GroupName: aws.String(groupName),
			}
			for item, err := range listGroupPolicyAttachments(ctx, conn, &attachInput) {
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}
				policyARN := aws.ToString(item.PolicyArn)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey("group"), groupName)
				ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("policy_arn"), policyARN)

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.Set("group", groupName)
				rd.Set("policy_arn", policyARN)
				rd.SetId(groupPolicyAttachmentImportID{}.Create(rd))

				if request.IncludeResource { //nolint:revive,staticcheck // Be explicit about IncludeResource handling
					// No-op, all attributes already set
				}

				result.DisplayName = fmt.Sprintf("Group: %s - Policy: %s", groupName, policyARN)

				l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

func listGroups(ctx context.Context, conn *iam.Client, input *iam.ListGroupsInput) iter.Seq2[awstypes.Group, error] {
	return func(yield func(awstypes.Group, error) bool) {
		pages := iam.NewListGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Group{}, err)
				return
			}

			for _, role := range page.Groups {
				if !yield(role, nil) {
					return
				}
			}
		}
	}
}

func listGroupPolicyAttachments(ctx context.Context, conn *iam.Client, input *iam.ListAttachedGroupPoliciesInput) iter.Seq2[awstypes.AttachedPolicy, error] {
	return func(yield func(awstypes.AttachedPolicy, error) bool) {
		pages := iam.NewListAttachedGroupPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AttachedPolicy{}, fmt.Errorf("listing IAM Group Policy Attachment resources: %w", err))
				return
			}

			for _, item := range page.AttachedPolicies {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
