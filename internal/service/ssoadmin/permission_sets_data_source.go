// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ssoadmin_permission_sets", name="Permission Sets")
func newPermissionSetsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &permissionSetsDataSource{}, nil
}

type permissionSetsDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *permissionSetsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARNs: schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrID: framework.IDAttribute(),
			"instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
		},
	}
}

func (d *permissionSetsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data permissionSetsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SSOAdminClient(ctx)

	var arns []string
	input := &ssoadmin.ListPermissionSetsInput{
		InstanceArn: fwflex.StringFromFramework(ctx, data.InstanceARN),
	}
	pages := ssoadmin.NewListPermissionSetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			response.Diagnostics.AddError("listing SSO Permission Sets", err.Error())

			return
		}

		arns = append(arns, page.PermissionSets...)
	}

	data.ID = fwflex.StringValueToFramework(ctx, data.InstanceARN.ValueString())
	data.ARNs = fwflex.FlattenFrameworkStringValueList(ctx, arns)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type permissionSetsDataSourceModel struct {
	ARNs        types.List   `tfsdk:"arns"`
	ID          types.String `tfsdk:"id"`
	InstanceARN fwtypes.ARN  `tfsdk:"instance_arn"`
}
