package fsx

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOpenzfsSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenzfsSnapshotCreate,
		Read:   resourceOpenzfsSnapshotRead,
		Update: resourceOpenzfsSnapshotUpdate,
		Delete: resourceOpenzfsSnapshotDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 203),
			},
			"tags":     tftags.TagsSchemaComputed(),
			"tags_all": tftags.TagsSchemaComputed(),
			"volume_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(23, 23),
			},
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),
	}
}

func resourceOpenzfsSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateSnapshotInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		Name:               aws.String(d.Get("name").(string)),
		VolumeId:           aws.String(d.Get("volume_id").(string)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	result, err := conn.CreateSnapshot(input)
	if err != nil {
		return fmt.Errorf("error creating FSx OpenZFS Snapshot: %w", err)
	}

	d.SetId(aws.StringValue(result.Snapshot.SnapshotId))

	log.Println("[DEBUG] Waiting for FSx OpenZFS Snapshot to become available")
	if _, err := waitSnapshotCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for FSx OpenZFS Snapshot (%s) to be available: %w", d.Id(), err)
	}

	return resourceOpenzfsSnapshotRead(d, meta)
}

func resourceOpenzfsSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	snapshot, err := FindSnapshotByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FSx Snapshot (%s): %w", d.Id(), err)
	}

	d.Set("arn", snapshot.ResourceARN)
	d.Set("volume_id", snapshot.VolumeId)
	d.Set("name", snapshot.Name)

	if err := d.Set("creation_time", snapshot.CreationTime.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting creation_time: %w", err)
	}

	//Snapshot tags do not get returned with describe call so need to make a separate list tags call
	tags, tagserr := ListTags(conn, *snapshot.ResourceARN)

	if tagserr != nil {
		return fmt.Errorf("error reading Tags for FSx OpenZFS Snapshot (%s): %w", d.Id(), err)
	} else {
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	}
	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceOpenzfsSnapshotUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Snapshot (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	if d.HasChangesExcept("tags_all", "tags") {
		input := &fsx.UpdateSnapshotInput{
			ClientRequestToken: aws.String(resource.UniqueId()),
			SnapshotId:         aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		_, err := conn.UpdateSnapshot(input)

		if err != nil {
			return fmt.Errorf("error updating FSx OpenZFS Snapshot (%s): %w", d.Id(), err)
		}

		if _, err := waitSnapshotUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for FSx OpenZFS Snapshot (%s) update: %w", d.Id(), err)
		}
	}

	return resourceOpenzfsSnapshotRead(d, meta)
}

func resourceOpenzfsSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	request := &fsx.DeleteSnapshotInput{
		SnapshotId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting FSx Snapshot: %s", d.Id())
	_, err := conn.DeleteSnapshot(request)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, fsx.ErrCodeSnapshotNotFound) {
			return nil
		}
		return fmt.Errorf("error deleting FSx Snapshot (%s): %w", d.Id(), err)
	}

	log.Println("[DEBUG] Waiting for snapshot to delete")
	if _, err := waitSnapshotDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for FSx Snapshot (%s) to deleted: %w", d.Id(), err)
	}

	return nil
}
