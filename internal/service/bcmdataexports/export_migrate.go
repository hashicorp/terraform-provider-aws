// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bcmdataexports

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func exportSchemaV0(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrID:      framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"export": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[exportData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"export_arn": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"data_query":                 exportDataQuerySchema(ctx),
						"destination_configurations": exportDestinationConfigurationsSchema(ctx),
						"refresh_cadence":            exportRefreshCadenceSchema(ctx),
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

type exportResourceModelV0 struct {
	Export   fwtypes.ListNestedObjectValueOf[exportData] `tfsdk:"export"`
	ID       types.String                                `tfsdk:"id"`
	Tags     tftags.Map                                  `tfsdk:"tags"`
	TagsAll  tftags.Map                                  `tfsdk:"tags_all"`
	Timeouts timeouts.Value                              `tfsdk:"timeouts"`
}

func upgradeExportResourceStateFromV0(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var exportDataV0 exportResourceModelV0
	response.Diagnostics.Append(request.State.Get(ctx, &exportDataV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	exportData := exportResourceModel{
		ARN:      exportDataV0.ID,
		Export:   exportDataV0.Export,
		ID:       exportDataV0.ID,
		Tags:     exportDataV0.Tags,
		TagsAll:  exportDataV0.TagsAll,
		Timeouts: exportDataV0.Timeouts,
	}

	response.Diagnostics.Append(response.State.Set(ctx, exportData)...)
}
