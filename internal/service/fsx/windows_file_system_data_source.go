// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_fsx_windows_file_system")
func DataSourceWindowsFileSystem() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWindowsFileSystemRead,

		Schema: map[string]*schema.Schema{
			"active_directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aliases": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"audit_log_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"audit_log_destination": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"file_access_audit_log_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"file_share_access_audit_log_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"automatic_backup_retention_days": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"backup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"copy_tags_to_backups": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"daily_automatic_backup_start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_iops_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preferred_file_server_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preferred_subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"skip_final_backup": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"storage_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"storage_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tags": tftags.TagsSchemaComputed(),
			"throughput_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weekly_maintenance_start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceWindowsFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).FSxConn(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Get("id").(string)
	filesystem, err := FindWindowsFileSystemByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for Windows File Server File System (%s): %s", id, err)
	}

	windowsConfig := filesystem.WindowsConfiguration

	d.SetId(aws.StringValue(filesystem.FileSystemId))
	d.Set("active_directory_id", windowsConfig.ActiveDirectoryId)
	d.Set("aliases", aws.StringValueSlice(expandAliasValues(windowsConfig.Aliases)))
	d.Set("arn", filesystem.ResourceARN)
	if err := d.Set("audit_log_configuration", flattenWindowsAuditLogConfiguration(windowsConfig.AuditLogConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting audit_log_configuration: %s", err)
	}
	d.Set("automatic_backup_retention_days", windowsConfig.AutomaticBackupRetentionDays)
	d.Set("copy_tags_to_backups", windowsConfig.CopyTagsToBackups)
	d.Set("daily_automatic_backup_start_time", windowsConfig.DailyAutomaticBackupStartTime)
	d.Set("deployment_type", windowsConfig.DeploymentType)
	if err := d.Set("disk_iops_configuration", flattenWindowsDiskIopsConfiguration(windowsConfig.DiskIopsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting disk_iops_configuration: %s", err)
	}
	d.Set("dns_name", filesystem.DNSName)
	d.Set("id", filesystem.FileSystemId)
	d.Set("kms_key_id", filesystem.KmsKeyId)
	d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds))
	d.Set("owner_id", filesystem.OwnerId)
	d.Set("preferred_file_server_ip", windowsConfig.PreferredFileServerIp)
	d.Set("preferred_subnet_id", windowsConfig.PreferredSubnetId)
	d.Set("storage_capacity", filesystem.StorageCapacity)
	d.Set("storage_type", filesystem.StorageType)
	d.Set("subnet_ids", aws.StringValueSlice(filesystem.SubnetIds))
	d.Set("throughput_capacity", windowsConfig.ThroughputCapacity)
	d.Set("vpc_id", filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", windowsConfig.WeeklyMaintenanceStartTime)

	tags := KeyValueTags(ctx, filesystem.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
