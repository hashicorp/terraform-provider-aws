// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
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
func DataSourceClusterSnapshot() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterSnapshotRead,

		Schema: map[string]*schema.Schema{
			"allocated_storage": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"availability_zones": {
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
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_id": {
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
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"port": {
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceClusterSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

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

	f := tfslices.PredicateTrue[*rds.DBClusterSnapshot]()
	if tags := getTagsIn(ctx); len(tags) > 0 {
		f = func(v *rds.DBClusterSnapshot) bool {
			return KeyValueTags(ctx, v.TagList).ContainsAll(KeyValueTags(ctx, tags))
		}
	}

	snapshots, err := findDBClusterSnapshots(ctx, conn, input, f)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Cluster Snapshots: %s", err)
	}

	if len(snapshots) < 1 {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
	}

	var snapshot *rds.DBClusterSnapshot
	if len(snapshots) > 1 {
		if d.Get("most_recent").(bool) {
			snapshot = mostRecentClusterSnapshot(snapshots)
		} else {
			return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more specific search criteria.")
		}
	} else {
		snapshot = snapshots[0]
	}

	d.SetId(aws.StringValue(snapshot.DBClusterSnapshotIdentifier))
	d.Set("allocated_storage", snapshot.AllocatedStorage)
	d.Set("availability_zones", aws.StringValueSlice(snapshot.AvailabilityZones))
	d.Set("db_cluster_identifier", snapshot.DBClusterIdentifier)
	d.Set("db_cluster_snapshot_arn", snapshot.DBClusterSnapshotArn)
	d.Set("db_cluster_snapshot_identifier", snapshot.DBClusterSnapshotIdentifier)
	d.Set("engine", snapshot.Engine)
	d.Set("engine_version", snapshot.EngineVersion)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("license_model", snapshot.LicenseModel)
	d.Set("port", snapshot.Port)
	if snapshot.SnapshotCreateTime != nil {
		d.Set("snapshot_create_time", snapshot.SnapshotCreateTime.Format(time.RFC3339))
	}
	d.Set("snapshot_type", snapshot.SnapshotType)
	d.Set("source_db_cluster_snapshot_arn", snapshot.SourceDBClusterSnapshotArn)
	d.Set("status", snapshot.Status)
	d.Set("storage_encrypted", snapshot.StorageEncrypted)
	d.Set("vpc_id", snapshot.VpcId)

	setTagsOut(ctx, snapshot.TagList)

	return diags
}

type rdsClusterSnapshotSort []*rds.DBClusterSnapshot

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

	return (*a[i].SnapshotCreateTime).Before(*a[j].SnapshotCreateTime)
}

func mostRecentClusterSnapshot(snapshots []*rds.DBClusterSnapshot) *rds.DBClusterSnapshot {
	sortedSnapshots := snapshots
	sort.Sort(rdsClusterSnapshotSort(sortedSnapshots))
	return sortedSnapshots[len(sortedSnapshots)-1]
}
