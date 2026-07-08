// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package fsx

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_fsx_lustre_file_system", name="Lustre File System")
// @Tags(identifierAttribute="arn")
func newLustreFileSystemDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &lustreFileSystemDataSource{}, nil
}

const (
	DSNameLustreFileSystem = "Lustre File System Data Source"
)

type lustreFileSystemDataSource struct {
	framework.DataSourceWithModel[lustreFileSystemDataSourceModel]
}

func (d *lustreFileSystemDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"auto_import_policy": schema.StringAttribute{
				Computed: true,
			},
			"automatic_backup_retention_days": schema.Int64Attribute{
				Computed: true,
			},
			"copy_tags_to_backups": schema.BoolAttribute{
				Computed: true,
			},
			"daily_automatic_backup_start_time": schema.StringAttribute{
				Computed: true,
			},
			"data_compression_type": schema.StringAttribute{
				Computed: true,
			},
			"data_read_cache_configuration": framework.DataSourceComputedListOfObjectAttribute[dsDataReadCacheConfigurationModel](ctx),
			"deployment_type": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDNSName: schema.StringAttribute{
				Computed: true,
			},
			"drive_cache_type": schema.StringAttribute{
				Computed: true,
			},
			"efa_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"export_path": schema.StringAttribute{
				Computed: true,
			},
			"file_system_type_version": schema.StringAttribute{
				Computed: true,
			},
			"imported_file_chunk_size": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Computed: true,
			},
			"log_configuration":      framework.DataSourceComputedListOfObjectAttribute[dsLogConfigurationModel](ctx),
			"metadata_configuration": framework.DataSourceComputedListOfObjectAttribute[dsMetadataConfigurationModel](ctx),
			"mount_name": schema.StringAttribute{
				Computed: true,
			},
			"network_interface_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
			},
			"per_unit_storage_throughput": schema.Int64Attribute{
				Computed: true,
			},
			"root_squash_configuration": framework.DataSourceComputedListOfObjectAttribute[dsRootSquashConfigurationModel](ctx),
			"storage_capacity": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrStorageType: schema.StringAttribute{
				Computed: true,
			},
			names.AttrSubnetIDs: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"throughput_capacity": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
			},
			"weekly_maintenance_start_time": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *lustreFileSystemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().FSxClient(ctx)

	var data lustreFileSystemDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	filesystem, err := findLustreFileSystemByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	data.ARN = flex.StringToFramework(ctx, filesystem.ResourceARN)
	data.DNSName = flex.StringToFramework(ctx, filesystem.DNSName)
	data.FileSystemTypeVersion = flex.StringToFramework(ctx, filesystem.FileSystemTypeVersion)
	data.KMSKeyID = types.StringValue(aws.ToString(filesystem.KmsKeyId))
	data.OwnerID = flex.StringToFramework(ctx, filesystem.OwnerId)
	data.StorageCapacity = flex.Int32ToFrameworkInt64(ctx, filesystem.StorageCapacity)
	data.StorageType = flex.StringValueToFramework(ctx, filesystem.StorageType)
	data.VpcID = flex.StringToFramework(ctx, filesystem.VpcId)

	resp.Diagnostics.Append(flex.Flatten(ctx, filesystem.NetworkInterfaceIds, &data.NetworkInterfaceIDs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, filesystem.SubnetIds, &data.SubnetIDs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if lustreConfig := filesystem.LustreConfiguration; lustreConfig != nil {
		data.AutomaticBackupRetentionDays = types.Int64Value(int64(aws.ToInt32(lustreConfig.AutomaticBackupRetentionDays)))
		data.CopyTagsToBackups = types.BoolValue(aws.ToBool(lustreConfig.CopyTagsToBackups))
		data.DailyAutomaticBackupStartTime = flex.StringToFramework(ctx, lustreConfig.DailyAutomaticBackupStartTime)
		data.DataCompressionType = flex.StringValueToFramework(ctx, lustreConfig.DataCompressionType)
		data.DeploymentType = flex.StringValueToFramework(ctx, lustreConfig.DeploymentType)
		data.DriveCacheType = flex.StringValueToFramework(ctx, lustreConfig.DriveCacheType)
		data.EfaEnabled = types.BoolValue(aws.ToBool(lustreConfig.EfaEnabled))
		data.MountName = flex.StringToFramework(ctx, lustreConfig.MountName)
		data.PerUnitStorageThroughput = types.Int64Value(int64(aws.ToInt32(lustreConfig.PerUnitStorageThroughput)))
		data.ThroughputCapacity = types.Int64Value(int64(aws.ToInt32(lustreConfig.ThroughputCapacity)))
		data.WeeklyMaintenanceStartTime = flex.StringToFramework(ctx, lustreConfig.WeeklyMaintenanceStartTime)

		if lustreConfig.DataRepositoryConfiguration != nil {
			data.AutoImportPolicy = flex.StringValueToFramework(ctx, lustreConfig.DataRepositoryConfiguration.AutoImportPolicy)
			data.ExportPath = flex.StringToFramework(ctx, lustreConfig.DataRepositoryConfiguration.ExportPath)
			data.ImportedFileChunkSize = flex.Int32ToFrameworkInt64(ctx, lustreConfig.DataRepositoryConfiguration.ImportedFileChunkSize)
		}

		if lustreConfig.DataReadCacheConfiguration != nil {
			drc := &dsDataReadCacheConfigurationModel{
				Size:       flex.Int32ToFrameworkInt64(ctx, lustreConfig.DataReadCacheConfiguration.SizeGiB),
				SizingMode: flex.StringValueToFramework(ctx, lustreConfig.DataReadCacheConfiguration.SizingMode),
			}
			data.DataReadCacheConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, drc)
		}

		if lustreConfig.LogConfiguration != nil {
			lc := &dsLogConfigurationModel{
				Destination: flex.StringToFramework(ctx, lustreConfig.LogConfiguration.Destination),
				Level:       flex.StringValueToFramework(ctx, lustreConfig.LogConfiguration.Level),
			}
			data.LogConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, lc)
		}

		if lustreConfig.MetadataConfiguration != nil {
			mc := &dsMetadataConfigurationModel{
				Mode: flex.StringValueToFramework(ctx, lustreConfig.MetadataConfiguration.Mode),
				Iops: flex.Int32ToFrameworkInt64(ctx, lustreConfig.MetadataConfiguration.Iops),
			}
			data.MetadataConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, mc)
		}

		if lustreConfig.RootSquashConfiguration != nil {
			rsc := &dsRootSquashConfigurationModel{
				RootSquash: flex.StringToFramework(ctx, lustreConfig.RootSquashConfiguration.RootSquash),
			}
			resp.Diagnostics.Append(flex.Flatten(ctx, lustreConfig.RootSquashConfiguration.NoSquashNids, &rsc.NoSquashNids)...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.RootSquashConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, rsc)
		}
	}

	setTagsOut(ctx, filesystem.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.ID.String())
}

