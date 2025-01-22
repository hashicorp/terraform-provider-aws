// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"slices"
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

// @SDKDataSource("aws_db_snapshot", name="DB Snapshot")
// @Tags
// @Testing(tagsTest=false)
func dataSourceSnapshot() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSnapshotRead,

		Schema: map[string]*schema.Schema{
			names.AttrAllocatedStorage: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_instance_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"db_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
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
			names.AttrIOPS: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
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
			"option_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"original_snapshot_create_time": {
				Type:     schema.TypeString,
				Computed: true,
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
			"source_db_snapshot_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
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

func dataSourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.DescribeDBSnapshotsInput{
		IncludePublic: aws.Bool(d.Get("include_public").(bool)),
		IncludeShared: aws.Bool(d.Get("include_shared").(bool)),
	}

	if v, ok := d.GetOk("db_instance_identifier"); ok {
		input.DBInstanceIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_snapshot_identifier"); ok {
		input.DBSnapshotIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_type"); ok {
		input.SnapshotType = aws.String(v.(string))
	}

	f := tfslices.PredicateTrue[*types.DBSnapshot]()
	if tags := getTagsIn(ctx); len(tags) > 0 {
		f = func(v *types.DBSnapshot) bool {
			return KeyValueTags(ctx, v.TagList).ContainsAll(KeyValueTags(ctx, tags))
		}
	}

	snapshots, err := findDBSnapshots(ctx, conn, input, f)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Snapshots: %s", err)
	}

	if len(snapshots) < 1 {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
	}

	if len(snapshots) > 1 && !d.Get(names.AttrMostRecent).(bool) {
		return sdkdiag.AppendErrorf(diags, "Your query returned more than one result. Please try a more specific search criteria.")
	}

	snapshot := slices.MaxFunc(snapshots, func(a, b types.DBSnapshot) int {
		if a.SnapshotCreateTime == nil || b.SnapshotCreateTime == nil {
			return 0
		}
		return a.SnapshotCreateTime.Compare(aws.ToTime(b.SnapshotCreateTime))
	})

	d.SetId(aws.ToString(snapshot.DBSnapshotIdentifier))
	d.Set(names.AttrAllocatedStorage, snapshot.AllocatedStorage)
	d.Set(names.AttrAvailabilityZone, snapshot.AvailabilityZone)
	d.Set("db_instance_identifier", snapshot.DBInstanceIdentifier)
	d.Set("db_snapshot_arn", snapshot.DBSnapshotArn)
	d.Set("db_snapshot_identifier", snapshot.DBSnapshotIdentifier)
	d.Set(names.AttrEncrypted, snapshot.Encrypted)
	d.Set(names.AttrEngine, snapshot.Engine)
	d.Set(names.AttrEngineVersion, snapshot.EngineVersion)
	d.Set(names.AttrIOPS, snapshot.Iops)
	d.Set(names.AttrKMSKeyID, snapshot.KmsKeyId)
	d.Set("license_model", snapshot.LicenseModel)
	d.Set("option_group_name", snapshot.OptionGroupName)
	if snapshot.OriginalSnapshotCreateTime != nil {
		d.Set("original_snapshot_create_time", snapshot.OriginalSnapshotCreateTime.Format(time.RFC3339))
	}
	d.Set(names.AttrPort, snapshot.Port)
	d.Set("source_db_snapshot_identifier", snapshot.SourceDBSnapshotIdentifier)
	d.Set("source_region", snapshot.SourceRegion)
	if snapshot.SnapshotCreateTime != nil {
		d.Set("snapshot_create_time", snapshot.SnapshotCreateTime.Format(time.RFC3339))
	}
	d.Set("snapshot_type", snapshot.SnapshotType)
	d.Set(names.AttrStatus, snapshot.Status)
	d.Set(names.AttrStorageType, snapshot.StorageType)
	d.Set(names.AttrVPCID, snapshot.VpcId)

	setTagsOut(ctx, snapshot.TagList)

	return diags
}
