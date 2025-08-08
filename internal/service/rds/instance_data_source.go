// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"

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

// @SDKDataSource("aws_db_instance", name="DB Instance")
// @Tags
// @Testing(tagsIdentifierAttribute="db_instance_arn")
func dataSourceInstance() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceRead,

		Schema: map[string]*schema.Schema{
			names.AttrAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAllocatedStorage: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrAutoMinorVersionUpgrade: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
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
			"database_insights_mode": {
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
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
			"master_username": {
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
			names.AttrPubliclyAccessible: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"replicate_source_db": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"storage_throughput": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
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

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	var instance *types.DBInstance

	filter := tfslices.PredicateTrue[*types.DBInstance]()
	if tags := getTagsIn(ctx); len(tags) > 0 {
		filter = func(v *types.DBInstance) bool {
			return keyValueTags(ctx, v.TagList).ContainsAll(keyValueTags(ctx, tags))
		}
	}

	if v, ok := d.GetOk("db_instance_identifier"); ok {
		id := v.(string)
		output, err := findDBInstanceByID(ctx, conn, id)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading RDS DB Instance (%s): %s", id, err)
		}

		if !filter(output) {
			return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again.")
		}

		instance = output
	} else {
		input := &rds.DescribeDBInstancesInput{}
		output, err := findDBInstance(ctx, conn, input, filter)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading RDS DB Instance: %s", err)
		}

		instance = output
	}

	d.SetId(aws.ToString(instance.DBInstanceIdentifier))
	d.Set(names.AttrAllocatedStorage, instance.AllocatedStorage)
	d.Set(names.AttrAutoMinorVersionUpgrade, instance.AutoMinorVersionUpgrade)
	d.Set(names.AttrAvailabilityZone, instance.AvailabilityZone)
	d.Set("backup_retention_period", instance.BackupRetentionPeriod)
	d.Set("ca_cert_identifier", instance.CACertificateIdentifier)
	d.Set("database_insights_mode", instance.DatabaseInsightsMode)
	d.Set("db_cluster_identifier", instance.DBClusterIdentifier)
	d.Set("db_instance_arn", instance.DBInstanceArn)
	d.Set("db_instance_class", instance.DBInstanceClass)
	d.Set("db_instance_port", instance.DbInstancePort)
	d.Set("db_name", instance.DBName)
	d.Set("db_parameter_groups", tfslices.ApplyToAll(instance.DBParameterGroups, func(v types.DBParameterGroupStatus) string {
		return aws.ToString(v.DBParameterGroupName)
	}))
	if instance.DBSubnetGroup != nil {
		d.Set("db_subnet_group", instance.DBSubnetGroup.DBSubnetGroupName)
	} else {
		d.Set("db_subnet_group", "")
	}
	d.Set("enabled_cloudwatch_logs_exports", instance.EnabledCloudwatchLogsExports)
	d.Set(names.AttrEngine, instance.Engine)
	d.Set(names.AttrEngineVersion, instance.EngineVersion)
	d.Set(names.AttrIOPS, instance.Iops)
	d.Set(names.AttrKMSKeyID, instance.KmsKeyId)
	d.Set("license_model", instance.LicenseModel)
	d.Set("master_username", instance.MasterUsername)
	if instance.MasterUserSecret != nil {
		if err := d.Set("master_user_secret", []any{flattenManagedMasterUserSecret(instance.MasterUserSecret)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_user_secret: %s", err)
		}
	}
	d.Set("max_allocated_storage", instance.MaxAllocatedStorage)
	d.Set("monitoring_interval", instance.MonitoringInterval)
	d.Set("monitoring_role_arn", instance.MonitoringRoleArn)
	d.Set("multi_az", instance.MultiAZ)
	d.Set("network_type", instance.NetworkType)
	d.Set("option_group_memberships", tfslices.ApplyToAll(instance.OptionGroupMemberships, func(v types.OptionGroupMembership) string {
		return aws.ToString(v.OptionGroupName)
	}))
	d.Set("preferred_backup_window", instance.PreferredBackupWindow)
	d.Set(names.AttrPreferredMaintenanceWindow, instance.PreferredMaintenanceWindow)
	d.Set(names.AttrPubliclyAccessible, instance.PubliclyAccessible)
	d.Set("replicate_source_db", instance.ReadReplicaSourceDBInstanceIdentifier)
	d.Set(names.AttrResourceID, instance.DbiResourceId)
	d.Set(names.AttrStorageEncrypted, instance.StorageEncrypted)
	d.Set("storage_throughput", instance.StorageThroughput)
	d.Set(names.AttrStorageType, instance.StorageType)
	d.Set("timezone", instance.Timezone)
	d.Set("vpc_security_groups", tfslices.ApplyToAll(instance.VpcSecurityGroups, func(v types.VpcSecurityGroupMembership) string {
		return aws.ToString(v.VpcSecurityGroupId)
	}))

	// Per AWS SDK Go docs:
	// The endpoint might not be shown for instances whose status is creating.
	if dbEndpoint := instance.Endpoint; dbEndpoint != nil {
		d.Set(names.AttrAddress, dbEndpoint.Address)
		d.Set(names.AttrEndpoint, fmt.Sprintf("%s:%d", aws.ToString(dbEndpoint.Address), aws.ToInt32(dbEndpoint.Port)))
		d.Set(names.AttrHostedZoneID, dbEndpoint.HostedZoneId)
		d.Set(names.AttrPort, dbEndpoint.Port)
	} else {
		d.Set(names.AttrAddress, nil)
		d.Set(names.AttrEndpoint, nil)
		d.Set(names.AttrHostedZoneID, nil)
		d.Set(names.AttrPort, nil)
	}

	setTagsOut(ctx, instance.TagList)

	return diags
}
