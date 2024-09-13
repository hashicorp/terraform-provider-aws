// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_snapshot_create_volume_permission", name="EBS Snapshot CreateVolume Permission")
func resourceSnapshotCreateVolumePermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotCreateVolumePermissionCreate,
		ReadWithoutTimeout:   resourceSnapshotCreateVolumePermissionRead,
		DeleteWithoutTimeout: resourceSnapshotCreateVolumePermissionDelete,

		CustomizeDiff: resourceSnapshotCreateVolumePermissionCustomizeDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrSnapshotID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSnapshotCreateVolumePermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	snapshotID := d.Get(names.AttrSnapshotID).(string)
	accountID := d.Get(names.AttrAccountID).(string)
	id := ebsSnapshotCreateVolumePermissionCreateResourceID(snapshotID, accountID)
	input := &ec2.ModifySnapshotAttributeInput{
		Attribute: awstypes.SnapshotAttributeNameCreateVolumePermission,
		CreateVolumePermission: &awstypes.CreateVolumePermissionModifications{
			Add: []awstypes.CreateVolumePermission{
				{UserId: aws.String(accountID)},
			},
		},
		SnapshotId: aws.String(snapshotID),
	}

	_, err := conn.ModifySnapshotAttribute(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EBS Snapshot CreateVolumePermission (%s): %s", id, err)
	}

	d.SetId(id)

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return findCreateSnapshotCreateVolumePermissionByTwoPartKey(ctx, conn, snapshotID, accountID)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot CreateVolumePermission create (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceSnapshotCreateVolumePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	snapshotID, accountID, err := ebsSnapshotCreateVolumePermissionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findCreateSnapshotCreateVolumePermissionByTwoPartKey(ctx, conn, snapshotID, accountID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS Snapshot CreateVolumePermission %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Snapshot CreateVolumePermission (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceSnapshotCreateVolumePermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	snapshotID, accountID, err := ebsSnapshotCreateVolumePermissionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EBS Snapshot CreateVolumePermission: %s", d.Id())
	_, err = conn.ModifySnapshotAttribute(ctx, &ec2.ModifySnapshotAttributeInput{
		Attribute: awstypes.SnapshotAttributeNameCreateVolumePermission,
		CreateVolumePermission: &awstypes.CreateVolumePermissionModifications{
			Remove: []awstypes.CreateVolumePermission{
				{UserId: aws.String(accountID)},
			},
		},
		SnapshotId: aws.String(snapshotID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSnapshotNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EBS Snapshot CreateVolumePermission (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return findCreateSnapshotCreateVolumePermissionByTwoPartKey(ctx, conn, snapshotID, accountID)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot CreateVolumePermission delete (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceSnapshotCreateVolumePermissionCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if diff.Id() == "" {
		if snapshotID := diff.Get(names.AttrSnapshotID).(string); snapshotID != "" {
			conn := meta.(*conns.AWSClient).EC2Client(ctx)

			snapshot, err := findSnapshotByID(ctx, conn, snapshotID)

			if err != nil {
				return fmt.Errorf("reading EBS Snapshot (%s): %w", snapshotID, err)
			}

			if accountID := diff.Get(names.AttrAccountID).(string); aws.ToString(snapshot.OwnerId) == accountID {
				return fmt.Errorf("AWS Account (%s) owns EBS Snapshot (%s)", accountID, snapshotID)
			}
		}
	}

	return nil
}

const ebsSnapshotCreateVolumePermissionIDSeparator = "-"

func ebsSnapshotCreateVolumePermissionCreateResourceID(snapshotID, accountID string) string {
	parts := []string{snapshotID, accountID}
	id := strings.Join(parts, ebsSnapshotCreateVolumePermissionIDSeparator)

	return id
}

func ebsSnapshotCreateVolumePermissionParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, ebsSnapshotCreateVolumePermissionIDSeparator, 3)

	if len(parts) != 3 || parts[0] != "snap" || parts[1] == "" || parts[2] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected SNAPSHOT_ID%[2]sACCOUNT_ID", id, ebsSnapshotCreateVolumePermissionIDSeparator)
	}

	return strings.Join([]string{parts[0], parts[1]}, ebsSnapshotCreateVolumePermissionIDSeparator), parts[2], nil
}
