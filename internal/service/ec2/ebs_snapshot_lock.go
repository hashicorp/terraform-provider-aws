// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cool_off_period": {
				Type:     schema.TypeInt,
				ForceNew: true,
			},
			"lock_duration": {
				Type:     schema.TypeInt,
				ForceNew: true,
			},
			"expiration_date": {
				Type:     schema.TypeString,
				ForceNew: true,
			},
		},
	}
}

func resourceEBSSnapshotLockCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
}

func resourceEBSSnapshotLockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
}

func resourceEBSSnapshotLockUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
}

func resourceEBSSnapshotLockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
}
