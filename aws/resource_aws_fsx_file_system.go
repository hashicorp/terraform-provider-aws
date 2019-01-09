package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"LUSTRE", "WINDOWS"}, false),
			},
			"capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(300),
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"lustre_configuration": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"windows_configuration"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"import_path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"chunk_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1024,
						},
						"weekly_maintenance_start_time": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "3:03:30",
						},
					},
				},
			},
			"windows_configuration": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"lustre_configuration"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active_directory_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"backup_retention": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  7,
						},
						"copy_tags_to_backups": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"daily_backup_start_time": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "06:00",
						},
						"throughput_capacity": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"weekly_maintenance_start_time": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "3:03:30",
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
		SubnetIds:       expandStringList(d.Get("subnet_ids").([]interface{})),
		KmsKeyId:        aws.String(d.Get("kms_key_id").(string)),
	}

	if value, ok := d.GetOk("security_group_ids"); ok {
		request.SecurityGroupIds = expandStringList(value.([]interface{}))
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

	log.Println("[DEBUG] Waiting for filesystem to become available")

	stateConf := &resource.StateChangeConf{
		Pending: []string{"CREATING"},
		Target:  []string{"AVAILABLE"},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeFileSystems(&fsx.DescribeFileSystemsInput{
				FileSystemIds: []*string{aws.String(*result.FileSystem.FileSystemId)},
			})

			if err != nil {
				if ec2err, ok := err.(awserr.Error); ok {
					log.Printf("Error on FSx State Refresh: message: \"%s\", code:\"%s\"", ec2err.Message(), ec2err.Code())
					resp = nil
					return nil, "", err
				} else {
					log.Printf("Error on FSx State Refresh: %s", err)
					return nil, "", err
				}
			}

			v := resp.FileSystems[0]
			return v, *v.Lifecycle, nil
		},
		Timeout:    60 * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for filesystem (%s) to become available: %s",
			*result.FileSystem.FileSystemId, err)
	}

	d.SetId(*result.FileSystem.FileSystemId)

	return resourceAwsFsxFileSystemRead(d, meta)
}

func resourceAwsFsxFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	if _, ok := d.GetOk("tags"); ok {
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
			return err
		}
	}

	return resourceAwsFsxFileSystemRead(d, meta)
}

func resourceAwsFsxFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	request := &fsx.DescribeFileSystemsInput{
		FileSystemIds: []*string{aws.String(d.Id())},
	}

	response, err := conn.DescribeFileSystems(request)
	if err != nil {
		return fmt.Errorf("Error reading FSx filesystem %s: %s", d.Id(), err)
	}

	d.Set("type", *response.FileSystems[0].FileSystemType)
	d.Set("capacity", *response.FileSystems[0].StorageCapacity)
	d.Set("arn", *response.FileSystems[0].ResourceARN)
	d.Set("dns_name", *response.FileSystems[0].DNSName)

	d.Set("tags", tagsToMapFSX(response.FileSystems[0].Tags))

	subnets := make([]string, 0)
	for _, subnet := range response.FileSystems[0].SubnetIds {
		subnets = append(subnets, aws.StringValue(subnet))
	}

	d.Set("subnet_ids", subnets)

	if response.FileSystems[0].KmsKeyId != nil {
		d.Set("kms_key_id", *response.FileSystems[0].KmsKeyId)
	}

	if _, ok := d.GetOk("lustre_configuration"); ok {
		err = d.Set("lustre_configuration", flattenLustreOptsConfig(response.FileSystems[0].LustreConfiguration))
		if err != nil {
			return err
		}
	}

	if _, ok := d.GetOk("windows_configuration"); ok {
		err = d.Set("windows_configuration", flattenWindowsOptsConfig(response.FileSystems[0].WindowsConfiguration))
		if err != nil {
			return err
		}
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
		Pending: []string{"AVAILABLE", "DELETING"},
		Target:  []string{},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeFileSystems(&fsx.DescribeFileSystemsInput{
				FileSystemIds: []*string{aws.String(d.Id())},
			})
			if err != nil {
				efsErr, ok := err.(awserr.Error)
				if ok && efsErr.Code() == "FileSystemNotFound" {
					return nil, "", nil
				}
				return nil, "error", err
			}

			if hasEmptyFsxFileSystems(resp) {
				return nil, "", nil
			}

			v := resp.FileSystems[0]
			log.Printf("[DEBUG] current status of %q: %q", *v.FileSystemId, *v.Lifecycle)
			return v, *v.Lifecycle, nil
		},
		Timeout:    60 * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 15 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for filesystem (%s) to delete: %s", d.Id(), err)
	}

	return nil
}

func hasEmptyFsxFileSystems(fs *fsx.DescribeFileSystemsOutput) bool {
	if fs != nil && len(fs.FileSystems) > 0 {
		return false
	}
	return true
}

func expandFsxLustreConfigurationCreate(l []interface{}) *fsx.CreateFileSystemLustreConfiguration {
	data := l[0].(map[string]interface{})
	req := &fsx.CreateFileSystemLustreConfiguration{}

	if data["import_path"].(string) != "" {
		req.ImportPath = aws.String(data["import_path"].(string))
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
	data := l[0].(map[string]interface{})
	req := &fsx.UpdateFileSystemLustreConfiguration{}

	if data["weekly_maintenance_start_time"].(string) != "" {
		req.WeeklyMaintenanceStartTime = aws.String(data["weekly_maintenance_start_time"].(string))
	}

	return req
}

func expandFsxWindowsConfigurationCreate(l []interface{}) *fsx.CreateFileSystemWindowsConfiguration {
	data := l[0].(map[string]interface{})
	req := &fsx.CreateFileSystemWindowsConfiguration{
		ThroughputCapacity: aws.Int64(int64(data["throughput_capacity"].(int))),
	}

	if data["active_directory_id"].(string) != "" {
		req.ActiveDirectoryId = aws.String(data["active_directory_id"].(string))
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

	return req
}

func flattenLustreOptsConfig(lopts *fsx.LustreFileSystemConfiguration) []map[string]interface{} {
	if lopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"import_path":                   aws.StringValue(lopts.DataRepositoryConfiguration.ImportPath),
		"chunk_size":                    aws.Int64Value(lopts.DataRepositoryConfiguration.ImportedFileChunkSize),
		"weekly_maintenance_start_time": aws.StringValue(lopts.WeeklyMaintenanceStartTime),
	}

	return []map[string]interface{}{m}
}

func flattenWindowsOptsConfig(wopts *fsx.WindowsFileSystemConfiguration) []map[string]interface{} {
	if wopts == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"active_directory_id":           aws.StringValue(wopts.ActiveDirectoryId),
		"backup_retention":              aws.Int64Value(wopts.AutomaticBackupRetentionDays),
		"copy_tags_to_backups":          aws.BoolValue(wopts.CopyTagsToBackups),
		"daily_backup_start_time":       aws.StringValue(wopts.DailyAutomaticBackupStartTime),
		"throughput_capacity":           aws.Int64Value(wopts.ThroughputCapacity),
		"weekly_maintenance_start_time": aws.StringValue(wopts.WeeklyMaintenanceStartTime),
	}

	return []map[string]interface{}{m}
}
