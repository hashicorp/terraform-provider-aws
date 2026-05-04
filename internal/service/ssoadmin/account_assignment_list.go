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

// @SDKListResource("aws_ssoadmin_account_assignment")
func newAccountAssignmentResourceAsListResource() inttypes.ListResourceForSDK {
	l := accountAssignmentListResource{}
	l.SetResourceSchema(resourceAccountAssignment())
	return &l
}

var _ list.ListResource = &accountAssignmentListResource{}
var _ list.ListResourceWithRawV5Schemas = &accountAssignmentListResource{}

type accountAssignmentListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type accountAssignmentListResourceModel struct {
	framework.WithRegionModel
	AccountID        types.String `tfsdk:"account_id"`
	InstanceARN      types.String `tfsdk:"instance_arn"`
	PermissionSetARN types.String `tfsdk:"permission_set_arn"`
}

func (l *accountAssignmentListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			names.AttrAccountID: listschema.StringAttribute{
				Required:    true,
				Description: "AWS account ID to list account assignments for.",
			},
			"instance_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the IAM Identity Center instance.",
			},
			"permission_set_arn": listschema.StringAttribute{
				Required:    true,
				Description: "ARN of the permission set whose assignments should be listed.",
			},
		},
	}
}

func (l *accountAssignmentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().SSOAdminClient(ctx)

	var query accountAssignmentListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	accountID := query.AccountID.ValueString()
	instanceARN := query.InstanceARN.ValueString()
	permissionSetARN := query.PermissionSetARN.ValueString()

	tflog.Info(ctx, "Listing SSO Account Assignments", map[string]any{
		logging.ResourceAttributeKey(names.AttrAccountID):  accountID,
		logging.ResourceAttributeKey("instance_arn"):       instanceARN,
		logging.ResourceAttributeKey("permission_set_arn"): permissionSetARN,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := ssoadmin.ListAccountAssignmentsInput{
			AccountId:        aws.String(accountID),
			InstanceArn:      aws.String(instanceARN),
			PermissionSetArn: aws.String(permissionSetARN),
		}

		for item, err := range listAccountAssignments(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := accountAssignmentCreateResourceID(
				aws.ToString(item.PrincipalId),
				string(item.PrincipalType),
				accountID,
				string(awstypes.TargetTypeAwsAccount),
				permissionSetARN,
				instanceARN,
			)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)

			// no need for IncludeResource since every identity attribute is required to be set.
			if err := resourceAccountAssignmentFlatten(rd, &item, instanceARN, string(awstypes.TargetTypeAwsAccount)); err != nil {
				tflog.Error(ctx, "Flattening SSO Account Assignment", map[string]any{
					"error":      err.Error(),
					names.AttrID: id,
				})
				continue
			}

			result.DisplayName = fmt.Sprintf("%s %s", item.PrincipalType, aws.ToString(item.PrincipalId))

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

func listAccountAssignments(ctx context.Context, conn *ssoadmin.Client, input *ssoadmin.ListAccountAssignmentsInput) iter.Seq2[awstypes.AccountAssignment, error] {
	return func(yield func(awstypes.AccountAssignment, error) bool) {
		pages := ssoadmin.NewListAccountAssignmentsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AccountAssignment{}, fmt.Errorf("listing SSO Account Assignments: %w", err))
				return
			}

			for _, item := range page.AccountAssignments {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
