package elasticache

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cache_nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"outpost_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
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
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
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
						"destination": {
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
			"parameter_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
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
			"security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"security_group_names": {
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
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterID := d.Get("cluster_id").(string)
	cluster, err := FindCacheClusterWithNodeInfoByID(ctx, conn, clusterID)
	if tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "Your query returned no results. Please change your search criteria and try again")
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Cache Cluster (%s): %s", clusterID, err)
	}

	d.SetId(aws.StringValue(cluster.CacheClusterId))

	d.Set("cluster_id", cluster.CacheClusterId)
	d.Set("node_type", cluster.CacheNodeType)
	d.Set("num_cache_nodes", cluster.NumCacheNodes)
	d.Set("subnet_group_name", cluster.CacheSubnetGroupName)
	d.Set("engine", cluster.Engine)
	d.Set("engine_version", cluster.EngineVersion)
	d.Set("ip_discovery", cluster.IpDiscovery)
	d.Set("network_type", cluster.NetworkType)
	d.Set("preferred_outpost_arn", cluster.PreferredOutpostArn)
	d.Set("security_group_names", flattenSecurityGroupNames(cluster.CacheSecurityGroups))
	d.Set("security_group_ids", flattenSecurityGroupIDs(cluster.SecurityGroups))

	if cluster.CacheParameterGroup != nil {
		d.Set("parameter_group_name", cluster.CacheParameterGroup.CacheParameterGroupName)
	}

	d.Set("replication_group_id", cluster.ReplicationGroupId)

	d.Set("log_delivery_configuration", flattenLogDeliveryConfigurations(cluster.LogDeliveryConfigurations))
	d.Set("maintenance_window", cluster.PreferredMaintenanceWindow)
	d.Set("snapshot_window", cluster.SnapshotWindow)
	d.Set("snapshot_retention_limit", cluster.SnapshotRetentionLimit)
	d.Set("availability_zone", cluster.PreferredAvailabilityZone)

	if cluster.NotificationConfiguration != nil {
		if aws.StringValue(cluster.NotificationConfiguration.TopicStatus) == "active" {
			d.Set("notification_topic_arn", cluster.NotificationConfiguration.TopicArn)
		}
	}

	if cluster.ConfigurationEndpoint != nil {
		d.Set("port", cluster.ConfigurationEndpoint.Port)
		d.Set("configuration_endpoint", aws.String(fmt.Sprintf("%s:%d", *cluster.ConfigurationEndpoint.Address, *cluster.ConfigurationEndpoint.Port)))
		d.Set("cluster_address", aws.String(*cluster.ConfigurationEndpoint.Address))
	}

	if err := setCacheNodeData(d, cluster); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Cache Cluster (%s): %s", clusterID, err)
	}

	d.Set("arn", cluster.ARN)

	tags, err := ListTags(ctx, conn, aws.StringValue(cluster.ARN))

	if err != nil && !verify.ErrorISOUnsupported(conn.PartitionID, err) {
		return sdkdiag.AppendErrorf(diags, "listing tags for ElastiCache Cluster (%s): %s", d.Id(), err)
	}

	if err != nil {
		log.Printf("[WARN] error listing tags for ElastiCache Cluster (%s): %s", d.Id(), err)
	}

	if tags != nil {
		if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}
	}

	return diags
}
