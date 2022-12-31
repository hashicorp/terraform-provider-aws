package fsx

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceWindowsFileSystem() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWindowsFileSystemRead,

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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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

func dataSourceWindowsFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Get("id").(string)

	filesystem, err := FindFileSystemByID(conn, id)

	if err != nil {
		return fmt.Errorf("error reading FSx Windows File System (%s): %w", d.Id(), err)
	}

	if filesystem.LustreConfiguration != nil {
		return fmt.Errorf("expected FSx Windows File System, found FSx Lustre File System: %s", d.Id())
	}

	if filesystem.WindowsConfiguration == nil {
		return fmt.Errorf("error describing FSx Windows File System (%s): empty Windows configuration", d.Id())
	}

	d.Set("active_directory_id", filesystem.WindowsConfiguration.ActiveDirectoryId)
	d.Set("arn", filesystem.ResourceARN)
	d.Set("automatic_backup_retention_days", filesystem.WindowsConfiguration.AutomaticBackupRetentionDays)
	d.Set("copy_tags_to_backups", filesystem.WindowsConfiguration.CopyTagsToBackups)
	d.Set("daily_automatic_backup_start_time", filesystem.WindowsConfiguration.DailyAutomaticBackupStartTime)
	d.Set("deployment_type", filesystem.WindowsConfiguration.DeploymentType)
	d.Set("dns_name", filesystem.DNSName)
	d.Set("id", filesystem.FileSystemId)
	d.Set("kms_key_id", filesystem.KmsKeyId)
	d.Set("owner_id", filesystem.OwnerId)
	d.Set("preferred_subnet_id", filesystem.WindowsConfiguration.PreferredSubnetId)
	d.Set("preferred_file_server_ip", filesystem.WindowsConfiguration.PreferredFileServerIp)
	d.Set("storage_capacity", filesystem.StorageCapacity)
	d.Set("storage_type", filesystem.StorageType)
	d.Set("throughput_capacity", filesystem.WindowsConfiguration.ThroughputCapacity)
	d.Set("vpc_id", filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", filesystem.WindowsConfiguration.WeeklyMaintenanceStartTime)

	if err := d.Set("aliases", aws.StringValueSlice(expandAliasValues(filesystem.WindowsConfiguration.Aliases))); err != nil {
		return fmt.Errorf("error setting aliases: %w", err)
	}

	if err := d.Set("audit_log_configuration", flattenWindowsAuditLogConfiguration(filesystem.WindowsConfiguration.AuditLogConfiguration)); err != nil {
		return fmt.Errorf("error setting audit_log_configuration: %w", err)
	}

	if err := d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds)); err != nil {
		return fmt.Errorf("error setting network_interface_ids: %w", err)
	}

	if err := d.Set("subnet_ids", aws.StringValueSlice(filesystem.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %w", err)
	}

	tags := KeyValueTags(filesystem.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}
