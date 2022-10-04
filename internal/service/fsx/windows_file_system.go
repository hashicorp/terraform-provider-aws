package fsx

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
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

func ResourceWindowsFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceWindowsFileSystemCreate,
		Read:   resourceWindowsFileSystemRead,
		Update: resourceWindowsFileSystemUpdate,
		Delete: resourceWindowsFileSystemDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
						// validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9]([.][A-Za-z0-9][A-Za-z0-9-]*[A-Za-z0-9])+$`), "must be in the fqdn format hostname.domain"),
					),
				},
			},
			"arn": {
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      fsx.WindowsAccessAuditLogLevelDisabled,
							ValidateFunc: validation.StringInSlice(fsx.WindowsAccessAuditLogLevel_Values(), false),
						},
						"file_share_access_audit_log_level": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      fsx.WindowsAccessAuditLogLevelDisabled,
							ValidateFunc: validation.StringInSlice(fsx.WindowsAccessAuditLogLevel_Values(), false),
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
					validation.StringMatch(regexp.MustCompile(`^([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format HH:MM"),
				),
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
				Type:     schema.TypeSet,
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
			"self_managed_active_directory": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"active_directory_id"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"domain_name": {
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
						"password": {
							Type:         schema.TypeString,
							Required:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
						"username": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
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
			"subnet_ids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"throughput_capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(8, 2048),
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
			"deployment_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.WindowsDeploymentTypeSingleAz1,
				ValidateFunc: validation.StringInSlice(fsx.WindowsDeploymentType_Values(), false),
			},
			"preferred_subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"preferred_file_server_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"remote_administration_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.StorageTypeSsd,
				ValidateFunc: validation.StringInSlice(fsx.StorageType_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWindowsFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemType:     aws.String(fsx.FileSystemTypeWindows),
		StorageCapacity:    aws.Int64(int64(d.Get("storage_capacity").(int))),
		SubnetIds:          flex.ExpandStringList(d.Get("subnet_ids").([]interface{})),
		WindowsConfiguration: &fsx.CreateFileSystemWindowsConfiguration{
			AutomaticBackupRetentionDays: aws.Int64(int64(d.Get("automatic_backup_retention_days").(int))),
			CopyTagsToBackups:            aws.Bool(d.Get("copy_tags_to_backups").(bool)),
			ThroughputCapacity:           aws.Int64(int64(d.Get("throughput_capacity").(int))),
		},
	}

	backupInput := &fsx.CreateFileSystemFromBackupInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		SubnetIds:          flex.ExpandStringList(d.Get("subnet_ids").([]interface{})),
		WindowsConfiguration: &fsx.CreateFileSystemWindowsConfiguration{
			AutomaticBackupRetentionDays: aws.Int64(int64(d.Get("automatic_backup_retention_days").(int))),
			CopyTagsToBackups:            aws.Bool(d.Get("copy_tags_to_backups").(bool)),
			ThroughputCapacity:           aws.Int64(int64(d.Get("throughput_capacity").(int))),
		},
	}

	if v, ok := d.GetOk("active_directory_id"); ok {
		input.WindowsConfiguration.ActiveDirectoryId = aws.String(v.(string))
		backupInput.WindowsConfiguration.ActiveDirectoryId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("aliases"); ok {
		input.WindowsConfiguration.Aliases = flex.ExpandStringSet(v.(*schema.Set))
		backupInput.WindowsConfiguration.Aliases = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("deployment_type"); ok {
		input.WindowsConfiguration.DeploymentType = aws.String(v.(string))
		backupInput.WindowsConfiguration.DeploymentType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_subnet_id"); ok {
		input.WindowsConfiguration.PreferredSubnetId = aws.String(v.(string))
		backupInput.WindowsConfiguration.PreferredSubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("daily_automatic_backup_start_time"); ok {
		input.WindowsConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
		backupInput.WindowsConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
		backupInput.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		backupInput.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("self_managed_active_directory"); ok {
		input.WindowsConfiguration.SelfManagedActiveDirectoryConfiguration = expandSelfManagedActiveDirectoryConfigurationCreate(v.([]interface{}))
		backupInput.WindowsConfiguration.SelfManagedActiveDirectoryConfiguration = expandSelfManagedActiveDirectoryConfigurationCreate(v.([]interface{}))
	}

	if v, ok := d.GetOk("audit_log_configuration"); ok && len(v.([]interface{})) > 0 {
		input.WindowsConfiguration.AuditLogConfiguration = expandWindowsAuditLogCreateConfiguration(v.([]interface{}))
		backupInput.WindowsConfiguration.AuditLogConfiguration = expandWindowsAuditLogCreateConfiguration(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
		backupInput.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		input.WindowsConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
		backupInput.WindowsConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("storage_type"); ok {
		input.StorageType = aws.String(v.(string))
		backupInput.StorageType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("backup_id"); ok {
		backupInput.BackupId = aws.String(v.(string))

		log.Printf("[DEBUG] Creating FSx Windows File System: %s", backupInput)
		result, err := conn.CreateFileSystemFromBackup(backupInput)

		if err != nil {
			return fmt.Errorf("error creating FSx Windows File System from backup: %w", err)
		}

		d.SetId(aws.StringValue(result.FileSystem.FileSystemId))
	} else {
		log.Printf("[DEBUG] Creating FSx Windows File System: %s", input)
		result, err := conn.CreateFileSystem(input)

		if err != nil {
			return fmt.Errorf("error creating FSx Windows File System: %w", err)
		}

		d.SetId(aws.StringValue(result.FileSystem.FileSystemId))
	}

	if _, err := waitFileSystemCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for FSx Windows File System (%s) create: %w", d.Id(), err)
	}

	return resourceWindowsFileSystemRead(d, meta)
}

func resourceWindowsFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Windows File System (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	if d.HasChange("aliases") {
		o, n := d.GetChange("aliases")

		if err := updateAliases(conn, d.Id(), o.(*schema.Set), n.(*schema.Set), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error updating FSx Windows File System (%s) aliases: %w", d.Id(), err)
		}
	}

	if d.HasChangesExcept("tags_all", "aliases") {
		input := &fsx.UpdateFileSystemInput{
			ClientRequestToken:   aws.String(resource.UniqueId()),
			FileSystemId:         aws.String(d.Id()),
			WindowsConfiguration: &fsx.UpdateFileSystemWindowsConfiguration{},
		}

		if d.HasChange("automatic_backup_retention_days") {
			input.WindowsConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(d.Get("automatic_backup_retention_days").(int)))
		}

		if d.HasChange("throughput_capacity") {
			input.WindowsConfiguration.ThroughputCapacity = aws.Int64(int64(d.Get("throughput_capacity").(int)))
		}

		if d.HasChange("storage_capacity") {
			input.StorageCapacity = aws.Int64(int64(d.Get("storage_capacity").(int)))
		}

		if d.HasChange("daily_automatic_backup_start_time") {
			input.WindowsConfiguration.DailyAutomaticBackupStartTime = aws.String(d.Get("daily_automatic_backup_start_time").(string))
		}

		if d.HasChange("self_managed_active_directory") {
			input.WindowsConfiguration.SelfManagedActiveDirectoryConfiguration = expandSelfManagedActiveDirectoryConfigurationUpdate(d.Get("self_managed_active_directory").([]interface{}))
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.WindowsConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		if d.HasChange("audit_log_configuration") {
			input.WindowsConfiguration.AuditLogConfiguration = expandWindowsAuditLogCreateConfiguration(d.Get("audit_log_configuration").([]interface{}))
		}

		_, err := conn.UpdateFileSystem(input)

		if err != nil {
			return fmt.Errorf("error updating FSx Windows File System (%s): %w", d.Id(), err)
		}

		if _, err := waitAdministrativeActionCompleted(conn, d.Id(), fsx.AdministrativeActionTypeFileSystemUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for FSx Windows File System (%s) update: %w", d.Id(), err)
		}
	}

	return resourceWindowsFileSystemRead(d, meta)
}

func resourceWindowsFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	filesystem, err := FindFileSystemByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Windows File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

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
	d.Set("preferred_subnet_id", filesystem.WindowsConfiguration.PreferredSubnetId)
	d.Set("preferred_file_server_ip", filesystem.WindowsConfiguration.PreferredFileServerIp)
	d.Set("remote_administration_endpoint", filesystem.WindowsConfiguration.RemoteAdministrationEndpoint)
	d.Set("dns_name", filesystem.DNSName)
	d.Set("kms_key_id", filesystem.KmsKeyId)
	d.Set("storage_type", filesystem.StorageType)

	if err := d.Set("aliases", aws.StringValueSlice(expandAliasValues(filesystem.WindowsConfiguration.Aliases))); err != nil {
		return fmt.Errorf("error setting aliases: %w", err)
	}

	if err := d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds)); err != nil {
		return fmt.Errorf("error setting network_interface_ids: %w", err)
	}

	d.Set("owner_id", filesystem.OwnerId)

	if err := d.Set("self_managed_active_directory", flattenSelfManagedActiveDirectoryConfiguration(d, filesystem.WindowsConfiguration.SelfManagedActiveDirectoryConfiguration)); err != nil {
		return fmt.Errorf("error setting self_managed_active_directory: %w", err)
	}

	if err := d.Set("audit_log_configuration", flattenWindowsAuditLogConfiguration(filesystem.WindowsConfiguration.AuditLogConfiguration)); err != nil {
		return fmt.Errorf("error setting audit_log_configuration: %w", err)
	}

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

	d.Set("throughput_capacity", filesystem.WindowsConfiguration.ThroughputCapacity)
	d.Set("vpc_id", filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", filesystem.WindowsConfiguration.WeeklyMaintenanceStartTime)

	return nil
}

func resourceWindowsFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	input := &fsx.DeleteFileSystemInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemId:       aws.String(d.Id()),
		WindowsConfiguration: &fsx.DeleteFileSystemWindowsConfiguration{
			SkipFinalBackup: aws.Bool(d.Get("skip_final_backup").(bool)),
		},
	}

	log.Printf("[DEBUG] Deleting FSx Windows File System: %s", d.Id())
	_, err := conn.DeleteFileSystem(input)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FSx Windows File System (%s): %w", d.Id(), err)
	}

	if _, err := waitFileSystemDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for FSx Windows File System (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func expandAliasValues(aliases []*fsx.Alias) []*string {
	var alternateDNSNames []*string

	for _, alias := range aliases {
		aName := alias.Name
		alternateDNSNames = append(alternateDNSNames, aName)
	}

	return alternateDNSNames
}

func updateAliases(conn *fsx.FSx, identifier string, oldSet *schema.Set, newSet *schema.Set, timeout time.Duration) error {
	if newSet.Len() > 0 {
		if newAliases := newSet.Difference(oldSet); newAliases.Len() > 0 {

			input := &fsx.AssociateFileSystemAliasesInput{
				FileSystemId: aws.String(identifier),
				Aliases:      flex.ExpandStringSet(newAliases),
			}

			_, err := conn.AssociateFileSystemAliases(input)

			if err != nil {
				return fmt.Errorf("error associating aliases to FSx file system (%s): %w", identifier, err)
			}

			if _, err := waitAdministrativeActionCompleted(conn, identifier, fsx.AdministrativeActionTypeFileSystemAliasAssociation, timeout); err != nil {
				return fmt.Errorf("error waiting for FSx Windows File System (%s) alias to be associated: %w", identifier, err)
			}
		}
	}

	if oldSet.Len() > 0 {
		if oldAliases := oldSet.Difference(newSet); oldAliases.Len() > 0 {
			input := &fsx.DisassociateFileSystemAliasesInput{
				FileSystemId: aws.String(identifier),
				Aliases:      flex.ExpandStringSet(oldAliases),
			}

			_, err := conn.DisassociateFileSystemAliases(input)

			if err != nil {
				return fmt.Errorf("error disassociating aliases from FSx file system (%s): %w", identifier, err)
			}

			if _, err := waitAdministrativeActionCompleted(conn, identifier, fsx.AdministrativeActionTypeFileSystemAliasDisassociation, timeout); err != nil {
				return fmt.Errorf("error waiting for FSx Windows File System (%s) alias to be disassociated: %w", identifier, err)
			}
		}
	}

	return nil
}

func expandSelfManagedActiveDirectoryConfigurationCreate(l []interface{}) *fsx.SelfManagedActiveDirectoryConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.SelfManagedActiveDirectoryConfiguration{
		DomainName: aws.String(data["domain_name"].(string)),
		DnsIps:     flex.ExpandStringSet(data["dns_ips"].(*schema.Set)),
		Password:   aws.String(data["password"].(string)),
		UserName:   aws.String(data["username"].(string)),
	}

	if v, ok := data["file_system_administrators_group"]; ok && v.(string) != "" {
		req.FileSystemAdministratorsGroup = aws.String(v.(string))
	}

	if v, ok := data["organizational_unit_distinguished_name"]; ok && v.(string) != "" {
		req.OrganizationalUnitDistinguishedName = aws.String(v.(string))
	}

	return req
}

func expandSelfManagedActiveDirectoryConfigurationUpdate(l []interface{}) *fsx.SelfManagedActiveDirectoryConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.SelfManagedActiveDirectoryConfigurationUpdates{}

	if v, ok := data["dns_ips"].(*schema.Set); ok && v.Len() > 0 {
		req.DnsIps = flex.ExpandStringSet(v)
	}

	if v, ok := data["password"].(string); ok && v != "" {
		req.Password = aws.String(v)
	}

	if v, ok := data["username"].(string); ok && v != "" {
		req.UserName = aws.String(v)
	}

	return req
}

func flattenSelfManagedActiveDirectoryConfiguration(d *schema.ResourceData, adopts *fsx.SelfManagedActiveDirectoryAttributes) []map[string]interface{} {
	if adopts == nil {
		return []map[string]interface{}{}
	}

	// Since we are in a configuration block and the FSx API does not return
	// the password, we need to set the value if we can or Terraform will
	// show a difference for the argument from empty string to the value.
	// This is not a pattern that should be used normally.
	// See also: flattenEmrKerberosAttributes

	m := map[string]interface{}{
		"dns_ips":                                aws.StringValueSlice(adopts.DnsIps),
		"domain_name":                            aws.StringValue(adopts.DomainName),
		"file_system_administrators_group":       aws.StringValue(adopts.FileSystemAdministratorsGroup),
		"organizational_unit_distinguished_name": aws.StringValue(adopts.OrganizationalUnitDistinguishedName),
		"password":                               d.Get("self_managed_active_directory.0.password").(string),
		"username":                               aws.StringValue(adopts.UserName),
	}

	return []map[string]interface{}{m}
}

func expandWindowsAuditLogCreateConfiguration(l []interface{}) *fsx.WindowsAuditLogCreateConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.WindowsAuditLogCreateConfiguration{
		FileAccessAuditLogLevel:      aws.String(data["file_access_audit_log_level"].(string)),
		FileShareAccessAuditLogLevel: aws.String(data["file_share_access_audit_log_level"].(string)),
	}

	if v, ok := data["audit_log_destination"].(string); ok && v != "" {
		req.AuditLogDestination = aws.String(windowsAuditLogStateFunc(v))
	}

	return req
}

func flattenWindowsAuditLogConfiguration(adopts *fsx.WindowsAuditLogConfiguration) []map[string]interface{} {
	if adopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"file_access_audit_log_level":       aws.StringValue(adopts.FileAccessAuditLogLevel),
		"file_share_access_audit_log_level": aws.StringValue(adopts.FileShareAccessAuditLogLevel),
	}

	if adopts.AuditLogDestination != nil {
		m["audit_log_destination"] = aws.StringValue(adopts.AuditLogDestination)
	}

	return []map[string]interface{}{m}
}

func windowsAuditLogStateFunc(v interface{}) string {
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
