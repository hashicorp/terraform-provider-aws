package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsFsxFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFsxFileSystemCreate,
		Read:   resourceAwsFsxFileSystemRead,
		Update: resourceAwsFsxFileSystemUpdate,
		Delete: resourceAwsFsxFileSystemDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					fsx.FileSystemTypeLustre,
					fsx.FileSystemTypeWindows,
				}, false),
			},
			"capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(300),
			},
			"kms_key_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"lustre_configuration"},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      120,
				ValidateFunc: validation.IntAtLeast(60),
			},
			"lustre_configuration": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"windows_configuration"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"import_path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"export_path": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"chunk_size": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"weekly_maintenance_start_time": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"windows_configuration": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"lustre_configuration"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active_directory_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"self_managed_active_directory": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"username": {
										Type:     schema.TypeString,
										Required: true,
									},
									"password": {
										Type:     schema.TypeString,
										Required: true,
									},
									"dns_ips": {
										Type:     schema.TypeSet,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"domain_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"administrators_group": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"ou_distinguished_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"backup_retention": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      7,
							ValidateFunc: validation.IntBetween(0, 35),
						},
						"copy_tags_to_backups": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
						"daily_backup_start_time": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"throughput_capacity": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"weekly_maintenance_start_time": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsFsxFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	request := &fsx.CreateFileSystemInput{
		FileSystemType:  aws.String(d.Get("type").(string)),
		StorageCapacity: aws.Int64(int64(d.Get("capacity").(int))),
		SubnetIds:       expandStringList(d.Get("subnet_ids").(*schema.Set).List()),
	}

	if _, ok := d.GetOk("kms_key_id"); ok {
		request.KmsKeyId = aws.String(d.Get("kms_key_id").(string))
	}

	if _, ok := d.GetOk("security_group_ids"); ok {
		request.SecurityGroupIds = expandStringList(d.Get("security_group_ids").(*schema.Set).List())
	}

	if _, ok := d.GetOk("lustre_configuration"); ok {
		request.LustreConfiguration = expandFsxLustreConfigurationCreate(d.Get("lustre_configuration").([]interface{}))
	}

	if _, ok := d.GetOk("windows_configuration"); ok {
		request.WindowsConfiguration = expandFsxWindowsConfigurationCreate(d.Get("windows_configuration").([]interface{}))
	}

	if value, ok := d.GetOk("tags"); ok {
		request.Tags = tagsFromMapFSX(value.(map[string]interface{}))
	}

	log.Printf("[DEBUG] FSx Filesystem create opts: %s", request)
	result, err := conn.CreateFileSystem(request)
	if err != nil {
		return fmt.Errorf("Error creating FSx filesystem: %s", err)
	}

	d.SetId(*result.FileSystem.FileSystemId)

	log.Println("[DEBUG] Waiting for filesystem to become available")

	stateConf := &resource.StateChangeConf{
		Pending:    []string{fsx.FileSystemLifecycleCreating},
		Target:     []string{fsx.FileSystemLifecycleAvailable},
		Refresh:    fsxStateRefreshFunc(conn, d.Id()),
		Timeout:    time.Duration(d.Get("timeout").(int)) * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for filesystem (%s) to become available: %s",
			*result.FileSystem.FileSystemId, err)
	}

	return resourceAwsFsxFileSystemRead(d, meta)
}

func resourceAwsFsxFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	if d.HasChange("tags") {
		if err := setTagsFSX(conn, d); err != nil {
			return fmt.Errorf("Error updating tags for FSx filesystem: %s", err)
		}
	}

	requestUpdate := false
	params := &fsx.UpdateFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	}

	if d.HasChange("lustre_configuration") {
		params.LustreConfiguration = expandFsxLustreConfigurationUpdate(d.Get("lustre_configuration").([]interface{}))
		requestUpdate = true
	}

	if d.HasChange("windows_configuration") {
		params.WindowsConfiguration = expandFsxWindowsConfigurationUpdate(d.Get("windows_configuration").([]interface{}))
		requestUpdate = true
	}

	if requestUpdate {
		_, err := conn.UpdateFileSystem(params)
		if err != nil {
			fmt.Errorf("error updating FSX File System (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsFsxFileSystemRead(d, meta)
}

func resourceAwsFsxFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	filesystem, err := describeFsxFileSystem(conn, d.Id())

	if isAWSErr(err, fsx.ErrCodeFileSystemNotFound, "") {
		log.Printf("[WARN] FSx File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading FSx File System (%s): %s", d.Id(), err)
	}

	if filesystem == nil {
		log.Printf("[WARN] FSx File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("type", filesystem.FileSystemType)
	d.Set("capacity", filesystem.StorageCapacity)
	d.Set("arn", filesystem.ResourceARN)
	d.Set("dns_name", filesystem.DNSName)
	d.Set("kms_key_id", filesystem.KmsKeyId)

	d.Set("tags", tagsToMapFSX(filesystem.Tags))

	err = d.Set("lustre_configuration", flattenLustreOptsConfig(filesystem.LustreConfiguration))
	if err != nil {
		return err
	}

	err = d.Set("windows_configuration", flattenWindowsOptsConfig(filesystem.WindowsConfiguration))
	if err != nil {
		return err
	}

	err = d.Set("subnet_ids", aws.StringValueSlice(filesystem.SubnetIds))
	if err != nil {
		return err
	}

	err = d.Set("security_group_ids", expandStringList(d.Get("security_group_ids").(*schema.Set).List()))
	if err != nil {
		return err
	}

	return nil
}

func resourceAwsFsxFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	request := &fsx.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	}

	_, err := conn.DeleteFileSystem(request)
	if err != nil {
		return fmt.Errorf("Error deleting FSx filesystem: %s", err)
	}

	log.Println("[DEBUG] Waiting for filesystem to delete")

	stateConf := &resource.StateChangeConf{
		Pending:    []string{fsx.FileSystemLifecycleAvailable, fsx.FileSystemLifecycleDeleting},
		Target:     []string{},
		Refresh:    fsxStateRefreshFunc(conn, d.Id()),
		Timeout:    time.Duration(d.Get("timeout").(int)) * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for filesystem (%s) to delete: %s", d.Id(), err)
	}

	return nil
}

func describeFsxFileSystem(conn *fsx.FSx, id string) (*fsx.FileSystem, error) {
	input := &fsx.DescribeFileSystemsInput{
		FileSystemIds: []*string{aws.String(id)},
	}
	var filesystem *fsx.FileSystem

	err := conn.DescribeFileSystemsPages(input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
		for _, fs := range page.FileSystems {
			if aws.StringValue(fs.FileSystemId) == id {
				filesystem = fs
				break
			}
		}

		return !lastPage
	})

	return filesystem, err
}

func fsxStateRefreshFunc(conn *fsx.FSx, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeFileSystems(&fsx.DescribeFileSystemsInput{
			FileSystemIds: []*string{aws.String(id)},
		})

		if resp == nil {
			return nil, "", nil
		}

		if isAWSErr(err, fsx.ErrCodeFileSystemNotFound, "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		filesystem := resp.FileSystems[0]
		return filesystem, *filesystem.Lifecycle, nil
	}
}

func hasEmptyFsxFileSystems(fs *fsx.DescribeFileSystemsOutput) bool {
	if fs != nil && len(fs.FileSystems) > 0 {
		return false
	}
	return true
}

func expandFsxLustreConfigurationCreate(l []interface{}) *fsx.CreateFileSystemLustreConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.CreateFileSystemLustreConfiguration{}

	if data["import_path"].(string) != "" {
		req.ImportPath = aws.String(data["import_path"].(string))
	}

	if data["export_path"].(string) != "" {
		req.ExportPath = aws.String(data["export_path"].(string))
	}

	if data["chunk_size"] != nil {
		req.ImportedFileChunkSize = aws.Int64(int64(data["chunk_size"].(int)))
	}

	if data["weekly_maintenance_start_time"].(string) != "" {
		req.WeeklyMaintenanceStartTime = aws.String(data["weekly_maintenance_start_time"].(string))
	}

	return req
}

