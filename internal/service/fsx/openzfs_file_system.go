// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_openzfs_file_system", name="OpenZFS File System")
// @Tags(identifierAttribute="arn")
func resourceOpenZFSFileSystem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOpenZFSFileSystemCreate,
		ReadWithoutTimeout:   resourceOpenZFSFileSystemRead,
		UpdateWithoutTimeout: resourceOpenZFSFileSystemUpdate,
		DeleteWithoutTimeout: resourceOpenZFSFileSystemDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("skip_final_backup", false)

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic_backup_retention_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 90),
			},
			"backup_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"copy_tags_to_backups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"copy_tags_to_volumes": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"daily_automatic_backup_start_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(5, 5),
					validation.StringMatch(regexache.MustCompile(`^([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format HH:MM"),
				),
			},
			"delete_options": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.DeleteFileSystemOpenZFSOption](),
				},
			},
			"deployment_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.OpenZFSDeploymentType](),
			},
			"disk_iops_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						names.AttrMode: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.DiskIopsConfigurationModeAutomatic,
							ValidateDiagFunc: enum.Validate[awstypes.DiskIopsConfigurationMode](),
						},
					},
				},
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_ip_address_range": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"final_backup_tags": tftags.TagsSchema(),
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"network_interface_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preferred_subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"root_volume_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"copy_tags_to_snapshots": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"data_compression_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OpenZFSDataCompressionType](),
						},
						"nfs_exports": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_configurations": {
										Type:     schema.TypeSet,
										Required: true,
										MaxItems: 25,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"clients": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 128),
														validation.StringMatch(regexache.MustCompile(`^[ -~]{1,128}$`), "must be either IP Address or CIDR"),
													),
												},
												"options": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 20,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringLenBetween(1, 128),
													},
												},
											},
										},
									},
								},
							},
						},
						"read_only": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"record_size_kib": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      128,
							ValidateFunc: validation.IntInSlice([]int{4, 8, 16, 32, 64, 128, 256, 512, 1024}),
						},
						"user_and_group_quotas": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							MaxItems: 100,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrID: {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(0, 2147483647),
									},
									"storage_capacity_quota_gib": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(0, 2147483647),
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.OpenZFSQuotaType](),
									},
								},
							},
						},
					},
				},
			},
			"root_volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"route_table_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"skip_final_backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"storage_capacity": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(64, 512*1024),
			},
			names.AttrStorageType: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.StorageTypeSsd,
				ValidateDiagFunc: enum.Validate[awstypes.StorageType](),
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"throughput_capacity": {
				Type:     schema.TypeInt,
				Required: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weekly_maintenance_start_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(7, 7),
					validation.StringMatch(regexache.MustCompile(`^[1-7]:([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format d:HH:MM"),
				),
			},
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			validateDiskConfigurationIOPS,
			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				var (
					singleAZ1ThroughputCapacityValues            = []int{64, 128, 256, 512, 1024, 2048, 3072, 4096}
					singleAZ2AndMultiAZ1ThroughputCapacityValues = []int{160, 320, 640, 1280, 2560, 3840, 5120, 7680, 10240}
				)

				switch deploymentType, throughputCapacity := d.Get("deployment_type").(string), d.Get("throughput_capacity").(int); deploymentType {
				case string(awstypes.OpenZFSDeploymentTypeSingleAz1):
					if !slices.Contains(singleAZ1ThroughputCapacityValues, throughputCapacity) {
						return fmt.Errorf("%d is not a valid value for `throughput_capacity` when `deployment_type` is %q. Valid values: %v", throughputCapacity, deploymentType, singleAZ1ThroughputCapacityValues)
					}
				case string(awstypes.OpenZFSDeploymentTypeSingleAz2), string(awstypes.OpenZFSDeploymentTypeMultiAz1):
					if !slices.Contains(singleAZ2AndMultiAZ1ThroughputCapacityValues, throughputCapacity) {
						return fmt.Errorf("%d is not a valid value for `throughput_capacity` when `deployment_type` is %q. Valid values: %v", throughputCapacity, deploymentType, singleAZ2AndMultiAZ1ThroughputCapacityValues)
					}
					// default:
					// Allow validation to pass for unknown/new types.
				}

				return nil
			},
		),
	}
}

