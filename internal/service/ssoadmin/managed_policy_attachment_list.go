// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
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

// @SDKListResource("aws_ssoadmin_managed_policy_attachment")
func newManagedPolicyAttachmentResourceAsListResource() inttypes.ListResourceForSDK {
	l := managedPolicyAttachmentListResource{}
	l.SetResourceSchema(resourceManagedPolicyAttachment())
	return &l
}

var _ list.ListResource = &managedPolicyAttachmentListResource{}
var _ list.ListResourceWithRawV5Schemas = &managedPolicyAttachmentListResource{}

type managedPolicyAttachmentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type managedPolicyAttachmentListResourceModel struct {
	framework.WithRegionModel
	InstanceARN      types.String `tfsdk:"instance_arn"`
	PermissionSetARN types.String `tfsdk:"permission_set_arn"`
}

func (l *managedPolicyAttachmentListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"instance_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the IAM Identity Center instance.",
			},
			"permission_set_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the permission set whose managed policy attachments should be listed.",
			},
		},
	}
}

func (l *managedPolicyAttachmentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SSOAdminClient(ctx)

	var query managedPolicyAttachmentListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	instanceARN := query.InstanceARN.ValueString()
	permissionSetARN := query.PermissionSetARN.ValueString()

	tflog.Info(ctx, "Listing SSO Managed Policy Attachments", map[string]any{
		logging.ResourceAttributeKey("instance_arn"):       instanceARN,
		logging.ResourceAttributeKey("permission_set_arn"): permissionSetARN,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := ssoadmin.ListManagedPoliciesInPermissionSetInput{
			InstanceArn:      aws.String(instanceARN),
			PermissionSetArn: aws.String(permissionSetARN),
		}

		for item, err := range listManagedPolicyAttachments(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := managedPolicyAttachmentCreateResourceID(aws.ToString(item.Arn), permissionSetARN, instanceARN)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)

			// No need to check IncludeResource: the list response contains every resource attribute.
			if err := resourceManagedPolicyAttachmentFlatten(rd, &item, instanceARN, permissionSetARN); err != nil {
				tflog.Error(ctx, "Flattening SSO Managed Policy Attachment", map[string]any{
					"error":      err.Error(),
					names.AttrID: id,
				})
				continue
			}

			result.DisplayName = aws.ToString(item.Name)

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

func listManagedPolicyAttachments(ctx context.Context, conn *ssoadmin.Client, input *ssoadmin.ListManagedPoliciesInPermissionSetInput) iter.Seq2[awstypes.AttachedManagedPolicy, error] {
	return func(yield func(awstypes.AttachedManagedPolicy, error) bool) {
		pages := ssoadmin.NewListManagedPoliciesInPermissionSetPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AttachedManagedPolicy{}, fmt.Errorf("listing SSO Managed Policy Attachments: %w", err))
				return
			}

			for _, item := range page.AttachedManagedPolicies {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
