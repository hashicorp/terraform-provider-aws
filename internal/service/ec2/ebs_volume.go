package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceEBSVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceEBSVolumeCreate,
		Read:   resourceEBSVolumeRead,
		Update: resourceEBSVolumeUpdate,
		Delete: resourceEBSVolumeDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			resourceEBSVolumeCustomizeDiff,
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"iops": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"multi_attach_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"outpost_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"size", "snapshot_id"},
			},
			"snapshot_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				AtLeastOneOf: []string{"size", "snapshot_id"},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"throughput": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(125, 1000),
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceEBSVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateVolumeInput{
		AvailabilityZone:  aws.String(d.Get("availability_zone").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVolume),
	}

	if value, ok := d.GetOk("encrypted"); ok {
		input.Encrypted = aws.Bool(value.(bool))
	}

	if value, ok := d.GetOk("iops"); ok {
		input.Iops = aws.Int64(int64(value.(int)))
	}

	if value, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(value.(string))
	}

	if value, ok := d.GetOk("multi_attach_enabled"); ok {
		input.MultiAttachEnabled = aws.Bool(value.(bool))
	}

	if value, ok := d.GetOk("outpost_arn"); ok {
		input.OutpostArn = aws.String(value.(string))
	}

	if value, ok := d.GetOk("size"); ok {
		input.Size = aws.Int64(int64(value.(int)))
	}

	if value, ok := d.GetOk("snapshot_id"); ok {
		input.SnapshotId = aws.String(value.(string))
	}

	if value, ok := d.GetOk("throughput"); ok {
		input.Throughput = aws.Int64(int64(value.(int)))
	}

	if value, ok := d.GetOk("type"); ok {
		input.VolumeType = aws.String(value.(string))
	}

	log.Printf("[DEBUG] Creating EBS Volume: %s", input)
	output, err := conn.CreateVolume(input)

	if err != nil {
		return fmt.Errorf("creating EBS Volume: %w", err)
	}

	d.SetId(aws.StringValue(output.VolumeId))

	if _, err := WaitVolumeCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for EBS Volume (%s) create: %w", d.Id(), err)
	}

	return resourceEBSVolumeRead(d, meta)
}

func resourceEBSVolumeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	volume, err := FindEBSVolumeByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS Volume %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EBS Volume (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("volume/%s", d.Id()),
	}
	d.Set("arn", arn.String())
	d.Set("availability_zone", volume.AvailabilityZone)
	d.Set("encrypted", volume.Encrypted)
	d.Set("iops", volume.Iops)
	d.Set("kms_key_id", volume.KmsKeyId)
	d.Set("multi_attach_enabled", volume.MultiAttachEnabled)
	d.Set("outpost_arn", volume.OutpostArn)
	d.Set("size", volume.Size)
	d.Set("snapshot_id", volume.SnapshotId)
	d.Set("throughput", volume.Throughput)
	d.Set("type", volume.VolumeType)

	tags := KeyValueTags(volume.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceEBSVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyVolumeInput{
			VolumeId: aws.String(d.Id()),
		}

		if d.HasChange("iops") {
			input.Iops = aws.Int64(int64(d.Get("iops").(int)))
		}

		if d.HasChange("size") {
			input.Size = aws.Int64(int64(d.Get("size").(int)))
		}

		// "If no throughput value is specified, the existing value is retained."
		// Not currently correct, so always specify any non-zero throughput value.
		// Throughput is valid only for gp3 volumes.
		if v := d.Get("throughput").(int); v > 0 && d.Get("type").(string) == ec2.VolumeTypeGp3 {
			input.Throughput = aws.Int64(int64(v))
		}

		if d.HasChange("type") {
			volumeType := d.Get("type").(string)
			input.VolumeType = aws.String(volumeType)

			// Get Iops value because in the ec2.ModifyVolumeInput API,
			// if you change the volume type to io1, io2, or gp3, the default is 3,000.
			// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifyVolume.html
			if volumeType == ec2.VolumeTypeIo1 || volumeType == ec2.VolumeTypeIo2 || volumeType == ec2.VolumeTypeGp3 {
				input.Iops = aws.Int64(int64(d.Get("iops").(int)))
			}
		}

		_, err := conn.ModifyVolume(input)

		if err != nil {
			return fmt.Errorf("modifying EBS Volume (%s): %w", d.Id(), err)
		}

		if _, err := WaitVolumeUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for EBS Volume (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating EBS Volume (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceEBSVolumeRead(d, meta)
}

func resourceEBSVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.Get("final_snapshot").(bool) {
		input := &ec2.CreateSnapshotInput{
			TagSpecifications: tagSpecificationsFromMap(d.Get("tags_all").(map[string]interface{}), ec2.ResourceTypeSnapshot),
			VolumeId:          aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Creating EBS Snapshot: %s", input)
		outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(1*time.Minute,
			func() (interface{}, error) {
				return conn.CreateSnapshot(input)
			},
			errCodeSnapshotCreationPerVolumeRateExceeded, "The maximum per volume CreateSnapshot request rate has been exceeded")

		if err != nil {
			return fmt.Errorf("creating EBS Snapshot (%s): %w", d.Id(), err)
		}

		snapshotID := aws.StringValue(outputRaw.(*ec2.Snapshot).SnapshotId)

		_, err = tfresource.RetryWhenAWSErrCodeEquals(d.Timeout(schema.TimeoutDelete),
			func() (interface{}, error) {
				return nil, conn.WaitUntilSnapshotCompleted(&ec2.DescribeSnapshotsInput{
					SnapshotIds: aws.StringSlice([]string{snapshotID}),
				})
			},
			errCodeResourceNotReady)

		if err != nil {
			return fmt.Errorf("waiting for EBS Snapshot (%s) create: %w", snapshotID, err)
		}
	}

	log.Printf("[DEBUG] Deleting EBS Volume: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return conn.DeleteVolume(&ec2.DeleteVolumeInput{
				VolumeId: aws.String(d.Id()),
			})
		},
		errCodeVolumeInUse)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVolumeNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EBS Volume (%s): %w", d.Id(), err)
	}

	if _, err := WaitVolumeDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for EBS Volume (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func resourceEBSVolumeCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	iops := diff.Get("iops").(int)
	multiAttachEnabled := diff.Get("multi_attach_enabled").(bool)
	throughput := diff.Get("throughput").(int)
	volumeType := diff.Get("type").(string)

	if diff.Id() == "" {
		// Create.

		// Iops is required for io1 and io2 volumes.
		// The default for gp3 volumes is 3,000 IOPS.
		// This parameter is not supported for gp2, st1, sc1, or standard volumes.
		// Hard validation in place to return an error if IOPs are provided
		// for an unsupported storage type.
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
		switch volumeType {
		case ec2.VolumeTypeIo1, ec2.VolumeTypeIo2:
			if iops == 0 {
				return fmt.Errorf("'iops' must be set when 'type' is '%s'", volumeType)
			}

		case ec2.VolumeTypeGp3:

		default:
			if iops != 0 {
				return fmt.Errorf("'iops' must not be set when 'type' is '%s'", volumeType)
			}
		}

		// MultiAttachEnabled is supported with io1 & io2 volumes only.
		if multiAttachEnabled && volumeType != ec2.VolumeTypeIo1 && volumeType != ec2.VolumeTypeIo2 {
			return fmt.Errorf("'multi_attach_enabled' must not be set when 'type' is '%s'", volumeType)
		}

		// Throughput is valid only for gp3 volumes.
		if throughput > 0 && volumeType != ec2.VolumeTypeGp3 {
			return fmt.Errorf("'throughput' must not be set when 'type' is '%s'", volumeType)
		}
	} else {
		// Update.

		// Setting 'iops = 0' is a no-op if the volume type does not require Iops to be specified.
		if diff.HasChange("iops") && volumeType != ec2.VolumeTypeIo1 && volumeType != ec2.VolumeTypeIo2 && volumeType != ec2.VolumeTypeGp3 && iops == 0 {
			return diff.Clear("iops")
		}
	}

	return nil
}
