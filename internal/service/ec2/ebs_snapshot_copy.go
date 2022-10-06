package ec2

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEBSSnapshotCopy() *schema.Resource {
	return &schema.Resource{
		Create: resourceEBSSnapshotCopyCreate,
		Read:   resourceEBSSnapshotRead,
		Update: resourceEBSSnapshotUpdate,
		Delete: resourceEBSSnapshotDelete,

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
				Optional: true,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"source_region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_snapshot_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_tier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(append(ec2.TargetStorageTier_Values(), TargetStorageTierStandard), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"temporary_restore_days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEBSSnapshotCopyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CopySnapshotInput{
		SourceRegion:      aws.String(d.Get("source_region").(string)),
		SourceSnapshotId:  aws.String(d.Get("source_snapshot_id").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeSnapshot),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encrypted"); ok {
		input.Encrypted = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	output, err := conn.CopySnapshot(input)

	if err != nil {
		return fmt.Errorf("creating EBS Snapshot Copy: %w", err)
	}

	d.SetId(aws.StringValue(output.SnapshotId))

	_, err = tfresource.RetryWhenAWSErrCodeEquals(d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return nil, conn.WaitUntilSnapshotCompleted(&ec2.DescribeSnapshotsInput{
				SnapshotIds: aws.StringSlice([]string{d.Id()}),
			})
		},
		errCodeResourceNotReady)

	if err != nil {
		return fmt.Errorf("waiting for EBS Snapshot Copy (%s) create: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("storage_tier"); ok && v.(string) == ec2.TargetStorageTierArchive {
		_, err = conn.ModifySnapshotTier(&ec2.ModifySnapshotTierInput{
			SnapshotId:  aws.String(d.Id()),
			StorageTier: aws.String(v.(string)),
		})

		if err != nil {
			return fmt.Errorf("setting EBS Snapshot Copy (%s) Storage Tier: %w", d.Id(), err)
		}

		_, err = waitEBSSnapshotTierArchive(conn, d.Id(), ebsSnapshotArchivedTimeout)

		if err != nil {
			return fmt.Errorf("waiting for EBS Snapshot Copy (%s) Storage Tier archive: %w", d.Id(), err)
		}
	}

	return resourceEBSSnapshotRead(d, meta)
}
