// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_iam_user_policy_attachment")
func newUserPolicyAttachmentResourceAsListResource() inttypes.ListResourceForSDK {
	l := userPolicyAttachmentListResource{}
	l.SetResourceSchema(resourceUserPolicyAttachment())

	return &l
}

var _ list.ListResource = &userPolicyAttachmentListResource{}
var _ list.ListResourceWithRawV5Schemas = &userPolicyAttachmentListResource{}

type userPolicyAttachmentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *userPolicyAttachmentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.IAMClient(ctx)

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for user, err := range listUsers(ctx, conn, &iam.ListUsersInput{}) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			tflog.Info(ctx, "Listing attached policies for user", map[string]any{
				logging.ResourceAttributeKey("user"): aws.ToString(user.UserName),
			})

			input := iam.ListAttachedUserPoliciesInput{
				UserName: user.UserName,
			}
			pages := iam.NewListAttachedUserPoliciesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if errs.IsA[*awstypes.NoSuchEntityException](err) {
					tflog.Warn(ctx, "Resource disappeared during listing, skipping", map[string]any{
						logging.ResourceAttributeKey("user"): aws.ToString(user.UserName),
					})
					break
				}
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}

				for _, attachedPolicy := range page.AttachedPolicies {
					ctx := resourceUserPolicyAttachmentListItemLoggingContext(ctx, user, attachedPolicy)

					result := request.NewListResult(ctx)

					rd := l.ResourceData()
					rd.SetId(resourceUserPolicyAttachmentImportIDFromData(user, attachedPolicy))
					rd.Set("user", aws.ToString(user.UserName))
					rd.Set("policy_arn", aws.ToString(attachedPolicy.PolicyArn))

					if request.IncludeResource {
						resourceUserPolicyAttachmentFlatten(rd, aws.ToString(user.UserName), &attachedPolicy)
					}

					result.DisplayName = resourceUserPolicyAttachmentDisplayName(user, attachedPolicy)

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
	}
}

func resourceUserPolicyAttachmentListItemLoggingContext(ctx context.Context, user awstypes.User, attachedPolicy awstypes.AttachedPolicy) context.Context {
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("user"), aws.ToString(user.UserName))
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("policy_arn"), aws.ToString(attachedPolicy.PolicyArn))
	return ctx
}

func resourceUserPolicyAttachmentImportIDFromData(user awstypes.User, attachedPolicy awstypes.AttachedPolicy) string {
	return (userPolicyAttachmentImportID{}).create(aws.ToString(user.UserName), aws.ToString(attachedPolicy.PolicyArn))
}

func resourceUserPolicyAttachmentDisplayName(user awstypes.User, attachedPolicy awstypes.AttachedPolicy) string {
	foo := aws.ToString(attachedPolicy.PolicyArn)
	policyARN, err := arn.Parse(foo)
	if err == nil {
		foo = strings.TrimPrefix(policyARN.Resource, "policy/")
	}

	return fmt.Sprintf("User: %s - Policy: %s", aws.ToString(user.UserName), foo)
}