func expandFsxLustreConfigurationUpdate(l []interface{}) *fsx.UpdateFileSystemLustreConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.UpdateFileSystemLustreConfiguration{}

	if data["weekly_maintenance_start_time"].(string) != "" {
		req.WeeklyMaintenanceStartTime = aws.String(data["weekly_maintenance_start_time"].(string))
	}

	return req
}

func expandFsxWindowsConfigurationCreate(l []interface{}) *fsx.CreateFileSystemWindowsConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.CreateFileSystemWindowsConfiguration{
		ThroughputCapacity: aws.Int64(int64(data["throughput_capacity"].(int))),
	}

	if data["active_directory_id"].(string) != "" {
		req.ActiveDirectoryId = aws.String(data["active_directory_id"].(string))
	}

	if data["self_managed_active_directory"] != nil {
		req.SelfManagedActiveDirectoryConfiguration = expandSelfManagedAdOptsCreate(data["self_managed_active_directory"].([]interface{}))
	}

	if data["backup_retention"] != nil {
		req.AutomaticBackupRetentionDays = aws.Int64(int64(data["backup_retention"].(int)))
	}

	if data["copy_tags_to_backups"] != nil {
		req.CopyTagsToBackups = aws.Bool(data["copy_tags_to_backups"].(bool))
	}

	if data["daily_backup_start_time"].(string) != "" {
		req.DailyAutomaticBackupStartTime = aws.String(data["daily_backup_start_time"].(string))
	}

	if data["weekly_maintenance_start_time"].(string) != "" {
		req.WeeklyMaintenanceStartTime = aws.String(data["weekly_maintenance_start_time"].(string))
	}

	return req
}

func expandFsxWindowsConfigurationUpdate(l []interface{}) *fsx.UpdateFileSystemWindowsConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.UpdateFileSystemWindowsConfiguration{}

	if data["backup_retention"] != nil {
		req.AutomaticBackupRetentionDays = aws.Int64(int64(data["backup_retention"].(int)))
	}

	if data["daily_backup_start_time"].(string) != "" {
		req.DailyAutomaticBackupStartTime = aws.String(data["daily_backup_start_time"].(string))
	}

	if data["weekly_maintenance_start_time"].(string) != "" {
		req.WeeklyMaintenanceStartTime = aws.String(data["weekly_maintenance_start_time"].(string))
	}

	if data["self_managed_active_directory"] != nil {
		req.SelfManagedActiveDirectoryConfiguration = expandSelfManagedAdOptsUpdate(data["self_managed_active_directory"].([]interface{}))
	}

	return req
}

func expandSelfManagedAdOptsCreate(l []interface{}) *fsx.SelfManagedActiveDirectoryConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.SelfManagedActiveDirectoryConfiguration{}

	if d, ok := data["dns_ips"]; ok {
		req.DnsIps = expandStringList(d.([]interface{}))
	}

	if data["domain_name"].(string) != "" {
		req.DomainName = aws.String(data["domain_name"].(string))
	}

	if data["administrators_group"].(string) != "" {
		req.FileSystemAdministratorsGroup = aws.String(data["administrators_group"].(string))
	}

	if data["ou_distinguished_name"].(string) != "" {
		req.OrganizationalUnitDistinguishedName = aws.String(data["ou_distinguished_name"].(string))
	}

	if data["password"].(string) != "" {
		req.Password = aws.String(data["password"].(string))
	}

	if data["username"].(string) != "" {
		req.UserName = aws.String(data["username"].(string))
	}

	return req
}

