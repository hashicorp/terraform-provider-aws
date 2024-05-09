// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_fsx_openzfs_snapshot", name="OpenZFS Snapshot")
// @Tags
func dataSourceOpenzfsSnapshot() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOpenZFSSnapshotRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": snapshotFiltersSchema(),
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceOpenZFSSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	input := &fsx.DescribeSnapshotsInput{}

	if v, ok := d.GetOk("snapshot_ids"); ok && len(v.([]interface{})) > 0 {
		input.SnapshotIds = flex.ExpandStringList(v.([]interface{}))
	}

	input.Filters = append(input.Filters, newSnapshotFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	snapshots, err := findSnapshots(ctx, conn, input, tfslices.PredicateTrue[*fsx.Snapshot]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx Snapshots: %s", err)
	}

	if len(snapshots) < 1 {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
	}

	if len(snapshots) > 1 {
		if !d.Get("most_recent").(bool) {
			return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more "+
				"specific search criteria, or set `most_recent` attribute to true.")
		}

		sort.Slice(snapshots, func(i, j int) bool {
			return aws.TimeValue(snapshots[i].CreationTime).Unix() > aws.TimeValue(snapshots[j].CreationTime).Unix()
		})
	}

	snapshot := snapshots[0]
	d.SetId(aws.StringValue(snapshot.SnapshotId))
	d.Set(names.AttrARN, snapshot.ResourceARN)
	d.Set("creation_time", snapshot.CreationTime.Format(time.RFC3339))
	d.Set(names.AttrName, snapshot.Name)
	d.Set("snapshot_id", snapshot.SnapshotId)
	d.Set("volume_id", snapshot.VolumeId)

	setTagsOut(ctx, snapshot.Tags)

	return diags
}
