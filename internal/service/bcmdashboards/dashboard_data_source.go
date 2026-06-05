// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bcmdashboards

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bcmdashboards/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_bcmdashboards_dashboard",name="Dashboard")
// @Tags(identifierAttribute="arn")
func newDashboardDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dashboardDataSource{}, nil
}

type dashboardDataSource struct {
	framework.DataSourceWithModel[dashboardDataSourceModel]
}

func (d *dashboardDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"dashboard_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DashboardType](),
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"widget": framework.DataSourceComputedListOfObjectAttribute[widgetModel](ctx),
		},
	}
}

func (d *dashboardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().BCMDashboardsClient(ctx)

	var data dashboardDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDashboardByARN(ctx, conn, data.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BCMDashboards, create.ErrActionReading, ResNameDashboard, data.ARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.DashboardType = fwtypes.StringEnumValue(out.Type)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dashboardDataSourceModel struct {
	ARN           fwtypes.ARN                                  `tfsdk:"arn"`
	CreatedAt     timetypes.RFC3339                            `tfsdk:"created_at"`
	DashboardType fwtypes.StringEnum[awstypes.DashboardType]   `tfsdk:"dashboard_type"`
	Description   types.String                                 `tfsdk:"description"`
	Name          types.String                                 `tfsdk:"name"`
	Tags          tftags.Map                                   `tfsdk:"tags"`
	UpdatedAt     timetypes.RFC3339                            `tfsdk:"updated_at"`
	Widgets       fwtypes.ListNestedObjectValueOf[widgetModel] `tfsdk:"widget"`
}
