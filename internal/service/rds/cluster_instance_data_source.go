// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_rds_cluster_instance")
func DataSourceClusterInstance() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterInstanceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_cert_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"copy_tags_to_snapshot": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"custom_iam_instance_profile": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_parameter_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dbi_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"instance_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"monitoring_interval": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"monitoring_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"performance_insights_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"performance_insights_kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"performance_insights_retention_period": {
				Type:     schema.TypeInt,
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
			"promotion_tier": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"writer": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceClusterInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	identifier := d.Get("identifier").(string)

	db, err := findDBInstanceByIDSDKv1(ctx, conn, identifier)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Instance (%s): %s", identifier, err)
	}

	dbClusterID := aws.StringValue(db.DBClusterIdentifier)

	if dbClusterID == "" {
		return sdkdiag.AppendErrorf(diags, "DBClusterIdentifier is missing from RDS Cluster Instance (%s). The aws_db_instance resource should be used for non-Aurora instances", d.Id())
	}

	dbc, err := FindDBClusterByID(ctx, conn, dbClusterID)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster (%s): %s", dbClusterID, err)
	}

	log.Printf("DB Cluster: %v", dbc)

	for _, m := range dbc.DBClusterMembers {
		if aws.StringValue(m.DBInstanceIdentifier) == identifier {
			if aws.BoolValue(m.IsClusterWriter) {
				d.Set("writer", true)
			} else {
				d.Set("writer", false)
			}
		}
	}

	if db.Endpoint != nil {
		d.Set("endpoint", db.Endpoint.Address)
		d.Set("port", db.Endpoint.Port)
	}

	d.SetId(identifier)
	d.Set("arn", db.DBInstanceArn)
	d.Set("auto_minor_version_upgrade", db.AutoMinorVersionUpgrade)
	d.Set("availability_zone", db.AvailabilityZone)
	d.Set("ca_cert_identifier", db.CACertificateIdentifier)
	d.Set("cluster_identifier", db.DBClusterIdentifier)
	d.Set("copy_tags_to_snapshot", db.CopyTagsToSnapshot)
	d.Set("custom_iam_instance_profile", db.CustomIamInstanceProfile)
	d.Set("db_parameter_group_name", db.DBParameterGroups[0].DBParameterGroupName)
	d.Set("db_subnet_group_name", db.DBSubnetGroup.DBSubnetGroupName)
	d.Set("dbi_resource_id", db.DbiResourceId)
	d.Set("engine", db.Engine)
	d.Set("engine_version", db.EngineVersion)
	d.Set("engine_version_actual", db.EngineVersion)
	d.Set("instance_class", db.DBInstanceClass)
	d.Set("kms_key_id", db.KmsKeyId)
	d.Set("monitoring_interval", db.MonitoringInterval)
	d.Set("monitoring_role_arn", db.MonitoringRoleArn)
	d.Set("network_type", db.NetworkType)
	d.Set("performance_insights_enabled", db.PerformanceInsightsEnabled)
	d.Set("performance_insights_kms_key_id", db.PerformanceInsightsKMSKeyId)
	d.Set("performance_insights_retention_period", db.PerformanceInsightsRetentionPeriod)
	d.Set("preferred_backup_window", db.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", db.PreferredMaintenanceWindow)
	d.Set("promotion_tier", db.PromotionTier)
	d.Set("publicly_accessible", db.PubliclyAccessible)
	d.Set("storage_encrypted", db.StorageEncrypted)

	tags := KeyValueTags(ctx, db.TagList)

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return nil
}
