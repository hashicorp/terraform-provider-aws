package elasticache

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(v interface{}) string {
					value := v.(string)
					return strings.ToLower(value)
				},
			},

			"node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"num_cache_nodes": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"subnet_group_name": {
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

			"parameter_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"replication_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_group_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"log_delivery_configuration": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"destination": {
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

			"snapshot_window": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"snapshot_retention_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"notification_topic_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"configuration_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cluster_address": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cache_nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterID := d.Get("cluster_id").(string)
	cluster, err := FindCacheClusterWithNodeInfoByID(conn, clusterID)
	if tfresource.NotFound(err) {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again")
	}
	if err != nil {
		return fmt.Errorf("error reading ElastiCache Cache Cluster (%s): %w", clusterID, err)
	}

	d.SetId(aws.StringValue(cluster.CacheClusterId))

	d.Set("cluster_id", cluster.CacheClusterId)
	d.Set("node_type", cluster.CacheNodeType)
	d.Set("num_cache_nodes", cluster.NumCacheNodes)
	d.Set("subnet_group_name", cluster.CacheSubnetGroupName)
	d.Set("engine", cluster.Engine)
	d.Set("engine_version", cluster.EngineVersion)
	d.Set("security_group_names", flattenSecurityGroupNames(cluster.CacheSecurityGroups))
	d.Set("security_group_ids", flattenSecurityGroupIDs(cluster.SecurityGroups))

	if cluster.CacheParameterGroup != nil {
		d.Set("parameter_group_name", cluster.CacheParameterGroup.CacheParameterGroupName)
	}

	if cluster.ReplicationGroupId != nil {
		d.Set("replication_group_id", cluster.ReplicationGroupId)
	}

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
		return err
	}

	d.Set("arn", cluster.ARN)

	tags, err := ListTags(conn, aws.StringValue(cluster.ARN))

	if err != nil && !verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		return fmt.Errorf("error listing tags for Elasticache Cluster (%s): %w", d.Id(), err)
	}

	if err != nil {
		log.Printf("[WARN] error listing tags for Elasticache Cluster (%s): %s", d.Id(), err)
	}

	if tags != nil {
		if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}
	}

	return nil
}
