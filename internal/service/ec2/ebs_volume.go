// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/datafy"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ebs_volume", name="EBS Volume")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceEBSVolume() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSVolumeCreate,
		ReadWithoutTimeout:   resourceEBSVolumeRead,
		UpdateWithoutTimeout: resourceEBSVolumeUpdate,
		DeleteWithoutTimeout: resourceEBSVolumeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			resourceEBSVolumeCustomizeDiff,
			verify.SetTagsDiff,
			func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				// once the volume is managed, datafy has control on the volume. And ONLY tags can be updated via terraform.
				changes := slices.DeleteFunc(diff.GetChangedKeysPrefix(""), func(s string) bool {
					return strings.HasPrefix(s, "tags.")
				})
				if len(changes) > 0 {
					dc := meta.(*conns.AWSClient).DatafyClient(ctx)
					if datafyVolume, datafyErr := dc.GetVolume(diff.Id()); datafyErr == nil {
						if datafyVolume.IsManaged {
							return fmt.Errorf("can't modify datafied EBS Volume (%s). Changed keys: (%s)", diff.Id(), strings.Join(changes, ","))
						}
					}
				}

				return nil
			},
		),

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrEncrypted: {
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
			names.AttrIOPS: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrKMSKeyID: {
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
			names.AttrSize: {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{names.AttrSize, names.AttrSnapshotID},
			},
			names.AttrSnapshotID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				AtLeastOneOf: []string{names.AttrSize, names.AttrSnapshotID},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrThroughput: {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(125, 1000),
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceEBSVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVolumeInput{
		AvailabilityZone:  aws.String(d.Get(names.AttrAvailabilityZone).(string)),
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeVolume),
	}

	if value, ok := d.GetOk(names.AttrEncrypted); ok {
		input.Encrypted = aws.Bool(value.(bool))
	}

	if value, ok := d.GetOk(names.AttrIOPS); ok {
		input.Iops = aws.Int32(int32(value.(int)))
	}

	if value, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(value.(string))
	}

	if value, ok := d.GetOk("multi_attach_enabled"); ok {
		input.MultiAttachEnabled = aws.Bool(value.(bool))
	}

	if value, ok := d.GetOk("outpost_arn"); ok {
		input.OutpostArn = aws.String(value.(string))
	}

	if value, ok := d.GetOk(names.AttrSize); ok {
		input.Size = aws.Int32(int32(value.(int)))
	}

	if value, ok := d.GetOk(names.AttrSnapshotID); ok {
		input.SnapshotId = aws.String(value.(string))
	}

	if value, ok := d.GetOk(names.AttrThroughput); ok {
		input.Throughput = aws.Int32(int32(value.(int)))
	}

	if value, ok := d.GetOk(names.AttrType); ok {
		input.VolumeType = awstypes.VolumeType(value.(string))
	}

	output, err := conn.CreateVolume(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EBS Volume: %s", err)
	}

	d.SetId(aws.ToString(output.VolumeId))

	if _, err := waitVolumeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EBS Volume (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEBSVolumeRead(ctx, d, meta)...)
}

func resourceEBSVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	volume, err := findEBSVolumeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		volumeId := d.Id()

		// if not found on aws, it may mean we datafied it and deleted the volume
		dc := meta.(*conns.AWSClient).DatafyClient(ctx)
		if datafyVolume, datafyErr := dc.GetVolume(volumeId); datafyErr == nil {
			// if we are managing this volume, just return the state as is after updating the tags
			if datafyVolume.IsManaged {
				dvo, err := conn.DescribeVolumes(ctx, datafy.DescribeDatafiedVolumesInput(volumeId))
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "can't find datafy volumes of EBS volume (%s): %s", volumeId, err)
				} else if len(dvo.Volumes) == 0 {
					return sdkdiag.AppendErrorf(diags, "can't find datafy volumes of EBS volume (%s)", volumeId)
				}

				setTagsOut(ctx, datafy.RemoveDatafyTags(dvo.Volumes[0].Tags))
				return diags
			}

			// if the volume was replaced (new source due to undatafy), it means the new
			// volume is now the source volume, and we need to set the "new" values from aws
			if datafyVolume.ReplacedBy != "" {
				d.SetId(datafyVolume.ReplacedBy)
				return append(
					sdkdiag.AppendWarningf(diags, "new EBS Volume (%s) has been created to replace the undatafied EBS Volume (%s)", datafyVolume.ReplacedBy, volumeId),
					resourceEBSVolumeRead(ctx, d, meta)...,
				)
			}
		} else if datafy.NotFound(datafyErr) {
			log.Printf("[WARN] EBS Volume %s not found, removing from state", volumeId)
			d.SetId("")
			return diags
		} else {
			err = datafyErr
		}
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Volume (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("volume/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())
	d.Set(names.AttrAvailabilityZone, volume.AvailabilityZone)
	d.Set(names.AttrEncrypted, volume.Encrypted)
	d.Set(names.AttrIOPS, volume.Iops)
	d.Set(names.AttrKMSKeyID, volume.KmsKeyId)
	d.Set("multi_attach_enabled", volume.MultiAttachEnabled)
	d.Set("outpost_arn", volume.OutpostArn)
	d.Set(names.AttrSize, volume.Size)
	d.Set(names.AttrSnapshotID, volume.SnapshotId)
	d.Set(names.AttrThroughput, volume.Throughput)
	d.Set(names.AttrType, volume.VolumeType)

	setTagsOut(ctx, volume.Tags)

	return diags
}

func resourceEBSVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		// once the volume is managed, datafy has control on the volume, and it can't be updated via terraform.
		// if it was replaced (new source due to undatafy), so we set the new id and the volume properties to the state
		// and give back control to terraform
		dc := meta.(*conns.AWSClient).DatafyClient(ctx)
		if datafyVolume, datafyErr := dc.GetVolume(d.Id()); datafyErr == nil {
			if datafyVolume.IsManaged {
				return sdkdiag.AppendErrorf(diags, "can't modify datafied EBS Volume (%s)", d.Id())
			}
			if datafyVolume.ReplacedBy != "" {
				diags = sdkdiag.AppendWarningf(diags, "new EBS Volume (%s) has been created to replace the undatafied EBS Volume (%s)", datafyVolume.ReplacedBy, d.Id())

				d.SetId(datafyVolume.ReplacedBy)
				if diags := resourceEBSVolumeRead(ctx, d, meta); diags.HasError() {
					return diags
				}

				return resourceEBSVolumeUpdate(ctx, d, meta)
			}
		} else if !datafy.NotFound(datafyErr) {
			return sdkdiag.AppendErrorf(diags, "modifying EBS Volume (%s): %s", d.Id(), datafyErr)
		}

		input := &ec2.ModifyVolumeInput{
			VolumeId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrIOPS) {
			input.Iops = aws.Int32(int32(d.Get(names.AttrIOPS).(int)))
		}

		if d.HasChange(names.AttrSize) {
			input.Size = aws.Int32(int32(d.Get(names.AttrSize).(int)))
		}

		// "If no throughput value is specified, the existing value is retained."
		// Not currently correct, so always specify any non-zero throughput value.
		// Throughput is valid only for gp3 volumes.
		if v := d.Get(names.AttrThroughput).(int); v > 0 && d.Get(names.AttrType).(string) == string(awstypes.VolumeTypeGp3) {
			input.Throughput = aws.Int32(int32(v))
		}

		if d.HasChange(names.AttrType) {
			volumeType := awstypes.VolumeType(d.Get(names.AttrType).(string))
			input.VolumeType = volumeType

			// Get Iops value because in the ec2.ModifyVolumeInput API,
			// if you change the volume type to io1, io2, or gp3, the default is 3,000.
			// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifyVolume.html
			if volumeType == awstypes.VolumeTypeIo1 || volumeType == awstypes.VolumeTypeIo2 || volumeType == awstypes.VolumeTypeGp3 {
				input.Iops = aws.Int32(int32(d.Get(names.AttrIOPS).(int)))
			}
		}

		_, err := conn.ModifyVolume(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EBS Volume (%s): %s", d.Id(), err)
		}

		if _, err := waitVolumeUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EBS Volume (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEBSVolumeRead(ctx, d, meta)...)
}

func resourceEBSVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	volumesIDs := []string{d.Id()}

	// once the volume is managed, datafy has control on the volume, and it can't be deleted via terraform
	// the call must go via datafy api - that will also create the snapshot if needed.
	// if it was replaced, set the new id to the state and give back control to terraform
	dc := meta.(*conns.AWSClient).DatafyClient(ctx)
	if datafyVolume, datafyErr := dc.GetVolume(d.Id()); datafyErr == nil {
		if datafyVolume.IsManaged {
			dvo, err := conn.DescribeVolumes(ctx, datafy.DescribeDatafiedVolumesInput(d.Id()))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "can't find datafy volumes of EBS volume (%s): %s", d.Id(), err)
			} else if len(dvo.Volumes) == 0 {
				return sdkdiag.AppendErrorf(diags, "can't find datafy volumes of EBS volume (%s)", d.Id())
			}

			if !datafyVolume.HasSource {
				volumesIDs = make([]string, 0, len(dvo.Volumes))
			}
			for _, volume := range dvo.Volumes {
				volumesIDs = append(volumesIDs, aws.ToString(volume.VolumeId))
			}
		}
		if datafyVolume.ReplacedBy != "" {
			diags = sdkdiag.AppendWarningf(diags, "new EBS Volume (%s) has been created to replace the undatafied EBS Volume (%s)", datafyVolume.ReplacedBy, d.Id())

			d.SetId(datafyVolume.ReplacedBy)
			return resourceEBSVolumeDelete(ctx, d, meta)
		}
	} else if !datafy.NotFound(datafyErr) {
		return sdkdiag.AppendErrorf(diags, "deleting EBS Volume (%s): %s", d.Id(), datafyErr)
	}

	if d.Get("final_snapshot").(bool) {
		input := &ec2.CreateSnapshotInput{
			TagSpecifications: tagSpecificationsFromMap(ctx, d.Get(names.AttrTagsAll).(map[string]interface{}), awstypes.ResourceTypeSnapshot),
			VolumeId:          aws.String(d.Id()),
		}

		outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 1*time.Minute,
			func() (interface{}, error) {
				return conn.CreateSnapshot(ctx, input)
			},
			errCodeSnapshotCreationPerVolumeRateExceeded, "The maximum per volume CreateSnapshot request rate has been exceeded")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EBS Snapshot (%s): %s", d.Id(), err)
		}

		snapshotID := aws.ToString(outputRaw.(*ec2.CreateSnapshotOutput).SnapshotId)

		_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete),
			func() (interface{}, error) {
				waiter := ec2.NewSnapshotCompletedWaiter(conn)
				return waiter.WaitForOutput(ctx, &ec2.DescribeSnapshotsInput{
					SnapshotIds: []string{snapshotID},
				}, d.Timeout(schema.TimeoutDelete))
			},
			errCodeResourceNotReady)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot (%s) create: %s", snapshotID, err)
		}
	}

	for _, vid := range volumesIDs {
		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete),
			func() (interface{}, error) {
				return conn.DeleteVolume(ctx, &ec2.DeleteVolumeInput{
					VolumeId: aws.String(vid),
				})
			},
			errCodeVolumeInUse,
		)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidVolumeNotFound) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting EBS Volume (%s): %s", vid, err)
		}

		if _, err := waitVolumeDeleted(ctx, conn, vid, d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EBS Volume (%s) delete: %s", d.Id(), err)
		}
		log.Printf("[DEBUG] Deleting EBS Volume: %s", vid)
	}

	return diags
}

func resourceEBSVolumeCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	iops := diff.Get(names.AttrIOPS).(int)
	multiAttachEnabled := diff.Get("multi_attach_enabled").(bool)
	throughput := diff.Get(names.AttrThroughput).(int)
	volumeType := awstypes.VolumeType(diff.Get(names.AttrType).(string))

	if diff.Id() == "" {
		// Create.

		// Iops is required for io1 and io2 volumes.
		// The default for gp3 volumes is 3,000 IOPS.
		// This parameter is not supported for gp2, st1, sc1, or standard volumes.
		// Hard validation in place to return an error if IOPs are provided
		// for an unsupported storage type.
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
		switch volumeType {
		case awstypes.VolumeTypeIo1, awstypes.VolumeTypeIo2:
			if iops == 0 {
				return fmt.Errorf("'iops' must be set when 'type' is '%s'", volumeType)
			}

		case awstypes.VolumeTypeGp3:

		default:
			if iops != 0 {
				return fmt.Errorf("'iops' must not be set when 'type' is '%s'", volumeType)
			}
		}

		// MultiAttachEnabled is supported with io1 & io2 volumes only.
		if multiAttachEnabled && volumeType != awstypes.VolumeTypeIo1 && volumeType != awstypes.VolumeTypeIo2 {
			return fmt.Errorf("'multi_attach_enabled' must not be set when 'type' is '%s'", volumeType)
		}

		// Throughput is valid only for gp3 volumes.
		if throughput > 0 && volumeType != awstypes.VolumeTypeGp3 {
			return fmt.Errorf("'throughput' must not be set when 'type' is '%s'", volumeType)
		}
	} else {
		// Update.

		// Setting 'iops = 0' is a no-op if the volume type does not require Iops to be specified.
		if diff.HasChange(names.AttrIOPS) && volumeType != awstypes.VolumeTypeIo1 && volumeType != awstypes.VolumeTypeIo2 && volumeType != awstypes.VolumeTypeGp3 && iops == 0 {
			return diff.Clear(names.AttrIOPS)
		}
	}

	return nil
}
