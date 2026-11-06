// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridgev2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_eventbridgev2_event_bus", name="Event Bus")
func newEventBusDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &eventBusDataSource{}, nil
}

type eventBusDataSource struct {
	framework.DataSourceWithModel[eventBusDataSourceModel]
}

func (d *eventBusDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrEncryptionConfiguration: schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[encryptionConfigurationModel](ctx),
				Computed:   true,
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.BusState](),
				Computed:   true,
			},
			"state_reason": schema.StringAttribute{
				Computed: true,
			},
			"storage_configuration": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[storageConfigurationModel](ctx),
				Computed:   true,
			},
		},
	}
}

func (d *eventBusDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data eventBusDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EventBridgeV2Client(ctx)

	out, err := findEventBusByARN(ctx, conn, data.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, tfresource.SingularDataSourceFindError("EventBridge V2 Event Bus", err))
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data, fwflex.WithFieldNamePrefix("EventBus")))
	if response.Diagnostics.HasError() {
		return
	}

	// The service returns an empty description when none is set.
	if aws.ToString(out.Description) == "" {
		data.Description = types.StringNull()
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type eventBusDataSourceModel struct {
	framework.WithRegionModel
	ARN                     fwtypes.ARN                                         `tfsdk:"arn"`
	CreationTime            timetypes.RFC3339                                   `tfsdk:"creation_time"`
	Description             types.String                                        `tfsdk:"description"`
	EncryptionConfiguration fwtypes.ObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	LastModifiedTime        timetypes.RFC3339                                   `tfsdk:"last_modified_time"`
	Name                    types.String                                        `tfsdk:"name"`
	State                   fwtypes.StringEnum[awstypes.BusState]               `tfsdk:"state"`
	StateReason             types.String                                        `tfsdk:"state_reason"`
	StorageConfiguration    fwtypes.ObjectValueOf[storageConfigurationModel]    `tfsdk:"storage_configuration"`
}
