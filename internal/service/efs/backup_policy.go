package efs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceBackupPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBackupPolicyCreate,
		ReadWithoutTimeout:   resourceBackupPolicyRead,
		UpdateWithoutTimeout: resourceBackupPolicyUpdate,
		DeleteWithoutTimeout: resourceBackupPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceBackupPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()

	fsID := d.Get("file_system_id").(string)

	if err := backupPolicyPut(ctx, conn, fsID, d.Get("backup_policy").([]interface{})[0].(map[string]interface{})); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EFS Backup Policy (%s): %s", fsID, err)
	}

	d.SetId(fsID)

	return append(diags, resourceBackupPolicyRead(ctx, d, meta)...)
}

func resourceBackupPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()

	output, err := FindBackupPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS Backup Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Backup Policy (%s): %s", d.Id(), err)
	}

	if err := d.Set("backup_policy", []interface{}{flattenBackupPolicy(output)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting backup_policy: %s", err)
	}

	d.Set("file_system_id", d.Id())

	return diags
}

func resourceBackupPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()

	if err := backupPolicyPut(ctx, conn, d.Id(), d.Get("backup_policy").([]interface{})[0].(map[string]interface{})); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EFS Backup Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBackupPolicyRead(ctx, d, meta)...)
}

func resourceBackupPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()

	err := backupPolicyPut(ctx, conn, d.Id(), map[string]interface{}{
		"status": efs.StatusDisabled,
	})

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EFS Backup Policy (%s): %s", d.Id(), err)
	}

	return diags
}

// backupPolicyPut attempts to update the file system's backup policy.
// Any error is returned.
func backupPolicyPut(ctx context.Context, conn *efs.EFS, fsID string, tfMap map[string]interface{}) error {
	input := &efs.PutBackupPolicyInput{
		BackupPolicy: expandBackupPolicy(tfMap),
		FileSystemId: aws.String(fsID),
	}

	log.Printf("[DEBUG] Putting EFS Backup Policy: %s", input)
	_, err := conn.PutBackupPolicyWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("error putting EFS Backup Policy (%s): %w", fsID, err)
	}

	if aws.StringValue(input.BackupPolicy.Status) == efs.StatusEnabled {
		if _, err := waitBackupPolicyEnabled(ctx, conn, fsID); err != nil {
			return fmt.Errorf("error waiting for EFS Backup Policy (%s) to enable: %w", fsID, err)
		}
	} else {
		if _, err := waitBackupPolicyDisabled(ctx, conn, fsID); err != nil {
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
