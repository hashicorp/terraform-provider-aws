package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/efs/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/efs/waiter"
)

func resourceAwsEfsFileSystemBackupPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEfsFileSystemBackupPolicyPut,
		Read:   resourceAwsEfsFileSystemBackupPolicyRead,
		Delete: resourceAwsEfsFileSystemBackupPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"backup_policy": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								efs.StatusEnabled,
							}, false),
						},
					},
				},
			},
		},
	}
}

func resourceAwsEfsFileSystemBackupPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	fsId := d.Get("file_system_id").(string)
	input := &efs.PutBackupPolicyInput{
		FileSystemId: aws.String(fsId),
	}

	if v, ok := d.GetOk("backup_policy"); ok {
		input.BackupPolicy = expandEfsFileSystemBackupPolicy(v.([]interface{}))
	}

	log.Printf("[DEBUG] Adding EFS File System Backup Policy: %#v", input)
	_, err := conn.PutBackupPolicy(input)
	if err != nil {
		return fmt.Errorf("error creating EFS File System Backup Policy %q: %s", fsId, err.Error())
	}

	d.SetId(fsId)

	if _, err := waiter.FileSystemBackupPolicyCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EFS File System Backup Policy (%q) creation : %s", d.Id(), err)
	}

	return resourceAwsEfsFileSystemBackupPolicyRead(d, meta)
}

func resourceAwsEfsFileSystemBackupPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	bp, err := finder.FileSystemBackupPolicyById(conn, d.Id())
	if err != nil {
		if isAWSErr(err, efs.ErrCodeFileSystemNotFound, "") {
			log.Printf("[WARN] EFS File System (%q) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if isAWSErr(err, efs.ErrCodePolicyNotFound, "") {
			log.Printf("[WARN] EFS File System Backup Policy (%q) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error describing policy for EFS File System Backup Policy (%q): %s", d.Id(), err)
	}

	if bp == nil || aws.StringValue(bp.Status) == efs.StatusDisabled {
		log.Printf("[WARN] EFS File System Backup Policy (%q) is disabled, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("file_system_id", d.Id())
	if err := d.Set("backup_policy", flattenEfsFileSystemBackupPolicy(bp)); err != nil {
		return fmt.Errorf("error setting backup_policy: %s", err)
	}

	return nil
}

func resourceAwsEfsFileSystemBackupPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	log.Printf("[DEBUG] Deleting EFS File System Backup Policy: %s", d.Id())
	_, err := conn.PutBackupPolicy(&efs.PutBackupPolicyInput{
		FileSystemId: aws.String(d.Id()),
		BackupPolicy: &efs.BackupPolicy{
			Status: aws.String(efs.StatusDisabled),
		},
	})

	if err != nil {
		return fmt.Errorf("error deleting EFS File System Backup Policy: %s with err %s", d.Id(), err.Error())
	}

	if _, err := waiter.FileSystemBackupPolicyDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EFS File System Backup Policy (%q) deletion : %s", d.Id(), err)
	}

	log.Printf("[DEBUG] EFS File System Backup Policy %q deleted.", d.Id())

	return nil
}

func expandEfsFileSystemBackupPolicy(tfList []interface{}) *efs.BackupPolicy {
	return &efs.BackupPolicy{
		Status: aws.String(tfList[0].(map[string]interface{})["status"].(string)),
	}
}

func flattenEfsFileSystemBackupPolicy(apiObjects *efs.BackupPolicy) []interface{} {
	return []interface{}{map[string]interface{}{"status": apiObjects.Status}}
}
