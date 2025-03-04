// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ssm_patch_baselines", name="Patch Baselines")
func newDataSourcePatchBaselines(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourcePatchBaselines{}, nil
}

const (
	DSNamePatchBaselines = "Patch Baselines Data Source"
)

type dataSourcePatchBaselines struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourcePatchBaselines) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"baseline_identities": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[baselineIdentityModel](ctx),
				Computed:    true,
				ElementType: fwtypes.NewObjectTypeOf[baselineIdentityModel](ctx),
			},
			"default_baselines": schema.BoolAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[filterModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKey: schema.StringAttribute{
							Required: true,
						},
						names.AttrValues: schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Required:   true,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourcePatchBaselines) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SSMClient(ctx)

	var data dataSourcePatchBaselinesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ssm.DescribePatchBaselinesInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPatchBaselines(ctx, conn, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSM, create.ErrActionReading, DSNamePatchBaselines, "", err),
			err.Error(),
		)
		return
	}

	if data.DefaultBaselines.ValueBool() {
		out = tfslices.Filter(out, func(v awstypes.PatchBaselineIdentity) bool {
			return v.DefaultBaseline
		})
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data.BaselineIdentities)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findPatchBaselines(ctx context.Context, conn *ssm.Client, input *ssm.DescribePatchBaselinesInput) ([]awstypes.PatchBaselineIdentity, error) {
	var baselines []awstypes.PatchBaselineIdentity
	pages := ssm.NewDescribePatchBaselinesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		baselines = append(baselines, page.BaselineIdentities...)
	}

	return baselines, nil
}

type dataSourcePatchBaselinesModel struct {
	BaselineIdentities fwtypes.ListNestedObjectValueOf[baselineIdentityModel] `tfsdk:"baseline_identities"`
	Filter             fwtypes.ListNestedObjectValueOf[filterModel]           `tfsdk:"filter"`
	DefaultBaselines   types.Bool                                             `tfsdk:"default_baselines"`
}

type baselineIdentityModel struct {
	BaselineDescription types.String `tfsdk:"baseline_description"`
	BaselineID          types.String `tfsdk:"baseline_id"`
	BaselineName        types.String `tfsdk:"baseline_name"`
	DefaultBaseline     types.Bool   `tfsdk:"default_baseline"`
	OperatingSystem     types.String `tfsdk:"operating_system"`
}

type filterModel struct {
	Key    types.String        `tfsdk:"key"`
	Values fwtypes.SetOfString `tfsdk:"values"`
}
