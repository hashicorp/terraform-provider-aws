// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_redshift_cluster", name="Cluster")
// @Tags
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"allow_version_upgrade": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"aqua_configuration_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automated_snapshot_retention_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_relocation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrBucketName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_namespace_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"node_role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"cluster_parameter_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_revision_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_subnet_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_iam_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elastic_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_logging": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enhanced_vpc_routing": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"iam_roles": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maintenance_track_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manual_snapshot_retention_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"number_of_nodes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
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
			names.AttrS3KeyPrefix: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_destination_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_exports": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	clusterID := d.Get(names.AttrClusterIdentifier).(string)
	rsc, err := findClusterByID(ctx, conn, clusterID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Cluster (%s): %s", clusterID, err)
	}

	d.SetId(clusterID)
	d.Set("allow_version_upgrade", rsc.AllowVersionUpgrade)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   redshift.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("cluster:%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("automated_snapshot_retention_period", rsc.AutomatedSnapshotRetentionPeriod)
	if rsc.AquaConfiguration != nil {
		d.Set("aqua_configuration_status", rsc.AquaConfiguration.AquaConfigurationStatus)
	}
	d.Set(names.AttrAvailabilityZone, rsc.AvailabilityZone)
	if v, err := clusterAvailabilityZoneRelocationStatus(rsc); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else {
		d.Set("availability_zone_relocation_enabled", v)
	}
	d.Set(names.AttrClusterIdentifier, rsc.ClusterIdentifier)
	d.Set("cluster_namespace_arn", rsc.ClusterNamespaceArn)
	if err := d.Set("cluster_nodes", flattenClusterNodes(rsc.ClusterNodes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_nodes: %s", err)
	}
	if len(rsc.ClusterParameterGroups) > 0 {
		d.Set("cluster_parameter_group_name", rsc.ClusterParameterGroups[0].ParameterGroupName)
	}
	d.Set("cluster_public_key", rsc.ClusterPublicKey)
	d.Set("cluster_revision_number", rsc.ClusterRevisionNumber)
	d.Set("cluster_subnet_group_name", rsc.ClusterSubnetGroupName)
	if len(rsc.ClusterNodes) > 1 {
		d.Set("cluster_type", clusterTypeMultiNode)
	} else {
		d.Set("cluster_type", clusterTypeSingleNode)
	}
	d.Set("cluster_version", rsc.ClusterVersion)
	d.Set(names.AttrDatabaseName, rsc.DBName)
	d.Set("default_iam_role_arn", rsc.DefaultIamRoleArn)
	if rsc.ElasticIpStatus != nil {
		d.Set("elastic_ip", rsc.ElasticIpStatus.ElasticIp)
	}
	d.Set(names.AttrEncrypted, rsc.Encrypted)
	if rsc.Endpoint != nil {
		d.Set(names.AttrEndpoint, rsc.Endpoint.Address)
		d.Set(names.AttrPort, rsc.Endpoint.Port)
	}
	d.Set("enhanced_vpc_routing", rsc.EnhancedVpcRouting)
	d.Set("iam_roles", tfslices.ApplyToAll(rsc.IamRoles, func(v *redshift.ClusterIamRole) string {
		return aws.StringValue(v.IamRoleArn)
	}))
	d.Set(names.AttrKMSKeyID, rsc.KmsKeyId)
	d.Set("maintenance_track_name", rsc.MaintenanceTrackName)
	d.Set("manual_snapshot_retention_period", rsc.ManualSnapshotRetentionPeriod)
	d.Set("master_username", rsc.MasterUsername)
	if v, err := clusterMultiAZStatus(rsc); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else {
		d.Set("multi_az", v)
	}
	d.Set("node_type", rsc.NodeType)
	d.Set("number_of_nodes", rsc.NumberOfNodes)
	d.Set(names.AttrPreferredMaintenanceWindow, rsc.PreferredMaintenanceWindow)
	d.Set(names.AttrPubliclyAccessible, rsc.PubliclyAccessible)
	d.Set(names.AttrVPCID, rsc.VpcId)
	d.Set(names.AttrVPCSecurityGroupIDs, tfslices.ApplyToAll(rsc.VpcSecurityGroups, func(v *redshift.VpcSecurityGroupMembership) string {
		return aws.StringValue(v.VpcSecurityGroupId)
	}))

	setTagsOut(ctx, rsc.Tags)

	loggingStatus, err := conn.DescribeLoggingStatusWithContext(ctx, &redshift.DescribeLoggingStatusInput{
		ClusterIdentifier: aws.String(clusterID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Cluster (%s) logging status: %s", d.Id(), err)
	}

	if loggingStatus != nil && aws.BoolValue(loggingStatus.LoggingEnabled) {
		d.Set(names.AttrBucketName, loggingStatus.BucketName)
		d.Set("enable_logging", loggingStatus.LoggingEnabled)
		d.Set("log_destination_type", loggingStatus.LogDestinationType)
		d.Set("log_exports", aws.StringValueSlice(loggingStatus.LogExports))
		d.Set(names.AttrS3KeyPrefix, loggingStatus.S3KeyPrefix)
	}

	return diags
}
