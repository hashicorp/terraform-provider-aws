// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrDescription: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrKMSKeyID: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrName: fwschema.StringAttribute{
				Computed: true,
			},
			"organization_id": fwschema.StringAttribute{
				Computed: true,
			},
			"ou_id": fwschema.StringAttribute{
				Computed: true,
			},
			"sharing_enabled": fwschema.BoolAttribute{
				Computed: true,
			},
			"system_id": fwschema.StringAttribute{
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

	arn := fwflex.StringValueFromFramework(ctx, data.SystemARN)
	system, err := findSystemByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, system, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, system.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type systemDataSourceModel struct {
	framework.WithRegionModel
	Description    types.String `tfsdk:"description"`
	KMSKeyID       types.String `tfsdk:"kms_key_id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	OUID           types.String `tfsdk:"ou_id"`
	SharingEnabled types.Bool   `tfsdk:"sharing_enabled"`
	SystemARN      fwtypes.ARN  `tfsdk:"arn"`
	SystemID       types.String `tfsdk:"system_id"`
	Tags           tftags.Map   `tfsdk:"tags"`
}
