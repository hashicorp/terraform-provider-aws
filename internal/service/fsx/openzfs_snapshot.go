// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_openzfs_snapshot", name="OpenZFS Snapshot")
// @Tags(identifierAttribute="arn")
func resourceOpenZFSSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOpenZFSSnapshotCreate,
		ReadWithoutTimeout:   resourceOpenZFSSnapshotRead,
		UpdateWithoutTimeout: resourceOpenZFSSnapshotUpdate,
		DeleteWithoutTimeout: resourceOpenZFSSnapshotDelete,

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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
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

func resourceOpenZFSSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := &fsx.CreateSnapshotInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		Name:               aws.String(d.Get(names.AttrName).(string)),
		Tags:               getTagsIn(ctx),
		VolumeId:           aws.String(d.Get("volume_id").(string)),
	}

	output, err := conn.CreateSnapshot(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx OpenZFS Snapshot: %s", err)
	}

	d.SetId(aws.ToString(output.Snapshot.SnapshotId))

	if _, err := waitSnapshotCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx OpenZFS Snapshot (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOpenZFSSnapshotRead(ctx, d, meta)...)
}

func resourceOpenZFSSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	snapshot, err := findSnapshotByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx Snapshot (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, snapshot.ResourceARN)
	d.Set(names.AttrCreationTime, snapshot.CreationTime.Format(time.RFC3339))
	d.Set(names.AttrName, snapshot.Name)
	d.Set("volume_id", snapshot.VolumeId)

	// Snapshot tags aren't set in the Describe response.
	// setTagsOut(ctx, snapshot.Tags)

	return diags
}

func resourceOpenZFSSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fsx.UpdateSnapshotInput{
			ClientRequestToken: aws.String(id.UniqueId()),
			SnapshotId:         aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		_, err := conn.UpdateSnapshot(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx OpenZFS Snapshot (%s): %s", d.Id(), err)
		}

		if _, err := waitSnapshotUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx OpenZFS Snapshot (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceOpenZFSSnapshotRead(ctx, d, meta)...)
}

func resourceOpenZFSSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	log.Printf("[INFO] Deleting FSx Snapshot: %s", d.Id())
	_, err := conn.DeleteSnapshot(ctx, &fsx.DeleteSnapshotInput{
		SnapshotId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.SnapshotNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx Snapshot (%s): %s", d.Id(), err)
	}

	if _, err := waitSnapshotDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx Snapshot (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findSnapshotByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.Snapshot, error) {
	input := &fsx.DescribeSnapshotsInput{
		SnapshotIds: []string{id},
	}

	return findSnapshot(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Snapshot]())
}

func findSnapshot(ctx context.Context, conn *fsx.Client, input *fsx.DescribeSnapshotsInput, filter tfslices.Predicate[*awstypes.Snapshot]) (*awstypes.Snapshot, error) {
	output, err := findSnapshots(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSnapshots(ctx context.Context, conn *fsx.Client, input *fsx.DescribeSnapshotsInput, filter tfslices.Predicate[*awstypes.Snapshot]) ([]awstypes.Snapshot, error) {
	var output []awstypes.Snapshot

	pages := fsx.NewDescribeSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.SnapshotNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Snapshots {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusSnapshot(ctx context.Context, conn *fsx.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSnapshotByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Lifecycle), nil
	}
}

func waitSnapshotCreated(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotLifecycleCreating, awstypes.SnapshotLifecyclePending),
		Target:  enum.Slice(awstypes.SnapshotLifecycleAvailable),
		Refresh: statusSnapshot(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		if output.LifecycleTransitionReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitSnapshotUpdated(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotLifecyclePending),
		Target:  enum.Slice(awstypes.SnapshotLifecycleAvailable),
		Refresh: statusSnapshot(ctx, conn, id),
		Timeout: timeout,
		Delay:   150 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		if output.LifecycleTransitionReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitSnapshotDeleted(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotLifecyclePending, awstypes.SnapshotLifecycleDeleting),
		Target:  []string{},
		Refresh: statusSnapshot(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Snapshot); ok {
		if output.LifecycleTransitionReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}
