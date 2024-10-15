// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Table")
func newDataSourceTable(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceTable{}, nil
}

type dataSourceTable struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceTable) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_timestreamwrite_table"
}

func (d *dataSourceTable) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrLastUpdatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},

			names.AttrDatabaseName: schema.StringAttribute{
				Required: true,
			},
			"magnetic_store_write_properties": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsMagneticProp](ctx),
				Computed:   true,
			},
			"retention_properties": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsRetentionProperties](ctx),
				Computed:   true,
			},

			names.AttrSchema: schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsSchema](ctx),
				Computed:   true,
			},
			"table_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TableStatus](),
				Computed:   true,
			},
		},
	}
}

func (d *dataSourceTable) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dsTable
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().TimestreamWriteClient(ctx)
	out, err := findTableByTwoPartKey(ctx, conn, data.DatabaseName.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppStream, create.ErrActionReading, data.Name.ValueString(), data.Arn.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dsTable struct {
	Arn                          types.String                                           `tfsdk:"arn"`
	CreationTime                 timetypes.RFC3339                                      `tfsdk:"creation_time"`
	DatabaseName                 types.String                                           `tfsdk:"database_name"`
	LastUpdatedTime              timetypes.RFC3339                                      `tfsdk:"last_updated_time"`
	MagneticStoreWriteProperties fwtypes.ListNestedObjectValueOf[dsMagneticProp]        `tfsdk:"magnetic_store_write_properties"`
	RetentionProperties          fwtypes.ListNestedObjectValueOf[dsRetentionProperties] `tfsdk:"retention_properties"`
	Schema                       fwtypes.ListNestedObjectValueOf[dsSchema]              `tfsdk:"schema"`
	Name                         types.String                                           `tfsdk:"name"`
	Status                       fwtypes.StringEnum[awstypes.TableStatus]               `tfsdk:"table_status"`
}

type dsMagneticProp struct {
	EnableMagneticStoreWrites         types.Bool                                                `tfsdk:"enable_magnetic_store_writes"`
	MagneticStoreRejectedDataLocation fwtypes.ListNestedObjectValueOf[dsMagneticRejectLocation] `tfsdk:"magnetic_store_rejected_data_location"`
}

type dsMagneticRejectLocation struct {
	S3Configuration fwtypes.ListNestedObjectValueOf[dsS3Config] `tfsdk:"s3_configuration"`
}

type dsS3Config struct {
	BucketName       types.String `tfsdk:"bucket_name"`
	EncryptionOption types.String `tfsdk:"encryption_option"`
	KmsKeyId         types.String `tfsdk:"kms_key_id"`
	ObjectKeyPrefix  types.String `tfsdk:"object_key_prefix"`
}

type dsRetentionProperties struct {
	MagneticStoreRetentionPeriodInDays types.Int64 `tfsdk:"magnetic_store_retention_period_in_days"`
	MemoryStoreRetentionPeriodInHours  types.Int64 `tfsdk:"memory_store_retention_period_in_hours"`
}

type dsSchema struct {
	CompositePartitionKey fwtypes.ListNestedObjectValueOf[dsPartition] `tfsdk:"composite_partition_key"`
}

type dsPartition struct {
	Type                fwtypes.StringEnum[awstypes.PartitionKeyType]             `tfsdk:"type"`
	EnforcementInRecord fwtypes.StringEnum[awstypes.PartitionKeyEnforcementLevel] `tfsdk:"enforcement_in_record"`
	Name                types.String                                              `tfsdk:"name"`
}
