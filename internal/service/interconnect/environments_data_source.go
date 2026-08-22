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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		Blocks: map[string]schema.Block{
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
	if !data.InterconnectProvider.IsNull() {
		filter, d := data.InterconnectProvider.ToPtr(ctx)
		smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}

		provider, d := filter.Expand(ctx)
		smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}

		input.Provider, _ = provider.(awstypes.Provider)
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
	Environments         fwtypes.ListNestedObjectValueOf[environmentSummaryModel] `tfsdk:"environments"`
	InterconnectProvider fwtypes.ListNestedObjectValueOf[providerFilterModel]     `tfsdk:"interconnect_provider"`
	Location             types.String                                             `tfsdk:"location"`
}

// providerFilterModel expands to the SDK Provider tagged union. The block is named
// "interconnect_provider" because "provider" is a reserved Terraform meta-argument.
type providerFilterModel struct {
	CloudServiceProvider types.String `tfsdk:"cloud_service_provider"`
	LastMileProvider     types.String `tfsdk:"last_mile_provider"`
}

var _ fwflex.Expander = providerFilterModel{}

func (m providerFilterModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.CloudServiceProvider.IsNull():
		return &awstypes.ProviderMemberCloudServiceProvider{
			Value: m.CloudServiceProvider.ValueString(),
		}, diags
	case !m.LastMileProvider.IsNull():
		return &awstypes.ProviderMemberLastMileProvider{
			Value: m.LastMileProvider.ValueString(),
		}, diags
	}

	return nil, diags
}

type bandwidthsModel struct {
	Available fwtypes.ListOfString `tfsdk:"available"`
	Supported fwtypes.ListOfString `tfsdk:"supported"`
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
