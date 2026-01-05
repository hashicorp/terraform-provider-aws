// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_networkmanager_core_network", name="Core Network")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newCoreNetworkDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &coreNetworkDataSource{}, nil
}

type coreNetworkDataSource struct {
	framework.DataSourceWithModel[coreNetworkDataSourceModel]
}

func (d *coreNetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"core_network_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"global_network_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"edges": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[edgeModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"asn": schema.Int64Attribute{
							Computed: true,
						},
						"edge_location": schema.StringAttribute{
							Computed: true,
						},
						"inside_cidr_blocks": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
			"network_function_groups": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[networkFunctionGroupModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"edge_locations": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"segments": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[networkFunctionGroupSegmentsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"send_to": schema.ListAttribute{
										ElementType: types.StringType,
										Computed:    true,
									},
									"send_via": schema.ListAttribute{
										ElementType: types.StringType,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"segments": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[segmentModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"edge_locations": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
						"shared_segments": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *coreNetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().NetworkManagerClient(ctx)
	var data coreNetworkDataSourceModel

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCoreNetworkByID(ctx, conn, data.CoreNetworkID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.CoreNetworkID.String())
		return
	}

	data.ID = data.CoreNetworkID

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data, fwflex.WithFieldNamePrefix("CoreNetwork")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.CoreNetworkID.String())
}

type coreNetworkDataSourceModel struct {
	ARN                   types.String                                               `tfsdk:"arn"`
	CoreNetworkID         types.String                                               `tfsdk:"core_network_id"`
	CreatedAt             timetypes.RFC3339                                          `tfsdk:"created_at"`
	Description           types.String                                               `tfsdk:"description"`
	Edges                 fwtypes.ListNestedObjectValueOf[edgeModel]                 `tfsdk:"edges"`
	GlobalNetworkID       types.String                                               `tfsdk:"global_network_id"`
	ID                    types.String                                               `tfsdk:"id"`
	NetworkFunctionGroups fwtypes.ListNestedObjectValueOf[networkFunctionGroupModel] `tfsdk:"network_function_groups"`
	Segments              fwtypes.ListNestedObjectValueOf[segmentModel]              `tfsdk:"segments"`
	State                 types.String                                               `tfsdk:"state"`
	Tags                  tftags.Map                                                 `tfsdk:"tags"`
}

type edgeModel struct {
	ASN              types.Int64                       `tfsdk:"asn"`
	EdgeLocation     types.String                      `tfsdk:"edge_location"`
	InsideCidrBlocks fwtypes.ListValueOf[types.String] `tfsdk:"inside_cidr_blocks"`
}

type networkFunctionGroupModel struct {
	EdgeLocations fwtypes.ListValueOf[types.String]                                  `tfsdk:"edge_locations"`
	Name          types.String                                                       `tfsdk:"name"`
	Segments      fwtypes.ListNestedObjectValueOf[networkFunctionGroupSegmentsModel] `tfsdk:"segments"`
}

type networkFunctionGroupSegmentsModel struct {
	SendTo  fwtypes.ListValueOf[types.String] `tfsdk:"send_to"`
	SendVia fwtypes.ListValueOf[types.String] `tfsdk:"send_via"`
}

type segmentModel struct {
	EdgeLocations  fwtypes.ListValueOf[types.String] `tfsdk:"edge_locations"`
	Name           types.String                      `tfsdk:"name"`
	SharedSegments fwtypes.ListValueOf[types.String] `tfsdk:"shared_segments"`
}
