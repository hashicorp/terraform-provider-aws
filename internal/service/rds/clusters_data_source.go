// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_clusters")
func DataSourceClusters() *schema.Resource {
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

const (
	DSNameClusters = "Clusters Data Source"
)

func dataSourceClustersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeDBClustersInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).RDSFilters()
	}

	var clusterArns []string
	var clusterIdentifiers []string

	err := conn.DescribeDBClustersPagesWithContext(ctx, input, func(page *rds.DescribeDBClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, dbCluster := range page.DBClusters {
			if dbCluster == nil {
				continue
			}

			clusterArns = append(clusterArns, aws.StringValue(dbCluster.DBClusterArn))
			clusterIdentifiers = append(clusterIdentifiers, aws.StringValue(dbCluster.DBClusterIdentifier))
		}

		return !lastPage
	})
	if err != nil {
		return create.AppendDiagError(diags, names.RDS, create.ErrActionReading, DSNameClusters, "", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("cluster_arns", clusterArns)
	d.Set("cluster_identifiers", clusterIdentifiers)

	return diags
}
