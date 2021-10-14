package backup

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVaultNotifications() *schema.Resource {
	return &schema.Resource{
		Create: resourceVaultNotificationsCreate,
		Read:   resourceVaultNotificationsRead,
		Delete: resourceVaultNotificationsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceVaultNotificationsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	input := &backup.PutBackupVaultNotificationsInput{
		BackupVaultName:   aws.String(d.Get("backup_vault_name").(string)),
		SNSTopicArn:       aws.String(d.Get("sns_topic_arn").(string)),
		BackupVaultEvents: flex.ExpandStringSet(d.Get("backup_vault_events").(*schema.Set)),
	}

	_, err := conn.PutBackupVaultNotifications(input)
	if err != nil {
		return fmt.Errorf("error creating Backup Vault Notifications (%s): %w", d.Id(), err)
	}

	d.SetId(d.Get("backup_vault_name").(string))

	return resourceVaultNotificationsRead(d, meta)
}

func resourceVaultNotificationsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	input := &backup.GetBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Id()),
	}

	resp, err := conn.GetBackupVaultNotifications(input)
	if tfawserr.ErrMessageContains(err, backup.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Backup Vault Notifcations %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Backup Vault Notifications (%s): %w", d.Id(), err)
	}
	d.Set("backup_vault_name", resp.BackupVaultName)
	d.Set("sns_topic_arn", resp.SNSTopicArn)
	d.Set("backup_vault_arn", resp.BackupVaultArn)
	if err := d.Set("backup_vault_events", flex.FlattenStringSet(resp.BackupVaultEvents)); err != nil {
		return fmt.Errorf("error setting backup_vault_events: %w", err)
	}

	return nil
}

func resourceVaultNotificationsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	input := &backup.DeleteBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Id()),
	}

	_, err := conn.DeleteBackupVaultNotifications(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, backup.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Backup Vault Notifications (%s): %w", d.Id(), err)
	}

	return nil
}
