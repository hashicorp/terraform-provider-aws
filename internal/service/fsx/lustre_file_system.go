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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			StateContext: schema.ImportStatePassthroughContext,
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(fsx.AutoImportPolicyType_Values(), false),
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(fsx.DataCompressionType_Values(), false),
				Default:      fsx.DataCompressionTypeNone,
			},
			"deployment_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.LustreDeploymentTypeScratch1,
				ValidateFunc: validation.StringInSlice(fsx.LustreDeploymentType_Values(), false),
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"drive_cache_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(fsx.DriveCacheType_Values(), false),
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

func resourceLustreFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	inputC := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		FileSystemType:     aws.String(fsx.FileSystemTypeLustre),
		LustreConfiguration: &fsx.CreateFileSystemLustreConfiguration{
			DeploymentType: aws.String(d.Get("deployment_type").(string)),
		},
		StorageCapacity: aws.Int64(int64(d.Get("storage_capacity").(int))),
		StorageType:     aws.String(d.Get("storage_type").(string)),
		SubnetIds:       flex.ExpandStringList(d.Get(names.AttrSubnetIDs).([]interface{})),
		Tags:            getTagsIn(ctx),
	}
	inputB := &fsx.CreateFileSystemFromBackupInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		LustreConfiguration: &fsx.CreateFileSystemLustreConfiguration{
			DeploymentType: aws.String(d.Get("deployment_type").(string)),
		},
		StorageType: aws.String(d.Get("storage_type").(string)),
		SubnetIds:   flex.ExpandStringList(d.Get(names.AttrSubnetIDs).([]interface{})),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("auto_import_policy"); ok {
		inputC.LustreConfiguration.AutoImportPolicy = aws.String(v.(string))
		inputB.LustreConfiguration.AutoImportPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("automatic_backup_retention_days"); ok {
		inputC.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(v.(int)))
		inputB.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(v.(int)))
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
		inputC.LustreConfiguration.DataCompressionType = aws.String(v.(string))
		inputB.LustreConfiguration.DataCompressionType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("drive_cache_type"); ok {
		inputC.LustreConfiguration.DriveCacheType = aws.String(v.(string))
		inputB.LustreConfiguration.DriveCacheType = aws.String(v.(string))
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
		inputC.LustreConfiguration.ImportedFileChunkSize = aws.Int64(int64(v.(int)))
		inputB.LustreConfiguration.ImportedFileChunkSize = aws.Int64(int64(v.(int)))
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

	if v, ok := d.GetOk("per_unit_storage_throughput"); ok {
		inputC.LustreConfiguration.PerUnitStorageThroughput = aws.Int64(int64(v.(int)))
		inputB.LustreConfiguration.PerUnitStorageThroughput = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("root_squash_configuration"); ok && len(v.([]interface{})) > 0 {
		inputC.LustreConfiguration.RootSquashConfiguration = expandLustreRootSquashConfiguration(v.([]interface{}))
		inputB.LustreConfiguration.RootSquashConfiguration = expandLustreRootSquashConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		inputC.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		inputB.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		inputC.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
		inputB.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("backup_id"); ok {
		backupID := v.(string)
		inputB.BackupId = aws.String(backupID)

		output, err := conn.CreateFileSystemFromBackupWithContext(ctx, inputB)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating FSx for Lustre File System from backup (%s): %s", backupID, err)
		}

		d.SetId(aws.StringValue(output.FileSystem.FileSystemId))
	} else {
		output, err := conn.CreateFileSystemWithContext(ctx, inputC)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating FSx for Lustre File System: %s", err)
		}

		d.SetId(aws.StringValue(output.FileSystem.FileSystemId))
	}

	if _, err := waitFileSystemCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre File System (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLustreFileSystemRead(ctx, d, meta)...)
}

func resourceLustreFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

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
		lustreConfig.DataRepositoryConfiguration = &fsx.DataRepositoryConfiguration{}
	}

	d.Set(names.AttrARN, filesystem.ResourceARN)
	d.Set("auto_import_policy", lustreConfig.DataRepositoryConfiguration.AutoImportPolicy)
	d.Set("automatic_backup_retention_days", lustreConfig.AutomaticBackupRetentionDays)
	d.Set("copy_tags_to_backups", lustreConfig.CopyTagsToBackups)
	d.Set("daily_automatic_backup_start_time", lustreConfig.DailyAutomaticBackupStartTime)
	d.Set("data_compression_type", lustreConfig.DataCompressionType)
	d.Set("deployment_type", lustreConfig.DeploymentType)
	d.Set("dns_name", filesystem.DNSName)
	d.Set("drive_cache_type", lustreConfig.DriveCacheType)
	d.Set("export_path", lustreConfig.DataRepositoryConfiguration.ExportPath)
	d.Set("file_system_type_version", filesystem.FileSystemTypeVersion)
	d.Set("import_path", lustreConfig.DataRepositoryConfiguration.ImportPath)
	d.Set("imported_file_chunk_size", lustreConfig.DataRepositoryConfiguration.ImportedFileChunkSize)
	d.Set(names.AttrKMSKeyID, filesystem.KmsKeyId)
	if err := d.Set("log_configuration", flattenLustreLogConfiguration(lustreConfig.LogConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_configuration: %s", err)
	}
	d.Set("mount_name", lustreConfig.MountName)
	d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds))
	d.Set(names.AttrOwnerID, filesystem.OwnerId)
	d.Set("per_unit_storage_throughput", lustreConfig.PerUnitStorageThroughput)
	if err := d.Set("root_squash_configuration", flattenLustreRootSquashConfiguration(lustreConfig.RootSquashConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting root_squash_configuration: %s", err)
	}
	d.Set("storage_capacity", filesystem.StorageCapacity)
	d.Set("storage_type", filesystem.StorageType)
	d.Set(names.AttrSubnetIDs, aws.StringValueSlice(filesystem.SubnetIds))
	d.Set(names.AttrVPCID, filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", lustreConfig.WeeklyMaintenanceStartTime)

	setTagsOut(ctx, filesystem.Tags)

	return diags
}

func resourceLustreFileSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fsx.UpdateFileSystemInput{
			ClientRequestToken:  aws.String(id.UniqueId()),
			FileSystemId:        aws.String(d.Id()),
			LustreConfiguration: &fsx.UpdateFileSystemLustreConfiguration{},
		}

		if d.HasChange("auto_import_policy") {
			input.LustreConfiguration.AutoImportPolicy = aws.String(d.Get("auto_import_policy").(string))
		}

		if d.HasChange("automatic_backup_retention_days") {
			input.LustreConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(d.Get("automatic_backup_retention_days").(int)))
		}

		if d.HasChange("daily_automatic_backup_start_time") {
			input.LustreConfiguration.DailyAutomaticBackupStartTime = aws.String(d.Get("daily_automatic_backup_start_time").(string))
		}

		if d.HasChange("data_compression_type") {
			input.LustreConfiguration.DataCompressionType = aws.String(d.Get("data_compression_type").(string))
		}

		if d.HasChange("log_configuration") {
			input.LustreConfiguration.LogConfiguration = expandLustreLogCreateConfiguration(d.Get("log_configuration").([]interface{}))
		}

		if d.HasChange("per_unit_storage_throughput") {
			input.LustreConfiguration.PerUnitStorageThroughput = aws.Int64(int64(d.Get("per_unit_storage_throughput").(int)))
		}

		if d.HasChange("root_squash_configuration") {
			input.LustreConfiguration.RootSquashConfiguration = expandLustreRootSquashConfiguration(d.Get("root_squash_configuration").([]interface{}))
		}

		if d.HasChange("storage_capacity") {
			input.StorageCapacity = aws.Int64(int64(d.Get("storage_capacity").(int)))
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.LustreConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		startTime := time.Now()
		_, err := conn.UpdateFileSystemWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSX for Lustre File System (%s): %s", d.Id(), err)
		}

		if _, err := waitFileSystemUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre File System (%s) update: %s", d.Id(), err)
		}

		if _, err := waitFileSystemAdministrativeActionCompleted(ctx, conn, d.Id(), fsx.AdministrativeActionTypeFileSystemUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre File System (%s) administrative action (%s) complete: %s", d.Id(), fsx.AdministrativeActionTypeFileSystemUpdate, err)
		}
	}

	return append(diags, resourceLustreFileSystemRead(ctx, d, meta)...)
}

func resourceLustreFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	log.Printf("[DEBUG] Deleting FSx for Lustre File System: %s", d.Id())
	_, err := conn.DeleteFileSystemWithContext(ctx, &fsx.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) {
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

func expandLustreRootSquashConfiguration(l []interface{}) *fsx.LustreRootSquashConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.LustreRootSquashConfiguration{}

	if v, ok := data["root_squash"].(string); ok && v != "" {
		req.RootSquash = aws.String(v)
	}

	if v, ok := data["no_squash_nids"].(*schema.Set); ok && v.Len() > 0 {
		req.NoSquashNids = flex.ExpandStringSet(v)
	}

	return req
}

func flattenLustreRootSquashConfiguration(adopts *fsx.LustreRootSquashConfiguration) []map[string]interface{} {
	if adopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if adopts.RootSquash != nil {
		m["root_squash"] = aws.StringValue(adopts.RootSquash)
	}

	if adopts.NoSquashNids != nil {
		m["no_squash_nids"] = flex.FlattenStringSet(adopts.NoSquashNids)
	}

	return []map[string]interface{}{m}
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

func findLustreFileSystemByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.FileSystem, error) {
	output, err := findFileSystemByIDAndType(ctx, conn, id, fsx.FileSystemTypeLustre)

	if err != nil {
		return nil, err
	}

	if output.LustreConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}

func findFileSystemByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.FileSystem, error) {
	input := &fsx.DescribeFileSystemsInput{
		FileSystemIds: aws.StringSlice([]string{id}),
	}

	return findFileSystem(ctx, conn, input, tfslices.PredicateTrue[*fsx.FileSystem]())
}

func findFileSystemByIDAndType(ctx context.Context, conn *fsx.FSx, fsID, fsType string) (*fsx.FileSystem, error) {
	input := &fsx.DescribeFileSystemsInput{
		FileSystemIds: aws.StringSlice([]string{fsID}),
	}
	filter := func(fs *fsx.FileSystem) bool {
		return aws.StringValue(fs.FileSystemType) == fsType
	}

	return findFileSystem(ctx, conn, input, filter)
}

func findFileSystem(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeFileSystemsInput, filter tfslices.Predicate[*fsx.FileSystem]) (*fsx.FileSystem, error) {
	output, err := findFileSystems(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findFileSystems(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeFileSystemsInput, filter tfslices.Predicate[*fsx.FileSystem]) ([]*fsx.FileSystem, error) {
	var output []*fsx.FileSystem

	err := conn.DescribeFileSystemsPagesWithContext(ctx, input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FileSystems {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusFileSystem(ctx context.Context, conn *fsx.FSx, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFileSystemByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Lifecycle), nil
	}
}

func waitFileSystemCreated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleCreating},
		Target:  []string{fsx.FileSystemLifecycleAvailable},
		Refresh: statusFileSystem(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.FileSystem); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileSystemLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(details.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitFileSystemUpdated(ctx context.Context, conn *fsx.FSx, id string, startTime time.Time, timeout time.Duration) (*fsx.FileSystem, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleUpdating},
		Target:  []string{fsx.FileSystemLifecycleAvailable},
		Refresh: statusFileSystem(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.FileSystem); ok {
		switch status := aws.StringValue(output.Lifecycle); status {
		case fsx.FileSystemLifecycleFailed, fsx.FileSystemLifecycleMisconfigured, fsx.FileSystemLifecycleMisconfiguredUnavailable:
			// Report any failed non-FILE_SYSTEM_UPDATE administrative actions.
			// See https://docs.aws.amazon.com/fsx/latest/APIReference/API_AdministrativeAction.html#FSx-Type-AdministrativeAction-AdministrativeActionType.
			administrativeActions := tfslices.Filter(output.AdministrativeActions, func(v *fsx.AdministrativeAction) bool {
				return v != nil && aws.StringValue(v.Status) == fsx.StatusFailed && aws.StringValue(v.AdministrativeActionType) != fsx.AdministrativeActionTypeFileSystemUpdate && v.FailureDetails != nil && startTime.Before(aws.TimeValue(v.RequestTime))
			})
			administrativeActionsError := errors.Join(tfslices.ApplyToAll(administrativeActions, func(v *fsx.AdministrativeAction) error {
				return fmt.Errorf("%s: %s", aws.StringValue(v.AdministrativeActionType), aws.StringValue(v.FailureDetails.Message))
			})...)

			if details := output.FailureDetails; details != nil {
				if message := aws.StringValue(details.Message); administrativeActionsError != nil {
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

func waitFileSystemDeleted(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleAvailable, fsx.FileSystemLifecycleDeleting},
		Target:  []string{},
		Refresh: statusFileSystem(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.FileSystem); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileSystemLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(details.Message)))
		}

		return output, err
	}

	return nil, err
}

func findFileSystemAdministrativeAction(ctx context.Context, conn *fsx.FSx, fsID, actionType string) (*fsx.AdministrativeAction, error) {
	output, err := findFileSystemByID(ctx, conn, fsID)

	if err != nil {
		return nil, err
	}

	for _, v := range output.AdministrativeActions {
		if v == nil {
			continue
		}

		if aws.StringValue(v.AdministrativeActionType) == actionType {
			return v, nil
		}
	}

	// If the administrative action isn't found, assume it's complete.
	return &fsx.AdministrativeAction{Status: aws.String(fsx.StatusCompleted)}, nil
}

func statusFileSystemAdministrativeAction(ctx context.Context, conn *fsx.FSx, fsID, actionType string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFileSystemAdministrativeAction(ctx, conn, fsID, actionType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitFileSystemAdministrativeActionCompleted(ctx context.Context, conn *fsx.FSx, fsID, actionType string, timeout time.Duration) (*fsx.AdministrativeAction, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.StatusInProgress, fsx.StatusPending},
		Target:  []string{fsx.StatusCompleted, fsx.StatusUpdatedOptimizing},
		Refresh: statusFileSystemAdministrativeAction(ctx, conn, fsID, actionType),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.AdministrativeAction); ok {
		if status, details := aws.StringValue(output.Status), output.FailureDetails; status == fsx.StatusFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}
