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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_fsx_openzfs_snapshot")
func DataSourceOpenzfsSnapshot() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOpenzfsSnapshotRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": DataSourceSnapshotFiltersSchema(),
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name": {
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
			"tags": tftags.TagsSchemaComputed(),
			"volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceOpenzfsSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &fsx.DescribeSnapshotsInput{}

	if v, ok := d.GetOk("snapshot_ids"); ok && len(v.([]interface{})) > 0 {
		input.SnapshotIds = flex.ExpandStringList(v.([]interface{}))
	}

	input.Filters = append(input.Filters, BuildSnapshotFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	snapshots, err := FindSnapshots(ctx, conn, input)

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
	d.Set("arn", snapshot.ResourceARN)
	d.Set("name", snapshot.Name)
	d.Set("snapshot_id", snapshot.SnapshotId)
	d.Set("volume_id", snapshot.VolumeId)

	if err := d.Set("creation_time", snapshot.CreationTime.Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting creation_time: %s", err)
	}

	//Snapshot tags do not get returned with describe call so need to make a separate list tags call
	tags, tagserr := listTags(ctx, conn, *snapshot.ResourceARN)

	if tagserr != nil {
		return sdkdiag.AppendErrorf(diags, "reading Tags for FSx OpenZFS Snapshot (%s): %s", d.Id(), err)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
