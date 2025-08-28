// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/memorydb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_memorydb_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"acl_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAutoMinorVersionUpgrade: {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"cluster_endpoint": endpointSchema(),
				"data_tiering": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"engine_patch_version": {
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
				"final_snapshot_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrKMSKeyARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"maintenance_window": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
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
				names.AttrParameterGroupName: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrPort: {
					Type:     schema.TypeInt,
					Computed: true,
				},
				names.AttrSecurityGroupIDs: {
					Type:     schema.TypeSet,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"shards": {
					Type:     schema.TypeSet,
					Computed: true,
					Set:      clusterShardHash,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrName: {
								Type:     schema.TypeString,
								Computed: true,
							},
							"nodes": {
								Type:     schema.TypeSet,
								Computed: true,
								Set:      clusterShardNodeHash,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrAvailabilityZone: {
											Type:     schema.TypeString,
											Computed: true,
										},
										names.AttrCreateTime: {
											Type:     schema.TypeString,
											Computed: true,
										},
										names.AttrEndpoint: endpointSchema(),
										names.AttrName: {
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
				names.AttrSNSTopicARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"subnet_group_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags: tftags.TagsSchemaComputed(),
				"tls_enabled": {
					Type:     schema.TypeBool,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	name := d.Get(names.AttrName).(string)
	cluster, err := findClusterByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("MemoryDB Cluster", err))
	}

	d.SetId(aws.ToString(cluster.Name))
	d.Set("acl_name", cluster.ACLName)
	d.Set(names.AttrARN, cluster.ARN)
	d.Set(names.AttrAutoMinorVersionUpgrade, cluster.AutoMinorVersionUpgrade)
	if v := cluster.ClusterEndpoint; v != nil {
		d.Set("cluster_endpoint", flattenEndpoint(v))
		d.Set(names.AttrPort, v.Port)
	}
	if v := string(cluster.DataTiering); v != "" {
		v, err := strconv.ParseBool(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set("data_tiering", v)
	}
	d.Set(names.AttrDescription, cluster.Description)
	d.Set("engine_patch_version", cluster.EnginePatchVersion)
	d.Set(names.AttrEngine, cluster.Engine)
	d.Set(names.AttrEngineVersion, cluster.EngineVersion)
	d.Set(names.AttrKMSKeyARN, cluster.KmsKeyId) // KmsKeyId is actually an ARN here.
	d.Set("maintenance_window", cluster.MaintenanceWindow)
	d.Set(names.AttrName, cluster.Name)
	d.Set("node_type", cluster.NodeType)
	if v, err := deriveClusterNumReplicasPerShard(cluster); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else {
		d.Set("num_replicas_per_shard", v)
	}
	d.Set("num_shards", cluster.NumberOfShards)
	d.Set(names.AttrParameterGroupName, cluster.ParameterGroupName)
	d.Set(names.AttrSecurityGroupIDs, tfslices.ApplyToAll(cluster.SecurityGroups, func(v awstypes.SecurityGroupMembership) string {
		return aws.ToString(v.SecurityGroupId)
	}))
	if err := d.Set("shards", flattenShards(cluster.Shards)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting shards: %s", err)
	}
	d.Set("snapshot_retention_limit", cluster.SnapshotRetentionLimit)
	d.Set("snapshot_window", cluster.SnapshotWindow)
	if aws.ToString(cluster.SnsTopicStatus) == clusterSNSTopicStatusActive {
		d.Set(names.AttrSNSTopicARN, cluster.SnsTopicArn)
	} else {
		d.Set(names.AttrSNSTopicARN, "")
	}
	d.Set("subnet_group_name", cluster.SubnetGroupName)
	d.Set("tls_enabled", cluster.TLSEnabled)

	return diags
}
