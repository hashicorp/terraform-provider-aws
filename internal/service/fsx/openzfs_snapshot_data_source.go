// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
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
func dataSourceOpenzfsSnapshot() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOpenZFSSnapshotRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: snapshotFiltersSchema(),
			names.AttrMostRecent: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrSnapshotID: {
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
	conn := meta.(*conns.AWSClient).FSxClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &fsx.DescribeSnapshotsInput{}

	if v, ok := d.GetOk("snapshot_ids"); ok && len(v.([]interface{})) > 0 {
		input.SnapshotIds = flex.ExpandStringValueList(v.([]interface{}))
	}

	input.Filters = append(input.Filters, newSnapshotFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	snapshots, err := findSnapshots(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Snapshot]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx Snapshots: %s", err)
	}

	if len(snapshots) < 1 {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
	}

	if len(snapshots) > 1 {
		if !d.Get(names.AttrMostRecent).(bool) {
			return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more "+
				"specific search criteria, or set `most_recent` attribute to true.")
		}

		sort.Slice(snapshots, func(i, j int) bool {
			return aws.ToTime(snapshots[i].CreationTime).Unix() > aws.ToTime(snapshots[j].CreationTime).Unix()
		})
	}

	snapshot := snapshots[0]
	d.SetId(aws.ToString(snapshot.SnapshotId))
	arn := aws.ToString(snapshot.ResourceARN)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreationTime, snapshot.CreationTime.Format(time.RFC3339))
	d.Set(names.AttrName, snapshot.Name)
	d.Set(names.AttrSnapshotID, snapshot.SnapshotId)
	d.Set("volume_id", snapshot.VolumeId)

	// Snapshot tags aren't set in the Describe response.
	// setTagsOut(ctx, snapshot.Tags)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for FSx OpenZFS Snapshot (%s): %s", arn, err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
