// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_openzfs_snapshot", name="OpenZFS Snapshot")
// @Tags(identifierAttribute="arn")
func ResourceOpenzfsSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOpenzfsSnapshotCreate,
		ReadWithoutTimeout:   resourceOpenzfsSnapshotRead,
		UpdateWithoutTimeout: resourceOpenzfsSnapshotUpdate,
		DeleteWithoutTimeout: resourceOpenzfsSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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

func resourceOpenzfsSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	input := &fsx.CreateSnapshotInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		Name:               aws.String(d.Get("name").(string)),
		Tags:               getTagsIn(ctx),
		VolumeId:           aws.String(d.Get("volume_id").(string)),
	}

	result, err := conn.CreateSnapshotWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx OpenZFS Snapshot: %s", err)
	}

	d.SetId(aws.StringValue(result.Snapshot.SnapshotId))

	log.Println("[DEBUG] Waiting for FSx OpenZFS Snapshot to become available")
	if _, err := waitSnapshotCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx OpenZFS Snapshot (%s) to be available: %s", d.Id(), err)
	}

	return append(diags, resourceOpenzfsSnapshotRead(ctx, d, meta)...)
}

func resourceOpenzfsSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	snapshot, err := FindSnapshotByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx Snapshot (%s): %s", d.Id(), err)
	}

	d.Set("arn", snapshot.ResourceARN)
	d.Set("volume_id", snapshot.VolumeId)
	d.Set("name", snapshot.Name)

	if err := d.Set("creation_time", snapshot.CreationTime.Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting creation_time: %s", err)
	}

	return diags
}

func resourceOpenzfsSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	if d.HasChangesExcept("tags_all", "tags") {
		input := &fsx.UpdateSnapshotInput{
			ClientRequestToken: aws.String(id.UniqueId()),
			SnapshotId:         aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		_, err := conn.UpdateSnapshotWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx OpenZFS Snapshot (%s): %s", d.Id(), err)
		}

		if _, err := waitSnapshotUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx OpenZFS Snapshot (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceOpenzfsSnapshotRead(ctx, d, meta)...)
}

func resourceOpenzfsSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	request := &fsx.DeleteSnapshotInput{
		SnapshotId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting FSx Snapshot: %s", d.Id())
	_, err := conn.DeleteSnapshotWithContext(ctx, request)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, fsx.ErrCodeSnapshotNotFound) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting FSx Snapshot (%s): %s", d.Id(), err)
	}

	log.Println("[DEBUG] Waiting for snapshot to delete")
	if _, err := waitSnapshotDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx Snapshot (%s) to deleted: %s", d.Id(), err)
	}

	return diags
}
