package backup

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVault() *schema.Resource {
	return &schema.Resource{
		Create: resourceVaultCreate,
		Read:   resourceVaultRead,
		Update: resourceVaultUpdate,
		Delete: resourceVaultDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 50),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_]*$`), "must consist of letters, numbers, and hyphens."),
				),
			},
			"recovery_points": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVaultCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &backup.CreateBackupVaultInput{
		BackupVaultName: aws.String(name),
		BackupVaultTags: Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.EncryptionKeyArn = aws.String(v.(string))
	}

	_, err := conn.CreateBackupVault(input)

	if err != nil {
		return fmt.Errorf("creating Backup Vault (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceVaultRead(d, meta)
}

func resourceVaultRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindVaultByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Vault (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Backup Vault (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.BackupVaultArn)
	d.Set("kms_key_arn", output.EncryptionKeyArn)
	d.Set("name", output.BackupVaultName)
	d.Set("recovery_points", output.NumberOfRecoveryPoints)

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return fmt.Errorf("listing tags for Backup Vault (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceVaultUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("updating tags for Backup Vault (%s): %w", d.Id(), err)
		}
	}

	return resourceVaultRead(d, meta)
}

func resourceVaultDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	if d.Get("force_destroy").(bool) {
		input := &backup.ListRecoveryPointsByBackupVaultInput{
			BackupVaultName: aws.String(d.Id()),
		}
		var recoveryPointErrs *multierror.Error

		err := conn.ListRecoveryPointsByBackupVaultPages(input, func(page *backup.ListRecoveryPointsByBackupVaultOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, v := range page.RecoveryPoints {
				recoveryPointARN := aws.StringValue(v.RecoveryPointArn)

				log.Printf("[DEBUG] Deleting Backup Vault recovery point: %s", recoveryPointARN)
				_, err := conn.DeleteRecoveryPoint(&backup.DeleteRecoveryPointInput{
					BackupVaultName:  aws.String(d.Id()),
					RecoveryPointArn: aws.String(recoveryPointARN),
				})

				if err != nil {
					recoveryPointErrs = multierror.Append(recoveryPointErrs, fmt.Errorf("deleting Backup Vault (%s) recovery point (%s): %w", d.Id(), recoveryPointARN, err))
					continue
				}

				if _, err := waitRecoveryPointDeleted(conn, d.Id(), recoveryPointARN, d.Timeout(schema.TimeoutDelete)); err != nil {
					recoveryPointErrs = multierror.Append(recoveryPointErrs, fmt.Errorf("waiting for Backup Vault (%s) recovery point (%s) delete: %w", d.Id(), recoveryPointARN, err))
				}
			}

			return !lastPage
		})

		if err != nil {
			return fmt.Errorf("listing Backup Vault (%s) recovery points: %w", d.Id(), err)
		}

		if err := recoveryPointErrs.ErrorOrNil(); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Deleting Backup Vault: %s", d.Id())
	_, err := conn.DeleteBackupVault(&backup.DeleteBackupVaultInput{
		BackupVaultName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Backup Vault (%s): %w", d.Id(), err)
	}

	return nil
}
