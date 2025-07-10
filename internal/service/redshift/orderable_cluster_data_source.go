// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_redshift_orderable_cluster", name="Orderable Cluster")
func dataSourceOrderableCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrderableClusterRead,

		Schema: map[string]*schema.Schema{
			names.AttrAvailabilityZones: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"cluster_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"node_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"preferred_node_types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceOrderableClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	input := &redshift.DescribeOrderableClusterOptionsInput{}

	if v, ok := d.GetOk("cluster_version"); ok {
		input.ClusterVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("node_type"); ok {
		input.NodeType = aws.String(v.(string))
	}

	var orderableClusterOptions []awstypes.OrderableClusterOption

	pages := redshift.NewDescribeOrderableClusterOptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Redshift Orderable Cluster Options: %s", err)
		}

		for _, orderableClusterOption := range page.OrderableClusterOptions {
			if v, ok := d.GetOk("cluster_type"); ok {
				if aws.ToString(orderableClusterOption.ClusterType) != v.(string) {
					continue
				}
			}

			orderableClusterOptions = append(orderableClusterOptions, orderableClusterOption)
		}
	}

	if len(orderableClusterOptions) == 0 {
		return sdkdiag.AppendErrorf(diags, "no Redshift Orderable Cluster Options found matching criteria; try different search")
	}

	var orderableClusterOption awstypes.OrderableClusterOption
	preferredNodeTypes := d.Get("preferred_node_types").([]any)
	if len(preferredNodeTypes) > 0 {
	listNodeTypes:
		for _, preferredNodeTypeRaw := range preferredNodeTypes {
			preferredNodeType, ok := preferredNodeTypeRaw.(string)

			if !ok {
				continue
			}

			for _, option := range orderableClusterOptions {
				if preferredNodeType == aws.ToString(option.NodeType) {
					orderableClusterOption = option
					break listNodeTypes
				}
			}
		}
	}

	if orderableClusterOption.NodeType == nil && len(orderableClusterOptions) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple Redshift Orderable Cluster Options (%v) match the criteria; try a different search", orderableClusterOptions)
	}

	if orderableClusterOption.NodeType == nil && len(orderableClusterOptions) == 1 {
		orderableClusterOption = orderableClusterOptions[0]
	}

	if orderableClusterOption.NodeType == nil {
		return sdkdiag.AppendErrorf(diags, "no Redshift Orderable Cluster Options match the criteria; try a different search")
	}

	d.SetId(aws.ToString(orderableClusterOption.NodeType))

	var availabilityZones []string
	for _, az := range orderableClusterOption.AvailabilityZones {
		availabilityZones = append(availabilityZones, aws.ToString(az.Name))
	}
	d.Set(names.AttrAvailabilityZones, availabilityZones)

	d.Set("cluster_type", orderableClusterOption.ClusterType)
	d.Set("cluster_version", orderableClusterOption.ClusterVersion)
	d.Set("node_type", orderableClusterOption.NodeType)

	return diags
}
