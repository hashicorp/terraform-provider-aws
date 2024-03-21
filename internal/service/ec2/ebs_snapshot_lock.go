// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.LockMode_Values(), false),
			},
			"cool_off_period": {
				Type:         schema.TypeInt,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 72),
			},
			"lock_duration": {
				Type:         schema.TypeInt,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 36500),
			},
			"expiration_date": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsRFC3339Time,
				ForceNew:     true,
			},
			"lock_created_on": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsRFC3339Time,
				computed: 	  true,
			},
			"cool_off_period_expires_on": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsRFC3339Time,
				computed: 	  true,
			},
			"lock_duration_start_time": {
				Type:         schema.TypeString,
				ValidateFunc: validation.IsRFC3339Time,
				computed: 	  true,
			},
			"lock_state": {
				Type:         schema.TypeString,
				computed: 	  true,
				ValidateFunc: validation.StringInSlice(ec2.LockState_Values(), false),
			},
		},
	}
}

func resourceEBSSnapshotLockCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.LockSnapshotInput{
		SnapshotId: aws.String(d.Get("snapshot_id").(string)),
		LockMode:   aws.String(d.Get("lock_mode").(string)),
	}

	if v, ok := d.GetOk("cool_off_period"); ok {
		input.CoolOffPeriod = aws.Int64(v.(int64))
	}

	if v, ok := d.GetOk("lock_duration"); ok {
		input.LockDuration = aws.Int64(v.(int64))
	}

	if v, ok := d.GetOk("expiration_date"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.ExpirationDate = aws.Time(t)
	}

	resp, err := conn.LockSnapshotWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating EBS snapshot lock: %s", err)
	}

	d.Set("lock_created_on", resp.LockCreatedOn.String() )
	d.Set("cool_off_period_expires_on", resp.CoolOffPeriodExpiresOn.String())
	d.Set("lock_duration_start_time", resp.LockDurationStartTime.String())
	d.Set("lock_state", aws.StringValue(resp.LockState))

	return append(diags, resourceEBSSnapshotLockRead(ctx, d, meta)...)
}

func resourceEBSSnapshotLockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeLockedSnapshotsInput{
		SnapshotIds: aws.StringSlice([]string{
			d.Get("snapshot_id").(string),
		}),
	}

	resp, err := conn.DescribeLockedSnapshotsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "EBS snapshot lock not found: %s", err)
	}

	d.Set("snapshot_id", aws.StringValue(resp.Snapshots[0].SnapshotId))
	d.Set("lock_mode", aws.StringValue(resp.Snapshots[0].))
	d.Set("cool_off_period", aws.StringValue(resp.Snapshots[0].CoolOffPeriod))
	d.Set("lock_duration", aws.StringValue(resp.Snapshots[0].LockDuration))
	d.Set("expiration_date", aws.StringValue(resp.Snapshots[0].))
	d.Set("lock_created_on", resp.LockCreatedOn.String())
	d.Set("cool_off_period_expires_on", resp.CoolOffPeriodExpiresOn.String())
	d.Set("lock_duration_start_time", resp.LockDurationStartTime.String())
	d.Set("lock_state", aws.StringValue(resp.LockState))

	return diags
}

func resourceEBSSnapshotLockUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	input := &ec2.LockSnapshotInput{
		SnapshotId: aws.String(d.Get("snapshot_id").(string)),
		LockMode:   aws.String(d.Get("lock_mode").(string)),
	}

	if v, ok := d.GetOk("cool_off_period"); ok {
		input.CoolOffPeriod = aws.Int64(v.(int64))
	}

	if v, ok := d.GetOk("lock_duration"); ok {
		input.LockDuration = aws.Int64(v.(int64))
	}

	if v, ok := d.GetOk("expiration_date"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.ExpirationDate = aws.Time(t)
	}

	resp, err := conn.LockSnapshotWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating EBS snapshot lock: %s", err)
	}

	return diags
}

func resourceEBSSnapshotLockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.UnlockSnapshotInput{
		SnapshotId: aws.String("snapshot_id")
	}

	resp, err := conn.UnlockSnapshotWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Unlock EBS snapshot: %s", err)
	}

	return diags
}
