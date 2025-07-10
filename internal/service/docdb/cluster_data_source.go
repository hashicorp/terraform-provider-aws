// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_docdb_cluster", name="Cluster")
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"backup_retention_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_members": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"cluster_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_cluster_parameter_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled_cloudwatch_logs_exports": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
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
			names.AttrFinalSnapshotIdentifier: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"preferred_backup_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPreferredMaintenanceWindow: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reader_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	dbClusterId := d.Get(names.AttrClusterIdentifier).(string)
	dbc, err := findDBClusterByID(ctx, conn, dbClusterId)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocDB Cluster (%s): %s", dbClusterId, err)
	}

	d.SetId(aws.ToString(dbc.DBClusterIdentifier))
	clusterARN := aws.ToString(dbc.DBClusterArn)
	d.Set(names.AttrARN, clusterARN)
	d.Set(names.AttrAvailabilityZones, dbc.AvailabilityZones)
	d.Set("backup_retention_period", dbc.BackupRetentionPeriod)
	d.Set(names.AttrClusterIdentifier, dbc.DBClusterIdentifier)
	d.Set("cluster_members", tfslices.ApplyToAll(dbc.DBClusterMembers, func(v types.DBClusterMember) string {
		return aws.ToString(v.DBInstanceIdentifier)
	}))
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)
	d.Set("db_cluster_parameter_group_name", dbc.DBClusterParameterGroup)
	d.Set("db_subnet_group_name", dbc.DBSubnetGroup)
	d.Set("enabled_cloudwatch_logs_exports", dbc.EnabledCloudwatchLogsExports)
	d.Set(names.AttrEndpoint, dbc.Endpoint)
	d.Set(names.AttrEngine, dbc.Engine)
	d.Set(names.AttrEngineVersion, dbc.EngineVersion)
	d.Set(names.AttrHostedZoneID, dbc.HostedZoneId)
	d.Set(names.AttrKMSKeyID, dbc.KmsKeyId)
	d.Set("master_username", dbc.MasterUsername)
	d.Set(names.AttrPort, dbc.Port)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set(names.AttrPreferredMaintenanceWindow, dbc.PreferredMaintenanceWindow)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	d.Set(names.AttrStorageEncrypted, dbc.StorageEncrypted)
	d.Set(names.AttrVPCSecurityGroupIDs, tfslices.ApplyToAll(dbc.VpcSecurityGroups, func(v types.VpcSecurityGroupMembership) string {
		return aws.ToString(v.VpcSecurityGroupId)
	}))

	tags, err := listTags(ctx, conn, clusterARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for DocumentDB Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
