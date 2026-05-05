// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_iam_role_policy_attachments", name="Role Policy Attachments")
func newRolePolicyAttachmentsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &rolePolicyAttachmentsDataSource{}, nil
}

type rolePolicyAttachmentsDataSource struct {
	framework.DataSourceWithModel[rolePolicyAttachmentsDataSourceModel]
}

func (d *rolePolicyAttachmentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"attached_policies": framework.DataSourceComputedListOfObjectAttribute[attachedPolicyModel](ctx),
			"path_prefix": schema.StringAttribute{
				Optional: true,
			},
			"role_name": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *rolePolicyAttachmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().IAMClient(ctx)

	var data rolePolicyAttachmentsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input iam.ListAttachedRolePoliciesInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	roleName := data.RoleName.String()
	out, err := findRolePolicyAttachments(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "role_name", roleName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data.AttachedPolicies), "role_name", roleName)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), "role_name", roleName)
}

func findRolePolicyAttachments(ctx context.Context, conn *iam.Client, input *iam.ListAttachedRolePoliciesInput) ([]awstypes.AttachedPolicy, error) {
	var output []awstypes.AttachedPolicy

	paginator := iam.NewListAttachedRolePoliciesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		output = append(output, page.AttachedPolicies...)
	}

	return output, nil
}

type rolePolicyAttachmentsDataSourceModel struct {
	AttachedPolicies fwtypes.ListNestedObjectValueOf[attachedPolicyModel] `tfsdk:"attached_policies"`
	PathPrefix       types.String                                         `tfsdk:"path_prefix"`
	RoleName         types.String                                         `tfsdk:"role_name"`
}

type attachedPolicyModel struct {
	PolicyArn  types.String `tfsdk:"policy_arn"`
	PolicyName types.String `tfsdk:"policy_name"`
}
