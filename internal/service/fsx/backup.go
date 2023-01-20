package fsx

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBackup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBackupCreate,
		ReadWithoutTimeout:   resourceBackupRead,
		UpdateWithoutTimeout: resourceBackupUpdate,
		DeleteWithoutTimeout: resourceBackupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchemaComputed(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),
	}
}

func resourceBackupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateBackupInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("file_system_id"); ok {
		input.FileSystemId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("volume_id"); ok {
		input.VolumeId = aws.String(v.(string))
	}

	if input.FileSystemId == nil && input.VolumeId == nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx Backup: %s", "must specify either file_system_id or volume_id")
	}

	if input.FileSystemId != nil && input.VolumeId != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx Backup: %s", "can only specify either file_system_id or volume_id")
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	result, err := conn.CreateBackupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx Backup: %s", err)
	}

	d.SetId(aws.StringValue(result.Backup.BackupId))

	log.Println("[DEBUG] Waiting for FSx backup to become available")
	if _, err := waitBackupAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx Backup (%s) to be available: %s", d.Id(), err)
	}

	return append(diags, resourceBackupRead(ctx, d, meta)...)
}

func resourceBackupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx Backup (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceBackupRead(ctx, d, meta)...)
}

func resourceBackupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	backup, err := FindBackupByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Backup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx Backup (%s): %s", d.Id(), err)
	}

	d.Set("arn", backup.ResourceARN)
	d.Set("type", backup.Type)

	if backup.FileSystem != nil {
		fs := backup.FileSystem
		d.Set("file_system_id", fs.FileSystemId)
	}

	d.Set("kms_key_id", backup.KmsKeyId)

	d.Set("owner_id", backup.OwnerId)

	if backup.Volume != nil {
		d.Set("volume_id", backup.Volume.VolumeId)
	}

	tags := KeyValueTags(backup.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceBackupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn()

	request := &fsx.DeleteBackupInput{
		BackupId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting FSx Backup: %s", d.Id())
	_, err := conn.DeleteBackupWithContext(ctx, request)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, fsx.ErrCodeBackupNotFound) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting FSx Backup (%s): %s", d.Id(), err)
	}

	log.Println("[DEBUG] Waiting for backup to delete")
	if _, err := waitBackupDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx Backup (%s) to deleted: %s", d.Id(), err)
	}

	return diags
}
