// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_bedrockagentcore_gateway_rate_limit", name="Gateway Rate Limit")
func newGatewayRateLimitDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &gatewayRateLimitDataSource{}, nil
}

type gatewayRateLimitDataSource struct {
	framework.DataSourceWithModel[gatewayRateLimitDataSourceModel]
}

func (d *gatewayRateLimitDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"dimension_keys": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
			// A computed list attribute rather than the resource's nested block:
			// fully computed blocks are not supported by Terraform protocol V6,
			// which the provider will adopt in a future major version.
			"entries": framework.DataSourceComputedListOfObjectAttribute[limitEntryDataSourceModel](ctx),
			"gateway_identifier": schema.StringAttribute{
				Required: true,
			},
			"rate_limit_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GatewayRateLimitStatus](),
				Computed:   true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *gatewayRateLimitDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data gatewayRateLimitDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier, rateLimitID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, data.RateLimitID)
	// Unlike the resource, a data source errors on not-found rather than
	// removing state. A cascade-deleted gateway surfaces as the same not-found
	// error as a missing rate limit, so the diagnostic ID carries both parts of
	// the key (the same two-part form as the resource's import ID).
	id := gatewayIdentifier + "," + rateLimitID
	out, err := findGatewayRateLimitByTwoPartKey(ctx, conn, gatewayIdentifier, rateLimitID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data), smerr.ID, id)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data), smerr.ID, id)
}

// Models.
//
// Deliberately separate from the resource's models: entries is a list here
// rather than a set, and this exposes status and the timestamps, which the
// resource omits because a transient status is state noise on a managed
// resource.

type gatewayRateLimitDataSourceModel struct {
	framework.WithRegionModel
	CreatedAt         timetypes.RFC3339                                          `tfsdk:"created_at"`
	Description       types.String                                               `tfsdk:"description"`
	DimensionKeys     fwtypes.ListOfString                                       `tfsdk:"dimension_keys"`
	Entries           fwtypes.ListNestedObjectValueOf[limitEntryDataSourceModel] `tfsdk:"entries"`
	GatewayIdentifier types.String                                               `tfsdk:"gateway_identifier"`
	RateLimitID       types.String                                               `tfsdk:"rate_limit_id"`
	Status            fwtypes.StringEnum[awstypes.GatewayRateLimitStatus]        `tfsdk:"status"`
	UpdatedAt         timetypes.RFC3339                                          `tfsdk:"updated_at"`
}

type limitEntryDataSourceModel struct {
	Connections fwtypes.ListNestedObjectValueOf[rateConfigDataSourceModel] `tfsdk:"connections"`
	Dimensions  fwtypes.MapOfString                                        `tfsdk:"dimensions"`
	Requests    fwtypes.ListNestedObjectValueOf[rateConfigDataSourceModel] `tfsdk:"requests"`
	Tokens      fwtypes.ListNestedObjectValueOf[rateConfigDataSourceModel] `tfsdk:"tokens"`
}

type rateConfigDataSourceModel struct {
	Period fwtypes.StringEnum[awstypes.Period] `tfsdk:"period"`
	Rate   types.Float64                       `tfsdk:"rate"`
}
