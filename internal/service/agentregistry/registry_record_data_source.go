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

// @FrameworkDataSource("aws_agentregistry_registry_record", name="Registry Record")
// @Tags(identifierAttribute="record_arn")
func newRegistryRecordDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &registryRecordDataSource{}, nil
}

type registryRecordDataSource struct {
	framework.DataSourceWithModel[registryRecordDataSourceModel]
}

func (d *registryRecordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"descriptors": framework.DataSourceComputedListOfObjectAttribute[descriptorsModel](ctx),
			names.AttrDisplayName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"record_arn": schema.StringAttribute{
				Computed: true,
			},
			"record_id": schema.StringAttribute{
				Required: true,
			},
			"record_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RecordType](),
				Computed:   true,
			},
			"record_version": schema.StringAttribute{
				Computed: true,
			},
			"registry_arn": schema.StringAttribute{
				Computed: true,
			},
			"registry_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RegistryRecordStatus](),
				Computed:   true,
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *registryRecordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data registryRecordDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AgentRegistryClient(ctx)

	registryID := fwflex.StringValueFromFramework(ctx, data.RegistryID)
	recordID := fwflex.StringValueFromFramework(ctx, data.RecordID)
	out, err := findRegistryRecordByTwoPartKey(ctx, conn, registryID, recordID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, recordID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type registryRecordDataSourceModel struct {
	framework.WithRegionModel
	CreatedAt     timetypes.RFC3339                                 `tfsdk:"created_at"`
	Description   types.String                                      `tfsdk:"description"`
	Descriptors   fwtypes.ListNestedObjectValueOf[descriptorsModel] `tfsdk:"descriptors"`
	DisplayName   types.String                                      `tfsdk:"display_name"`
	Name          types.String                                      `tfsdk:"name"`
	RecordARN     types.String                                      `tfsdk:"record_arn"`
	RecordID      types.String                                      `tfsdk:"record_id"`
	RecordType    fwtypes.StringEnum[awstypes.RecordType]           `tfsdk:"record_type"`
	RecordVersion types.String                                      `tfsdk:"record_version"`
	RegistryARN   types.String                                      `tfsdk:"registry_arn"`
	RegistryID    types.String                                      `tfsdk:"registry_id"`
	Status        fwtypes.StringEnum[awstypes.RegistryRecordStatus] `tfsdk:"status"`
	StatusReason  types.String                                      `tfsdk:"status_reason"`
	Tags          tftags.Map                                        `tfsdk:"tags"`
	UpdatedAt     timetypes.RFC3339                                 `tfsdk:"updated_at"`
}
