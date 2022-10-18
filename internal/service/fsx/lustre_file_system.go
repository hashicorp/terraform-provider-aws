package fsx

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLustreFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceLustreFileSystemCreate,
		Read:   resourceLustreFileSystemRead,
		Update: resourceLustreFileSystemUpdate,
		Delete: resourceLustreFileSystemDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backup_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"export_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 900),
					validation.StringMatch(regexp.MustCompile(`^s3://`), "must begin with s3://"),
				),
			},
			"import_path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 900),
					validation.StringMatch(regexp.MustCompile(`^s3://`), "must begin with s3://"),
				),
			},
			"imported_file_chunk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 512000),
			},
			"mount_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_ids": {
				// As explained in https://docs.aws.amazon.com/fsx/latest/LustreGuide/mounting-on-premises.html, the first
				// network_interface_id is the primary one, so ordering matters. Use TypeList instead of TypeSet to preserve it.
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owner_id": {
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
				ValidateFunc: validation.IntAtLeast(1200),
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
			"deployment_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.LustreDeploymentTypeScratch1,
				ValidateFunc: validation.StringInSlice(fsx.LustreDeploymentType_Values(), false),
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"per_unit_storage_throughput": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.IntInSlice([]int{
					12,
					40,
					50,
					100,
					125,
					200,
					250,
					500,
					1000,
				}),
			},
			"automatic_backup_retention_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 90),
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
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.StorageTypeSsd,
				ValidateFunc: validation.StringInSlice(fsx.StorageType_Values(), false),
			},
			"drive_cache_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(fsx.DriveCacheType_Values(), false),
			},
			"auto_import_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(fsx.AutoImportPolicyType_Values(), false),
			},
			"copy_tags_to_backups": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"data_compression_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(fsx.DataCompressionType_Values(), false),
				Default:      fsx.DataCompressionTypeNone,
			},
			"file_system_type_version": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 20),
					validation.StringMatch(regexp.MustCompile(`^[0-9].[0-9]+$`), "must be in format x.y"),
				),
			},
			"log_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidARN,
							StateFunc:    logStateFunc,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return strings.HasPrefix(old, fmt.Sprintf("%s:", new))
							},
						},
						"level": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(fsx.LustreAccessAuditLogLevel_Values(), false),
						},
					},
				},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			resourceLustreFileSystemSchemaCustomizeDiff,
		),
	}
}

func resourceLustreFileSystemSchemaCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// we want to force a new resource if the new storage capacity is less than the old one
	if d.HasChange("storage_capacity") {
		o, n := d.GetChange("storage_capacity")
		if n.(int) < o.(int) || d.Get("deployment_type").(string) == fsx.LustreDeploymentTypeScratch1 {
			if err := d.ForceNew("storage_capacity"); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceLustreFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemType:     aws.String(fsx.FileSystemTypeLustre),
		StorageCapacity:    aws.Int64(int64(d.Get("storage_capacity").(int))),
		StorageType:        aws.String(d.Get("storage_type").(string)),
		SubnetIds:          flex.ExpandStringList(d.Get("subnet_ids").([]interface{})),
		LustreConfiguration: &fsx.CreateFileSystemLustreConfiguration{
			DeploymentType: aws.String(d.Get("deployment_type").(string)),
		},
	}

	backupInput := &fsx.CreateFileSystemFromBackupInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		StorageType:        aws.String(d.Get("storage_type").(string)),
		SubnetIds:          flex.ExpandStringList(d.Get("subnet_ids").([]interface{})),
		LustreConfiguration: &fsx.CreateFileSystemLustreConfiguration{
			DeploymentType: aws.String(d.Get("deployment_type").(string)),
		},
	}

	//Applicable only for TypePersistent1 and TypePersistent2
	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
		backupInput.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("automatic_backup_retention_days"); ok {
		input.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(v.(int)))
		backupInput.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("daily_automatic_backup_start_time"); ok {
		input.LustreConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
		backupInput.LustreConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("export_path"); ok {
		input.LustreConfiguration.ExportPath = aws.String(v.(string))
		backupInput.LustreConfiguration.ExportPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("import_path"); ok {
		input.LustreConfiguration.ImportPath = aws.String(v.(string))
		backupInput.LustreConfiguration.ImportPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("imported_file_chunk_size"); ok {
		input.LustreConfiguration.ImportedFileChunkSize = aws.Int64(int64(v.(int)))
		backupInput.LustreConfiguration.ImportedFileChunkSize = aws.Int64(int64(v.(int)))
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
		input.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
		backupInput.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("per_unit_storage_throughput"); ok {
		input.LustreConfiguration.PerUnitStorageThroughput = aws.Int64(int64(v.(int)))
		backupInput.LustreConfiguration.PerUnitStorageThroughput = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("drive_cache_type"); ok {
		input.LustreConfiguration.DriveCacheType = aws.String(v.(string))
		backupInput.LustreConfiguration.DriveCacheType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("auto_import_policy"); ok {
		input.LustreConfiguration.AutoImportPolicy = aws.String(v.(string))
		backupInput.LustreConfiguration.AutoImportPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("copy_tags_to_backups"); ok {
		input.LustreConfiguration.CopyTagsToBackups = aws.Bool(v.(bool))
		backupInput.LustreConfiguration.CopyTagsToBackups = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("data_compression_type"); ok {
		input.LustreConfiguration.DataCompressionType = aws.String(v.(string))
		backupInput.LustreConfiguration.DataCompressionType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("file_system_type_version"); ok {
		input.FileSystemTypeVersion = aws.String(v.(string))
		backupInput.FileSystemTypeVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_configuration"); ok && len(v.([]interface{})) > 0 {
		input.LustreConfiguration.LogConfiguration = expandLustreLogCreateConfiguration(v.([]interface{}))
		backupInput.LustreConfiguration.LogConfiguration = expandLustreLogCreateConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("backup_id"); ok {
		backupInput.BackupId = aws.String(v.(string))

		log.Printf("[DEBUG] Creating FSx Lustre File System: %s", backupInput)
		result, err := conn.CreateFileSystemFromBackup(backupInput)

		if err != nil {
			return fmt.Errorf("error creating FSx Lustre File System from backup: %w", err)
		}

		d.SetId(aws.StringValue(result.FileSystem.FileSystemId))
	} else {
		log.Printf("[DEBUG] Creating FSx Lustre File System: %s", input)
		result, err := conn.CreateFileSystem(input)

		if err != nil {
			return fmt.Errorf("error creating FSx Lustre File System: %w", err)
		}

		d.SetId(aws.StringValue(result.FileSystem.FileSystemId))
	}

	if _, err := waitFileSystemCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for FSx Lustre File System (%s) create: %w", d.Id(), err)
	}

	return resourceLustreFileSystemRead(d, meta)
}

func resourceLustreFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Lustre File System (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	if d.HasChangesExcept("tags_all", "tags") {
		var waitAdminAction = false
		input := &fsx.UpdateFileSystemInput{
			ClientRequestToken:  aws.String(resource.UniqueId()),
			FileSystemId:        aws.String(d.Id()),
			LustreConfiguration: &fsx.UpdateFileSystemLustreConfiguration{},
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		if d.HasChange("automatic_backup_retention_days") {
			input.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(d.Get("automatic_backup_retention_days").(int)))
		}

		if d.HasChange("daily_automatic_backup_start_time") {
			input.LustreConfiguration.DailyAutomaticBackupStartTime = aws.String(d.Get("daily_automatic_backup_start_time").(string))
		}

		if d.HasChange("auto_import_policy") {
			input.LustreConfiguration.AutoImportPolicy = aws.String(d.Get("auto_import_policy").(string))
		}

		if d.HasChange("storage_capacity") {
			input.StorageCapacity = aws.Int64(int64(d.Get("storage_capacity").(int)))
		}

		if v, ok := d.GetOk("data_compression_type"); ok {
			input.LustreConfiguration.DataCompressionType = aws.String(v.(string))
		}

		if d.HasChange("log_configuration") {
			input.LustreConfiguration.LogConfiguration = expandLustreLogCreateConfiguration(d.Get("log_configuration").([]interface{}))
			waitAdminAction = true
		}

		_, err := conn.UpdateFileSystem(input)
		if err != nil {
			return fmt.Errorf("error updating FSX Lustre File System (%s): %w", d.Id(), err)
		}

		if _, err := waitFileSystemUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for FSx Lustre File System (%s) update: %w", d.Id(), err)
		}

		if waitAdminAction {
			if _, err := waitAdministrativeActionCompleted(conn, d.Id(), fsx.AdministrativeActionTypeFileSystemUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return fmt.Errorf("error waiting for FSx Lustre File System (%s) Log Configuratio to be updated: %w", d.Id(), err)
			}
		}
	}

	return resourceLustreFileSystemRead(d, meta)
}

func resourceLustreFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	filesystem, err := FindFileSystemByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Lustre File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FSx Lustre File System (%s): %w", d.Id(), err)
	}

	lustreConfig := filesystem.LustreConfiguration

	if filesystem.WindowsConfiguration != nil {
		return fmt.Errorf("expected FSx Lustre File System, found FSx Windows File System: %s", d.Id())
	}

	if lustreConfig == nil {
		return fmt.Errorf("error describing FSx Lustre File System (%s): empty Lustre configuration", d.Id())
	}

	if lustreConfig.DataRepositoryConfiguration == nil {
		// Initialize an empty structure to simplify d.Set() handling
		lustreConfig.DataRepositoryConfiguration = &fsx.DataRepositoryConfiguration{}
	}

	d.Set("arn", filesystem.ResourceARN)
	d.Set("dns_name", filesystem.DNSName)
	d.Set("export_path", lustreConfig.DataRepositoryConfiguration.ExportPath)
	d.Set("import_path", lustreConfig.DataRepositoryConfiguration.ImportPath)
	d.Set("auto_import_policy", lustreConfig.DataRepositoryConfiguration.AutoImportPolicy)
	d.Set("imported_file_chunk_size", lustreConfig.DataRepositoryConfiguration.ImportedFileChunkSize)
	d.Set("deployment_type", lustreConfig.DeploymentType)
	if lustreConfig.PerUnitStorageThroughput != nil {
		d.Set("per_unit_storage_throughput", lustreConfig.PerUnitStorageThroughput)
	}
	d.Set("mount_name", lustreConfig.MountName)
	d.Set("storage_type", filesystem.StorageType)
	if lustreConfig.DriveCacheType != nil {
		d.Set("drive_cache_type", lustreConfig.DriveCacheType)
	}

	if filesystem.KmsKeyId != nil {
		d.Set("kms_key_id", filesystem.KmsKeyId)
	}

	if err := d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds)); err != nil {
		return fmt.Errorf("error setting network_interface_ids: %w", err)
	}

	d.Set("owner_id", filesystem.OwnerId)
	d.Set("storage_capacity", filesystem.StorageCapacity)

	if err := d.Set("subnet_ids", aws.StringValueSlice(filesystem.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %w", err)
	}

	if err := d.Set("log_configuration", flattenLustreLogConfiguration(lustreConfig.LogConfiguration)); err != nil {
		return fmt.Errorf("error setting log_configuration: %w", err)
	}

	tags := KeyValueTags(filesystem.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("vpc_id", filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", lustreConfig.WeeklyMaintenanceStartTime)
	d.Set("automatic_backup_retention_days", lustreConfig.AutomaticBackupRetentionDays)
	d.Set("daily_automatic_backup_start_time", lustreConfig.DailyAutomaticBackupStartTime)
	d.Set("copy_tags_to_backups", lustreConfig.CopyTagsToBackups)
	d.Set("data_compression_type", lustreConfig.DataCompressionType)
	d.Set("file_system_type_version", filesystem.FileSystemTypeVersion)

	return nil
}

func resourceLustreFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	request := &fsx.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting FSx Lustre File System: %s", d.Id())
	_, err := conn.DeleteFileSystem(request)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FSx Lustre File System (%s): %w", d.Id(), err)
	}

	if _, err := waitFileSystemDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for FSx Lustre File System (%s) to deleted: %w", d.Id(), err)
	}

	return nil
}

func expandLustreLogCreateConfiguration(l []interface{}) *fsx.LustreLogCreateConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.LustreLogCreateConfiguration{
		Level: aws.String(data["level"].(string)),
	}

	if v, ok := data["destination"].(string); ok && v != "" {
		req.Destination = aws.String(logStateFunc(v))
	}

	return req
}

func flattenLustreLogConfiguration(adopts *fsx.LustreLogConfiguration) []map[string]interface{} {
	if adopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"level": aws.StringValue(adopts.Level),
	}

	if adopts.Destination != nil {
		m["destination"] = aws.StringValue(adopts.Destination)
	}

	return []map[string]interface{}{m}
}

func logStateFunc(v interface{}) string {
	value := v.(string)
	// API returns the specific log stream arn instead of provided log group
	logArn, _ := arn.Parse(value)
	if logArn.Service == "logs" {
		parts := strings.SplitN(logArn.Resource, ":", 3)
		if len(parts) == 3 {
			return strings.TrimSuffix(value, fmt.Sprintf(":%s", parts[2]))
		} else {
			return value
		}
	}
	return value
}
