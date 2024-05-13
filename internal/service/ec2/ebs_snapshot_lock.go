// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

		CustomizeDiff: customdiff.Sequence(
			resourcEBSSnapshotLockCustomizeDiff,
		),

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
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 72)),
			},
			"lock_duration": {
				Type:             schema.TypeInt,
				Optional:         true,
				ConflictsWith:    []string{"expiration_date"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 36500)),
			},
			"expiration_date": {
				Type:             schema.TypeString,
				Optional:         true,
				ConflictsWith:    []string{"lock_duration"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsRFC3339Time),
			},
			"lock_created_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cool_off_period_expires_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lock_duration_start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lock_state": {
				Type:     schema.TypeString,
				Computed: true,
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
		return sdkdiag.AppendErrorf(diags, "creating EBS snapshot lock: %s", err)
	}

	d.SetId(aws.ToString(resp.SnapshotId))
	d.Set("lock_created_on", aws.ToTime(resp.LockCreatedOn).Format(time.RFC3339))
	d.Set("cool_off_period_expires_on", aws.ToTime(resp.CoolOffPeriodExpiresOn).Format(time.RFC3339))
	d.Set("lock_duration_start_time", aws.ToTime(resp.LockDurationStartTime).Format(time.RFC3339))
	d.Set("lock_state", resp.LockState)

	return append(diags, resourceEBSSnapshotLockRead(ctx, d, meta)...)
}

func resourceEBSSnapshotLockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	resp, err := FindSnapshotLockByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS snapshot lock %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS snapshot lock (%s): %s", d.Id(), err)
	}

	d.Set("cool_off_period", resp.Snapshots[0].CoolOffPeriod)
	d.Set("cool_off_period_expires_on", aws.ToTime(resp.Snapshots[0].CoolOffPeriodExpiresOn).Format(time.RFC3339))
	d.Set("lock_created_on", aws.ToTime(resp.Snapshots[0].LockCreatedOn).Format(time.RFC3339))
	d.Set("lock_duration", resp.Snapshots[0].LockDuration)
	d.Set("lock_duration_start_time", aws.ToTime(resp.Snapshots[0].LockDurationStartTime).Format(time.RFC3339))
	d.Set("expiration_date", aws.ToTime(resp.Snapshots[0].LockExpiresOn).Format(time.RFC3339))
	d.Set("lock_state", resp.Snapshots[0].LockState)
	d.Set("snapshot_id", resp.Snapshots[0].SnapshotId)

	return diags
}

func resourceEBSSnapshotLockUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.LockSnapshotInput{
		SnapshotId: aws.String(d.Id()),
	}

	if d.HasChange("cool_off_period") {
		input.CoolOffPeriod = aws.Int32(d.Get("cool_off_period").(int32))
	}

	if d.HasChange("lock_duration") {
		input.LockDuration = aws.Int32(d.Get("lock_duration").(int32))
	}

	if d.HasChange("expiration_date") {
		t, _ := time.Parse(time.RFC3339, d.Get("expiration_date").(string))
		input.ExpirationDate = aws.Time(t)
	}

	resp, err := conn.LockSnapshot(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EBS snapshot lock (%s): %s", d.Id(), err)
	}

	d.Set("lock_created_on", aws.ToTime(resp.LockCreatedOn).Format(time.RFC3339))
	d.Set("cool_off_period_expires_on", aws.ToTime(resp.CoolOffPeriodExpiresOn).Format(time.RFC3339))
	d.Set("lock_duration_start_time", aws.ToTime(resp.LockDurationStartTime).Format(time.RFC3339))
	d.Set("lock_state", resp.LockState)

	return append(diags, resourceEBSSnapshotLockRead(ctx, d, meta)...)
}

func resourceEBSSnapshotLockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.UnlockSnapshotInput{
		SnapshotId: aws.String(d.Id()),
	}

	_, err := conn.UnlockSnapshot(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "unlocking EBS snapshot (%s): %s", d.Id(), err)
	}

	return diags
}

func resourcEBSSnapshotLockCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	lockMode := diff.Get("lock_mode").(string)
	coolOffPeriod := diff.Get("cool_off_period").(string)

	if lockMode == "governance" && coolOffPeriod != "" {
		return fmt.Errorf("'cool_off_period' must not be set when 'lockMode' is '%s'", lockMode)
	}

	return nil
}
