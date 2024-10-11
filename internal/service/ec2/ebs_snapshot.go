// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ebs_snapshot", name="EBS Snapshot")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceEBSSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSSnapshotCreate,
		ReadWithoutTimeout:   resourceEBSSnapshotRead,
		UpdateWithoutTimeout: resourceEBSSnapshotUpdate,
		DeleteWithoutTimeout: resourceEBSSnapshotDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_encryption_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrKMSKeyID: {
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
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permanent_restore": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"storage_tier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(enum.Slice(append(awstypes.TargetStorageTier.Values(""), targetStorageTierStandard)...), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"temporary_restore_days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrVolumeSize: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEBSSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	volumeID := d.Get("volume_id").(string)
	input := &ec2.CreateSnapshotInput{
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeSnapshot),
		VolumeId:          aws.String(volumeID),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		input.OutpostArn = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 1*time.Minute,
		func() (interface{}, error) {
			return conn.CreateSnapshot(ctx, input)
		},
		errCodeSnapshotCreationPerVolumeRateExceeded, "The maximum per volume CreateSnapshot request rate has been exceeded")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EBS Snapshot (%s): %s", volumeID, err)
	}

	d.SetId(aws.ToString(outputRaw.(*ec2.CreateSnapshotOutput).SnapshotId))

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			waiter := ec2.NewSnapshotCompletedWaiter(conn)
			return waiter.WaitForOutput(ctx, &ec2.DescribeSnapshotsInput{
				SnapshotIds: []string{d.Id()},
			}, d.Timeout(schema.TimeoutCreate))
		},
		errCodeResourceNotReady)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("storage_tier"); ok && v.(string) == string(awstypes.TargetStorageTierArchive) {
		_, err = conn.ModifySnapshotTier(ctx, &ec2.ModifySnapshotTierInput{
			SnapshotId:  aws.String(d.Id()),
			StorageTier: awstypes.TargetStorageTier(v.(string)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EBS Snapshot (%s) Storage Tier: %s", d.Id(), err)
		}

		if _, err := waitEBSSnapshotTierArchive(ctx, conn, d.Id(), ebsSnapshotArchivedTimeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot (%s) Storage Tier archive: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEBSSnapshotRead(ctx, d, meta)...)
}

func resourceEBSSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	snapshot, err := findSnapshotByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS Snapshot %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Snapshot (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("snapshot/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("data_encryption_key_id", snapshot.DataEncryptionKeyId)
	d.Set(names.AttrDescription, snapshot.Description)
	d.Set(names.AttrEncrypted, snapshot.Encrypted)
	d.Set(names.AttrKMSKeyID, snapshot.KmsKeyId)
	d.Set("outpost_arn", snapshot.OutpostArn)
	d.Set("owner_alias", snapshot.OwnerAlias)
	d.Set(names.AttrOwnerID, snapshot.OwnerId)
	d.Set("storage_tier", snapshot.StorageTier)
	d.Set("volume_id", snapshot.VolumeId)
	d.Set(names.AttrVolumeSize, snapshot.VolumeSize)

	setTagsOut(ctx, snapshot.Tags)

	return diags
}

func resourceEBSSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange("storage_tier") {
		if tier := d.Get("storage_tier").(string); tier == string(awstypes.TargetStorageTierArchive) {
			_, err := conn.ModifySnapshotTier(ctx, &ec2.ModifySnapshotTierInput{
				SnapshotId:  aws.String(d.Id()),
				StorageTier: awstypes.TargetStorageTier(tier),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EBS Snapshot (%s) Storage Tier: %s", d.Id(), err)
			}

			if _, err := waitEBSSnapshotTierArchive(ctx, conn, d.Id(), ebsSnapshotArchivedTimeout); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot (%s) Storage Tier archive: %s", d.Id(), err)
			}
		} else {
			input := &ec2.RestoreSnapshotTierInput{
				SnapshotId: aws.String(d.Id()),
			}

			if v, ok := d.GetOk("permanent_restore"); ok {
				input.PermanentRestore = aws.Bool(v.(bool))
			}

			if v, ok := d.GetOk("temporary_restore_days"); ok {
				input.TemporaryRestoreDays = aws.Int32(int32(v.(int)))
			}

			//Skipping waiter as restoring a snapshot takes 24-72 hours so state will reamin (https://aws.amazon.com/blogs/aws/new-amazon-ebs-snapshots-archive/)
			_, err := conn.RestoreSnapshotTier(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "restoring EBS Snapshot (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceEBSSnapshotRead(ctx, d, meta)...)
}

func resourceEBSSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EBS Snapshot: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteSnapshot(ctx, &ec2.DeleteSnapshotInput{
			SnapshotId: aws.String(d.Id()),
		})
	}, errCodeInvalidSnapshotInUse)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSnapshotNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EBS Snapshot (%s): %s", d.Id(), err)
	}

	return diags
}
