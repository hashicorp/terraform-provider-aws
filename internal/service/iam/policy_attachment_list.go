// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"iter"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_iam_policy_attachment")
func newPolicyAttachmentResourceAsListResource() inttypes.ListResourceForSDK {
	l := policyAttachmentListResource{}
	l.SetResourceSchema(resourcePolicyAttachment())

	return &l
}

var _ list.ListResource = &policyAttachmentListResource{}

type policyAttachmentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type policyAttachmentListItem struct {
	name      string
	policyARN string
	groups    []string
	roles     []string
	users     []string
}

func (l *policyAttachmentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.IAMClient(ctx)

	tflog.Info(ctx, "Listing IAM Policy Attachment resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		items, err := listPolicyAttachmentListItems(ctx, conn)
		if err != nil {
			result := fwdiag.NewListResultErrorDiagnostic(err)
			yield(result)
			return
		}

		for _, item := range items {
			ctx := resourcePolicyAttachmentListItemLoggingContext(ctx, item)
			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(resourcePolicyAttachmentImportIDFromData(item))
			rd.Set(names.AttrName, item.name)
			rd.Set("policy_arn", item.policyARN)

			if request.IncludeResource {
				resourcePolicyAttachmentFlatten(rd, item.groups, item.roles, item.users)
			}

			result.DisplayName = item.name

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

func listPolicyAttachmentListItems(ctx context.Context, conn *iam.Client) ([]policyAttachmentListItem, error) {
	items := make(map[string]*policyAttachmentListItem)

	if err := collectPolicyAttachmentsForGroups(ctx, conn, items); err != nil {
		return nil, err
	}
	if err := collectPolicyAttachmentsForRoles(ctx, conn, items); err != nil {
		return nil, err
	}
	if err := collectPolicyAttachmentsForUsers(ctx, conn, items); err != nil {
		return nil, err
	}

	result := make([]policyAttachmentListItem, 0, len(items))
	for _, item := range items {
		slices.Sort(item.groups)
		slices.Sort(item.roles)
		slices.Sort(item.users)
		result = append(result, *item)
	}

	slices.SortFunc(result, func(a, b policyAttachmentListItem) int {
		if a.name != b.name {
			if a.name < b.name {
				return -1
			}
			return 1
		}
		if a.policyARN < b.policyARN {
			return -1
		} else if a.policyARN > b.policyARN {
			return 1
		}
		return 0
	})

	return result, nil
}

func collectPolicyAttachmentsForGroups(ctx context.Context, conn *iam.Client, items map[string]*policyAttachmentListItem) error {
	for group, err := range listGroups(ctx, conn, &iam.ListGroupsInput{}) {
		if err != nil {
			return err
		}

		groupName := aws.ToString(group.GroupName)
		tflog.Info(ctx, "Listing attached policies for group", map[string]any{
			logging.ResourceAttributeKey("group"): groupName,
		})

		input := iam.ListAttachedGroupPoliciesInput{GroupName: group.GroupName}
		pages := iam.NewListAttachedGroupPoliciesPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if errs.IsA[*awstypes.NoSuchEntityException](err) {
				tflog.Warn(ctx, "Resource disappeared during listing, skipping", map[string]any{
					logging.ResourceAttributeKey("group"): groupName,
				})
				break
			}
			if err != nil {
				return fmt.Errorf("listing attached policies for IAM Group (%s): %w", groupName, err)
			}

			for _, attachedPolicy := range page.AttachedPolicies {
				item := putPolicyAttachmentListItem(items, attachedPolicy)
				item.groups = tfslices.AppendUnique(item.groups, groupName)
			}
		}
	}

	return nil
}

func collectPolicyAttachmentsForRoles(ctx context.Context, conn *iam.Client, items map[string]*policyAttachmentListItem) error {
	for role, err := range listNonServiceLinkedRoles(ctx, conn, &iam.ListRolesInput{}) {
		if err != nil {
			return err
		}

		roleName := aws.ToString(role.RoleName)
		tflog.Info(ctx, "Listing attached policies for role", map[string]any{
			logging.ResourceAttributeKey(names.AttrRole): roleName,
		})

		input := iam.ListAttachedRolePoliciesInput{RoleName: role.RoleName}
		pages := iam.NewListAttachedRolePoliciesPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if errs.IsA[*awstypes.NoSuchEntityException](err) {
				tflog.Warn(ctx, "Resource disappeared during listing, skipping", map[string]any{
					logging.ResourceAttributeKey(names.AttrRole): roleName,
				})
				break
			}
			if err != nil {
				return fmt.Errorf("listing attached policies for IAM Role (%s): %w", roleName, err)
			}

			for _, attachedPolicy := range page.AttachedPolicies {
				item := putPolicyAttachmentListItem(items, attachedPolicy)
				item.roles = tfslices.AppendUnique(item.roles, roleName)
			}
		}
	}

	return nil
}

func collectPolicyAttachmentsForUsers(ctx context.Context, conn *iam.Client, items map[string]*policyAttachmentListItem) error {
	for user, err := range listUsers(ctx, conn, &iam.ListUsersInput{}) {
		if err != nil {
			return err
		}

		userName := aws.ToString(user.UserName)
		tflog.Info(ctx, "Listing attached policies for user", map[string]any{
			logging.ResourceAttributeKey("user"): userName,
		})

		input := iam.ListAttachedUserPoliciesInput{UserName: user.UserName}
		pages := iam.NewListAttachedUserPoliciesPaginator(conn, &input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if errs.IsA[*awstypes.NoSuchEntityException](err) {
				tflog.Warn(ctx, "Resource disappeared during listing, skipping", map[string]any{
					logging.ResourceAttributeKey("user"): userName,
				})
				break
			}
			if err != nil {
				return fmt.Errorf("listing attached policies for IAM User (%s): %w", userName, err)
			}

			for _, attachedPolicy := range page.AttachedPolicies {
				item := putPolicyAttachmentListItem(items, attachedPolicy)
				item.users = tfslices.AppendUnique(item.users, userName)
			}
		}
	}

	return nil
}

func putPolicyAttachmentListItem(items map[string]*policyAttachmentListItem, attachedPolicy awstypes.AttachedPolicy) *policyAttachmentListItem {
	policyARN := aws.ToString(attachedPolicy.PolicyArn)
	if item, ok := items[policyARN]; ok {
		return item
	}

	name := aws.ToString(attachedPolicy.PolicyName)
	if name == "" {
		name = policyARN
	}

	item := &policyAttachmentListItem{
		name:      name,
		policyARN: policyARN,
	}
	items[policyARN] = item

	return item
}

func resourcePolicyAttachmentListItemLoggingContext(ctx context.Context, item policyAttachmentListItem) context.Context {
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), item.name)
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("policy_arn"), item.policyARN)
	return ctx
}

func resourcePolicyAttachmentImportIDFromData(item policyAttachmentListItem) string {
	return (policyAttachmentImportID{}).create(item.name, item.policyARN)
}

func listGroups(ctx context.Context, conn *iam.Client, input *iam.ListGroupsInput) iter.Seq2[awstypes.Group, error] {
	return func(yield func(awstypes.Group, error) bool) {
		pages := iam.NewListGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Group{}, fmt.Errorf("listing IAM Groups: %w", err))
				return
			}

			for _, group := range page.Groups {
				if !yield(group, nil) {
					return
				}
			}
		}
	}
}
