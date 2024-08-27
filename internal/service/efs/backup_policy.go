// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_efs_backup_policy", name="Backup Policy")
func resourceBackupPolicy() *schema.Resource {
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

	if err := putBackupPolicy(ctx, conn, fsID, d.Get("backup_policy").([]interface{})[0].(map[string]interface{})); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(fsID)

	return append(diags, resourceBackupPolicyRead(ctx, d, meta)...)
}

func resourceBackupPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	output, err := findBackupPolicyByID(ctx, conn, d.Id())

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

	if err := putBackupPolicy(ctx, conn, d.Id(), d.Get("backup_policy").([]interface{})[0].(map[string]interface{})); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceBackupPolicyRead(ctx, d, meta)...)
}

func resourceBackupPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	err := putBackupPolicy(ctx, conn, d.Id(), map[string]interface{}{
		names.AttrStatus: string(awstypes.StatusDisabled),
	})

	if errs.IsA[*awstypes.FileSystemNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func putBackupPolicy(ctx context.Context, conn *efs.Client, fsID string, tfMap map[string]interface{}) error {
	input := &efs.PutBackupPolicyInput{
		BackupPolicy: expandBackupPolicy(tfMap),
		FileSystemId: aws.String(fsID),
	}

	_, err := conn.PutBackupPolicy(ctx, input)

	if err != nil {
		return fmt.Errorf("putting EFS Backup Policy (%s): %w", fsID, err)
	}

	if input.BackupPolicy.Status == awstypes.StatusEnabled {
		if _, err := waitBackupPolicyEnabled(ctx, conn, fsID); err != nil {
			return fmt.Errorf("waiting for EFS Backup Policy (%s) enable: %w", fsID, err)
		}
	} else {
		if _, err := waitBackupPolicyDisabled(ctx, conn, fsID); err != nil {
			return fmt.Errorf("waiting for EFS Backup Policy (%s) disable: %w", fsID, err)
		}
	}

	return nil
}

func findBackupPolicyByID(ctx context.Context, conn *efs.Client, id string) (*awstypes.BackupPolicy, error) {
	input := &efs.DescribeBackupPolicyInput{
		FileSystemId: aws.String(id),
	}

	output, err := conn.DescribeBackupPolicy(ctx, input)

	if errs.IsA[*awstypes.FileSystemNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.BackupPolicy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.BackupPolicy, nil
}

func statusBackupPolicy(ctx context.Context, conn *efs.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findBackupPolicyByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitBackupPolicyEnabled(ctx context.Context, conn *efs.Client, id string) (*awstypes.BackupPolicy, error) {
	const (
		backupPoltimeoutcyEnabledTimeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusEnabling),
		Target:  enum.Slice(awstypes.StatusEnabled),
		Refresh: statusBackupPolicy(ctx, conn, id),
		Timeout: backupPoltimeoutcyEnabledTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.BackupPolicy); ok {
		return output, err
	}

	return nil, err
}

func waitBackupPolicyDisabled(ctx context.Context, conn *efs.Client, id string) (*awstypes.BackupPolicy, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusDisabling),
		Target:  enum.Slice(awstypes.StatusDisabled),
		Refresh: statusBackupPolicy(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.BackupPolicy); ok {
		return output, err
	}

	return nil, err
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

	tfMap[names.AttrStatus] = apiObject.Status

	return tfMap
}
