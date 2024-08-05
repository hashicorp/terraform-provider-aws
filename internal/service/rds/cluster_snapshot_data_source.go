// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_db_cluster_snapshot", name="DB Cluster Snapshot")
// @Tags
// @Testing(tagsTest=false)
func dataSourceClusterSnapshot() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterSnapshotRead,

		Schema: map[string]*schema.Schema{
			names.AttrAllocatedStorage: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"db_cluster_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"db_cluster_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_cluster_snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"include_public": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"include_shared": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"license_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrMostRecent: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"snapshot_create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_db_cluster_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceClusterSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.DescribeDBClusterSnapshotsInput{
		IncludePublic: aws.Bool(d.Get("include_public").(bool)),
		IncludeShared: aws.Bool(d.Get("include_shared").(bool)),
	}

	if v, ok := d.GetOk("db_cluster_identifier"); ok {
		input.DBClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_cluster_snapshot_identifier"); ok {
		input.DBClusterSnapshotIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_type"); ok {
		input.SnapshotType = aws.String(v.(string))
	}

	f := tfslices.PredicateTrue[*types.DBClusterSnapshot]()
	if tags := getTagsInV2(ctx); len(tags) > 0 {
		f = func(v *types.DBClusterSnapshot) bool {
			return keyValueTagsV2(ctx, v.TagList).ContainsAll(keyValueTagsV2(ctx, tags))
		}
	}

	snapshots, err := findDBClusterSnapshots(ctx, conn, input, f)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Cluster Snapshots: %s", err)
	}

	if len(snapshots) < 1 {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
	}

	var snapshot *types.DBClusterSnapshot
	if len(snapshots) > 1 {
		if d.Get(names.AttrMostRecent).(bool) {
			snapshot = mostRecentClusterSnapshot(snapshots)
		} else {
			return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more specific search criteria.")
		}
	} else {
		snapshot = &snapshots[0]
	}

	d.SetId(aws.ToString(snapshot.DBClusterSnapshotIdentifier))
	d.Set(names.AttrAllocatedStorage, snapshot.AllocatedStorage)
	d.Set(names.AttrAvailabilityZones, snapshot.AvailabilityZones)
	d.Set("db_cluster_identifier", snapshot.DBClusterIdentifier)
	d.Set("db_cluster_snapshot_arn", snapshot.DBClusterSnapshotArn)
	d.Set("db_cluster_snapshot_identifier", snapshot.DBClusterSnapshotIdentifier)
	d.Set(names.AttrEngine, snapshot.Engine)
	d.Set(names.AttrEngineVersion, snapshot.EngineVersion)
	d.Set(names.AttrKMSKeyID, snapshot.KmsKeyId)
	d.Set("license_model", snapshot.LicenseModel)
	d.Set(names.AttrPort, snapshot.Port)
	if snapshot.SnapshotCreateTime != nil {
		d.Set("snapshot_create_time", snapshot.SnapshotCreateTime.Format(time.RFC3339))
	}
	d.Set("snapshot_type", snapshot.SnapshotType)
	d.Set("source_db_cluster_snapshot_arn", snapshot.SourceDBClusterSnapshotArn)
	d.Set(names.AttrStatus, snapshot.Status)
	d.Set(names.AttrStorageEncrypted, snapshot.StorageEncrypted)
	d.Set(names.AttrVPCID, snapshot.VpcId)

	setTagsOutV2(ctx, snapshot.TagList)

	return diags
}

type rdsClusterSnapshotSort []types.DBClusterSnapshot

func (a rdsClusterSnapshotSort) Len() int      { return len(a) }
func (a rdsClusterSnapshotSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a rdsClusterSnapshotSort) Less(i, j int) bool {
	// Snapshot creation can be in progress
	if a[i].SnapshotCreateTime == nil {
		return true
	}
	if a[j].SnapshotCreateTime == nil {
		return false
	}

	return (aws.ToTime(a[i].SnapshotCreateTime)).Before(aws.ToTime(a[j].SnapshotCreateTime))
}

func mostRecentClusterSnapshot(snapshots []types.DBClusterSnapshot) *types.DBClusterSnapshot {
	sortedSnapshots := snapshots
	sort.Sort(rdsClusterSnapshotSort(sortedSnapshots))
	return &sortedSnapshots[len(sortedSnapshots)-1]
}
