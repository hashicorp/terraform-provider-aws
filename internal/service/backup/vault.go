package backup

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVault() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVaultCreate,
		ReadWithoutTimeout:   resourceVaultRead,
		UpdateWithoutTimeout: resourceVaultUpdate,
		DeleteWithoutTimeout: resourceVaultDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceVaultCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn()
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

	_, err := conn.CreateBackupVaultWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Backup Vault (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceVaultRead(ctx, d, meta)...)
}

func resourceVaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindVaultByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Backup Vault (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Vault (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.BackupVaultArn)
	d.Set("kms_key_arn", output.EncryptionKeyArn)
	d.Set("name", output.BackupVaultName)
	d.Set("recovery_points", output.NumberOfRecoveryPoints)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Backup Vault (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceVaultUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags for Backup Vault (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVaultRead(ctx, d, meta)...)
}

func resourceVaultDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn()

	if d.Get("force_destroy").(bool) {
		input := &backup.ListRecoveryPointsByBackupVaultInput{
			BackupVaultName: aws.String(d.Id()),
		}
		var recoveryPointErrs *multierror.Error

		err := conn.ListRecoveryPointsByBackupVaultPagesWithContext(ctx, input, func(page *backup.ListRecoveryPointsByBackupVaultOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, v := range page.RecoveryPoints {
				recoveryPointARN := aws.StringValue(v.RecoveryPointArn)

				log.Printf("[DEBUG] Deleting Backup Vault recovery point: %s", recoveryPointARN)
				_, err := conn.DeleteRecoveryPointWithContext(ctx, &backup.DeleteRecoveryPointInput{
					BackupVaultName:  aws.String(d.Id()),
					RecoveryPointArn: aws.String(recoveryPointARN),
				})

				if err != nil {
					recoveryPointErrs = multierror.Append(recoveryPointErrs, fmt.Errorf("deleting recovery point (%s): %w", recoveryPointARN, err))
					continue
				}

				if _, err := waitRecoveryPointDeleted(ctx, conn, d.Id(), recoveryPointARN, d.Timeout(schema.TimeoutDelete)); err != nil {
					recoveryPointErrs = multierror.Append(recoveryPointErrs, fmt.Errorf("deleting recovery point (%s): waiting for completion: %w", recoveryPointARN, err))
				}
			}

			return !lastPage
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Backup Vault (%s) recovery points: %s", d.Id(), err)
		}

		if err := recoveryPointErrs.ErrorOrNil(); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Backup Vault (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting Backup Vault: %s", d.Id())
	_, err := conn.DeleteBackupVaultWithContext(ctx, &backup.DeleteBackupVaultInput{
		BackupVaultName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Backup Vault (%s): %s", d.Id(), err)
	}

	return diags
}