type lustreFileSystemDataSourceModel struct {
	framework.WithRegionModel
	ARN                           types.String                                                       `tfsdk:"arn"`
	AutoImportPolicy              types.String                                                       `tfsdk:"auto_import_policy"`
	AutomaticBackupRetentionDays  types.Int64                                                        `tfsdk:"automatic_backup_retention_days"`
	CopyTagsToBackups             types.Bool                                                         `tfsdk:"copy_tags_to_backups"`
	DailyAutomaticBackupStartTime types.String                                                       `tfsdk:"daily_automatic_backup_start_time"`
	DataCompressionType           types.String                                                       `tfsdk:"data_compression_type"`
	DataReadCacheConfiguration    fwtypes.ListNestedObjectValueOf[dsDataReadCacheConfigurationModel] `tfsdk:"data_read_cache_configuration"`
	DeploymentType                types.String                                                       `tfsdk:"deployment_type"`
	DNSName                       types.String                                                       `tfsdk:"dns_name"`
	DriveCacheType                types.String                                                       `tfsdk:"drive_cache_type"`
	EfaEnabled                    types.Bool                                                         `tfsdk:"efa_enabled"`
	ExportPath                    types.String                                                       `tfsdk:"export_path"`
	FileSystemTypeVersion         types.String                                                       `tfsdk:"file_system_type_version"`
	ID                            types.String                                                       `tfsdk:"id"`
	ImportedFileChunkSize         types.Int64                                                        `tfsdk:"imported_file_chunk_size"`
	KMSKeyID                      types.String                                                       `tfsdk:"kms_key_id"`
	LogConfiguration              fwtypes.ListNestedObjectValueOf[dsLogConfigurationModel]           `tfsdk:"log_configuration"`
	MetadataConfiguration         fwtypes.ListNestedObjectValueOf[dsMetadataConfigurationModel]      `tfsdk:"metadata_configuration"`
	MountName                     types.String                                                       `tfsdk:"mount_name"`
	NetworkInterfaceIDs           fwtypes.ListOfString                                               `tfsdk:"network_interface_ids"`
	OwnerID                       types.String                                                       `tfsdk:"owner_id"`
	PerUnitStorageThroughput      types.Int64                                                        `tfsdk:"per_unit_storage_throughput"`
	RootSquashConfiguration       fwtypes.ListNestedObjectValueOf[dsRootSquashConfigurationModel]    `tfsdk:"root_squash_configuration"`
	StorageCapacity               types.Int64                                                        `tfsdk:"storage_capacity"`
	StorageType                   types.String                                                       `tfsdk:"storage_type"`
	SubnetIDs                     fwtypes.ListOfString                                               `tfsdk:"subnet_ids"`
	Tags                          tftags.Map                                                         `tfsdk:"tags"`
	ThroughputCapacity            types.Int64                                                        `tfsdk:"throughput_capacity"`
	VpcID                         types.String                                                       `tfsdk:"vpc_id"`
	WeeklyMaintenanceStartTime    types.String                                                       `tfsdk:"weekly_maintenance_start_time"`
}

type dsDataReadCacheConfigurationModel struct {
	Size       types.Int64  `tfsdk:"size"`
	SizingMode types.String `tfsdk:"sizing_mode"`
}

type dsLogConfigurationModel struct {
	Destination types.String `tfsdk:"destination"`
	Level       types.String `tfsdk:"level"`
}

type dsMetadataConfigurationModel struct {
	Mode types.String `tfsdk:"mode"`
	Iops types.Int64  `tfsdk:"iops"`
}

type dsRootSquashConfigurationModel struct {
	NoSquashNids fwtypes.SetOfString `tfsdk:"no_squash_nids"`
	RootSquash   types.String        `tfsdk:"root_squash"`
}
