// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_elasticache_replication_group", name="Replication Group")
func dataSourceReplicationGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReplicationGroupRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_token_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"automatic_failover_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"cluster_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
			"member_clusters": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"multi_az_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"num_cache_clusters": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"num_node_groups": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"primary_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reader_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateReplicationGroupID,
			},
			"replicas_per_node_group": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"snapshot_retention_limit": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"snapshot_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceReplicationGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	rg, err := findReplicationGroupByID(ctx, conn, d.Get("replication_group_id").(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("ElastiCache Replication Group", err))
	}

	d.SetId(aws.ToString(rg.ReplicationGroupId))
	d.Set(names.AttrDescription, rg.Description)
	d.Set(names.AttrARN, rg.ARN)
	d.Set("auth_token_enabled", rg.AuthTokenEnabled)

	switch rg.AutomaticFailover {
	case awstypes.AutomaticFailoverStatusDisabled, awstypes.AutomaticFailoverStatusDisabling:
		d.Set("automatic_failover_enabled", false)
	case awstypes.AutomaticFailoverStatusEnabled, awstypes.AutomaticFailoverStatusEnabling:
		d.Set("automatic_failover_enabled", true)
	}

	switch rg.MultiAZ {
	case awstypes.MultiAZStatusEnabled:
		d.Set("multi_az_enabled", true)
	case awstypes.MultiAZStatusDisabled:
		d.Set("multi_az_enabled", false)
	default:
		log.Printf("Unknown MultiAZ state %q", string(rg.MultiAZ))
	}

	if rg.ConfigurationEndpoint != nil {
		d.Set(names.AttrPort, rg.ConfigurationEndpoint.Port)
		d.Set("configuration_endpoint_address", rg.ConfigurationEndpoint.Address)
	} else {
		if rg.NodeGroups == nil {
			d.SetId("")
			return sdkdiag.AppendErrorf(diags, "ElastiCache Replication Group (%s) doesn't have node groups", aws.ToString(rg.ReplicationGroupId))
		}
		d.Set(names.AttrPort, rg.NodeGroups[0].PrimaryEndpoint.Port)
		d.Set("primary_endpoint_address", rg.NodeGroups[0].PrimaryEndpoint.Address)
		d.Set("reader_endpoint_address", rg.NodeGroups[0].ReaderEndpoint.Address)
	}

	d.Set("num_cache_clusters", len(rg.MemberClusters))
	if err := d.Set("member_clusters", flex.FlattenStringValueList(rg.MemberClusters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting member_clusters: %s", err)
	}
	d.Set("node_type", rg.CacheNodeType)
	d.Set("num_node_groups", len(rg.NodeGroups))
	d.Set("replicas_per_node_group", len(rg.NodeGroups[0].NodeGroupMembers)-1)
	d.Set("cluster_mode", rg.ClusterMode)
	d.Set("log_delivery_configuration", flattenLogDeliveryConfigurations(rg.LogDeliveryConfigurations))
	d.Set("snapshot_window", rg.SnapshotWindow)
	d.Set("snapshot_retention_limit", rg.SnapshotRetentionLimit)

	return diags
}
