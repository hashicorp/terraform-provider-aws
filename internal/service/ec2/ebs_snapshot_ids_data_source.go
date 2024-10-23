// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ebs_snapshot_ids", name="EBS Snapshot IDs")
func dataSourceEBSSnapshotIDs() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEBSSnapshotIDsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owners": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"restorable_by_user_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceEBSSnapshotIDsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeSnapshotsInput{}

	if v, ok := d.GetOk("owners"); ok && len(v.([]interface{})) > 0 {
		input.OwnerIds = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("restorable_by_user_ids"); ok && len(v.([]interface{})) > 0 {
		input.RestorableByUserIds = flex.ExpandStringValueList(v.([]interface{}))
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	snapshots, err := findSnapshots(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Snapshots: %s", err)
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return aws.ToTime(snapshots[i].StartTime).Unix() > aws.ToTime(snapshots[j].StartTime).Unix()
	})

	var snapshotIDs []string

	for _, v := range snapshots {
		snapshotIDs = append(snapshotIDs, aws.ToString(v.SnapshotId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrIDs, snapshotIDs)

	return diags
}
