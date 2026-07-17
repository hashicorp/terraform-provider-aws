// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_lambdaweb_function", name="Function")
func newFunctionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &functionDataSource{}, nil
}

type functionDataSource struct {
	framework.DataSourceWithModel[functionDataSourceModel]
}

func (d *functionDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"function_name": schema.StringAttribute{
				Required: true,
			},
			"latest_revision_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			"state_reason": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *functionDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data functionDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().LambdaWebClient(ctx)
	name := data.FunctionName.ValueString()

	out, err := findFunctionByName(ctx, conn, name)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}
	data.ARN = fwflex.StringToFramework(ctx, out.FunctionArn)

	if revisionID, err := findLatestRevisionID(ctx, conn, name); err == nil && revisionID != nil {
		data.LatestRevisionID = fwflex.StringToFramework(ctx, revisionID)
	} else if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type functionDataSourceModel struct {
	framework.WithRegionModel
	ARN              types.String `tfsdk:"arn"`
	FunctionName     types.String `tfsdk:"function_name"`
	LatestRevisionID types.String `tfsdk:"latest_revision_id"`
	State            types.String `tfsdk:"state"`
	StateReason      types.String `tfsdk:"state_reason"`
}
