package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsBackupVault() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBackupVaultCreate,
		Read:   resourceAwsBackupVaultRead,
		Update: resourceAwsBackupVaultUpdate,
		Delete: resourceAwsBackupVaultDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]{1,50}$`), "must consist of lowercase letters, numbers, and hyphens."),
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recovery_points": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsBackupVaultCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.CreateBackupVaultInput{
		BackupVaultName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("tags"); ok {
		input.BackupVaultTags = tagsFromMapGeneric(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.EncryptionKeyArn = aws.String(v.(string))
	}

	_, err := conn.CreateBackupVault(input)
	if err != nil {
		return fmt.Errorf("error creating Backup Vault (%s): %s", d.Id(), err)
	}

	d.SetId(d.Get("name").(string))

	return resourceAwsBackupVaultRead(d, meta)
}

func resourceAwsBackupVaultRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.DescribeBackupVaultInput{
		BackupVaultName: aws.String(d.Id()),
	}

	resp, err := conn.DescribeBackupVault(input)
	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Backup Vault %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading Backup Vault (%s): %s", d.Id(), err)
	}
	d.Set("name", resp.BackupVaultName)
	d.Set("kms_key_arn", resp.EncryptionKeyArn)
	d.Set("arn", resp.BackupVaultArn)
	d.Set("recovery_points", resp.NumberOfRecoveryPoints)

	tresp, err := conn.ListTags(&backup.ListTagsInput{
		ResourceArn: aws.String(*resp.BackupVaultArn),
	})

	if err != nil {
		return fmt.Errorf("error retrieving Backup Vault (%s) tags: %s", aws.StringValue(resp.BackupVaultArn), err)
	}

	if err := d.Set("tags", tagsToMapGeneric(tresp.Tags)); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsBackupVaultUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	if d.HasChange("tags") {
		resourceArn := d.Get("arn").(string)
		oraw, nraw := d.GetChange("tags")
		create, remove := diffTagsGeneric(oraw.(map[string]interface{}), nraw.(map[string]interface{}))

		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			keys := make([]*string, 0, len(remove))
			for k := range remove {
				keys = append(keys, aws.String(k))
			}

			_, err := conn.UntagResource(&backup.UntagResourceInput{
				ResourceArn: aws.String(resourceArn),
				TagKeyList:  keys,
			})
			if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
				log.Printf("[WARN] Backup Vault %s not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			if err != nil {
				return fmt.Errorf("Error removing tags for (%s): %s", d.Id(), err)
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagResource(&backup.TagResourceInput{
				ResourceArn: aws.String(resourceArn),
				Tags:        create,
			})
			if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
				log.Printf("[WARN] Backup Vault %s not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			if err != nil {
				return fmt.Errorf("Error setting tags for (%s): %s", d.Id(), err)
			}
		}
	}

	return resourceAwsBackupVaultRead(d, meta)
}

func resourceAwsBackupVaultDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.DeleteBackupVaultInput{
		BackupVaultName: aws.String(d.Get("name").(string)),
	}

	_, err := conn.DeleteBackupVault(input)
	if err != nil {
		return fmt.Errorf("error deleting Backup Vault (%s): %s", d.Id(), err)
	}

	return nil
}
