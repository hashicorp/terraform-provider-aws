// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_elasticache_clusters", name="Clusters")
func dataSourceClusters() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClustersRead,

		Schema: map[string]*schema.Schema{
			"cluster_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"cluster_identifiers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: namevaluesfilters.Schema(),
		},
	}
}

func dataSourceClustersRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	input := &elasticache.DescribeCacheClustersInput{}

	clusters, err := findCacheClusters(ctx, conn, input, tfslices.PredicateTrue[*awstypes.CacheCluster]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticache Clusters: %s", err)
	}

	var clusterARNs []string
	var clusterIdentifiers []string

	for _, cluster := range clusters {
		clusterARNs = append(clusterARNs, aws.ToString(cluster.ARN))
		clusterIdentifiers = append(clusterIdentifiers, aws.ToString(cluster.CacheClusterId))
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set("cluster_arns", clusterARNs)
	d.Set("cluster_identifiers", clusterIdentifiers)

	return diags
}
