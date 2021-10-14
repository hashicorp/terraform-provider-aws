package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/fsx/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/fsx/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBackup() *schema.Resource {
	return &schema.Resource{
		Create: resourceBackupCreate,
		Read:   resourceBackupRead,
		Update: resourceBackupUpdate,
		Delete: resourceBackupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Required: true,
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
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),
	}
}

func resourceBackupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateBackupInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemId:       aws.String(d.Get("file_system_id").(string)),
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().FsxTags()
	}

	result, err := conn.CreateBackup(input)
	if err != nil {
		return fmt.Errorf("error creating FSx Backup: %w", err)
	}

	d.SetId(aws.StringValue(result.Backup.BackupId))

	log.Println("[DEBUG] Waiting for FSx backup to become available")
	if _, err := waiter.BackupAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for FSx Backup (%s) to be available: %w", d.Id(), err)
	}

	return resourceBackupRead(d, meta)
}

func resourceBackupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.FsxUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Backup (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceBackupRead(d, meta)
}

func resourceBackupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	backup, err := finder.BackupByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Backup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FSx Backup (%s): %w", d.Id(), err)
	}

	d.Set("arn", backup.ResourceARN)
	d.Set("type", backup.Type)

	fs := backup.FileSystem
	d.Set("file_system_id", fs.FileSystemId)

	d.Set("kms_key_id", backup.KmsKeyId)

	d.Set("owner_id", backup.OwnerId)

	tags := tftags.FsxKeyValueTags(backup.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceBackupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	request := &fsx.DeleteBackupInput{
		BackupId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting FSx Backup: %s", d.Id())
	_, err := conn.DeleteBackup(request)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, fsx.ErrCodeBackupNotFound) {
			return nil
		}
		return fmt.Errorf("error deleting FSx Backup (%s): %w", d.Id(), err)
	}

	log.Println("[DEBUG] Waiting for backup to delete")
	if _, err := waiter.BackupDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for FSx Backup (%s) to deleted: %w", d.Id(), err)
	}

	return nil
}
