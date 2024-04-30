// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_ebs_snapshot_lock", name="EBS Snapshot Lock")
func ResourceEBSSnapshotLock() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSSnapshotLockCreate,
		ReadWithoutTimeout:   resourceEBSSnapshotLockRead,
		UpdateWithoutTimeout: resourceEBSSnapshotLockUpdate,
		DeleteWithoutTimeout: resourceEBSSnapshotLockDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"snapshot_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"lock_mode": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.LockMode](),
			},
			"cool_off_period": {
				Type:             schema.TypeInt,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 72)),
			},
			"lock_duration": {
				Type:             schema.TypeInt,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 36500)),
			},
			"expiration_date": {
				Type:             schema.TypeString,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
				ForceNew:         true,
			},
			"lock_created_on": {
				Type:             schema.TypeString,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
				Computed:         true,
			},
			"cool_off_period_expires_on": {
				Type:             schema.TypeString,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
				Computed:         true,
			},
			"lock_duration_start_time": {
				Type:             schema.TypeString,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
				Computed:         true,
			},
			"lock_state": {
				Type:             schema.TypeString,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.LockState](),
			},
		},
	}
}

func resourceEBSSnapshotLockCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.LockSnapshotInput{
		SnapshotId: aws.String(d.Get("snapshot_id").(string)),
		LockMode:   types.LockMode(d.Get("lock_mode").(string)),
	}

	if v, ok := d.GetOk("cool_off_period"); ok {
		input.CoolOffPeriod = aws.Int32(v.(int32))
	}

	if v, ok := d.GetOk("lock_duration"); ok {
		input.LockDuration = aws.Int32(v.(int32))
	}

	if v, ok := d.GetOk("expiration_date"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.ExpirationDate = aws.Time(t)
	}

	resp, err := conn.LockSnapshot(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating EBS snapshot lock: %s", err)
	}

	d.Set("lock_created_on", resp.LockCreatedOn)
	d.Set("cool_off_period_expires_on", resp.CoolOffPeriodExpiresOn)
	d.Set("lock_duration_start_time", resp.LockDurationStartTime)
	d.Set("lock_state", resp.LockState)

	return append(diags, resourceEBSSnapshotLockRead(ctx, d, meta)...)
}

func resourceEBSSnapshotLockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeLockedSnapshotsInput{
		SnapshotIds: []string{
			d.Get("snapshot_id").(string),
		},
	}

	resp, err := conn.DescribeLockedSnapshots(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "EBS snapshot lock not found: %s", err)
	}

	d.Set("cool_off_period", resp.Snapshots[0].CoolOffPeriod)
	d.Set("cool_off_period_expires_on", resp.Snapshots[0].CoolOffPeriodExpiresOn)
	d.Set("lock_created_on", resp.Snapshots[0].LockCreatedOn)
	d.Set("lock_duration", resp.Snapshots[0].LockDuration)
	d.Set("lock_duration_start_time", resp.Snapshots[0].LockDurationStartTime)
	d.Set("expiration_date", resp.Snapshots[0].LockExpiresOn)
	d.Set("lock_state", resp.Snapshots[0].LockState)
	d.Set("snapshot_id", resp.Snapshots[0].SnapshotId)

	return diags
}

func resourceEBSSnapshotLockUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.LockSnapshotInput{
		SnapshotId: aws.String(d.Get("snapshot_id").(string)),
		LockMode:   types.LockMode(d.Get("lock_mode").(string)),
	}

	if v, ok := d.GetOk("cool_off_period"); ok {
		input.CoolOffPeriod = aws.Int32(v.(int32))
	}

	if v, ok := d.GetOk("lock_duration"); ok {
		input.LockDuration = aws.Int32(v.(int32))
	}

	if v, ok := d.GetOk("expiration_date"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.ExpirationDate = aws.Time(t)
	}

	resp, err := conn.LockSnapshot(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating EBS snapshot lock: %s", err)
	}

	d.Set("lock_created_on", resp.LockCreatedOn)
	d.Set("cool_off_period_expires_on", resp.CoolOffPeriodExpiresOn)
	d.Set("lock_duration_start_time", resp.LockDurationStartTime)
	d.Set("lock_state", resp.LockState)

	return diags
}

func resourceEBSSnapshotLockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.UnlockSnapshotInput{
		SnapshotId: aws.String(d.Get("snapshot_id").(string)),
	}

	_, err := conn.UnlockSnapshot(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Unlock EBS snapshot: %s", err)
	}

	return diags
}
