// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_db_instance")
func DataSourceInstance() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceRead,

		Schema: map[string]*schema.Schema{
			"address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"allocated_storage": {
				Type:     schema.TypeInt,
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
			"backup_retention_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ca_cert_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_cluster_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_instance_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_instance_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_instance_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"db_instance_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"db_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_parameter_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"db_subnet_group": {
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
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iops": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_user_secret": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
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
			"max_allocated_storage": {
				Type:     schema.TypeInt,
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
			"multi_az": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"network_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"option_group_memberships": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"replicate_source_db": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"storage_throughput": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"storage_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"timezone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_security_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	v, err := findDBInstanceByIDSDKv1(ctx, conn, d.Get("db_instance_identifier").(string))
	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("RDS DB Instance", err))
	}

	d.SetId(aws.StringValue(v.DBInstanceIdentifier))
	d.Set("allocated_storage", v.AllocatedStorage)
	d.Set("auto_minor_version_upgrade", v.AutoMinorVersionUpgrade)
	d.Set("availability_zone", v.AvailabilityZone)
	d.Set("backup_retention_period", v.BackupRetentionPeriod)
	d.Set("ca_cert_identifier", v.CACertificateIdentifier)
	d.Set("db_cluster_identifier", v.DBClusterIdentifier)
	d.Set("db_instance_arn", v.DBInstanceArn)
	d.Set("db_instance_class", v.DBInstanceClass)
	d.Set("db_instance_port", v.DbInstancePort)
	d.Set("db_name", v.DBName)
	var parameterGroupNames []string
	for _, v := range v.DBParameterGroups {
		parameterGroupNames = append(parameterGroupNames, aws.StringValue(v.DBParameterGroupName))
	}
	d.Set("db_parameter_groups", parameterGroupNames)
	if v.DBSubnetGroup != nil {
		d.Set("db_subnet_group", v.DBSubnetGroup.DBSubnetGroupName)
	} else {
		d.Set("db_subnet_group", "")
	}
	d.Set("enabled_cloudwatch_logs_exports", aws.StringValueSlice(v.EnabledCloudwatchLogsExports))
	d.Set("engine", v.Engine)
	d.Set("engine_version", v.EngineVersion)
	d.Set("iops", v.Iops)
	d.Set("kms_key_id", v.KmsKeyId)
	d.Set("license_model", v.LicenseModel)
	d.Set("master_username", v.MasterUsername)
	if v.MasterUserSecret != nil {
		if err := d.Set("master_user_secret", []interface{}{flattenManagedMasterUserSecret(v.MasterUserSecret)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_user_secret: %s", err)
		}
	}
	d.Set("max_allocated_storage", v.MaxAllocatedStorage)
	d.Set("monitoring_interval", v.MonitoringInterval)
	d.Set("monitoring_role_arn", v.MonitoringRoleArn)
	d.Set("multi_az", v.MultiAZ)
	d.Set("network_type", v.NetworkType)
	var optionGroupNames []string
	for _, v := range v.OptionGroupMemberships {
		optionGroupNames = append(optionGroupNames, aws.StringValue(v.OptionGroupName))
	}
	d.Set("option_group_memberships", optionGroupNames)
	d.Set("preferred_backup_window", v.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", v.PreferredMaintenanceWindow)
	d.Set("publicly_accessible", v.PubliclyAccessible)
	d.Set("replicate_source_db", v.ReadReplicaSourceDBInstanceIdentifier)
	d.Set("resource_id", v.DbiResourceId)
	d.Set("storage_encrypted", v.StorageEncrypted)
	d.Set("storage_throughput", v.StorageThroughput)
	d.Set("storage_type", v.StorageType)
	d.Set("timezone", v.Timezone)
	var vpcSecurityGroupIDs []string
	for _, v := range v.VpcSecurityGroups {
		vpcSecurityGroupIDs = append(vpcSecurityGroupIDs, aws.StringValue(v.VpcSecurityGroupId))
	}
	d.Set("vpc_security_groups", vpcSecurityGroupIDs)

	// Per AWS SDK Go docs:
	// The endpoint might not be shown for instances whose status is creating.
	if dbEndpoint := v.Endpoint; dbEndpoint != nil {
		d.Set("address", dbEndpoint.Address)
		d.Set("endpoint", fmt.Sprintf("%s:%d", aws.StringValue(dbEndpoint.Address), aws.Int64Value(dbEndpoint.Port)))
		d.Set("hosted_zone_id", dbEndpoint.HostedZoneId)
		d.Set("port", dbEndpoint.Port)
	} else {
		d.Set("address", nil)
		d.Set("endpoint", nil)
		d.Set("hosted_zone_id", nil)
		d.Set("port", nil)
	}

	tags := KeyValueTags(ctx, v.TagList)

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
