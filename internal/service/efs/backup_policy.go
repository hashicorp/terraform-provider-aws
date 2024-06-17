// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_efs_backup_policy")
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
						names.AttrStatus: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(enum.Slice(
								awstypes.StatusDisabled,
								awstypes.StatusEnabled,
							), false),
						},
					},
				},
			},

			names.AttrFileSystemID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBackupPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	fsID := d.Get(names.AttrFileSystemID).(string)

	if err := backupPolicyPut(ctx, conn, fsID, d.Get("backup_policy").([]interface{})[0].(map[string]interface{})); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EFS Backup Policy (%s): %s", fsID, err)
	}

	d.SetId(fsID)

	return append(diags, resourceBackupPolicyRead(ctx, d, meta)...)
}

func resourceBackupPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

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

	d.Set(names.AttrFileSystemID, d.Id())

	return diags
}

func resourceBackupPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	if err := backupPolicyPut(ctx, conn, d.Id(), d.Get("backup_policy").([]interface{})[0].(map[string]interface{})); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EFS Backup Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBackupPolicyRead(ctx, d, meta)...)
}

func resourceBackupPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	err := backupPolicyPut(ctx, conn, d.Id(), map[string]interface{}{
		names.AttrStatus: awstypes.StatusDisabled,
	})

	if errs.IsA[*awstypes.FileSystemNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EFS Backup Policy (%s): %s", d.Id(), err)
	}

	return diags
}

// backupPolicyPut attempts to update the file system's backup policy.
// Any error is returned.
func backupPolicyPut(ctx context.Context, conn *efs.Client, fsID string, tfMap map[string]interface{}) error {
	input := &efs.PutBackupPolicyInput{
		BackupPolicy: expandBackupPolicy(tfMap),
		FileSystemId: aws.String(fsID),
	}

	log.Printf("[DEBUG] Putting EFS Backup Policy: %+v", input)
	_, err := conn.PutBackupPolicy(ctx, input)

	if err != nil {
		return fmt.Errorf("putting EFS Backup Policy (%s): %w", fsID, err)
	}

	if input.BackupPolicy.Status == awstypes.StatusEnabled {
		if _, err := waitBackupPolicyEnabled(ctx, conn, fsID); err != nil {
			return fmt.Errorf("waiting for EFS Backup Policy (%s) to enable: %w", fsID, err)
		}
	} else {
		if _, err := waitBackupPolicyDisabled(ctx, conn, fsID); err != nil {
			return fmt.Errorf("waiting for EFS Backup Policy (%s) to disable: %w", fsID, err)
		}
	}

	return nil
}

func expandBackupPolicy(tfMap map[string]interface{}) *awstypes.BackupPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.BackupPolicy{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = awstypes.Status(v)
	}

	return apiObject
}

func flattenBackupPolicy(apiObject *awstypes.BackupPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrStatus] = string(apiObject.Status)

	return tfMap
}
