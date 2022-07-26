package backup

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVaultPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceVaultPolicyPut,
		Update: resourceVaultPolicyPut,
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
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceVaultPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
	}

	name := d.Get("backup_vault_name").(string)
	input := &backup.PutBackupVaultAccessPolicyInput{
		BackupVaultName: aws.String(name),
		Policy:          aws.String(policy),
	}

	_, err = conn.PutBackupVaultAccessPolicy(input)

	if err != nil {
		return fmt.Errorf("error creating Backup Vault Policy (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceVaultPolicyRead(d, meta)
}

func resourceVaultPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	output, err := FindVaultAccessPolicyByName(conn, d.Id())

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

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.StringValue(output.Policy))

	if err != nil {
		return fmt.Errorf("while setting policy (%s), encountered: %w", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policyToSet, err)
	}

	d.Set("policy", policyToSet)

	return nil
}

func resourceVaultPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	log.Printf("[DEBUG] Deleting Backup Vault Policy (%s)", d.Id())
	_, err := conn.DeleteBackupVaultAccessPolicy(&backup.DeleteBackupVaultAccessPolicyInput{
		BackupVaultName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) || tfawserr.ErrCodeEquals(err, errCodeAccessDeniedException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Backup Vault Policy (%s): %w", d.Id(), err)
	}

	return nil
}
