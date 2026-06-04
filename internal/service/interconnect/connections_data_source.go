// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/interconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/interconnect/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_interconnect_connections", name="Connections")
func newConnectionsDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &connectionsDataSource{}, nil
}

type connectionsDataSource struct {
	framework.DataSourceWithModel[connectionsDataSourceModel]
}

func (d *connectionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				Optional: true,
			},
			"connections": framework.DataSourceComputedListOfObjectAttribute[connectionSummaryModel](ctx),
		},
	}
}

func (d *connectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().InterconnectClient(ctx)

	var data connectionsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input interconnect.ListConnectionsInput
	if !data.EnvironmentID.IsNull() {
		input.EnvironmentId = data.EnvironmentID.ValueStringPointer()
	}

	var connections []awstypes.ConnectionSummary
	paginator := interconnect.NewListConnectionsPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		connections = append(connections, page.Connections...)
	}

	summaries := make([]*connectionSummaryModel, 0, len(connections))
	for _, c := range connections {
		var m connectionSummaryModel
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, c, &m))
		if resp.Diagnostics.HasError() {
			return
		}
		m.InterconnectProvider = flattenProvider(c.Provider)
		summaries = append(summaries, &m)
	}

	data.Connections = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, summaries)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type connectionsDataSourceModel struct {
	framework.WithRegionModel
	Connections   fwtypes.ListNestedObjectValueOf[connectionSummaryModel] `tfsdk:"connections"`
	EnvironmentID types.String                                            `tfsdk:"environment_id"`
}

type connectionSummaryModel struct {
	ARN                  types.String                                      `tfsdk:"arn"`
	AttachPoint          fwtypes.ListNestedObjectValueOf[attachPointModel] `tfsdk:"attach_point"`
	Bandwidth            types.String                                      `tfsdk:"bandwidth"`
	BillingTier          types.Int32                                       `tfsdk:"billing_tier"`
	Description          types.String                                      `tfsdk:"description"`
	EnvironmentID        types.String                                      `tfsdk:"environment_id"`
	ID                   types.String                                      `tfsdk:"id"`
	InterconnectProvider types.String                                      `tfsdk:"interconnect_provider" autoflex:"-"`
	Location             types.String                                      `tfsdk:"location"`
	SharedID             types.String                                      `tfsdk:"shared_id"`
	State                fwtypes.StringEnum[awstypes.ConnectionState]      `tfsdk:"state"`
	Type                 types.String                                      `tfsdk:"type"`
}
