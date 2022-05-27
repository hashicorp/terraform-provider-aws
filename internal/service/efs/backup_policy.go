package efs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceBackupPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceBackupPolicyCreate,
		Read:   resourceBackupPolicyRead,
		Update: resourceBackupPolicyUpdate,
		Delete: resourceBackupPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"backup_policy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								efs.StatusDisabled,
								efs.StatusEnabled,
							}, false),
						},
					},
				},
			},

			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBackupPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	fsID := d.Get("file_system_id").(string)

	if err := backupPolicyPut(conn, fsID, d.Get("backup_policy").([]interface{})[0].(map[string]interface{})); err != nil {
		return err
	}

	d.SetId(fsID)

	return resourceBackupPolicyRead(d, meta)
}

func resourceBackupPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	output, err := FindBackupPolicyByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS Backup Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EFS Backup Policy (%s): %w", d.Id(), err)
	}

	if err := d.Set("backup_policy", []interface{}{flattenBackupPolicy(output)}); err != nil {
		return fmt.Errorf("error setting backup_policy: %w", err)
	}

	d.Set("file_system_id", d.Id())

	return nil
}

func resourceBackupPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	if err := backupPolicyPut(conn, d.Id(), d.Get("backup_policy").([]interface{})[0].(map[string]interface{})); err != nil {
		return err
	}

	return resourceBackupPolicyRead(d, meta)
}

func resourceBackupPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EFSConn

	err := backupPolicyPut(conn, d.Id(), map[string]interface{}{
		"status": efs.StatusDisabled,
	})

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

// backupPolicyPut attempts to update the file system's backup policy.
// Any error is returned.
func backupPolicyPut(conn *efs.EFS, fsID string, tfMap map[string]interface{}) error {
	input := &efs.PutBackupPolicyInput{
		BackupPolicy: expandBackupPolicy(tfMap),
		FileSystemId: aws.String(fsID),
	}

	log.Printf("[DEBUG] Putting EFS Backup Policy: %s", input)
	_, err := conn.PutBackupPolicy(input)

	if err != nil {
		return fmt.Errorf("error putting EFS Backup Policy (%s): %w", fsID, err)
	}

	if aws.StringValue(input.BackupPolicy.Status) == efs.StatusEnabled {
		if _, err := waitBackupPolicyEnabled(conn, fsID); err != nil {
			return fmt.Errorf("error waiting for EFS Backup Policy (%s) to enable: %w", fsID, err)
		}
	} else {
		if _, err := waitBackupPolicyDisabled(conn, fsID); err != nil {
			return fmt.Errorf("error waiting for EFS Backup Policy (%s) to disable: %w", fsID, err)
		}
	}

	return nil
}

func expandBackupPolicy(tfMap map[string]interface{}) *efs.BackupPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &efs.BackupPolicy{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		apiObject.Status = aws.String(v)
	}

	return apiObject
}

func flattenBackupPolicy(apiObject *efs.BackupPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	return tfMap
}
