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

	return diags
}

func resourceEBSSnapshotLockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
}

func resourceEBSSnapshotLockUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
}

func resourceEBSSnapshotLockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
}
