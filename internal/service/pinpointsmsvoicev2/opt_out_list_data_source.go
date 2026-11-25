// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_pinpointsmsvoicev2_opt_out_list", name="Opt-out List")
// @Tags(identifierAttribute="arn")
// @Testing(preCheck="testAccPreCheckOptOutList")
func newOptOutListDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &optOutListDataSource{}, nil
}

type optOutListDataSource struct {
	framework.DataSourceWithModel[optOutListDataSourceModel]
}

func (d *optOutListDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *optOutListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().PinpointSMSVoiceV2Client(ctx)

	var data optOutListDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findOptOutListByID(ctx, conn, data.OptOutListName.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.OptOutListName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.OptOutListName.ValueString())
}

type optOutListDataSourceModel struct {
	framework.WithRegionModel
	OptOutListARN  types.String `tfsdk:"arn"`
	OptOutListName types.String `tfsdk:"name"`
	Tags           tftags.Map   `tfsdk:"tags"`
}
