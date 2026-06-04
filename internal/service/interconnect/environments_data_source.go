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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_interconnect_environments", name="Environments")
func newEnvironmentsDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &environmentsDataSource{}, nil
}

type environmentsDataSource struct {
	framework.DataSourceWithModel[environmentsDataSourceModel]
}

func (d *environmentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrLocation: schema.StringAttribute{
				Optional: true,
			},
			"environments": framework.DataSourceComputedListOfObjectAttribute[environmentSummaryModel](ctx),
		},
	}
}

func (d *environmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().InterconnectClient(ctx)

	var data environmentsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input interconnect.ListEnvironmentsInput
	if !data.Location.IsNull() {
		input.Location = data.Location.ValueStringPointer()
	}

	var environments []awstypes.Environment
	paginator := interconnect.NewListEnvironmentsPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		environments = append(environments, page.Environments...)
	}

	summaries := make([]*environmentSummaryModel, 0, len(environments))
	for _, e := range environments {
		var m environmentSummaryModel
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, e, &m))
		if resp.Diagnostics.HasError() {
			return
		}
		m.InterconnectProvider = flattenProvider(e.Provider)
		summaries = append(summaries, &m)
	}

	data.Environments = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, summaries)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type environmentsDataSourceModel struct {
	framework.WithRegionModel
	Environments fwtypes.ListNestedObjectValueOf[environmentSummaryModel] `tfsdk:"environments"`
	Location     types.String                                             `tfsdk:"location"`
}

type environmentSummaryModel struct {
	ActivationPageURL    types.String                                             `tfsdk:"activation_page_url"`
	Bandwidths           fwtypes.ListNestedObjectValueOf[bandwidthsModel]         `tfsdk:"bandwidths"`
	EnvironmentID        types.String                                             `tfsdk:"environment_id"`
	InterconnectProvider types.String                                             `tfsdk:"interconnect_provider" autoflex:"-"`
	Location             types.String                                             `tfsdk:"location"`
	RemoteIdentifierType fwtypes.StringEnum[awstypes.RemoteAccountIdentifierType] `tfsdk:"remote_identifier_type"`
	State                fwtypes.StringEnum[awstypes.EnvironmentState]            `tfsdk:"state"`
	Type                 types.String                                             `tfsdk:"type"`
}
