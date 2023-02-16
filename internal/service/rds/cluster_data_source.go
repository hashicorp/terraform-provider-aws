package rds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zones": {
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
			"cluster_identifier": {
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
			"database_name": {
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
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"final_snapshot_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
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
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"preferred_backup_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preferred_maintenance_window": {
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
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dbClusterID := d.Get("cluster_identifier").(string)
	dbc, err := FindDBClusterByID(ctx, conn, dbClusterID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster (%s): %s", dbClusterID, err)
	}

	d.SetId(aws.StringValue(dbc.DBClusterIdentifier))

	clusterARN := aws.StringValue(dbc.DBClusterArn)
	d.Set("arn", clusterARN)
	d.Set("availability_zones", aws.StringValueSlice(dbc.AvailabilityZones))
	d.Set("backtrack_window", dbc.BacktrackWindow)
	d.Set("backup_retention_period", dbc.BackupRetentionPeriod)
	d.Set("cluster_identifier", dbc.DBClusterIdentifier)
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
		d.Set("database_name", dbc.DatabaseName)
	}
	d.Set("db_cluster_parameter_group_name", dbc.DBClusterParameterGroup)
	d.Set("db_subnet_group_name", dbc.DBSubnetGroup)
	d.Set("enabled_cloudwatch_logs_exports", aws.StringValueSlice(dbc.EnabledCloudwatchLogsExports))
	d.Set("endpoint", dbc.Endpoint)
	d.Set("engine", dbc.Engine)
	d.Set("engine_mode", dbc.EngineMode)
	d.Set("engine_version", dbc.EngineVersion)
	d.Set("hosted_zone_id", dbc.HostedZoneId)
	d.Set("iam_database_authentication_enabled", dbc.IAMDatabaseAuthenticationEnabled)
	var iamRoleARNs []string
	for _, v := range dbc.AssociatedRoles {
		iamRoleARNs = append(iamRoleARNs, aws.StringValue(v.RoleArn))
	}
	d.Set("iam_roles", iamRoleARNs)
	d.Set("kms_key_id", dbc.KmsKeyId)
	d.Set("master_username", dbc.MasterUsername)
	d.Set("network_type", dbc.NetworkType)
	d.Set("port", dbc.Port)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", dbc.PreferredMaintenanceWindow)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	d.Set("replication_source_identifier", dbc.ReplicationSourceIdentifier)
	d.Set("storage_encrypted", dbc.StorageEncrypted)
	var securityGroupIDs []string
	for _, v := range dbc.VpcSecurityGroups {
		securityGroupIDs = append(securityGroupIDs, aws.StringValue(v.VpcSecurityGroupId))
	}
	d.Set("vpc_security_group_ids", securityGroupIDs)

	tags, err := ListTags(ctx, conn, clusterARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for RDS Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