func expandSelfManagedAdOptsUpdate(l []interface{}) *fsx.SelfManagedActiveDirectoryConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.SelfManagedActiveDirectoryConfigurationUpdates{}

	if d, ok := data["dns_ips"]; ok {
		req.DnsIps = expandStringList(d.([]interface{}))
	}

	if data["password"].(string) != "" {
		req.Password = aws.String(data["password"].(string))
	}

	if data["username"].(string) != "" {
		req.UserName = aws.String(data["username"].(string))
	}

	return req
}

func flattenLustreOptsConfig(lopts *fsx.LustreFileSystemConfiguration) []map[string]interface{} {
	if lopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if lopts.DataRepositoryConfiguration != nil && *lopts.DataRepositoryConfiguration.ImportPath != "" {
		m["import_path"] = aws.StringValue(lopts.DataRepositoryConfiguration.ImportPath)
	}
	if lopts.DataRepositoryConfiguration != nil && *lopts.DataRepositoryConfiguration.ExportPath != "" {
		m["export_path"] = aws.StringValue(lopts.DataRepositoryConfiguration.ExportPath)
	}
	if lopts.DataRepositoryConfiguration != nil && *lopts.DataRepositoryConfiguration.ImportedFileChunkSize != 0 {
		m["chunk_size"] = aws.Int64Value(lopts.DataRepositoryConfiguration.ImportedFileChunkSize)
	}
	if lopts.WeeklyMaintenanceStartTime != nil {
		m["weekly_maintenance_start_time"] = aws.StringValue(lopts.WeeklyMaintenanceStartTime)
	}

	return []map[string]interface{}{m}
}

func flattenWindowsOptsConfig(wopts *fsx.WindowsFileSystemConfiguration) []map[string]interface{} {
	if wopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if wopts.ActiveDirectoryId != nil {
		m["active_directory_id"] = aws.StringValue(wopts.ActiveDirectoryId)
	}
	if wopts.AutomaticBackupRetentionDays != nil {
		m["backup_retention"] = aws.Int64Value(wopts.AutomaticBackupRetentionDays)
	}
	if wopts.CopyTagsToBackups != nil {
		m["copy_tags_to_backups"] = aws.BoolValue(wopts.CopyTagsToBackups)
	}
	if wopts.DailyAutomaticBackupStartTime != nil {
		m["daily_backup_start_time"] = aws.StringValue(wopts.DailyAutomaticBackupStartTime)
	}
	if wopts.ThroughputCapacity != nil {
		m["throughput_capacity"] = aws.Int64Value(wopts.ThroughputCapacity)
	}
	if wopts.WeeklyMaintenanceStartTime != nil {
		m["weekly_maintenance_start_time"] = aws.StringValue(wopts.WeeklyMaintenanceStartTime)
	}
	if wopts.SelfManagedActiveDirectoryConfiguration != nil {
		m["self_managed_active_directory"] = flattenSelfManagedAdOptsConfig(wopts.SelfManagedActiveDirectoryConfiguration)
	}

	return []map[string]interface{}{m}
}

func flattenSelfManagedAdOptsConfig(adopts *fsx.SelfManagedActiveDirectoryAttributes) []map[string]interface{} {
	if adopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if adopts.UserName != nil {
		m["username"] = aws.StringValue(adopts.UserName)
	}
	if adopts.DnsIps != nil {
		m["dns_ips"] = aws.StringValueSlice(adopts.DnsIps)
	}
	if adopts.DomainName != nil {
		m["domain_name"] = aws.StringValue(adopts.DomainName)
	}
	if adopts.FileSystemAdministratorsGroup != nil {
		m["administrators_group"] = aws.StringValue(adopts.FileSystemAdministratorsGroup)
	}
	if adopts.OrganizationalUnitDistinguishedName != nil {
		m["ou_distinguished_name"] = aws.StringValue(adopts.OrganizationalUnitDistinguishedName)
	}

	return []map[string]interface{}{m}
}
