// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ebs_snapshot", name="EBS Snapshot")
// @Tags(identifierAttribute="id")
func ResourceEBSSnapshot() *schema.Resource {
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(append(ec2.TargetStorageTier_Values(), TargetStorageTierStandard), false),
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
			"volume_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEBSSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	volumeID := d.Get("volume_id").(string)
	input := &ec2.CreateSnapshotInput{
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeSnapshot),
		VolumeId:          aws.String(volumeID),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		input.OutpostArn = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EBS Snapshot: %s", input)
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 1*time.Minute,
		func() (interface{}, error) {
			return conn.CreateSnapshotWithContext(ctx, input)
		},
		errCodeSnapshotCreationPerVolumeRateExceeded, "The maximum per volume CreateSnapshot request rate has been exceeded")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EBS Snapshot (%s): %s", volumeID, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*ec2.Snapshot).SnapshotId))

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return nil, conn.WaitUntilSnapshotCompletedWithContext(ctx, &ec2.DescribeSnapshotsInput{
				SnapshotIds: aws.StringSlice([]string{d.Id()}),
			})
		},
		errCodeResourceNotReady)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("storage_tier"); ok && v.(string) == ec2.TargetStorageTierArchive {
		_, err = conn.ModifySnapshotTierWithContext(ctx, &ec2.ModifySnapshotTierInput{
			SnapshotId:  aws.String(d.Id()),
			StorageTier: aws.String(v.(string)),
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
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	snapshot, err := FindSnapshotByID(ctx, conn, d.Id())

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
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("snapshot/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
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

	setTagsOut(ctx, snapshot.Tags)

	return diags
}

func resourceEBSSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	if d.HasChange("storage_tier") {
		if tier := d.Get("storage_tier").(string); tier == ec2.TargetStorageTierArchive {
			_, err := conn.ModifySnapshotTierWithContext(ctx, &ec2.ModifySnapshotTierInput{
				SnapshotId:  aws.String(d.Id()),
				StorageTier: aws.String(tier),
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
				input.TemporaryRestoreDays = aws.Int64(int64(v.(int)))
			}

			//Skipping waiter as restoring a snapshot takes 24-72 hours so state will reamin (https://aws.amazon.com/blogs/aws/new-amazon-ebs-snapshots-archive/)
			_, err := conn.RestoreSnapshotTierWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "restoring EBS Snapshot (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceEBSSnapshotRead(ctx, d, meta)...)
}

func resourceEBSSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[INFO] Deleting EBS Snapshot: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteSnapshotWithContext(ctx, &ec2.DeleteSnapshotInput{
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
