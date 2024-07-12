// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_cluster")
func DataSourceCluster() *schema.Resource {
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
			"backtrack_window": {
				Type:     schema.TypeInt,
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
			names.AttrDatabaseName: {
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
			"db_system_id": {
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
			"engine_mode": {
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
			"iam_database_authentication_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"iam_roles": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_user_secret": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"master_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_type": {
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
			"replication_source_identifier": {
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

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dbClusterID := d.Get(names.AttrClusterIdentifier).(string)
	dbc, err := FindDBClusterByID(ctx, conn, dbClusterID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster (%s): %s", dbClusterID, err)
	}

	d.SetId(aws.StringValue(dbc.DBClusterIdentifier))

	clusterARN := aws.StringValue(dbc.DBClusterArn)
	d.Set(names.AttrARN, clusterARN)
	d.Set(names.AttrAvailabilityZones, aws.StringValueSlice(dbc.AvailabilityZones))
	d.Set("backtrack_window", dbc.BacktrackWindow)
	d.Set("backup_retention_period", dbc.BackupRetentionPeriod)
	d.Set(names.AttrClusterIdentifier, dbc.DBClusterIdentifier)
	var clusterMembers []string
	for _, v := range dbc.DBClusterMembers {
		clusterMembers = append(clusterMembers, aws.StringValue(v.DBInstanceIdentifier))
	}
	d.Set("cluster_members", clusterMembers)
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)
	// Only set the DatabaseName if it is not nil. There is a known API bug where
	// RDS accepts a DatabaseName but does not return it, causing a perpetual
	// diff.
	//	See https://github.com/hashicorp/terraform/issues/4671 for backstory
	if dbc.DatabaseName != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set(names.AttrDatabaseName, dbc.DatabaseName)
	}
	d.Set("db_cluster_parameter_group_name", dbc.DBClusterParameterGroup)
	d.Set("db_subnet_group_name", dbc.DBSubnetGroup)
	d.Set("db_system_id", dbc.DBSystemId)
	d.Set("enabled_cloudwatch_logs_exports", aws.StringValueSlice(dbc.EnabledCloudwatchLogsExports))
	d.Set(names.AttrEndpoint, dbc.Endpoint)
	d.Set(names.AttrEngine, dbc.Engine)
	d.Set("engine_mode", dbc.EngineMode)
	d.Set(names.AttrEngineVersion, dbc.EngineVersion)
	d.Set(names.AttrHostedZoneID, dbc.HostedZoneId)
	d.Set("iam_database_authentication_enabled", dbc.IAMDatabaseAuthenticationEnabled)
	var iamRoleARNs []string
	for _, v := range dbc.AssociatedRoles {
		iamRoleARNs = append(iamRoleARNs, aws.StringValue(v.RoleArn))
	}
	d.Set("iam_roles", iamRoleARNs)
	d.Set(names.AttrKMSKeyID, dbc.KmsKeyId)
	if dbc.MasterUserSecret != nil {
		if err := d.Set("master_user_secret", []interface{}{flattenManagedMasterUserSecret(dbc.MasterUserSecret)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_user_secret: %s", err)
		}
	}
	d.Set("master_username", dbc.MasterUsername)
	d.Set("network_type", dbc.NetworkType)
	d.Set(names.AttrPort, dbc.Port)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set(names.AttrPreferredMaintenanceWindow, dbc.PreferredMaintenanceWindow)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	d.Set("replication_source_identifier", dbc.ReplicationSourceIdentifier)
	d.Set(names.AttrStorageEncrypted, dbc.StorageEncrypted)
	var securityGroupIDs []string
	for _, v := range dbc.VpcSecurityGroups {
		securityGroupIDs = append(securityGroupIDs, aws.StringValue(v.VpcSecurityGroupId))
	}
	d.Set(names.AttrVPCSecurityGroupIDs, securityGroupIDs)

	tags := KeyValueTags(ctx, dbc.TagList)

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
