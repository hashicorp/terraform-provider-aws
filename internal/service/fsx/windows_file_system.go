// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package fsx

import (
	"context"
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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_windows_file_system", name="Windows File System")
// @Tags(identifierAttribute="arn")
func resourceWindowsFileSystem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWindowsFileSystemCreate,
		ReadWithoutTimeout:   resourceWindowsFileSystemRead,
		UpdateWithoutTimeout: resourceWindowsFileSystemUpdate,
		DeleteWithoutTimeout: resourceWindowsFileSystemDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				d.Set("skip_final_backup", false)

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(45 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"active_directory_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"self_managed_active_directory"},
			},
			"aliases": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(4, 253),
						// validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]([.][0-9A-Za-z][0-9A-Za-z-]*[0-9A-Za-z])+$`), "must be in the fqdn format hostname.domain"),
					),
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"audit_log_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"audit_log_destination": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidARN,
							StateFunc:    windowsAuditLogStateFunc,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return strings.HasPrefix(old, fmt.Sprintf("%s:", new))
							},
						},
						"file_access_audit_log_level": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.WindowsAccessAuditLogLevelDisabled,
							ValidateDiagFunc: enum.Validate[awstypes.WindowsAccessAuditLogLevel](),
						},
						"file_share_access_audit_log_level": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.WindowsAccessAuditLogLevelDisabled,
							ValidateDiagFunc: enum.Validate[awstypes.WindowsAccessAuditLogLevel](),
						},
					},
				},
			},
			"automatic_backup_retention_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      7,
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
				ForceNew: true,
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.WindowsDeploymentTypeSingleAz1,
				ValidateDiagFunc: enum.Validate[awstypes.WindowsDeploymentType](),
			},
			"disk_iops_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrIOPS: {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(0, 350000),
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
			"final_backup_tags": tftags.TagsSchema(),
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"network_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preferred_file_server_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preferred_subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"remote_administration_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"self_managed_active_directory": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"active_directory_id"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_join_service_account_secret": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
							ConflictsWith: []string{
								"self_managed_active_directory.0.password",
								"self_managed_active_directory.0.username",
							},
						},
						"dns_ips": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							MaxItems: 2,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPAddress,
							},
						},
						names.AttrDomainName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"file_system_administrators_group": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Domain Admins",
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
						"organizational_unit_distinguished_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 2000),
						},
						names.AttrPassword: {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 256),
							ConflictsWith: []string{
								"self_managed_active_directory.0.domain_join_service_account_secret",
							},
						},
						names.AttrUsername: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
							ConflictsWith: []string{
								"self_managed_active_directory.0.domain_join_service_account_secret",
							},
						},
					},
				},
			},
			"skip_final_backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"storage_capacity": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(32, 65536),
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
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntInSlice([]int{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4608, 6144, 9216, 12228}),
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
	}
}

func resourceWindowsFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	inputCFS := fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(sdkid.UniqueId()),
		FileSystemType:     awstypes.FileSystemTypeWindows,
		StorageCapacity:    aws.Int32(int32(d.Get("storage_capacity").(int))),
		SubnetIds:          flex.ExpandStringValueList(d.Get(names.AttrSubnetIDs).([]any)),
		Tags:               getTagsIn(ctx),
		WindowsConfiguration: &awstypes.CreateFileSystemWindowsConfiguration{
			AutomaticBackupRetentionDays: aws.Int32(int32(d.Get("automatic_backup_retention_days").(int))),
			CopyTagsToBackups:            aws.Bool(d.Get("copy_tags_to_backups").(bool)),
			ThroughputCapacity:           aws.Int32(int32(d.Get("throughput_capacity").(int))),
		},
	}
	inputCFSFB := fsx.CreateFileSystemFromBackupInput{
		ClientRequestToken: aws.String(sdkid.UniqueId()),
		SubnetIds:          flex.ExpandStringValueList(d.Get(names.AttrSubnetIDs).([]any)),
		Tags:               getTagsIn(ctx),
		WindowsConfiguration: &awstypes.CreateFileSystemWindowsConfiguration{
			AutomaticBackupRetentionDays: aws.Int32(int32(d.Get("automatic_backup_retention_days").(int))),
			CopyTagsToBackups:            aws.Bool(d.Get("copy_tags_to_backups").(bool)),
			ThroughputCapacity:           aws.Int32(int32(d.Get("throughput_capacity").(int))),
		},
	}

	if v, ok := d.GetOk("active_directory_id"); ok {
		inputCFS.WindowsConfiguration.ActiveDirectoryId = aws.String(v.(string))
		inputCFSFB.WindowsConfiguration.ActiveDirectoryId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("aliases"); ok {
		inputCFS.WindowsConfiguration.Aliases = flex.ExpandStringValueSet(v.(*schema.Set))
		inputCFSFB.WindowsConfiguration.Aliases = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("audit_log_configuration"); ok && len(v.([]any)) > 0 {
		inputCFS.WindowsConfiguration.AuditLogConfiguration = expandWindowsAuditLogCreateConfiguration(v.([]any))
		inputCFSFB.WindowsConfiguration.AuditLogConfiguration = expandWindowsAuditLogCreateConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("daily_automatic_backup_start_time"); ok {
		inputCFS.WindowsConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
		inputCFSFB.WindowsConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disk_iops_configuration"); ok && len(v.([]any)) > 0 {
		inputCFS.WindowsConfiguration.DiskIopsConfiguration = expandWindowsFileSystemDiskIopsConfiguration(v.([]any))
		inputCFSFB.WindowsConfiguration.DiskIopsConfiguration = expandWindowsFileSystemDiskIopsConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("deployment_type"); ok {
		inputCFS.WindowsConfiguration.DeploymentType = awstypes.WindowsDeploymentType(v.(string))
		inputCFSFB.WindowsConfiguration.DeploymentType = awstypes.WindowsDeploymentType(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		inputCFS.KmsKeyId = aws.String(v.(string))
		inputCFSFB.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_subnet_id"); ok {
		inputCFS.WindowsConfiguration.PreferredSubnetId = aws.String(v.(string))
		inputCFSFB.WindowsConfiguration.PreferredSubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok {
		inputCFS.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
		inputCFSFB.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("self_managed_active_directory"); ok {
		inputCFS.WindowsConfiguration.SelfManagedActiveDirectoryConfiguration = expandWindowsFileSystemSelfManagedActiveDirectoryConfiguration(v.([]any))
		inputCFSFB.WindowsConfiguration.SelfManagedActiveDirectoryConfiguration = expandWindowsFileSystemSelfManagedActiveDirectoryConfiguration(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrStorageType); ok {
		inputCFS.StorageType = awstypes.StorageType(v.(string))
		inputCFSFB.StorageType = awstypes.StorageType(v.(string))
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		inputCFS.WindowsConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
		inputCFSFB.WindowsConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("backup_id"); ok {
		backupID := v.(string)
		inputCFSFB.BackupId = aws.String(backupID)

		output, err := conn.CreateFileSystemFromBackup(ctx, &inputCFSFB)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating FSx for Windows File Server File System from backup (%s): %s", backupID, err)
		}

		d.SetId(aws.ToString(output.FileSystem.FileSystemId))
	} else {
		output, err := conn.CreateFileSystem(ctx, &inputCFS)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating FSx for Windows File Server File System: %s", err)
		}

		d.SetId(aws.ToString(output.FileSystem.FileSystemId))
	}

	if _, err := waitFileSystemCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for Windows File Server File System (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceWindowsFileSystemRead(ctx, d, meta)...)
}

func resourceWindowsFileSystemRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	filesystem, err := findWindowsFileSystemByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] FSx for Windows File Server File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for Windows File Server File System (%s): %s", d.Id(), err)
	}

	windowsConfig := filesystem.WindowsConfiguration
	d.Set("active_directory_id", windowsConfig.ActiveDirectoryId)
	d.Set("aliases", expandAliases(windowsConfig.Aliases))
	d.Set(names.AttrARN, filesystem.ResourceARN)
	if err := d.Set("audit_log_configuration", flattenWindowsAuditLogConfiguration(windowsConfig.AuditLogConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting audit_log_configuration: %s", err)
	}
	d.Set("automatic_backup_retention_days", windowsConfig.AutomaticBackupRetentionDays)
	d.Set("copy_tags_to_backups", windowsConfig.CopyTagsToBackups)
	d.Set("daily_automatic_backup_start_time", windowsConfig.DailyAutomaticBackupStartTime)
	d.Set("deployment_type", windowsConfig.DeploymentType)
	if err := d.Set("disk_iops_configuration", flattenWindowsFileSystemDiskIopsConfiguration(windowsConfig.DiskIopsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting disk_iops_configuration: %s", err)
	}
	d.Set(names.AttrDNSName, filesystem.DNSName)
	d.Set(names.AttrKMSKeyID, filesystem.KmsKeyId)
	d.Set("network_interface_ids", filesystem.NetworkInterfaceIds)
	d.Set(names.AttrOwnerID, filesystem.OwnerId)
	d.Set("preferred_file_server_ip", windowsConfig.PreferredFileServerIp)
	d.Set("preferred_subnet_id", windowsConfig.PreferredSubnetId)
	d.Set("remote_administration_endpoint", windowsConfig.RemoteAdministrationEndpoint)
	if err := d.Set("self_managed_active_directory", flattenWindowsFileSystemSelfManagedActiveDirectoryAttributes(d, windowsConfig.SelfManagedActiveDirectoryConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting self_managed_active_directory: %s", err)
	}
	d.Set("storage_capacity", filesystem.StorageCapacity)
	d.Set(names.AttrStorageType, filesystem.StorageType)
	d.Set(names.AttrSubnetIDs, filesystem.SubnetIds)
	d.Set("throughput_capacity", windowsConfig.ThroughputCapacity)
	d.Set(names.AttrVPCID, filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", windowsConfig.WeeklyMaintenanceStartTime)

	setTagsOut(ctx, filesystem.Tags)

	return diags
}

func resourceWindowsFileSystemUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	if d.HasChange("aliases") {
		o, n := d.GetChange("aliases")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

		if len(add) > 0 {
			input := fsx.AssociateFileSystemAliasesInput{
				Aliases:      add,
				FileSystemId: aws.String(d.Id()),
			}

			_, err := conn.AssociateFileSystemAliases(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating FSx for Windows File Server File System (%s) aliases: %s", d.Id(), err)
			}

			if _, err := waitFileSystemAdministrativeActionCompleted(ctx, conn, d.Id(), awstypes.AdministrativeActionTypeFileSystemAliasAssociation, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for FSx for Windows File Server File System (%s) administrative action (%s) complete: %s", d.Id(), awstypes.AdministrativeActionTypeFileSystemAliasAssociation, err)
			}
		}

		if len(del) > 0 {
			input := fsx.DisassociateFileSystemAliasesInput{
				Aliases:      del,
				FileSystemId: aws.String(d.Id()),
			}

			_, err := conn.DisassociateFileSystemAliases(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating FSx for Windows File Server File System (%s) aliases: %s", d.Id(), err)
			}

			if _, err := waitFileSystemAdministrativeActionCompleted(ctx, conn, d.Id(), awstypes.AdministrativeActionTypeFileSystemAliasDisassociation, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for FSx for Windows File Server File System (%s) administrative action (%s) complete: %s", d.Id(), awstypes.AdministrativeActionTypeFileSystemAliasDisassociation, err)
			}
		}
	}

	// Increase ThroughputCapacity first to avoid errors like
	// "BadRequest: Unable to perform the storage capacity update. Updating storage capacity requires your file system to have at least 16 MB/s of throughput capacity."
	if d.HasChange("throughput_capacity") {
		o, n := d.GetChange("throughput_capacity")
		if o, n := o.(int), n.(int); n > o {
			input := fsx.UpdateFileSystemInput{
				ClientRequestToken: aws.String(sdkid.UniqueId()),
				FileSystemId:       aws.String(d.Id()),
				WindowsConfiguration: &awstypes.UpdateFileSystemWindowsConfiguration{
					ThroughputCapacity: aws.Int32(int32(n)),
				},
			}

			startTime := time.Now()
			_, err := conn.UpdateFileSystem(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating FSx for Windows File Server File System (%s) ThroughputCapacity: %s", d.Id(), err)
			}

			if _, err := waitFileSystemUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for FSx Windows File Server File System (%s) update: %s", d.Id(), err)
			}

			if _, err := waitFileSystemAdministrativeActionCompleted(ctx, conn, d.Id(), awstypes.AdministrativeActionTypeFileSystemUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for FSx Windows File Server File System (%s) administrative action (%s) complete: %s", d.Id(), awstypes.AdministrativeActionTypeFileSystemUpdate, err)
			}
		}
	}

	if d.HasChangesExcept(
		"aliases",
		"final_backup_tags",
		"skip_final_backup",
		names.AttrTags,
		names.AttrTagsAll,
	) {
		input := fsx.UpdateFileSystemInput{
			ClientRequestToken:   aws.String(sdkid.UniqueId()),
			FileSystemId:         aws.String(d.Id()),
			WindowsConfiguration: &awstypes.UpdateFileSystemWindowsConfiguration{},
		}

		if d.HasChange("audit_log_configuration") {
			input.WindowsConfiguration.AuditLogConfiguration = expandWindowsAuditLogCreateConfiguration(d.Get("audit_log_configuration").([]any))
		}

		if d.HasChange("automatic_backup_retention_days") {
			input.WindowsConfiguration.AutomaticBackupRetentionDays = aws.Int32(int32(d.Get("automatic_backup_retention_days").(int)))
		}

		if d.HasChange("daily_automatic_backup_start_time") {
			input.WindowsConfiguration.DailyAutomaticBackupStartTime = aws.String(d.Get("daily_automatic_backup_start_time").(string))
		}

		if d.HasChange("disk_iops_configuration") {
			input.WindowsConfiguration.DiskIopsConfiguration = expandWindowsFileSystemDiskIopsConfiguration(d.Get("disk_iops_configuration").([]any))
		}

		if d.HasChange("self_managed_active_directory") {
			input.WindowsConfiguration.SelfManagedActiveDirectoryConfiguration = expandWindowsFileSystemSelfManagedActiveDirectoryConfigurationUpdates(d.Get("self_managed_active_directory").([]any))
		}

		if d.HasChange("storage_capacity") {
			input.StorageCapacity = aws.Int32(int32(d.Get("storage_capacity").(int)))
		}

		if d.HasChange("throughput_capacity") {
			input.WindowsConfiguration.ThroughputCapacity = aws.Int32(int32(d.Get("throughput_capacity").(int)))
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.WindowsConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		startTime := time.Now()
		_, err := conn.UpdateFileSystem(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx for Windows File Server File System (%s): %s", d.Id(), err)
		}

		if _, err := waitFileSystemUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx Windows File Server File System (%s) update: %s", d.Id(), err)
		}

		if _, err := waitFileSystemAdministrativeActionCompleted(ctx, conn, d.Id(), awstypes.AdministrativeActionTypeFileSystemUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx Windows File Server File System (%s) administrative action (%s) complete: %s", d.Id(), awstypes.AdministrativeActionTypeFileSystemUpdate, err)
		}
	}

	return append(diags, resourceWindowsFileSystemRead(ctx, d, meta)...)
}

func resourceWindowsFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := fsx.DeleteFileSystemInput{
		ClientRequestToken: aws.String(sdkid.UniqueId()),
		FileSystemId:       aws.String(d.Id()),
		WindowsConfiguration: &awstypes.DeleteFileSystemWindowsConfiguration{
			SkipFinalBackup: aws.Bool(d.Get("skip_final_backup").(bool)),
		},
	}

	if v, ok := d.GetOk("final_backup_tags"); ok && len(v.(map[string]any)) > 0 {
		input.WindowsConfiguration.FinalBackupTags = svcTags(tftags.New(ctx, v))
	}

	log.Printf("[DEBUG] Deleting FSx for Windows File Server File System: %s", d.Id())
	_, err := conn.DeleteFileSystem(ctx, &input)

	if errs.IsA[*awstypes.FileSystemNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx for Windows File Server File System (%s): %s", d.Id(), err)
	}

	if _, err := waitFileSystemDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for Windows File Server File System (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandAliases(apiObjects []awstypes.Alias) []string {
	return tfslices.ApplyToAll(apiObjects, func(v awstypes.Alias) string {
		return aws.ToString(v.Name)
	})
}

func expandWindowsFileSystemSelfManagedActiveDirectoryConfiguration(tfList []any) *awstypes.SelfManagedActiveDirectoryConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.SelfManagedActiveDirectoryConfiguration{
		DomainName: aws.String(tfMap[names.AttrDomainName].(string)),
		DnsIps:     flex.ExpandStringValueSet(tfMap["dns_ips"].(*schema.Set)),
	}

	if v, ok := tfMap["domain_join_service_account_secret"].(string); ok && v != "" {
		apiObject.DomainJoinServiceAccountSecret = aws.String(v)
	}

	if v, ok := tfMap["file_system_administrators_group"]; ok && v.(string) != "" {
		apiObject.FileSystemAdministratorsGroup = aws.String(v.(string))
	}

	if v, ok := tfMap["organizational_unit_distinguished_name"]; ok && v.(string) != "" {
		apiObject.OrganizationalUnitDistinguishedName = aws.String(v.(string))
	}

	if v, ok := tfMap[names.AttrPassword].(string); ok && v != "" {
		apiObject.Password = aws.String(v)
	}

	if v, ok := tfMap[names.AttrUsername].(string); ok && v != "" {
		apiObject.UserName = aws.String(v)
	}

	return apiObject
}

func expandWindowsFileSystemSelfManagedActiveDirectoryConfigurationUpdates(tfList []any) *awstypes.SelfManagedActiveDirectoryConfigurationUpdates {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.SelfManagedActiveDirectoryConfigurationUpdates{}

	if v, ok := tfMap["domain_join_service_account_secret"].(string); ok && v != "" {
		apiObject.DomainJoinServiceAccountSecret = aws.String(v)
	}

	if v, ok := tfMap["dns_ips"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.DnsIps = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrPassword].(string); ok && v != "" {
		apiObject.Password = aws.String(v)
	}

	if v, ok := tfMap[names.AttrUsername].(string); ok && v != "" {
		apiObject.UserName = aws.String(v)
	}

	return apiObject
}

func flattenWindowsFileSystemSelfManagedActiveDirectoryAttributes(d *schema.ResourceData, apiObject *awstypes.SelfManagedActiveDirectoryAttributes) []any {
	if apiObject == nil {
		return []any{}
	}

	// Since we are in a configuration block and the FSx API does not return
	// the password, we need to set the value if we can or Terraform will
	// show a difference for the argument from empty string to the value.
	// This is not a pattern that should be used normally.
	// See also: flattenEmrKerberosAttributes

	tfMap := map[string]any{
		"domain_join_service_account_secret":     aws.ToString(apiObject.DomainJoinServiceAccountSecret),
		"dns_ips":                                apiObject.DnsIps,
		names.AttrDomainName:                     aws.ToString(apiObject.DomainName),
		"file_system_administrators_group":       aws.ToString(apiObject.FileSystemAdministratorsGroup),
		"organizational_unit_distinguished_name": aws.ToString(apiObject.OrganizationalUnitDistinguishedName),
		names.AttrPassword:                       d.Get("self_managed_active_directory.0.password").(string),
		names.AttrUsername:                       aws.ToString(apiObject.UserName),
	}

	return []any{tfMap}
}

func expandWindowsAuditLogCreateConfiguration(tfList []any) *awstypes.WindowsAuditLogCreateConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	fileAccessAuditLogLevel, ok1 := tfMap["file_access_audit_log_level"].(string)
	fileShareAccessAuditLogLevel, ok2 := tfMap["file_share_access_audit_log_level"].(string)

	if !ok1 || !ok2 {
		return nil
	}

	apiObject := &awstypes.WindowsAuditLogCreateConfiguration{
		FileAccessAuditLogLevel:      awstypes.WindowsAccessAuditLogLevel(fileAccessAuditLogLevel),
		FileShareAccessAuditLogLevel: awstypes.WindowsAccessAuditLogLevel(fileShareAccessAuditLogLevel),
	}

	// audit_log_destination cannot be included in the request if the log levels are disabled.
	if fileAccessAuditLogLevel == string(awstypes.WindowsAccessAuditLogLevelDisabled) && fileShareAccessAuditLogLevel == string(awstypes.WindowsAccessAuditLogLevelDisabled) {
		return apiObject
	}

	if v, ok := tfMap["audit_log_destination"].(string); ok && v != "" {
		apiObject.AuditLogDestination = aws.String(windowsAuditLogStateFunc(v))
	}

	return apiObject
}

func flattenWindowsAuditLogConfiguration(apiObject *awstypes.WindowsAuditLogConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"file_access_audit_log_level":       apiObject.FileAccessAuditLogLevel,
		"file_share_access_audit_log_level": apiObject.FileShareAccessAuditLogLevel,
	}

	if apiObject.AuditLogDestination != nil {
		tfMap["audit_log_destination"] = aws.ToString(apiObject.AuditLogDestination)
	}

	return []any{tfMap}
}

func expandWindowsFileSystemDiskIopsConfiguration(tfList []any) *awstypes.DiskIopsConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.DiskIopsConfiguration{}

	if v, ok := tfMap[names.AttrIOPS].(int); ok {
		apiObject.Iops = aws.Int64(int64(v))
	}

	if v, ok := tfMap[names.AttrMode].(string); ok && v != "" {
		apiObject.Mode = awstypes.DiskIopsConfigurationMode(v)
	}

	return apiObject
}

func flattenWindowsFileSystemDiskIopsConfiguration(apiObject *awstypes.DiskIopsConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrMode: apiObject.Mode,
	}

	if apiObject.Iops != nil {
		tfMap[names.AttrIOPS] = aws.ToInt64(apiObject.Iops)
	}

	return []any{tfMap}
}

func windowsAuditLogStateFunc(v any) string {
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

func findWindowsFileSystemByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.FileSystem, error) {
	output, err := findFileSystemByIDAndType(ctx, conn, id, awstypes.FileSystemTypeWindows)

	if err != nil {
		return nil, err
	}

	if output.WindowsConfiguration == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
