package backup

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]{1,50}$`), "must consist of lowercase letters, numbers, and hyphens."),
			},
			"sns_topic_arn": {
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
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(backup.VaultEvent_Values(), false),
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
	conn := meta.(*conns.AWSClient).BackupConn()

	input := &backup.PutBackupVaultNotificationsInput{
		BackupVaultName:   aws.String(d.Get("backup_vault_name").(string)),
		SNSTopicArn:       aws.String(d.Get("sns_topic_arn").(string)),
		BackupVaultEvents: flex.ExpandStringSet(d.Get("backup_vault_events").(*schema.Set)),
	}

	_, err := conn.PutBackupVaultNotificationsWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Vault Notifications (%s): %s", d.Id(), err)
	}

	d.SetId(d.Get("backup_vault_name").(string))

	return append(diags, resourceVaultNotificationsRead(ctx, d, meta)...)
}

func resourceVaultNotificationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn()

	input := &backup.GetBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Id()),
	}

	resp, err := conn.GetBackupVaultNotificationsWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Backup Vault Notifcations %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Vault Notifications (%s): %s", d.Id(), err)
	}
	d.Set("backup_vault_name", resp.BackupVaultName)
	d.Set("sns_topic_arn", resp.SNSTopicArn)
	d.Set("backup_vault_arn", resp.BackupVaultArn)
	if err := d.Set("backup_vault_events", flex.FlattenStringSet(resp.BackupVaultEvents)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting backup_vault_events: %s", err)
	}

	return diags
}

func resourceVaultNotificationsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn()

	input := &backup.DeleteBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Id()),
	}

	_, err := conn.DeleteBackupVaultNotificationsWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Backup Vault Notifications (%s): %s", d.Id(), err)
	}

	return diags
}
