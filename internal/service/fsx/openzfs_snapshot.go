// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"creation_time": {
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
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	input := &fsx.CreateSnapshotInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		Name:               aws.String(d.Get(names.AttrName).(string)),
		Tags:               getTagsIn(ctx),
		VolumeId:           aws.String(d.Get("volume_id").(string)),
	}

	output, err := conn.CreateSnapshotWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx OpenZFS Snapshot: %s", err)
	}

	d.SetId(aws.StringValue(output.Snapshot.SnapshotId))

	if _, err := waitSnapshotCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx OpenZFS Snapshot (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOpenZFSSnapshotRead(ctx, d, meta)...)
}

func resourceOpenZFSSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

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
	d.Set("creation_time", snapshot.CreationTime.Format(time.RFC3339))
	d.Set(names.AttrName, snapshot.Name)
	d.Set("volume_id", snapshot.VolumeId)

	// Snapshot tags aren't set in the Describe response.
	// setTagsOut(ctx, snapshot.Tags)

	return diags
}

func resourceOpenZFSSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fsx.UpdateSnapshotInput{
			ClientRequestToken: aws.String(id.UniqueId()),
			SnapshotId:         aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		_, err := conn.UpdateSnapshotWithContext(ctx, input)

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
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	log.Printf("[INFO] Deleting FSx Snapshot: %s", d.Id())
	_, err := conn.DeleteSnapshotWithContext(ctx, &fsx.DeleteSnapshotInput{
		SnapshotId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeSnapshotNotFound) {
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

func findSnapshotByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.Snapshot, error) {
	input := &fsx.DescribeSnapshotsInput{
		SnapshotIds: aws.StringSlice([]string{id}),
	}

	return findSnapshot(ctx, conn, input, tfslices.PredicateTrue[*fsx.Snapshot]())
}

func findSnapshot(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeSnapshotsInput, filter tfslices.Predicate[*fsx.Snapshot]) (*fsx.Snapshot, error) {
	output, err := findSnapshots(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findSnapshots(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeSnapshotsInput, filter tfslices.Predicate[*fsx.Snapshot]) ([]*fsx.Snapshot, error) {
	var output []*fsx.Snapshot

	err := conn.DescribeSnapshotsPagesWithContext(ctx, input, func(page *fsx.DescribeSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Snapshots {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeSnapshotNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusSnapshot(ctx context.Context, conn *fsx.FSx, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSnapshotByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Lifecycle), nil
	}
}

func waitSnapshotCreated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.SnapshotLifecycleCreating, fsx.SnapshotLifecyclePending},
		Target:  []string{fsx.SnapshotLifecycleAvailable},
		Refresh: statusSnapshot(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.Snapshot); ok {
		if output.LifecycleTransitionReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitSnapshotUpdated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.SnapshotLifecyclePending},
		Target:  []string{fsx.SnapshotLifecycleAvailable},
		Refresh: statusSnapshot(ctx, conn, id),
		Timeout: timeout,
		Delay:   150 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.Snapshot); ok {
		if output.LifecycleTransitionReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitSnapshotDeleted(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.SnapshotLifecyclePending, fsx.SnapshotLifecycleDeleting},
		Target:  []string{},
		Refresh: statusSnapshot(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.Snapshot); ok {
		if output.LifecycleTransitionReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}
