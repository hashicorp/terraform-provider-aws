// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_elasticache_cluster", name="Cluster")
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cache_nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAddress: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"outpost_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"cluster_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(v interface{}) string {
					value := v.(string)
					return strings.ToLower(value)
				},
			},
			"configuration_endpoint": {
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
			"ip_discovery": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_delivery_configuration": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDestination: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"destination_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"log_format": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"log_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"maintenance_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"notification_topic_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"num_cache_nodes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrParameterGroupName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"preferred_outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"snapshot_retention_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"snapshot_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	partition := meta.(*conns.AWSClient).Partition

	clusterID := d.Get("cluster_id").(string)
	cluster, err := findCacheClusterWithNodeInfoByID(ctx, conn, clusterID)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("ElastiCache Cluster", err))
	}

	d.SetId(aws.ToString(cluster.CacheClusterId))
	d.Set(names.AttrARN, cluster.ARN)
	d.Set(names.AttrAvailabilityZone, cluster.PreferredAvailabilityZone)
	if cluster.ConfigurationEndpoint != nil {
		clusterAddress, port := aws.ToString(cluster.ConfigurationEndpoint.Address), aws.ToInt32(cluster.ConfigurationEndpoint.Port)
		d.Set("cluster_address", clusterAddress)
		d.Set("configuration_endpoint", fmt.Sprintf("%s:%d", clusterAddress, port))
		d.Set(names.AttrPort, port)
	}
	d.Set("cluster_id", cluster.CacheClusterId)
	d.Set(names.AttrEngine, cluster.Engine)
	d.Set(names.AttrEngineVersion, cluster.EngineVersion)
	d.Set("ip_discovery", cluster.IpDiscovery)
	d.Set("log_delivery_configuration", flattenLogDeliveryConfigurations(cluster.LogDeliveryConfigurations))
	d.Set("maintenance_window", cluster.PreferredMaintenanceWindow)
	d.Set("network_type", cluster.NetworkType)
	d.Set("node_type", cluster.CacheNodeType)
	if cluster.NotificationConfiguration != nil {
		if aws.ToString(cluster.NotificationConfiguration.TopicStatus) == "active" {
			d.Set("notification_topic_arn", cluster.NotificationConfiguration.TopicArn)
		}
	}
	d.Set("num_cache_nodes", cluster.NumCacheNodes)
	if cluster.CacheParameterGroup != nil {
		d.Set(names.AttrParameterGroupName, cluster.CacheParameterGroup.CacheParameterGroupName)
	}
	d.Set("preferred_outpost_arn", cluster.PreferredOutpostArn)
	d.Set("replication_group_id", cluster.ReplicationGroupId)
	d.Set(names.AttrSecurityGroupIDs, flattenSecurityGroupIDs(cluster.SecurityGroups))
	d.Set("snapshot_retention_limit", cluster.SnapshotRetentionLimit)
	d.Set("snapshot_window", cluster.SnapshotWindow)
	d.Set("subnet_group_name", cluster.CacheSubnetGroupName)

	if err := setCacheNodeData(d, cluster); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	tags, err := listTags(ctx, conn, aws.ToString(cluster.ARN))

	if err != nil && !errs.IsUnsupportedOperationInPartitionError(partition, err) {
		return sdkdiag.AppendErrorf(diags, "listing tags for ElastiCache Cluster (%s): %s", d.Id(), err)
	}

	if err != nil {
		log.Printf("[WARN] error listing tags for ElastiCache Cluster (%s): %s", d.Id(), err)
	}

	if tags != nil {
		if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}
	}

	return diags
}
