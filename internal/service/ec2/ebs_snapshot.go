package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEBSSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceEBSSnapshotCreate,
		Read:   resourceEBSSnapshotRead,
		Update: resourceEBSSnapshotUpdate,
		Delete: resourceEBSSnapshotDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_encryption_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outpost_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"owner_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permanent_restore": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"storage_tier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.Any(
					validation.StringInSlice(ec2.TargetStorageTier_Values(), false),
					validation.StringInSlice([]string{"standard"}, false), //Enum slice does not include `standard` type.
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"temporary_restore_days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"volume_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEBSSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	request := &ec2.CreateSnapshotInput{
		VolumeId:          aws.String(d.Get("volume_id").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeSnapshot),
	}
	if v, ok := d.GetOk("description"); ok {
		request.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		request.OutpostArn = aws.String(v.(string))
	}

	var res *ec2.Snapshot
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		res, err = conn.CreateSnapshot(request)

		if tfawserr.ErrMessageContains(err, "SnapshotCreationPerVolumeRateExceeded", "The maximum per volume CreateSnapshot request rate has been exceeded") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		res, err = conn.CreateSnapshot(request)
	}
	if err != nil {
		return fmt.Errorf("error creating EBS Snapshot: %w", err)
	}

	d.SetId(aws.StringValue(res.SnapshotId))

	err = resourceEBSSnapshotWaitForAvailable(d, conn)
	if err != nil {
		return err
	}

	if v, ok := d.GetOk("storage_tier"); ok && v.(string) == ec2.TargetStorageTierArchive {
		_, err = conn.ModifySnapshotTier(&ec2.ModifySnapshotTierInput{
			SnapshotId:  aws.String(d.Id()),
			StorageTier: aws.String(v.(string)),
		})

		if err != nil {
			return fmt.Errorf("error setting EBS Snapshot (%s) Storage Tier: %w", d.Id(), err)
		}

		_, err = WaitEBSSnapshotTierArchive(conn, d.Id())
		if err != nil {
			return fmt.Errorf("Error waiting for EBS Snapshot (%s) Storage Tier to be archived: %w", d.Id(), err)
		}
	}

	return resourceEBSSnapshotRead(d, meta)
}

func resourceEBSSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	snapshot, err := FindSnapshotById(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS Snapshot (%s) Not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EBS Snapshot (%s): %w", d.Id(), err)
	}

	d.Set("data_encryption_key_id", snapshot.DataEncryptionKeyId)
	d.Set("description", snapshot.Description)
	d.Set("encrypted", snapshot.Encrypted)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("outpost_arn", snapshot.OutpostArn)
	d.Set("owner_alias", snapshot.OwnerAlias)
	d.Set("owner_id", snapshot.OwnerId)
	d.Set("storage_tier", snapshot.StorageTier)
	d.Set("volume_id", snapshot.VolumeId)
	d.Set("volume_size", snapshot.VolumeSize)

	tags := KeyValueTags(snapshot.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	snapshotArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("snapshot/%s", d.Id()),
		Service:   ec2.ServiceName,
	}.String()

	d.Set("arn", snapshotArn)

	return nil
}

func resourceEBSSnapshotUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("storage_tier") {
		tier := d.Get("storage_tier").(string)
		if tier == ec2.TargetStorageTierArchive {
			_, err := conn.ModifySnapshotTier(&ec2.ModifySnapshotTierInput{
				SnapshotId:  aws.String(d.Id()),
				StorageTier: aws.String(tier),
			})

			if err != nil {
				return fmt.Errorf("error upadating EBS Snapshot (%s) Storage Tier: %w", d.Id(), err)
			}

			_, err = WaitEBSSnapshotTierArchive(conn, d.Id())
			if err != nil {
				return fmt.Errorf("Error waiting for EBS Snapshot (%s) Storage Tier to be archived: %w", d.Id(), err)
			}
		} else {
			input := &ec2.RestoreSnapshotTierInput{
				SnapshotId: aws.String(d.Id()),
			}

			if v, ok := d.GetOk("permanent_restore"); ok {
				input.PermanentRestore = aws.Bool(v.(bool))
			}

			if v, ok := d.GetOk("temporary_restore_days"); ok {
				input.TemporaryRestoreDays = aws.Int64(int64(v.(int)))
			}

			//Skipping waiter as restoring a snapshot takes 24-72 hours so state will reamin (https://aws.amazon.com/blogs/aws/new-amazon-ebs-snapshots-archive/)
			_, err := conn.RestoreSnapshotTier(input)

			if err != nil {
				return fmt.Errorf("error restoring EBS Snapshot (%s): %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceEBSSnapshotRead(d, meta)
}

func resourceEBSSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting EBS Snapshot: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteSnapshot(&ec2.DeleteSnapshotInput{
			SnapshotId: aws.String(d.Id()),
		})
	}, errCodeInvalidSnapshotInUse)

	if err != nil {
		return fmt.Errorf("error deleting EBS Snapshot (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceEBSSnapshotWaitForAvailable(d *schema.ResourceData, conn *ec2.EC2) error {
	log.Printf("Waiting for Snapshot %s to become available...", d.Id())
	input := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{aws.String(d.Id())},
	}
	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		err := conn.WaitUntilSnapshotCompleted(input)
		if err == nil {
			return nil
		}
		if tfawserr.ErrCodeEquals(err, "ResourceNotReady") {
			return resource.RetryableError(fmt.Errorf("EBS Snapshot - waiting for snapshot to become available"))
		}
		return resource.NonRetryableError(err)
	})
	if tfresource.TimedOut(err) {
		err = conn.WaitUntilSnapshotCompleted(input)
	}
	if err != nil {
		return fmt.Errorf("Error waiting for EBS snapshot to complete: %w", err)
	}
	return nil
}
