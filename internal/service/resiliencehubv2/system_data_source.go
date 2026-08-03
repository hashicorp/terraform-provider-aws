// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_resiliencehubv2_system", name="System")
// @Tags(identifierAttribute="arn")
func newSystemDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &systemDataSource{}, nil
}

type systemDataSource struct {
	framework.DataSourceWithModel[systemDataSourceModel]
}

func (d *systemDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: fwschema.StringAttribute{
				Required: true,
			},
			names.AttrName: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: fwschema.StringAttribute{
				Computed: true,
			},
			"sharing_enabled": fwschema.BoolAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *systemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data systemDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ResilienceHubV2Client(ctx)

	system, err := findSystemByARN(ctx, conn, data.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, system, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	data.ARN = types.StringValue(aws.ToString(system.SystemArn))

	tags, err := listTags(ctx, conn, data.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}
	setTagsOut(ctx, tags.Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type systemDataSourceModel struct {
	framework.WithRegionModel
	ARN            types.String `tfsdk:"arn"`
	Description    types.String `tfsdk:"description"`
	Name           types.String `tfsdk:"name"`
	SharingEnabled types.Bool   `tfsdk:"sharing_enabled"`
	Tags           tftags.Map   `tfsdk:"tags"`
}
