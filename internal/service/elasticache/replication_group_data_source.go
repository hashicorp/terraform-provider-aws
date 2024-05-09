// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
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
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	groupID := d.Get("replication_group_id").(string)

	rg, err := findReplicationGroupByID(ctx, conn, groupID)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("ElastiCache Replication Group", err))
	}

	d.SetId(aws.StringValue(rg.ReplicationGroupId))
	d.Set(names.AttrDescription, rg.Description)
	d.Set(names.AttrARN, rg.ARN)
	d.Set("auth_token_enabled", rg.AuthTokenEnabled)

	if rg.AutomaticFailover != nil {
		switch aws.StringValue(rg.AutomaticFailover) {
		case elasticache.AutomaticFailoverStatusDisabled, elasticache.AutomaticFailoverStatusDisabling:
			d.Set("automatic_failover_enabled", false)
		case elasticache.AutomaticFailoverStatusEnabled, elasticache.AutomaticFailoverStatusEnabling:
			d.Set("automatic_failover_enabled", true)
		}
	}

	if rg.MultiAZ != nil {
		switch strings.ToLower(aws.StringValue(rg.MultiAZ)) {
		case elasticache.MultiAZStatusEnabled:
			d.Set("multi_az_enabled", true)
		case elasticache.MultiAZStatusDisabled:
			d.Set("multi_az_enabled", false)
		default:
			log.Printf("Unknown MultiAZ state %q", aws.StringValue(rg.MultiAZ))
		}
	}

	if rg.ConfigurationEndpoint != nil {
		d.Set(names.AttrPort, rg.ConfigurationEndpoint.Port)
		d.Set("configuration_endpoint_address", rg.ConfigurationEndpoint.Address)
	} else {
		if rg.NodeGroups == nil {
			d.SetId("")
			return sdkdiag.AppendErrorf(diags, "ElastiCache Replication Group (%s) doesn't have node groups", aws.StringValue(rg.ReplicationGroupId))
		}
		d.Set(names.AttrPort, rg.NodeGroups[0].PrimaryEndpoint.Port)
		d.Set("primary_endpoint_address", rg.NodeGroups[0].PrimaryEndpoint.Address)
		d.Set("reader_endpoint_address", rg.NodeGroups[0].ReaderEndpoint.Address)
	}

	d.Set("num_cache_clusters", len(rg.MemberClusters))
	if err := d.Set("member_clusters", flex.FlattenStringList(rg.MemberClusters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting member_clusters: %s", err)
	}
	d.Set("node_type", rg.CacheNodeType)
	d.Set("num_node_groups", len(rg.NodeGroups))
	d.Set("replicas_per_node_group", len(rg.NodeGroups[0].NodeGroupMembers)-1)
	d.Set("log_delivery_configuration", flattenLogDeliveryConfigurations(rg.LogDeliveryConfigurations))
	d.Set("snapshot_window", rg.SnapshotWindow)
	d.Set("snapshot_retention_limit", rg.SnapshotRetentionLimit)

	return diags
}
