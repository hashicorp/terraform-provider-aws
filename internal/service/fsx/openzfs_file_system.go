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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"deployment_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(fsx.OpenZFSDeploymentType_Values(), false),
			},
			"disk_iops_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iops": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      fsx.DiskIopsConfigurationModeAutomatic,
							ValidateFunc: validation.StringInSlice(fsx.DiskIopsConfigurationMode_Values(), false),
						},
					},
				},
			},
			"dns_name": {
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(fsx.OpenZFSDataCompressionType_Values(), false),
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(fsx.OpenZFSQuotaType_Values(), false),
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
			"security_group_ids": {
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
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.StorageTypeSsd,
				ValidateFunc: validation.StringInSlice(fsx.StorageType_Values(), false),
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
				case fsx.OpenZFSDeploymentTypeSingleAz1:
					if !slices.Contains(singleAZ1ThroughputCapacityValues, throughputCapacity) {
						return fmt.Errorf("%d is not a valid value for `throughput_capacity` when `deployment_type` is %q. Valid values: %v", throughputCapacity, deploymentType, singleAZ1ThroughputCapacityValues)
					}
				case fsx.OpenZFSDeploymentTypeSingleAz2, fsx.OpenZFSDeploymentTypeMultiAz1:
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

			if v, ok := m["iops"].(int); ok {
				if deploymentType == fsx.OpenZFSDeploymentTypeSingleAz1 {
					if v < 0 || v > 160000 {
						return fmt.Errorf("expected disk_iops_configuration.0.iops to be in the range (0 - 160000) when deployment_type (%s), got %d", fsx.OpenZFSDeploymentTypeSingleAz1, v)
					}
				} else if deploymentType == fsx.OpenZFSDeploymentTypeSingleAz2 {
					if v < 0 || v > 350000 {
						return fmt.Errorf("expected disk_iops_configuration.0.iops to be in the range (0 - 350000) when deployment_type (%s), got %d", fsx.OpenZFSDeploymentTypeSingleAz2, v)
					}
				}
			}
		}
	}

	return nil
}

func resourceOpenZFSFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	inputC := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		FileSystemType:     aws.String(fsx.FileSystemTypeOpenzfs),
		OpenZFSConfiguration: &fsx.CreateFileSystemOpenZFSConfiguration{
			DeploymentType:               aws.String(d.Get("deployment_type").(string)),
			AutomaticBackupRetentionDays: aws.Int64(int64(d.Get("automatic_backup_retention_days").(int))),
		},
		StorageCapacity: aws.Int64(int64(d.Get("storage_capacity").(int))),
		StorageType:     aws.String(d.Get("storage_type").(string)),
		SubnetIds:       flex.ExpandStringList(d.Get(names.AttrSubnetIDs).([]interface{})),
		Tags:            getTagsIn(ctx),
	}
	inputB := &fsx.CreateFileSystemFromBackupInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		OpenZFSConfiguration: &fsx.CreateFileSystemOpenZFSConfiguration{
			DeploymentType:               aws.String(d.Get("deployment_type").(string)),
			AutomaticBackupRetentionDays: aws.Int64(int64(d.Get("automatic_backup_retention_days").(int))),
		},
		StorageType: aws.String(d.Get("storage_type").(string)),
		SubnetIds:   flex.ExpandStringList(d.Get(names.AttrSubnetIDs).([]interface{})),
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
		inputC.OpenZFSConfiguration.RouteTableIds = flex.ExpandStringSet(v.(*schema.Set))
		inputB.OpenZFSConfiguration.RouteTableIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		inputC.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		inputB.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("throughput_capacity"); ok {
		inputC.OpenZFSConfiguration.ThroughputCapacity = aws.Int64(int64(v.(int)))
		inputB.OpenZFSConfiguration.ThroughputCapacity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		inputC.OpenZFSConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
		inputB.OpenZFSConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("backup_id"); ok {
		backupID := v.(string)
		inputB.BackupId = aws.String(backupID)

		output, err := conn.CreateFileSystemFromBackupWithContext(ctx, inputB)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating FSx for OpenZFS File System from backup (%s): %s", backupID, err)
		}

		d.SetId(aws.StringValue(output.FileSystem.FileSystemId))
	} else {
		output, err := conn.CreateFileSystemWithContext(ctx, inputC)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating FSx for OpenZFS File System: %s", err)
		}

		d.SetId(aws.StringValue(output.FileSystem.FileSystemId))
	}

	if _, err := waitFileSystemCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS File System (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOpenZFSFileSystemRead(ctx, d, meta)...)
}

func resourceOpenZFSFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

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
	d.Set("dns_name", filesystem.DNSName)
	d.Set("endpoint_ip_address", openZFSConfig.EndpointIpAddress)
	d.Set("endpoint_ip_address_range", openZFSConfig.EndpointIpAddressRange)
	d.Set(names.AttrKMSKeyID, filesystem.KmsKeyId)
	d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds))
	d.Set(names.AttrOwnerID, filesystem.OwnerId)
	d.Set("preferred_subnet_id", openZFSConfig.PreferredSubnetId)
	rootVolumeID := aws.StringValue(openZFSConfig.RootVolumeId)
	d.Set("root_volume_id", rootVolumeID)
	d.Set("route_table_ids", aws.StringValueSlice(openZFSConfig.RouteTableIds))
	d.Set("storage_capacity", filesystem.StorageCapacity)
	d.Set("storage_type", filesystem.StorageType)
	d.Set(names.AttrSubnetIDs, aws.StringValueSlice(filesystem.SubnetIds))
	d.Set("throughput_capacity", openZFSConfig.ThroughputCapacity)
	d.Set(names.AttrVPCID, filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", openZFSConfig.WeeklyMaintenanceStartTime)

	setTagsOut(ctx, filesystem.Tags)

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
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fsx.UpdateFileSystemInput{
			ClientRequestToken:   aws.String(id.UniqueId()),
			FileSystemId:         aws.String(d.Id()),
			OpenZFSConfiguration: &fsx.UpdateFileSystemOpenZFSConfiguration{},
		}

		if d.HasChange("automatic_backup_retention_days") {
			input.OpenZFSConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(d.Get("automatic_backup_retention_days").(int)))
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
				input.OpenZFSConfiguration.AddRouteTableIds = aws.StringSlice(add)
			}
			if len(del) > 0 {
				input.OpenZFSConfiguration.RemoveRouteTableIds = aws.StringSlice(del)
			}
		}

		if d.HasChange("storage_capacity") {
			input.StorageCapacity = aws.Int64(int64(d.Get("storage_capacity").(int)))
		}

		if d.HasChange("throughput_capacity") {
			input.OpenZFSConfiguration.ThroughputCapacity = aws.Int64(int64(d.Get("throughput_capacity").(int)))
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.OpenZFSConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		startTime := time.Now()
		_, err := conn.UpdateFileSystemWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx for OpenZFS File System (%s): %s", d.Id(), err)
		}

		if _, err := waitFileSystemUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS File System (%s) update: %s", d.Id(), err)
		}

		if _, err := waitFileSystemAdministrativeActionCompleted(ctx, conn, d.Id(), fsx.AdministrativeActionTypeFileSystemUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS File System (%s) administrative action (%s) complete: %s", d.Id(), fsx.AdministrativeActionTypeFileSystemUpdate, err)
		}

		if d.HasChange("root_volume_configuration") {
			rootVolumeID := d.Get("root_volume_id").(string)
			input := &fsx.UpdateVolumeInput{
				ClientRequestToken:   aws.String(id.UniqueId()),
				OpenZFSConfiguration: expandUpdateOpenZFSVolumeConfiguration(d.Get("root_volume_configuration").([]interface{})),
				VolumeId:             aws.String(rootVolumeID),
			}

			startTime := time.Now()
			_, err := conn.UpdateVolumeWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating FSx for OpenZFS Root Volume (%s): %s", rootVolumeID, err)
			}

			if _, err := waitVolumeUpdated(ctx, conn, rootVolumeID, startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS Root Volume (%s) update: %s", rootVolumeID, err)
			}

			if _, err := waitVolumeAdministrativeActionCompleted(ctx, conn, rootVolumeID, fsx.AdministrativeActionTypeVolumeUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS Volume (%s) administrative action (%s) complete: %s", rootVolumeID, fsx.AdministrativeActionTypeVolumeUpdate, err)
			}
		}
	}

	return append(diags, resourceOpenZFSFileSystemRead(ctx, d, meta)...)
}

func resourceOpenZFSFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	log.Printf("[DEBUG] Deleting FSx for OpenZFS File System: %s", d.Id())
	_, err := conn.DeleteFileSystemWithContext(ctx, &fsx.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
		OpenZFSConfiguration: &fsx.DeleteFileSystemOpenZFSConfiguration{
			SkipFinalBackup: aws.Bool(d.Get("skip_final_backup").(bool)),
		},
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) {
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

func expandDiskIopsConfiguration(cfg []interface{}) *fsx.DiskIopsConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.DiskIopsConfiguration{}

	if v, ok := conf["mode"].(string); ok && len(v) > 0 {
		out.Mode = aws.String(v)
	}

	if v, ok := conf["iops"].(int); ok {
		out.Iops = aws.Int64(int64(v))
	}

	return &out
}

func expandOpenZFSCreateRootVolumeConfiguration(cfg []interface{}) *fsx.OpenZFSCreateRootVolumeConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.OpenZFSCreateRootVolumeConfiguration{}

	if v, ok := conf["copy_tags_to_snapshots"].(bool); ok {
		out.CopyTagsToSnapshots = aws.Bool(v)
	}

	if v, ok := conf["data_compression_type"].(string); ok {
		out.DataCompressionType = aws.String(v)
	}

	if v, ok := conf["read_only"].(bool); ok {
		out.ReadOnly = aws.Bool(v)
	}

	if v, ok := conf["record_size_kib"].(int); ok {
		out.RecordSizeKiB = aws.Int64(int64(v))
	}

	if v, ok := conf["user_and_group_quotas"]; ok {
		out.UserAndGroupQuotas = expandOpenZFSUserOrGroupQuotas(v.(*schema.Set).List())
	}

	if v, ok := conf["nfs_exports"].([]interface{}); ok {
		out.NfsExports = expandOpenZFSNfsExports(v)
	}

	return &out
}

func expandUpdateOpenZFSVolumeConfiguration(cfg []interface{}) *fsx.UpdateOpenZFSVolumeConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.UpdateOpenZFSVolumeConfiguration{}

	if v, ok := conf["data_compression_type"].(string); ok {
		out.DataCompressionType = aws.String(v)
	}

	if v, ok := conf["read_only"].(bool); ok {
		out.ReadOnly = aws.Bool(v)
	}

	if v, ok := conf["record_size_kib"].(int); ok {
		out.RecordSizeKiB = aws.Int64(int64(v))
	}

	if v, ok := conf["user_and_group_quotas"]; ok {
		out.UserAndGroupQuotas = expandOpenZFSUserOrGroupQuotas(v.(*schema.Set).List())
	}

	if v, ok := conf["nfs_exports"].([]interface{}); ok {
		out.NfsExports = expandOpenZFSNfsExports(v)
	}

	return &out
}

func flattenDiskIopsConfiguration(rs *fsx.DiskIopsConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.Mode != nil {
		m["mode"] = aws.StringValue(rs.Mode)
	}
	if rs.Iops != nil {
		m["iops"] = aws.Int64Value(rs.Iops)
	}

	return []interface{}{m}
}

func flattenOpenZFSFileSystemRootVolume(rs *fsx.Volume) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.OpenZFSConfiguration.CopyTagsToSnapshots != nil {
		m["copy_tags_to_snapshots"] = aws.BoolValue(rs.OpenZFSConfiguration.CopyTagsToSnapshots)
	}
	if rs.OpenZFSConfiguration.DataCompressionType != nil {
		m["data_compression_type"] = aws.StringValue(rs.OpenZFSConfiguration.DataCompressionType)
	}
	if rs.OpenZFSConfiguration.NfsExports != nil {
		m["nfs_exports"] = flattenOpenZFSNfsExports(rs.OpenZFSConfiguration.NfsExports)
	}
	if rs.OpenZFSConfiguration.ReadOnly != nil {
		m["read_only"] = aws.BoolValue(rs.OpenZFSConfiguration.ReadOnly)
	}
	if rs.OpenZFSConfiguration.RecordSizeKiB != nil {
		m["record_size_kib"] = aws.Int64Value(rs.OpenZFSConfiguration.RecordSizeKiB)
	}
	if rs.OpenZFSConfiguration.UserAndGroupQuotas != nil {
		m["user_and_group_quotas"] = flattenOpenZFSUserOrGroupQuotas(rs.OpenZFSConfiguration.UserAndGroupQuotas)
	}

	return []interface{}{m}
}

func findOpenZFSFileSystemByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.FileSystem, error) {
	output, err := findFileSystemByIDAndType(ctx, conn, id, fsx.FileSystemTypeOpenzfs)

	if err != nil {
		return nil, err
	}

	if output.OpenZFSConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}
