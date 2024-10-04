// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_lustre_file_system", name="Lustre File System")
// @Tags(identifierAttribute="arn")
func resourceLustreFileSystem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLustreFileSystemCreate,
		ReadWithoutTimeout:   resourceLustreFileSystemRead,
		UpdateWithoutTimeout: resourceLustreFileSystemUpdate,
		DeleteWithoutTimeout: resourceLustreFileSystemDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("skip_final_backup", true)

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_import_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AutoImportPolicyType](),
			},
			"automatic_backup_retention_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
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
				ForceNew: true,
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
			"data_compression_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DataCompressionType](),
				Default:          awstypes.DataCompressionTypeNone,
			},
			"deployment_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.LustreDeploymentTypeScratch1,
				ValidateDiagFunc: enum.Validate[awstypes.LustreDeploymentType](),
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"drive_cache_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DriveCacheType](),
			},
			"export_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 900),
					validation.StringMatch(regexache.MustCompile(`^s3://`), "must begin with s3://"),
				),
			},
			"file_system_type_version": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 20),
					validation.StringMatch(regexache.MustCompile(`^[0-9].[0-9]+$`), "must be in format x.y"),
				),
			},
			"final_backup_tags": tftags.TagsSchema(),
			"import_path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 900),
					validation.StringMatch(regexache.MustCompile(`^s3://`), "must begin with s3://"),
				),
			},
			"imported_file_chunk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 512000),
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"log_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDestination: {
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LustreAccessAuditLogLevel](),
						},
					},
				},
			},
			"metadata_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMode: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.MetadataConfigurationMode](),
						},
						names.AttrIOPS: {
							Type:             schema.TypeInt,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntInSlice([]int{1500, 3000, 6000, 12000, 24000, 36000, 48000, 60000, 72000, 84000, 96000, 108000, 120000, 132000, 144000, 156000, 168000, 180000, 192000})),
						},
					},
				},
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
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"per_unit_storage_throughput": {
				Type:     schema.TypeInt,
				Optional: true,
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
			"root_squash_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"no_squash_nids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([0-9\[\]\-]*\.){3}([0-9\[\]\-]*)@tcp$`), "must be in the standard Lustre NID foramt"),
							},
						},
						"root_squash": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([0-9]{1,10}):([0-9]{1,10})$`), "must be in the format UID:GID"),
						},
					},
				},
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
				Default:  true,
			},
			"storage_capacity": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1200),
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
				MaxItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			resourceLustreFileSystemStorageCapacityCustomizeDiff,
			resourceLustreFileSystemMetadataConfigCustomizeDiff,
		),
	}
}

func resourceLustreFileSystemStorageCapacityCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta any) error {
	// we want to force a new resource if the new storage capacity is less than the old one
	if d.HasChange("storage_capacity") {
		o, n := d.GetChange("storage_capacity")
		if n.(int) < o.(int) || d.Get("deployment_type").(string) == string(awstypes.LustreDeploymentTypeScratch1) {
			if err := d.ForceNew("storage_capacity"); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceLustreFileSystemMetadataConfigCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta any) error {
	//metadata_configuration is only supported when deployment_type is persistent2
	if v, ok := d.GetOk("metadata_configuration"); ok {
		if len(v.([]interface{})) > 0 {
			if deploymentType := awstypes.LustreDeploymentType(d.Get("deployment_type").(string)); deploymentType != awstypes.LustreDeploymentTypePersistent2 {
				return fmt.Errorf("metadata_configuration can only be set when deployment type is %s", awstypes.LustreDeploymentTypePersistent2)
			}
		}
	}

	// we want to force a new resource if the new Iops is less than the old one
	if d.HasChange("metadata_configuration") {
		if v, ok := d.GetOk("metadata_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			if mode := awstypes.MetadataConfigurationMode(d.Get("metadata_configuration.0.mode").(string)); mode == awstypes.MetadataConfigurationModeUserProvisioned {
				o, n := d.GetChange("metadata_configuration")

				oldV := o.([]interface{})
				newV := n.([]interface{})
				var metaOld map[string]interface{}
				var metaNew map[string]interface{}

				for _, v := range oldV {
					metaOld = v.(map[string]interface{})
				}

				for _, v := range newV {
					metaNew = v.(map[string]interface{})
				}

				if len(metaNew) > 0 && len(metaOld) > 0 {
					if metaNew[names.AttrIOPS].(int) < metaOld[names.AttrIOPS].(int) {
						log.Printf("[DEBUG] Forcing new due to metadata iops decrease. old iops: %d new iops: %d", metaOld[names.AttrIOPS].(int), metaNew[names.AttrIOPS].(int))
						if err := d.ForceNew("metadata_configuration.0.iops"); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func resourceLustreFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	inputC := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		FileSystemType:     awstypes.FileSystemTypeLustre,
		LustreConfiguration: &awstypes.CreateFileSystemLustreConfiguration{
			DeploymentType: awstypes.LustreDeploymentType(d.Get("deployment_type").(string)),
		},
		StorageCapacity: aws.Int32(int32(d.Get("storage_capacity").(int))),
		StorageType:     awstypes.StorageType(d.Get(names.AttrStorageType).(string)),
		SubnetIds:       flex.ExpandStringValueList(d.Get(names.AttrSubnetIDs).([]interface{})),
		Tags:            getTagsIn(ctx),
	}
	inputB := &fsx.CreateFileSystemFromBackupInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		LustreConfiguration: &awstypes.CreateFileSystemLustreConfiguration{
			DeploymentType: awstypes.LustreDeploymentType(d.Get("deployment_type").(string)),
		},
		StorageType: awstypes.StorageType(d.Get(names.AttrStorageType).(string)),
		SubnetIds:   flex.ExpandStringValueList(d.Get(names.AttrSubnetIDs).([]interface{})),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("auto_import_policy"); ok {
		inputC.LustreConfiguration.AutoImportPolicy = awstypes.AutoImportPolicyType(v.(string))
		inputB.LustreConfiguration.AutoImportPolicy = awstypes.AutoImportPolicyType(v.(string))
	}

	if v, ok := d.GetOk("automatic_backup_retention_days"); ok {
		inputC.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int32(int32(v.(int)))
		inputB.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("copy_tags_to_backups"); ok {
		inputC.LustreConfiguration.CopyTagsToBackups = aws.Bool(v.(bool))
		inputB.LustreConfiguration.CopyTagsToBackups = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("daily_automatic_backup_start_time"); ok {
		inputC.LustreConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
		inputB.LustreConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_compression_type"); ok {
		inputC.LustreConfiguration.DataCompressionType = awstypes.DataCompressionType(v.(string))
		inputB.LustreConfiguration.DataCompressionType = awstypes.DataCompressionType(v.(string))
	}

	if v, ok := d.GetOk("drive_cache_type"); ok {
		inputC.LustreConfiguration.DriveCacheType = awstypes.DriveCacheType(v.(string))
		inputB.LustreConfiguration.DriveCacheType = awstypes.DriveCacheType(v.(string))
	}

	if v, ok := d.GetOk("export_path"); ok {
		inputC.LustreConfiguration.ExportPath = aws.String(v.(string))
		inputB.LustreConfiguration.ExportPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("file_system_type_version"); ok {
		inputC.FileSystemTypeVersion = aws.String(v.(string))
		inputB.FileSystemTypeVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("import_path"); ok {
		inputC.LustreConfiguration.ImportPath = aws.String(v.(string))
		inputB.LustreConfiguration.ImportPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("imported_file_chunk_size"); ok {
		inputC.LustreConfiguration.ImportedFileChunkSize = aws.Int32(int32(v.(int)))
		inputB.LustreConfiguration.ImportedFileChunkSize = aws.Int32(int32(v.(int)))
	}

	// Applicable only for TypePersistent1 and TypePersistent2.
	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		inputC.KmsKeyId = aws.String(v.(string))
		inputB.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_configuration"); ok && len(v.([]interface{})) > 0 {
		inputC.LustreConfiguration.LogConfiguration = expandLustreLogCreateConfiguration(v.([]interface{}))
		inputB.LustreConfiguration.LogConfiguration = expandLustreLogCreateConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("metadata_configuration"); ok && len(v.([]interface{})) > 0 {
		inputC.LustreConfiguration.MetadataConfiguration = expandLustreMetadataCreateConfiguration(v.([]interface{}))
		inputB.LustreConfiguration.MetadataConfiguration = expandLustreMetadataCreateConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("per_unit_storage_throughput"); ok {
		inputC.LustreConfiguration.PerUnitStorageThroughput = aws.Int32(int32(v.(int)))
		inputB.LustreConfiguration.PerUnitStorageThroughput = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("root_squash_configuration"); ok && len(v.([]interface{})) > 0 {
		inputC.LustreConfiguration.RootSquashConfiguration = expandLustreRootSquashConfiguration(v.([]interface{}))
		inputB.LustreConfiguration.RootSquashConfiguration = expandLustreRootSquashConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok {
		inputC.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
		inputB.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		inputC.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
		inputB.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("backup_id"); ok {
		backupID := v.(string)
		inputB.BackupId = aws.String(backupID)

		output, err := conn.CreateFileSystemFromBackup(ctx, inputB)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating FSx for Lustre File System from backup (%s): %s", backupID, err)
		}

		d.SetId(aws.ToString(output.FileSystem.FileSystemId))
	} else {
		output, err := conn.CreateFileSystem(ctx, inputC)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating FSx for Lustre File System: %s", err)
		}

		d.SetId(aws.ToString(output.FileSystem.FileSystemId))
	}

	if _, err := waitFileSystemCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre File System (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLustreFileSystemRead(ctx, d, meta)...)
}

func resourceLustreFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	filesystem, err := findLustreFileSystemByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx for Lustre File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for Lustre File System (%s): %s", d.Id(), err)
	}

	lustreConfig := filesystem.LustreConfiguration
	if lustreConfig.DataRepositoryConfiguration == nil {
		// Initialize an empty structure to simplify d.Set() handling.
		lustreConfig.DataRepositoryConfiguration = &awstypes.DataRepositoryConfiguration{}
	}

	d.Set(names.AttrARN, filesystem.ResourceARN)
	d.Set("auto_import_policy", lustreConfig.DataRepositoryConfiguration.AutoImportPolicy)
	d.Set("automatic_backup_retention_days", lustreConfig.AutomaticBackupRetentionDays)
	d.Set("copy_tags_to_backups", lustreConfig.CopyTagsToBackups)
	d.Set("daily_automatic_backup_start_time", lustreConfig.DailyAutomaticBackupStartTime)
	d.Set("data_compression_type", lustreConfig.DataCompressionType)
	d.Set("deployment_type", lustreConfig.DeploymentType)
	d.Set(names.AttrDNSName, filesystem.DNSName)
	d.Set("drive_cache_type", lustreConfig.DriveCacheType)
	d.Set("export_path", lustreConfig.DataRepositoryConfiguration.ExportPath)
	d.Set("file_system_type_version", filesystem.FileSystemTypeVersion)
	d.Set("import_path", lustreConfig.DataRepositoryConfiguration.ImportPath)
	d.Set("imported_file_chunk_size", lustreConfig.DataRepositoryConfiguration.ImportedFileChunkSize)
	d.Set(names.AttrKMSKeyID, filesystem.KmsKeyId)
	if err := d.Set("log_configuration", flattenLustreLogConfiguration(lustreConfig.LogConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_configuration: %s", err)
	}
	if err := d.Set("metadata_configuration", flattenLustreMetadataConfiguration(lustreConfig.MetadataConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting metadata_configuration: %s", err)
	}
	d.Set("mount_name", lustreConfig.MountName)
	d.Set("network_interface_ids", filesystem.NetworkInterfaceIds)
	d.Set(names.AttrOwnerID, filesystem.OwnerId)
	d.Set("per_unit_storage_throughput", lustreConfig.PerUnitStorageThroughput)
	if err := d.Set("root_squash_configuration", flattenLustreRootSquashConfiguration(lustreConfig.RootSquashConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting root_squash_configuration: %s", err)
	}
	d.Set("storage_capacity", filesystem.StorageCapacity)
	d.Set(names.AttrStorageType, filesystem.StorageType)
	d.Set(names.AttrSubnetIDs, filesystem.SubnetIds)
	d.Set(names.AttrVPCID, filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", lustreConfig.WeeklyMaintenanceStartTime)

	setTagsOut(ctx, filesystem.Tags)

	return diags
}

func resourceLustreFileSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	if d.HasChangesExcept(
		"final_backup_tags",
		"skip_final_backup",
		names.AttrTags,
		names.AttrTagsAll,
	) {
		input := &fsx.UpdateFileSystemInput{
			ClientRequestToken:  aws.String(id.UniqueId()),
			FileSystemId:        aws.String(d.Id()),
			LustreConfiguration: &awstypes.UpdateFileSystemLustreConfiguration{},
		}

		if d.HasChange("auto_import_policy") {
			input.LustreConfiguration.AutoImportPolicy = awstypes.AutoImportPolicyType(d.Get("auto_import_policy").(string))
		}

		if d.HasChange("automatic_backup_retention_days") {
			input.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int32(int32(d.Get("automatic_backup_retention_days").(int)))
		}

		if d.HasChange("daily_automatic_backup_start_time") {
			input.LustreConfiguration.DailyAutomaticBackupStartTime = aws.String(d.Get("daily_automatic_backup_start_time").(string))
		}

		if d.HasChange("data_compression_type") {
			input.LustreConfiguration.DataCompressionType = awstypes.DataCompressionType(d.Get("data_compression_type").(string))
		}

		if d.HasChange("log_configuration") {
			input.LustreConfiguration.LogConfiguration = expandLustreLogCreateConfiguration(d.Get("log_configuration").([]interface{}))
		}

		if d.HasChange("metadata_configuration") {
			input.LustreConfiguration.MetadataConfiguration = expandLustreMetadataUpdateConfiguration(d.Get("metadata_configuration").([]interface{}))
		}

		if d.HasChange("per_unit_storage_throughput") {
			input.LustreConfiguration.PerUnitStorageThroughput = aws.Int32(int32(d.Get("per_unit_storage_throughput").(int)))
		}

		if d.HasChange("root_squash_configuration") {
			input.LustreConfiguration.RootSquashConfiguration = expandLustreRootSquashConfiguration(d.Get("root_squash_configuration").([]interface{}))
		}

		if d.HasChange("storage_capacity") {
			input.StorageCapacity = aws.Int32(int32(d.Get("storage_capacity").(int)))
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		startTime := time.Now()
		_, err := conn.UpdateFileSystem(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSX for Lustre File System (%s): %s", d.Id(), err)
		}

		if _, err := waitFileSystemUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre File System (%s) update: %s", d.Id(), err)
		}

		if _, err := waitFileSystemAdministrativeActionCompleted(ctx, conn, d.Id(), awstypes.AdministrativeActionTypeFileSystemUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre File System (%s) administrative action (%s) complete: %s", d.Id(), awstypes.AdministrativeActionTypeFileSystemUpdate, err)
		}
	}

	return append(diags, resourceLustreFileSystemRead(ctx, d, meta)...)
}

func resourceLustreFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := &fsx.DeleteFileSystemInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		FileSystemId:       aws.String(d.Id()),
	}

	// Final backup during delete is not supported on file systems using the Scratch deployment type
	// LustreConfiguration cannot be supplied at all, even when empty, in this scenario
	if v, ok := d.GetOk("deployment_type"); ok && !strings.HasPrefix(v.(string), "SCRATCH_") {
		lustreConfig := &awstypes.DeleteFileSystemLustreConfiguration{
			SkipFinalBackup: aws.Bool(d.Get("skip_final_backup").(bool)),
		}

		if v, ok := d.GetOk("final_backup_tags"); ok && len(v.(map[string]interface{})) > 0 {
			lustreConfig.FinalBackupTags = Tags(tftags.New(ctx, v))
		}

		input.LustreConfiguration = lustreConfig
	}

	log.Printf("[DEBUG] Deleting FSx for Lustre File System: %s", d.Id())
	_, err := conn.DeleteFileSystem(ctx, input)

	if errs.IsA[*awstypes.FileSystemNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx for Lustre File System (%s): %s", d.Id(), err)
	}

	if _, err := waitFileSystemDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre File System (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findLustreFileSystemByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.FileSystem, error) {
	output, err := findFileSystemByIDAndType(ctx, conn, id, awstypes.FileSystemTypeLustre)

	if err != nil {
		return nil, err
	}

	if output.LustreConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}

func findFileSystemByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.FileSystem, error) {
	input := &fsx.DescribeFileSystemsInput{
		FileSystemIds: []string{id},
	}

	return findFileSystem(ctx, conn, input, tfslices.PredicateTrue[*awstypes.FileSystem]())
}

func findFileSystemByIDAndType(ctx context.Context, conn *fsx.Client, fsID string, fsType awstypes.FileSystemType) (*awstypes.FileSystem, error) {
	input := &fsx.DescribeFileSystemsInput{
		FileSystemIds: []string{fsID},
	}
	filter := func(v *awstypes.FileSystem) bool {
		return v.FileSystemType == fsType
	}

	return findFileSystem(ctx, conn, input, filter)
}

func findFileSystem(ctx context.Context, conn *fsx.Client, input *fsx.DescribeFileSystemsInput, filter tfslices.Predicate[*awstypes.FileSystem]) (*awstypes.FileSystem, error) {
	output, err := findFileSystems(ctx, conn, input, filter)

	if err != nil {
		return &awstypes.FileSystem{}, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFileSystems(ctx context.Context, conn *fsx.Client, input *fsx.DescribeFileSystemsInput, filter tfslices.Predicate[*awstypes.FileSystem]) ([]awstypes.FileSystem, error) {
	var output []awstypes.FileSystem

	pages := fsx.NewDescribeFileSystemsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.FileSystemNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.FileSystems {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusFileSystem(ctx context.Context, conn *fsx.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFileSystemByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Lifecycle), nil
	}
}

func waitFileSystemCreated(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.FileSystem, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FileSystemLifecycleCreating),
		Target:  enum.Slice(awstypes.FileSystemLifecycleAvailable),
		Refresh: statusFileSystem(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,

		// When the filesystem is used by another service, e.g. an M2 Environment, it is not immediately available
		ContinuousTargetOccurence: 3,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FileSystem); ok {
		if status, details := output.Lifecycle, output.FailureDetails; status == awstypes.FileSystemLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(details.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitFileSystemUpdated(ctx context.Context, conn *fsx.Client, id string, startTime time.Time, timeout time.Duration) (*awstypes.FileSystem, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FileSystemLifecycleUpdating),
		Target:  enum.Slice(awstypes.FileSystemLifecycleAvailable),
		Refresh: statusFileSystem(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FileSystem); ok {
		switch status := output.Lifecycle; status {
		case awstypes.FileSystemLifecycleFailed, awstypes.FileSystemLifecycleMisconfigured, awstypes.FileSystemLifecycleMisconfiguredUnavailable:
			// Report any failed non-FILE_SYSTEM_UPDATE administrative actions.
			// See https://docs.aws.amazon.com/fsx/latest/APIReference/API_AdministrativeAction.html#FSx-Type-AdministrativeAction-AdministrativeActionType.
			administrativeActions := tfslices.Filter(output.AdministrativeActions, func(v awstypes.AdministrativeAction) bool {
				return v.Status == awstypes.StatusFailed && v.AdministrativeActionType != awstypes.AdministrativeActionTypeFileSystemUpdate && v.FailureDetails != nil && startTime.Before(aws.ToTime(v.RequestTime))
			})
			administrativeActionsError := errors.Join(tfslices.ApplyToAll(administrativeActions, func(v awstypes.AdministrativeAction) error {
				return fmt.Errorf("%s: %s", string(v.AdministrativeActionType), aws.ToString(v.FailureDetails.Message))
			})...)

			if details := output.FailureDetails; details != nil {
				if message := aws.ToString(details.Message); administrativeActionsError != nil {
					tfresource.SetLastError(err, fmt.Errorf("%s: %w", message, administrativeActionsError))
				} else {
					tfresource.SetLastError(err, errors.New(message))
				}
			} else {
				tfresource.SetLastError(err, administrativeActionsError)
			}
		}

		return output, err
	}

	return nil, err
}

func waitFileSystemDeleted(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.FileSystem, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FileSystemLifecycleAvailable, awstypes.FileSystemLifecycleDeleting),
		Target:  []string{},
		Refresh: statusFileSystem(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FileSystem); ok {
		if status, details := output.Lifecycle, output.FailureDetails; status == awstypes.FileSystemLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(details.Message)))
		}

		return output, err
	}

	return nil, err
}

func findFileSystemAdministrativeAction(ctx context.Context, conn *fsx.Client, fsID string, actionType awstypes.AdministrativeActionType) (awstypes.AdministrativeAction, error) {
	output, err := findFileSystemByID(ctx, conn, fsID)

	if err != nil {
		return awstypes.AdministrativeAction{}, err
	}

	for _, v := range output.AdministrativeActions {
		if v.AdministrativeActionType == actionType {
			return v, nil
		}
	}

	// If the administrative action isn't found, assume it's complete.
	return awstypes.AdministrativeAction{Status: awstypes.StatusCompleted}, nil
}

func statusFileSystemAdministrativeAction(ctx context.Context, conn *fsx.Client, fsID string, actionType awstypes.AdministrativeActionType) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFileSystemAdministrativeAction(ctx, conn, fsID, actionType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitFileSystemAdministrativeActionCompleted(ctx context.Context, conn *fsx.Client, fsID string, actionType awstypes.AdministrativeActionType, timeout time.Duration) (*awstypes.AdministrativeAction, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusInProgress, awstypes.StatusPending),
		Target:  enum.Slice(awstypes.StatusCompleted, awstypes.StatusUpdatedOptimizing),
		Refresh: statusFileSystemAdministrativeAction(ctx, conn, fsID, actionType),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AdministrativeAction); ok {
		if status, details := output.Status, output.FailureDetails; status == awstypes.StatusFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func expandLustreRootSquashConfiguration(l []interface{}) *awstypes.LustreRootSquashConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &awstypes.LustreRootSquashConfiguration{}

	if v, ok := data["root_squash"].(string); ok && v != "" {
		req.RootSquash = aws.String(v)
	}

	if v, ok := data["no_squash_nids"].(*schema.Set); ok && v.Len() > 0 {
		req.NoSquashNids = flex.ExpandStringValueSet(v)
	}

	return req
}

func flattenLustreRootSquashConfiguration(adopts *awstypes.LustreRootSquashConfiguration) []map[string]interface{} {
	if adopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if adopts.RootSquash != nil {
		m["root_squash"] = aws.ToString(adopts.RootSquash)
	}

	if adopts.NoSquashNids != nil {
		m["no_squash_nids"] = flex.FlattenStringValueSet(adopts.NoSquashNids)
	}

	return []map[string]interface{}{m}
}

func expandLustreLogCreateConfiguration(l []interface{}) *awstypes.LustreLogCreateConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &awstypes.LustreLogCreateConfiguration{
		Level: awstypes.LustreAccessAuditLogLevel(data["level"].(string)),
	}

	if v, ok := data[names.AttrDestination].(string); ok && v != "" {
		req.Destination = aws.String(logStateFunc(v))
	}

	return req
}

func flattenLustreLogConfiguration(adopts *awstypes.LustreLogConfiguration) []map[string]interface{} {
	if adopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"level": string(adopts.Level),
	}

	if adopts.Destination != nil {
		m[names.AttrDestination] = aws.ToString(adopts.Destination)
	}

	return []map[string]interface{}{m}
}

func expandLustreMetadataCreateConfiguration(l []interface{}) *awstypes.CreateFileSystemLustreMetadataConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &awstypes.CreateFileSystemLustreMetadataConfiguration{
		Mode: awstypes.MetadataConfigurationMode(data[names.AttrMode].(string)),
	}

	if v, ok := data[names.AttrIOPS].(int); ok && v != 0 {
		req.Iops = aws.Int32(int32(v))
	}

	return req
}

func expandLustreMetadataUpdateConfiguration(l []interface{}) *awstypes.UpdateFileSystemLustreMetadataConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &awstypes.UpdateFileSystemLustreMetadataConfiguration{
		Mode: awstypes.MetadataConfigurationMode(data[names.AttrMode].(string)),
	}

	if v, ok := data[names.AttrIOPS].(int); ok && v != 0 {
		req.Iops = aws.Int32(int32(v))
	}

	return req
}

func flattenLustreMetadataConfiguration(adopts *awstypes.FileSystemLustreMetadataConfiguration) []map[string]interface{} {
	if adopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrMode: string(adopts.Mode),
	}

	if adopts.Iops != nil {
		m[names.AttrIOPS] = aws.ToInt32(adopts.Iops)
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
