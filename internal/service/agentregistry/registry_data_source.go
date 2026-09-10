// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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

// @FrameworkDataSource("aws_agentregistry_registry", name="Registry")
// @Tags(identifierAttribute="registry_arn")
func newRegistryDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &registryDataSource{}, nil
}

type registryDataSource struct {
	framework.DataSourceWithModel[registryDataSourceModel]
}

func (d *registryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"approval_configuration":       framework.DataSourceComputedListOfObjectAttribute[approvalConfigurationModel](ctx),
			"auto_detection_configuration": framework.DataSourceComputedListOfObjectAttribute[autoDetectionConfigurationModel](ctx),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"discovery_configuration": framework.DataSourceComputedListOfObjectAttribute[discoveryConfigurationModel](ctx),
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"registry_arn": schema.StringAttribute{
				Computed: true,
			},
			"registry_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RegistryStatus](),
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *registryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data registryDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AgentRegistryClient(ctx)

	registryID := fwflex.StringValueFromFramework(ctx, data.RegistryID)
	out, err := findRegistryByID(ctx, conn, registryID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, registryID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type registryDataSourceModel struct {
	framework.WithRegionModel
	ApprovalConfiguration      fwtypes.ListNestedObjectValueOf[approvalConfigurationModel]      `tfsdk:"approval_configuration"`
	AutoDetectionConfiguration fwtypes.ListNestedObjectValueOf[autoDetectionConfigurationModel] `tfsdk:"auto_detection_configuration"`
	CreatedAt                  timetypes.RFC3339                                                `tfsdk:"created_at"`
	Description                types.String                                                     `tfsdk:"description"`
	DiscoveryConfiguration     fwtypes.ListNestedObjectValueOf[discoveryConfigurationModel]     `tfsdk:"discovery_configuration"`
	Name                       types.String                                                     `tfsdk:"name"`
	RegistryARN                types.String                                                     `tfsdk:"registry_arn"`
	RegistryID                 types.String                                                     `tfsdk:"registry_id"`
	Status                     fwtypes.StringEnum[awstypes.RegistryStatus]                      `tfsdk:"status"`
	Tags                       tftags.Map                                                       `tfsdk:"tags"`
	UpdatedAt                  timetypes.RFC3339                                                `tfsdk:"updated_at"`
}
