// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_memorydb_cluster")
func DataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"acl_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cluster_endpoint": endpointSchema(),
			"data_tiering": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_patch_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"final_snapshot_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maintenance_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"num_replicas_per_shard": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"num_shards": {
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
			"security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"shards": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      shardHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"nodes": {
							Type:     schema.TypeSet,
							Computed: true,
							Set:      nodeHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"create_time": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"endpoint": endpointSchema(),
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"num_nodes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"slots": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"snapshot_retention_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"snapshot_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sns_topic_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"tls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	cluster, err := FindClusterByName(ctx, conn, name)

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("MemoryDB Cluster", err))
	}

	d.SetId(aws.StringValue(cluster.Name))

	d.Set("acl_name", cluster.ACLName)
	d.Set("arn", cluster.ARN)
	d.Set("auto_minor_version_upgrade", cluster.AutoMinorVersionUpgrade)

	if v := cluster.ClusterEndpoint; v != nil {
		d.Set("cluster_endpoint", flattenEndpoint(v))
		d.Set("port", v.Port)
	}

	if v := aws.StringValue(cluster.DataTiering); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return diag.Errorf("reading data_tiering for MemoryDB Cluster (%s): %s", d.Id(), err)
		}

		d.Set("data_tiering", b)
	}

	d.Set("description", cluster.Description)
	d.Set("engine_patch_version", cluster.EnginePatchVersion)
	d.Set("engine_version", cluster.EngineVersion)
	d.Set("kms_key_arn", cluster.KmsKeyId) // KmsKeyId is actually an ARN here.
	d.Set("maintenance_window", cluster.MaintenanceWindow)
	d.Set("name", cluster.Name)
	d.Set("node_type", cluster.NodeType)

	numReplicasPerShard, err := deriveClusterNumReplicasPerShard(cluster)
	if err != nil {
		return diag.Errorf("reading num_replicas_per_shard for MemoryDB Cluster (%s): %s", d.Id(), err)
	}
	d.Set("num_replicas_per_shard", numReplicasPerShard)

	d.Set("num_shards", cluster.NumberOfShards)
	d.Set("parameter_group_name", cluster.ParameterGroupName)

	var securityGroupIds []*string
	for _, v := range cluster.SecurityGroups {
		securityGroupIds = append(securityGroupIds, v.SecurityGroupId)
	}
	d.Set("security_group_ids", flex.FlattenStringSet(securityGroupIds))

	if err := d.Set("shards", flattenShards(cluster.Shards)); err != nil {
		return diag.Errorf("failed to set shards for MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	d.Set("snapshot_retention_limit", cluster.SnapshotRetentionLimit)
	d.Set("snapshot_window", cluster.SnapshotWindow)

	if aws.StringValue(cluster.SnsTopicStatus) == ClusterSNSTopicStatusActive {
		d.Set("sns_topic_arn", cluster.SnsTopicArn)
	} else {
		d.Set("sns_topic_arn", "")
	}

	d.Set("subnet_group_name", cluster.SubnetGroupName)
	d.Set("tls_enabled", cluster.TLSEnabled)

	tags, err := listTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("listing tags for MemoryDB Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
