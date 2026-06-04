// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/interconnect/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_interconnect_connection", name="Connection")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newConnectionDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &connectionDataSource{}, nil
}

type connectionDataSource struct {
	framework.DataSourceWithModel[connectionDataSourceModel]
}

func (d *connectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"activation_key": schema.StringAttribute{
				Computed: true,
			},
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"bandwidth": schema.StringAttribute{
				Computed: true,
			},
			"billing_tier": schema.Int32Attribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"environment_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			"interconnect_provider": schema.StringAttribute{
				Computed: true,
			},
			names.AttrLocation: schema.StringAttribute{
				Computed: true,
			},
			"owner_account": schema.StringAttribute{
				Computed: true,
			},
			"shared_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrState: schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ConnectionState](),
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				Computed: true,
			},
			"attach_point": framework.DataSourceComputedListOfObjectAttribute[attachPointModel](ctx),
		},
	}
}

func (d *connectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().InterconnectClient(ctx)

	var data connectionDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findConnectionByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}
	data.InterconnectProvider = flattenProvider(out.Provider)

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type connectionDataSourceModel struct {
	framework.WithRegionModel
	ActivationKey        types.String                                      `tfsdk:"activation_key"`
	ARN                  types.String                                      `tfsdk:"arn"`
	AttachPoint          fwtypes.ListNestedObjectValueOf[attachPointModel] `tfsdk:"attach_point"`
	Bandwidth            types.String                                      `tfsdk:"bandwidth"`
	BillingTier          types.Int32                                       `tfsdk:"billing_tier"`
	Description          types.String                                      `tfsdk:"description"`
	EnvironmentID        types.String                                      `tfsdk:"environment_id"`
	ID                   types.String                                      `tfsdk:"id"`
	InterconnectProvider types.String                                      `tfsdk:"interconnect_provider" autoflex:"-"`
	Location             types.String                                      `tfsdk:"location"`
	OwnerAccount         types.String                                      `tfsdk:"owner_account"`
	SharedID             types.String                                      `tfsdk:"shared_id"`
	State                fwtypes.StringEnum[awstypes.ConnectionState]      `tfsdk:"state"`
	Tags                 tftags.Map                                        `tfsdk:"tags"`
	Type                 types.String                                      `tfsdk:"type"`
}
