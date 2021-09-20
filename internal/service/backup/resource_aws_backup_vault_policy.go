package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfbackup "github.com/hashicorp/terraform-provider-aws/aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/backup/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVaultPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBackupVaultPolicyPut,
		Update: resourceAwsBackupVaultPolicyPut,
		Read:   resourceVaultPolicyRead,
		Delete: resourceVaultPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"backup_vault_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backup_vault_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
			},
		},
	}
}

func resourceAwsBackupVaultPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	name := d.Get("backup_vault_name").(string)
	input := &backup.PutBackupVaultAccessPolicyInput{
		BackupVaultName: aws.String(name),
		Policy:          aws.String(d.Get("policy").(string)),
	}

	_, err := conn.PutBackupVaultAccessPolicy(input)

	if err != nil {
		return fmt.Errorf("error creating Backup Vault Policy (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceVaultPolicyRead(d, meta)
}

func resourceVaultPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	output, err := finder.FindBackupVaultAccessPolicyByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Vault Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Backup Vault Policy (%s): %w", d.Id(), err)
	}

	d.Set("backup_vault_arn", output.BackupVaultArn)
	d.Set("backup_vault_name", output.BackupVaultName)
	d.Set("policy", output.Policy)

	return nil
}

func resourceVaultPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	log.Printf("[DEBUG] Deleting Backup Vault Policy (%s)", d.Id())
	_, err := conn.DeleteBackupVaultAccessPolicy(&backup.DeleteBackupVaultAccessPolicyInput{
		BackupVaultName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) || tfawserr.ErrCodeEquals(err, tfbackup.errCodeAccessDeniedException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Backup Vault Policy (%s): %w", d.Id(), err)
	}

	return nil
}