func validateDiskConfigurationIOPS(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	deploymentType := d.Get("deployment_type").(string)

	if diskConfiguration, ok := d.GetOk("disk_iops_configuration"); ok {
		if len(diskConfiguration.([]interface{})) > 0 {
			m := diskConfiguration.([]interface{})[0].(map[string]interface{})

			if v, ok := m[names.AttrIOPS].(int); ok {
				if deploymentType == string(awstypes.OpenZFSDeploymentTypeSingleAz1) {
					if v < 0 || v > 160000 {
						return fmt.Errorf("expected disk_iops_configuration.0.iops to be in the range (0 - 160000) when deployment_type (%s), got %d", awstypes.OpenZFSDeploymentTypeSingleAz1, v)
					}
				} else if deploymentType == string(awstypes.OpenZFSDeploymentTypeSingleAz2) {
					if v < 0 || v > 350000 {
						return fmt.Errorf("expected disk_iops_configuration.0.iops to be in the range (0 - 350000) when deployment_type (%s), got %d", awstypes.OpenZFSDeploymentTypeSingleAz2, v)
					}
				}
			}
		}
	}

	return nil
}

func resourceOpenZFSFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	inputC := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		FileSystemType:     awstypes.FileSystemTypeOpenzfs,
		OpenZFSConfiguration: &awstypes.CreateFileSystemOpenZFSConfiguration{
			DeploymentType:               awstypes.OpenZFSDeploymentType(d.Get("deployment_type").(string)),
			AutomaticBackupRetentionDays: aws.Int32(int32(d.Get("automatic_backup_retention_days").(int))),
		},
		StorageCapacity: aws.Int32(int32(d.Get("storage_capacity").(int))),
		StorageType:     awstypes.StorageType(d.Get(names.AttrStorageType).(string)),
		SubnetIds:       flex.ExpandStringValueList(d.Get(names.AttrSubnetIDs).([]interface{})),
		Tags:            getTagsIn(ctx),
	}
	inputB := &fsx.CreateFileSystemFromBackupInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		OpenZFSConfiguration: &awstypes.CreateFileSystemOpenZFSConfiguration{
			DeploymentType:               awstypes.OpenZFSDeploymentType(d.Get("deployment_type").(string)),
			AutomaticBackupRetentionDays: aws.Int32(int32(d.Get("automatic_backup_retention_days").(int))),
		},
		StorageType: awstypes.StorageType(d.Get(names.AttrStorageType).(string)),
		SubnetIds:   flex.ExpandStringValueList(d.Get(names.AttrSubnetIDs).([]interface{})),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("copy_tags_to_backups"); ok {
		inputC.OpenZFSConfiguration.CopyTagsToBackups = aws.Bool(v.(bool))
		inputB.OpenZFSConfiguration.CopyTagsToBackups = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("copy_tags_to_volumes"); ok {
		inputC.OpenZFSConfiguration.CopyTagsToVolumes = aws.Bool(v.(bool))
		inputB.OpenZFSConfiguration.CopyTagsToVolumes = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("daily_automatic_backup_start_time"); ok {
		inputC.OpenZFSConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
		inputB.OpenZFSConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disk_iops_configuration"); ok {
		inputC.OpenZFSConfiguration.DiskIopsConfiguration = expandDiskIopsConfiguration(v.([]interface{}))
		inputB.OpenZFSConfiguration.DiskIopsConfiguration = expandDiskIopsConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("endpoint_ip_address_range"); ok {
		inputC.OpenZFSConfiguration.EndpointIpAddressRange = aws.String(v.(string))
		inputB.OpenZFSConfiguration.EndpointIpAddressRange = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		inputC.KmsKeyId = aws.String(v.(string))
		inputB.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_subnet_id"); ok {
		inputC.OpenZFSConfiguration.PreferredSubnetId = aws.String(v.(string))
		inputB.OpenZFSConfiguration.PreferredSubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("root_volume_configuration"); ok {
		inputC.OpenZFSConfiguration.RootVolumeConfiguration = expandOpenZFSCreateRootVolumeConfiguration(v.([]interface{}))
		inputB.OpenZFSConfiguration.RootVolumeConfiguration = expandOpenZFSCreateRootVolumeConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("route_table_ids"); ok {
		inputC.OpenZFSConfiguration.RouteTableIds = flex.ExpandStringValueSet(v.(*schema.Set))
		inputB.OpenZFSConfiguration.RouteTableIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok {
		inputC.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
		inputB.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("throughput_capacity"); ok {
		inputC.OpenZFSConfiguration.ThroughputCapacity = aws.Int32(int32(v.(int)))
		inputB.OpenZFSConfiguration.ThroughputCapacity = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		inputC.OpenZFSConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
		inputB.OpenZFSConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("backup_id"); ok {
		backupID := v.(string)
		inputB.BackupId = aws.String(backupID)

		output, err := conn.CreateFileSystemFromBackup(ctx, inputB)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating FSx for OpenZFS File System from backup (%s): %s", backupID, err)
		}

		d.SetId(aws.ToString(output.FileSystem.FileSystemId))
	} else {
		output, err := conn.CreateFileSystem(ctx, inputC)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating FSx for OpenZFS File System: %s", err)
		}

		d.SetId(aws.ToString(output.FileSystem.FileSystemId))
	}

	if _, err := waitFileSystemCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS File System (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOpenZFSFileSystemRead(ctx, d, meta)...)
}

func resourceOpenZFSFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	filesystem, err := findOpenZFSFileSystemByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx for OpenZFS File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for OpenZFS File System (%s): %s", d.Id(), err)
	}

	openZFSConfig := filesystem.OpenZFSConfiguration

	d.Set(names.AttrARN, filesystem.ResourceARN)
	d.Set("automatic_backup_retention_days", openZFSConfig.AutomaticBackupRetentionDays)
	d.Set("copy_tags_to_backups", openZFSConfig.CopyTagsToBackups)
	d.Set("copy_tags_to_volumes", openZFSConfig.CopyTagsToVolumes)
	d.Set("daily_automatic_backup_start_time", openZFSConfig.DailyAutomaticBackupStartTime)
	d.Set("deployment_type", openZFSConfig.DeploymentType)
	if err := d.Set("disk_iops_configuration", flattenDiskIopsConfiguration(openZFSConfig.DiskIopsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting disk_iops_configuration: %s", err)
	}
	d.Set(names.AttrDNSName, filesystem.DNSName)
	d.Set("endpoint_ip_address", openZFSConfig.EndpointIpAddress)
	d.Set("endpoint_ip_address_range", openZFSConfig.EndpointIpAddressRange)
	d.Set(names.AttrKMSKeyID, filesystem.KmsKeyId)
	d.Set("network_interface_ids", filesystem.NetworkInterfaceIds)
	d.Set(names.AttrOwnerID, filesystem.OwnerId)
	d.Set("preferred_subnet_id", openZFSConfig.PreferredSubnetId)
	rootVolumeID := aws.ToString(openZFSConfig.RootVolumeId)
	d.Set("root_volume_id", rootVolumeID)
	d.Set("route_table_ids", openZFSConfig.RouteTableIds)
	d.Set("storage_capacity", filesystem.StorageCapacity)
	d.Set(names.AttrStorageType, filesystem.StorageType)
	d.Set(names.AttrSubnetIDs, filesystem.SubnetIds)
	d.Set("throughput_capacity", openZFSConfig.ThroughputCapacity)
	d.Set(names.AttrVPCID, filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", openZFSConfig.WeeklyMaintenanceStartTime)

	// FS tags aren't set in the Describe response.
	// setTagsOut(ctx, filesystem.Tags)

	rootVolume, err := findOpenZFSVolumeByID(ctx, conn, rootVolumeID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for OpenZFS File System (%s) root volume (%s): %s", d.Id(), rootVolumeID, err)
	}

	if err := d.Set("root_volume_configuration", flattenOpenZFSFileSystemRootVolume(rootVolume)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting root_volume_configuration: %s", err)
	}

	return diags
}

func resourceOpenZFSFileSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	if d.HasChangesExcept(
		"delete_options",
		"final_backup_tags",
		"skip_final_backup",
		names.AttrTags,
		names.AttrTagsAll,
	) {
		input := &fsx.UpdateFileSystemInput{
			ClientRequestToken:   aws.String(id.UniqueId()),
			FileSystemId:         aws.String(d.Id()),
			OpenZFSConfiguration: &awstypes.UpdateFileSystemOpenZFSConfiguration{},
		}

		if d.HasChange("automatic_backup_retention_days") {
			input.OpenZFSConfiguration.AutomaticBackupRetentionDays = aws.Int32(int32(d.Get("automatic_backup_retention_days").(int)))
		}

		if d.HasChange("copy_tags_to_backups") {
			input.OpenZFSConfiguration.CopyTagsToBackups = aws.Bool(d.Get("copy_tags_to_backups").(bool))
		}

		if d.HasChange("copy_tags_to_volumes") {
			input.OpenZFSConfiguration.CopyTagsToVolumes = aws.Bool(d.Get("copy_tags_to_volumes").(bool))
		}

		if d.HasChange("daily_automatic_backup_start_time") {
			input.OpenZFSConfiguration.DailyAutomaticBackupStartTime = aws.String(d.Get("daily_automatic_backup_start_time").(string))
		}

		if d.HasChange("disk_iops_configuration") {
			input.OpenZFSConfiguration.DiskIopsConfiguration = expandDiskIopsConfiguration(d.Get("disk_iops_configuration").([]interface{}))
		}

		if d.HasChange("route_table_ids") {
			o, n := d.GetChange("route_table_ids")
			os, ns := o.(*schema.Set), n.(*schema.Set)
			add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

			if len(add) > 0 {
				input.OpenZFSConfiguration.AddRouteTableIds = add
			}
			if len(del) > 0 {
				input.OpenZFSConfiguration.RemoveRouteTableIds = del
			}
		}

		if d.HasChange("storage_capacity") {
			input.StorageCapacity = aws.Int32(int32(d.Get("storage_capacity").(int)))
		}

		if d.HasChange("throughput_capacity") {
			input.OpenZFSConfiguration.ThroughputCapacity = aws.Int32(int32(d.Get("throughput_capacity").(int)))
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.OpenZFSConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		startTime := time.Now()
		_, err := conn.UpdateFileSystem(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx for OpenZFS File System (%s): %s", d.Id(), err)
		}

		if _, err := waitFileSystemUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS File System (%s) update: %s", d.Id(), err)
		}

		if _, err := waitFileSystemAdministrativeActionCompleted(ctx, conn, d.Id(), awstypes.AdministrativeActionTypeFileSystemUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS File System (%s) administrative action (%s) complete: %s", d.Id(), awstypes.AdministrativeActionTypeFileSystemUpdate, err)
		}

		if d.HasChange("root_volume_configuration") {
			rootVolumeID := d.Get("root_volume_id").(string)
			input := &fsx.UpdateVolumeInput{
				ClientRequestToken:   aws.String(id.UniqueId()),
				OpenZFSConfiguration: expandUpdateOpenZFSVolumeConfiguration(d.Get("root_volume_configuration").([]interface{})),
				VolumeId:             aws.String(rootVolumeID),
			}

			startTime := time.Now()
			_, err := conn.UpdateVolume(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating FSx for OpenZFS Root Volume (%s): %s", rootVolumeID, err)
			}

			if _, err := waitVolumeUpdated(ctx, conn, rootVolumeID, startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS Root Volume (%s) update: %s", rootVolumeID, err)
			}

			if _, err := waitVolumeAdministrativeActionCompleted(ctx, conn, rootVolumeID, awstypes.AdministrativeActionTypeVolumeUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS Volume (%s) administrative action (%s) complete: %s", rootVolumeID, awstypes.AdministrativeActionTypeVolumeUpdate, err)
			}
		}
	}

	return append(diags, resourceOpenZFSFileSystemRead(ctx, d, meta)...)
}

func resourceOpenZFSFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := &fsx.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
		OpenZFSConfiguration: &awstypes.DeleteFileSystemOpenZFSConfiguration{
			SkipFinalBackup: aws.Bool(d.Get("skip_final_backup").(bool)),
		},
	}

	if v, ok := d.GetOk("delete_options"); ok {
		input.OpenZFSConfiguration.Options = flex.ExpandStringyValueSet[awstypes.DeleteFileSystemOpenZFSOption](v.(*schema.Set))
	}

	if v, ok := d.GetOk("final_backup_tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.OpenZFSConfiguration.FinalBackupTags = Tags(tftags.New(ctx, v))
	}

	log.Printf("[DEBUG] Deleting FSx for OpenZFS File System: %s", d.Id())
	_, err := conn.DeleteFileSystem(ctx, input)

	if errs.IsA[*awstypes.FileSystemNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx for OpenZFS File System (%s): %s", d.Id(), err)
	}

	if _, err := waitFileSystemDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS File System (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandDiskIopsConfiguration(cfg []interface{}) *awstypes.DiskIopsConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := awstypes.DiskIopsConfiguration{}

	if v, ok := conf[names.AttrMode].(string); ok && len(v) > 0 {
		out.Mode = awstypes.DiskIopsConfigurationMode(v)
	}

	if v, ok := conf[names.AttrIOPS].(int); ok {
		out.Iops = aws.Int64(int64(v))
	}

	return &out
}

func expandOpenZFSCreateRootVolumeConfiguration(tfList []interface{}) *awstypes.OpenZFSCreateRootVolumeConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.OpenZFSCreateRootVolumeConfiguration{}

	if v, ok := tfMap["copy_tags_to_snapshots"].(bool); ok {
		apiObject.CopyTagsToSnapshots = aws.Bool(v)
	}

	if v, ok := tfMap["data_compression_type"].(string); ok {
		apiObject.DataCompressionType = awstypes.OpenZFSDataCompressionType(v)
	}

	if v, ok := tfMap["nfs_exports"].([]interface{}); ok {
		apiObject.NfsExports = expandOpenZFSNfsExports(v)
	}

	if v, ok := tfMap["read_only"].(bool); ok {
		apiObject.ReadOnly = aws.Bool(v)
	}

	if v, ok := tfMap["record_size_kib"].(int); ok {
		apiObject.RecordSizeKiB = aws.Int32(int32(v))
	}

	if v, ok := tfMap["user_and_group_quotas"]; ok {
		apiObject.UserAndGroupQuotas = expandOpenZFSUserOrGroupQuotas(v.(*schema.Set).List())
	}

	return apiObject
}

func expandUpdateOpenZFSVolumeConfiguration(tfList []interface{}) *awstypes.UpdateOpenZFSVolumeConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.UpdateOpenZFSVolumeConfiguration{}

	if v, ok := tfMap["data_compression_type"].(string); ok {
		apiObject.DataCompressionType = awstypes.OpenZFSDataCompressionType(v)
	}

	if v, ok := tfMap["nfs_exports"].([]interface{}); ok {
		apiObject.NfsExports = expandOpenZFSNfsExports(v)
	}

	if v, ok := tfMap["read_only"].(bool); ok {
		apiObject.ReadOnly = aws.Bool(v)
	}

	if v, ok := tfMap["record_size_kib"].(int); ok {
		apiObject.RecordSizeKiB = aws.Int32(int32(v))
	}

	if v, ok := tfMap["user_and_group_quotas"]; ok {
		apiObject.UserAndGroupQuotas = expandOpenZFSUserOrGroupQuotas(v.(*schema.Set).List())
	}

	return apiObject
}

func flattenDiskIopsConfiguration(rs *awstypes.DiskIopsConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	m[names.AttrMode] = string(rs.Mode)
	if rs.Iops != nil {
		m[names.AttrIOPS] = aws.ToInt64(rs.Iops)
	}

	return []interface{}{m}
}

func flattenOpenZFSFileSystemRootVolume(apiObject *awstypes.Volume) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := make(map[string]interface{})

	if apiObject.OpenZFSConfiguration.CopyTagsToSnapshots != nil {
		tfMap["copy_tags_to_snapshots"] = aws.ToBool(apiObject.OpenZFSConfiguration.CopyTagsToSnapshots)
	}
	tfMap["data_compression_type"] = string(apiObject.OpenZFSConfiguration.DataCompressionType)
	if apiObject.OpenZFSConfiguration.NfsExports != nil {
		tfMap["nfs_exports"] = flattenOpenZFSNfsExports(apiObject.OpenZFSConfiguration.NfsExports)
	}
	if apiObject.OpenZFSConfiguration.ReadOnly != nil {
		tfMap["read_only"] = aws.ToBool(apiObject.OpenZFSConfiguration.ReadOnly)
	}
	if apiObject.OpenZFSConfiguration.RecordSizeKiB != nil {
		tfMap["record_size_kib"] = aws.ToInt32(apiObject.OpenZFSConfiguration.RecordSizeKiB)
	}
	if apiObject.OpenZFSConfiguration.UserAndGroupQuotas != nil {
		tfMap["user_and_group_quotas"] = flattenOpenZFSUserOrGroupQuotas(apiObject.OpenZFSConfiguration.UserAndGroupQuotas)
	}

	return []interface{}{tfMap}
}

func findOpenZFSFileSystemByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.FileSystem, error) {
	output, err := findFileSystemByIDAndType(ctx, conn, id, awstypes.FileSystemTypeOpenzfs)

	if err != nil {
		return nil, err
	}

	if output.OpenZFSConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}
