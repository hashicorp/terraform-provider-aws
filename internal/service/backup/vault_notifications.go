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
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_backup_vault_notifications", name="Vault Notifications")
func resourceVaultNotifications() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVaultNotificationsCreate,
		ReadWithoutTimeout:   resourceVaultNotificationsRead,
		DeleteWithoutTimeout: resourceVaultNotificationsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"backup_vault_arn": {
				Type:     schema.TypeString,
				Computed: true,
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
		},
	}
}

func resourceVaultNotificationsCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	name := d.Get("backup_vault_name").(string)
	input := &backup.PutBackupVaultNotificationsInput{
		BackupVaultEvents: flex.ExpandStringyValueSet[awstypes.BackupVaultEvent](d.Get("backup_vault_events").(*schema.Set)),
		BackupVaultName:   aws.String(name),
		SNSTopicArn:       aws.String(d.Get(names.AttrSNSTopicARN).(string)),
	}

	_, err := conn.PutBackupVaultNotifications(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Vault Notifications (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceVaultNotificationsRead(ctx, d, meta)...)
}

func resourceVaultNotificationsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	output, err := findVaultNotificationsByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Vault Notifications (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Vault Notifications (%s): %s", d.Id(), err)
	}

	d.Set("backup_vault_arn", output.BackupVaultArn)
	d.Set("backup_vault_events", output.BackupVaultEvents)
	d.Set("backup_vault_name", output.BackupVaultName)
	d.Set(names.AttrSNSTopicARN, output.SNSTopicArn)

	return diags
}

func resourceVaultNotificationsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	log.Printf("[DEBUG] Deleting Backup Vault Notifications: %s", d.Id())
	input := backup.DeleteBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Id()),
	}
	_, err := conn.DeleteBackupVaultNotifications(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Vault Notifications (%s): %s", d.Id(), err)
	}

	return diags
}

func findVaultNotificationsByName(ctx context.Context, conn *backup.Client, name string) (*backup.GetBackupVaultNotificationsOutput, error) {
	input := &backup.GetBackupVaultNotificationsInput{
		BackupVaultName: aws.String(name),
	}

	return findVaultNotifications(ctx, conn, input)
}

func findVaultNotifications(ctx context.Context, conn *backup.Client, input *backup.GetBackupVaultNotificationsInput) (*backup.GetBackupVaultNotificationsOutput, error) {
	output, err := conn.GetBackupVaultNotifications(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || tfawserr.ErrCodeEquals(err, errCodeAccessDeniedException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
