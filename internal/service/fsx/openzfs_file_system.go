package fsx

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOpenzfsFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceOepnzfsFileSystemCreate,
		Read:   resourceOpenzfsFileSystemRead,
		Update: resourceOpenzfsFileSystemUpdate,
		Delete: resourceOpenzfsFileSystemDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
					validation.StringMatch(regexp.MustCompile(`^([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format HH:MM"),
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
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(0, 160000),
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
														validation.StringMatch(regexp.MustCompile(`^[ -~]{1,128}$`), "must be either IP Address or CIDR"),
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
						"user_and_group_quotas": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							MaxItems: 100,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(0, 2147483647),
									},
									"storage_capacity_quota_gib": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(0, 2147483647),
									},
									"type": {
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
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_id": {
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
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"storage_capacity": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(64, 512*1024),
			},
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.StorageTypeSsd,
				ValidateFunc: validation.StringInSlice(fsx.StorageType_Values(), false),
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"throughput_capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntInSlice([]int{64, 128, 256, 512, 1024, 2048, 3072, 4096}),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weekly_maintenance_start_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(7, 7),
					validation.StringMatch(regexp.MustCompile(`^[1-7]:([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format d:HH:MM"),
				),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOepnzfsFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemType:     aws.String(fsx.FileSystemTypeOpenzfs),
		StorageCapacity:    aws.Int64(int64(d.Get("storage_capacity").(int))),
		StorageType:        aws.String(d.Get("storage_type").(string)),
		SubnetIds:          flex.ExpandStringList(d.Get("subnet_ids").([]interface{})),
		OpenZFSConfiguration: &fsx.CreateFileSystemOpenZFSConfiguration{
			DeploymentType:               aws.String(d.Get("deployment_type").(string)),
			AutomaticBackupRetentionDays: aws.Int64(int64(d.Get("automatic_backup_retention_days").(int))),
		},
	}

	backupInput := &fsx.CreateFileSystemFromBackupInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		StorageType:        aws.String(d.Get("storage_type").(string)),
		SubnetIds:          flex.ExpandStringList(d.Get("subnet_ids").([]interface{})),
		OpenZFSConfiguration: &fsx.CreateFileSystemOpenZFSConfiguration{
			DeploymentType:               aws.String(d.Get("deployment_type").(string)),
			AutomaticBackupRetentionDays: aws.Int64(int64(d.Get("automatic_backup_retention_days").(int))),
		},
	}

	if v, ok := d.GetOk("disk_iops_configuration"); ok {
		input.OpenZFSConfiguration.DiskIopsConfiguration = expandOpenzfsFileDiskIopsConfiguration(v.([]interface{}))
		backupInput.OpenZFSConfiguration.DiskIopsConfiguration = expandOpenzfsFileDiskIopsConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("root_volume_configuration"); ok {
		input.OpenZFSConfiguration.RootVolumeConfiguration = expandOpenzfsRootVolumeConfiguration(v.([]interface{}))
		backupInput.OpenZFSConfiguration.RootVolumeConfiguration = expandOpenzfsRootVolumeConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
		backupInput.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("daily_automatic_backup_start_time"); ok {
		input.OpenZFSConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
		backupInput.OpenZFSConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		backupInput.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
		backupInput.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		input.OpenZFSConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
		backupInput.OpenZFSConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("throughput_capacity"); ok {
		input.OpenZFSConfiguration.ThroughputCapacity = aws.Int64(int64(v.(int)))
		backupInput.OpenZFSConfiguration.ThroughputCapacity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("copy_tags_to_backups"); ok {
		input.OpenZFSConfiguration.CopyTagsToBackups = aws.Bool(v.(bool))
		backupInput.OpenZFSConfiguration.CopyTagsToBackups = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("copy_tags_to_volumes"); ok {
		input.OpenZFSConfiguration.CopyTagsToVolumes = aws.Bool(v.(bool))
		backupInput.OpenZFSConfiguration.CopyTagsToVolumes = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("backup_id"); ok {
		backupInput.BackupId = aws.String(v.(string))

		log.Printf("[DEBUG] Creating FSx OpenZFS File System: %s", backupInput)
		result, err := conn.CreateFileSystemFromBackup(backupInput)

		if err != nil {
			return fmt.Errorf("error creating FSx OpenZFS File System from backup: %w", err)
		}

		d.SetId(aws.StringValue(result.FileSystem.FileSystemId))
	} else {
		log.Printf("[DEBUG] Creating FSx OpenZFS File System: %s", input)
		result, err := conn.CreateFileSystem(input)

		if err != nil {
			return fmt.Errorf("error creating FSx OpenZFS File System: %w", err)
		}

		d.SetId(aws.StringValue(result.FileSystem.FileSystemId))
	}

	if _, err := waitFileSystemCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for FSx OpenZFS File System (%s) create: %w", d.Id(), err)
	}

	return resourceOpenzfsFileSystemRead(d, meta)
}

func resourceOpenzfsFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	filesystem, err := FindFileSystemByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx OpenZFS File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FSx OpenZFS File System (%s): %w", d.Id(), err)
	}

	openzfsConfig := filesystem.OpenZFSConfiguration

	if filesystem.WindowsConfiguration != nil {
		return fmt.Errorf("expected FSx OpenZFS File System, found FSx Windows File System: %s", d.Id())
	}

	if filesystem.LustreConfiguration != nil {
		return fmt.Errorf("expected FSx OpenZFS File System, found FSx Lustre File System: %s", d.Id())
	}

	if filesystem.OntapConfiguration != nil {
		return fmt.Errorf("expected FSx OpeZFS File System, found FSx ONTAP File System: %s", d.Id())
	}

	if openzfsConfig == nil {
		return fmt.Errorf("error describing FSx OpenZFS File System (%s): empty Openzfs configuration", d.Id())
	}

	d.Set("arn", filesystem.ResourceARN)
	d.Set("dns_name", filesystem.DNSName)
	d.Set("deployment_type", openzfsConfig.DeploymentType)
	if openzfsConfig.ThroughputCapacity != nil {
		d.Set("throughput_capacity", openzfsConfig.ThroughputCapacity)
	}
	d.Set("storage_type", filesystem.StorageType)

	if filesystem.KmsKeyId != nil {
		d.Set("kms_key_id", filesystem.KmsKeyId)
	}

	if err := d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds)); err != nil {
		return fmt.Errorf("error setting network_interface_ids: %w", err)
	}

	d.Set("owner_id", filesystem.OwnerId)
	d.Set("root_volume_id", openzfsConfig.RootVolumeId)
	d.Set("storage_capacity", filesystem.StorageCapacity)

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

	if err := d.Set("disk_iops_configuration", flattenOpenzfsFileDiskIopsConfiguration(openzfsConfig.DiskIopsConfiguration)); err != nil {
		return fmt.Errorf("error setting disk_iops_configuration: %w", err)
	}

	d.Set("vpc_id", filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", openzfsConfig.WeeklyMaintenanceStartTime)
	d.Set("automatic_backup_retention_days", openzfsConfig.AutomaticBackupRetentionDays)
	d.Set("daily_automatic_backup_start_time", openzfsConfig.DailyAutomaticBackupStartTime)
	d.Set("copy_tags_to_backups", openzfsConfig.CopyTagsToBackups)
	d.Set("copy_tags_to_volumes", openzfsConfig.CopyTagsToVolumes)

	rootVolume, err := FindVolumeByID(conn, *openzfsConfig.RootVolumeId)

	if err != nil {
		return fmt.Errorf("error reading FSx OpenZFS Root Volume Configuration (%s): %w", *openzfsConfig.RootVolumeId, err)
	}

	if err := d.Set("root_volume_configuration", flattenOpenzfsRootVolumeConfiguration(rootVolume)); err != nil {
		return fmt.Errorf("error setting root_volume_configuration: %w", err)
	}

	return nil
}

func resourceOpenzfsFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx OpenZFS File System (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	if d.HasChangesExcept("tags_all", "tags") {
		input := &fsx.UpdateFileSystemInput{
			ClientRequestToken:   aws.String(resource.UniqueId()),
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

		if d.HasChange("throughput_capacity") {
			input.OpenZFSConfiguration.ThroughputCapacity = aws.Int64(int64(d.Get("throughput_capacity").(int)))
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.OpenZFSConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		_, err := conn.UpdateFileSystem(input)

		if err != nil {
			return fmt.Errorf("error updating FSx OpenZFS File System (%s): %w", d.Id(), err)
		}

		if _, err := waitFileSystemUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for FSx OpenZFS File System (%s) update: %w", d.Id(), err)
		}

		if d.HasChange("root_volume_configuration") {
			input := &fsx.UpdateVolumeInput{
				ClientRequestToken:   aws.String(resource.UniqueId()),
				VolumeId:             aws.String(d.Get("root_volume_id").(string)),
				OpenZFSConfiguration: &fsx.UpdateOpenZFSVolumeConfiguration{},
			}

			input.OpenZFSConfiguration = expandOpenzfsUpdateRootVolumeConfiguration(d.Get("root_volume_configuration").([]interface{}))

			_, err := conn.UpdateVolume(input)

			if err != nil {
				return fmt.Errorf("error updating FSx OpenZFS Root Volume (%s): %w", d.Get("root_volume_id").(string), err)
			}

			if _, err := waitVolumeUpdated(conn, d.Get("root_volume_id").(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return fmt.Errorf("error waiting for FSx OpenZFS Root Volume (%s) update: %w", d.Get("root_volume_id").(string), err)
			}
		}

	}

	return resourceOpenzfsFileSystemRead(d, meta)
}

func resourceOpenzfsFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	log.Printf("[DEBUG] Deleting FSx OpenZFS File System: %s", d.Id())
	_, err := conn.DeleteFileSystem(&fsx.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FSx OpenZFS File System (%s): %w", d.Id(), err)
	}

	if _, err := waitFileSystemDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for FSx OpenZFS File System (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandOpenzfsFileDiskIopsConfiguration(cfg []interface{}) *fsx.DiskIopsConfiguration {
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

func expandOpenzfsRootVolumeConfiguration(cfg []interface{}) *fsx.OpenZFSCreateRootVolumeConfiguration {
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

	if v, ok := conf["user_and_group_quotas"]; ok {
		out.UserAndGroupQuotas = expandOpenzfsUserAndGroupQuotas(v.(*schema.Set).List())
	}

	if v, ok := conf["nfs_exports"].([]interface{}); ok {
		out.NfsExports = expandOpenzfsNFSExports(v)
	}

	return &out
}

func expandOpenzfsUpdateRootVolumeConfiguration(cfg []interface{}) *fsx.UpdateOpenZFSVolumeConfiguration {
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

	if v, ok := conf["user_and_group_quotas"]; ok {
		out.UserAndGroupQuotas = expandOpenzfsUserAndGroupQuotas(v.(*schema.Set).List())
	}

	if v, ok := conf["nfs_exports"].([]interface{}); ok {
		out.NfsExports = expandOpenzfsNFSExports(v)
	}

	return &out
}

func expandOpenzfsUserAndGroupQuotas(cfg []interface{}) []*fsx.OpenZFSUserOrGroupQuota {
	quotas := []*fsx.OpenZFSUserOrGroupQuota{}

	for _, quota := range cfg {
		expandedQuota := expandOpenzfsUserAndGroupQuota(quota.(map[string]interface{}))
		if expandedQuota != nil {
			quotas = append(quotas, expandedQuota)
		}
	}

	return quotas

}

func expandOpenzfsUserAndGroupQuota(conf map[string]interface{}) *fsx.OpenZFSUserOrGroupQuota {
	if len(conf) < 1 {
		return nil
	}

	out := fsx.OpenZFSUserOrGroupQuota{}

	if v, ok := conf["id"].(int); ok {
		out.Id = aws.Int64(int64(v))
	}

	if v, ok := conf["storage_capacity_quota_gib"].(int); ok {
		out.StorageCapacityQuotaGiB = aws.Int64(int64(v))
	}

	if v, ok := conf["type"].(string); ok {
		out.Type = aws.String(v)
	}

	return &out

}

func expandOpenzfsNFSExports(cfg []interface{}) []*fsx.OpenZFSNfsExport {
	exports := []*fsx.OpenZFSNfsExport{}

	for _, export := range cfg {
		expandedExport := expandOpenzfsNFSExport(export.(map[string]interface{}))
		if expandedExport != nil {
			exports = append(exports, expandedExport)
		}
	}

	return exports

}

func expandOpenzfsNFSExport(cfg map[string]interface{}) *fsx.OpenZFSNfsExport {
	out := fsx.OpenZFSNfsExport{}

	if v, ok := cfg["client_configurations"]; ok {
		out.ClientConfigurations = expandOpenzfsClinetConfigurations(v.(*schema.Set).List())
	}

	return &out
}

func expandOpenzfsClinetConfigurations(cfg []interface{}) []*fsx.OpenZFSClientConfiguration {
	configurations := []*fsx.OpenZFSClientConfiguration{}

	for _, configuration := range cfg {
		expandedConfiguration := expandOpenzfsClientConfiguration(configuration.(map[string]interface{}))
		if expandedConfiguration != nil {
			configurations = append(configurations, expandedConfiguration)
		}
	}

	return configurations

}

func expandOpenzfsClientConfiguration(conf map[string]interface{}) *fsx.OpenZFSClientConfiguration {
	out := fsx.OpenZFSClientConfiguration{}

	if v, ok := conf["clients"].(string); ok && len(v) > 0 {
		out.Clients = aws.String(v)
	}

	if v, ok := conf["options"].([]interface{}); ok {
		out.Options = flex.ExpandStringList(v)
	}

	return &out
}

func flattenOpenzfsFileDiskIopsConfiguration(rs *fsx.DiskIopsConfiguration) []interface{} {
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

func flattenOpenzfsRootVolumeConfiguration(rs *fsx.Volume) []interface{} {
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
		m["nfs_exports"] = flattenOpenzfsFileNFSExports(rs.OpenZFSConfiguration.NfsExports)
	}
	if rs.OpenZFSConfiguration.ReadOnly != nil {
		m["read_only"] = aws.BoolValue(rs.OpenZFSConfiguration.ReadOnly)
	}
	if rs.OpenZFSConfiguration.UserAndGroupQuotas != nil {
		m["user_and_group_quotas"] = flattenOpenzfsFileUserAndGroupQuotas(rs.OpenZFSConfiguration.UserAndGroupQuotas)
	}

	return []interface{}{m}
}

func flattenOpenzfsFileNFSExports(rs []*fsx.OpenZFSNfsExport) []map[string]interface{} {
	exports := make([]map[string]interface{}, 0)

	for _, export := range rs {
		if export != nil {
			cfg := make(map[string]interface{})
			cfg["client_configurations"] = flattenOpenzfsClientConfigurations(export.ClientConfigurations)
			exports = append(exports, cfg)
		}
	}

	if len(exports) > 0 {
		return exports
	}

	return nil
}

func flattenOpenzfsClientConfigurations(rs []*fsx.OpenZFSClientConfiguration) []map[string]interface{} {
	configurations := make([]map[string]interface{}, 0)

	for _, configuration := range rs {
		if configuration != nil {
			cfg := make(map[string]interface{})
			cfg["clients"] = aws.StringValue(configuration.Clients)
			cfg["options"] = flex.FlattenStringList(configuration.Options)
			configurations = append(configurations, cfg)
		}
	}

	if len(configurations) > 0 {
		return configurations
	}

	return nil
}

func flattenOpenzfsFileUserAndGroupQuotas(rs []*fsx.OpenZFSUserOrGroupQuota) []map[string]interface{} {
	quotas := make([]map[string]interface{}, 0)

	for _, quota := range rs {
		if quota != nil {
			cfg := make(map[string]interface{})
			cfg["id"] = aws.Int64Value(quota.Id)
			cfg["storage_capacity_quota_gib"] = aws.Int64Value(quota.StorageCapacityQuotaGiB)
			cfg["type"] = aws.StringValue(quota.Type)
			quotas = append(quotas, cfg)
		}
	}

	if len(quotas) > 0 {
		return quotas
	}

	return nil
}
