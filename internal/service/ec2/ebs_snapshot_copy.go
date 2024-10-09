// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

// @SDKResource("aws_ebs_snapshot_copy", name="EBS Snapshot Copy")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceEBSSnapshotCopy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSSnapshotCopyCreate,
		ReadWithoutTimeout:   resourceEBSSnapshotRead,
		UpdateWithoutTimeout: resourceEBSSnapshotUpdate,
		DeleteWithoutTimeout: resourceEBSSnapshotDelete,

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
				Optional: true,
				ForceNew: true,
			},
			names.AttrKMSKeyID: {
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
			names.AttrOwnerID: {
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
				Computed: true,
			},
			names.AttrVolumeSize: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEBSSnapshotCopyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CopySnapshotInput{
		SourceRegion:      aws.String(d.Get("source_region").(string)),
		SourceSnapshotId:  aws.String(d.Get("source_snapshot_id").(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeSnapshot),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEncrypted); ok {
		input.Encrypted = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	output, err := conn.CopySnapshot(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EBS Snapshot Copy: %s", err)
	}

	d.SetId(aws.ToString(output.SnapshotId))

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			waiter := ec2.NewSnapshotCompletedWaiter(conn)
			return waiter.WaitForOutput(ctx, &ec2.DescribeSnapshotsInput{
				SnapshotIds: []string{d.Id()},
			}, d.Timeout(schema.TimeoutCreate))
		},
		errCodeResourceNotReady)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot Copy (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("storage_tier"); ok && v.(string) == string(awstypes.TargetStorageTierArchive) {
		_, err = conn.ModifySnapshotTier(ctx, &ec2.ModifySnapshotTierInput{
			SnapshotId:  aws.String(d.Id()),
			StorageTier: awstypes.TargetStorageTier(v.(string)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting EBS Snapshot Copy (%s) Storage Tier: %s", d.Id(), err)
		}

		_, err = waitEBSSnapshotTierArchive(ctx, conn, d.Id(), ebsSnapshotArchivedTimeout)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot Copy (%s) Storage Tier archive: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEBSSnapshotRead(ctx, d, meta)...)
}
