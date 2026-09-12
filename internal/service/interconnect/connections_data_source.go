// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/interconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/interconnect/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			names.AttrState: schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ConnectionState](),
			},
			"connections": framework.DataSourceComputedListOfObjectAttribute[connectionSummaryModel](ctx),
		},
		Blocks: map[string]schema.Block{
			"attach_point": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[attachPointModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName(names.AttrARN),
									path.MatchRelative().AtParent().AtName("direct_connect_gateway"),
								),
							},
						},
						"direct_connect_gateway": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"interconnect_provider": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[providerFilterModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cloud_service_provider": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("cloud_service_provider"),
									path.MatchRelative().AtParent().AtName("last_mile_provider"),
								),
							},
						},
						"last_mile_provider": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
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

	// ListConnections combines all supplied filters, returning only Connections that
	// match every one of them.
	var input interconnect.ListConnectionsInput
	if !data.EnvironmentID.IsNull() {
		input.EnvironmentId = data.EnvironmentID.ValueStringPointer()
	}
	if !data.State.IsNull() {
		input.State = data.State.ValueEnum()
	}
	if !data.AttachPoint.IsNull() {
		filter, d := data.AttachPoint.ToPtr(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}

		attachPoint, d := filter.Expand(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}

		input.AttachPoint, _ = attachPoint.(awstypes.AttachPoint)
	}
	if !data.InterconnectProvider.IsNull() {
		filter, d := data.InterconnectProvider.ToPtr(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}

		provider, d := filter.Expand(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}

		input.Provider, _ = provider.(awstypes.Provider)
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
	AttachPoint          fwtypes.ListNestedObjectValueOf[attachPointModel]       `tfsdk:"attach_point"`
	Connections          fwtypes.ListNestedObjectValueOf[connectionSummaryModel] `tfsdk:"connections"`
	EnvironmentID        types.String                                            `tfsdk:"environment_id"`
	InterconnectProvider fwtypes.ListNestedObjectValueOf[providerFilterModel]    `tfsdk:"interconnect_provider"`
	State                fwtypes.StringEnum[awstypes.ConnectionState]            `tfsdk:"state"`
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
