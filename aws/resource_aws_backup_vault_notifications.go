package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsBackupVaultNotifications() *schema.Resource {

	return &schema.Resource{
		Create: resourceAwsBackupVaultNotificationsCreate,
		Read:   resourceAwsBackupVaultNotificationsRead,
		Delete: resourceAwsBackupVaultNotificationsDelete,

		Schema: map[string]*schema.Schema{
			"vault_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"events": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						backup.VaultEventBackupJobStarted,
						backup.VaultEventBackupJobCompleted,
						backup.VaultEventBackupJobSuccessful,
						backup.VaultEventBackupJobFailed,
						backup.VaultEventBackupJobExpired,
						backup.VaultEventRestoreJobStarted,
						backup.VaultEventRestoreJobCompleted,
						backup.VaultEventRestoreJobSuccessful,
						backup.VaultEventRestoreJobFailed,
						backup.VaultEventCopyJobStarted,
						backup.VaultEventCopyJobSuccessful,
						backup.VaultEventCopyJobFailed,
						backup.VaultEventRecoveryPointModified,
						backup.VaultEventBackupPlanCreated,
						backup.VaultEventBackupPlanModified,
					}, false),
				},
			},
		},
	}
}

func resourceAwsBackupVaultNotificationsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.GetBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Id()),
	}

	resp, err := conn.GetBackupVaultNotifications(input)

	if err != nil {
		return fmt.Errorf("error reading Backup Vault Notifications (%s): %s", d.Id(), err)
	}

	d.Set("vault_name", resp.BackupVaultName)
	d.Set("sns_topic_arn", resp.SNSTopicArn)
	d.Set("events", resp.BackupVaultEvents)

	return nil
}

func resourceAwsBackupVaultNotificationsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn
	input := &backup.PutBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Get("vault_name").(string)),
	}

	if v, ok := d.GetOk("sns_topic_arn"); ok {
		input.SNSTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("events"); ok {
		input.BackupVaultEvents = expandStringList(v.(*schema.Set).List())
	}

	_, err := conn.PutBackupVaultNotifications(input)

	if err != nil {
		return fmt.Errorf("error creating Backup Vault Notifications(%s): %s", d.Id(), err)
	}

	d.SetId(d.Get("vault_name").(string))

	return resourceAwsBackupVaultNotificationsRead(d, meta)
}

func resourceAwsBackupVaultNotificationsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.DeleteBackupVaultNotificationsInput{
		BackupVaultName: aws.String(d.Get("vault_name").(string)),
	}

	_, err := conn.DeleteBackupVaultNotifications(input)

	if err != nil {
		return fmt.Errorf("error deleting Backup Vault Notifications (%s): %s", d.Id(), err)
	}

	return nil
}
