// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_backup_vault_notifications")
func ResourceVaultNotifications() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVaultNotificationsCreate,
		ReadWithoutTimeout:   resourceVaultNotificationsRead,
		DeleteWithoutTimeout: resourceVaultNotificationsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"backup_vault_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]{1,50}$`), "must consist of lowercase letters, numbers, and hyphens."),
			},
			names.AttrSNSTopicARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"backup_vault_events": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.BackupVaultEvent](),
				},
			},
			"backup_vault_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVaultNotificationsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	input := &backup.PutBackupVaultNotificationsInput{
		BackupVaultName:   aws.String(d.Get("backup_vault_name").(string)),
		SNSTopicArn:       aws.String(d.Get(names.AttrSNSTopicARN).(string)),
		BackupVaultEvents: flex.ExpandStringyValueSet[awstypes.BackupVaultEvent](d.Get("backup_vault_events").(*schema.Set)),
	}

	_, err := conn.PutBackupVaultNotifications(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Vault Notifications (%s): %s", d.Id(), err)
	}

	d.SetId(d.Get("backup_vault_name").(string))

	return append(diags, resourceVaultNotificationsRead(ctx, d, meta)...)
}

func resourceVaultNotificationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	input := &backup.GetBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Id()),
	}

	resp, err := conn.GetBackupVaultNotifications(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Backup Vault Notifcations %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Vault Notifications (%s): %s", d.Id(), err)
	}
	d.Set("backup_vault_name", resp.BackupVaultName)
	d.Set(names.AttrSNSTopicARN, resp.SNSTopicArn)
	d.Set("backup_vault_arn", resp.BackupVaultArn)
	if err := d.Set("backup_vault_events", flex.FlattenStringyValueSet(resp.BackupVaultEvents)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting backup_vault_events: %s", err)
	}

	return diags
}

func resourceVaultNotificationsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	input := &backup.DeleteBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Id()),
	}

	_, err := conn.DeleteBackupVaultNotifications(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Backup Vault Notifications (%s): %s", d.Id(), err)
	}

	return diags
}
