// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_redshift_orderable_cluster", name="Orderable Cluster Options")
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

func dataSourceOrderableClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	input := &redshift.DescribeOrderableClusterOptionsInput{}

	if v, ok := d.GetOk("cluster_version"); ok {
		input.ClusterVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("node_type"); ok {
		input.NodeType = aws.String(v.(string))
	}

	var orderableClusterOptions []*redshift.OrderableClusterOption

	err := conn.DescribeOrderableClusterOptionsPagesWithContext(ctx, input, func(page *redshift.DescribeOrderableClusterOptionsOutput, lastPage bool) bool {
		for _, orderableClusterOption := range page.OrderableClusterOptions {
			if orderableClusterOption == nil {
				continue
			}

			if v, ok := d.GetOk("cluster_type"); ok {
				if aws.StringValue(orderableClusterOption.ClusterType) != v.(string) {
					continue
				}
			}

			orderableClusterOptions = append(orderableClusterOptions, orderableClusterOption)
		}
		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Orderable Cluster Options: %s", err)
	}

	if len(orderableClusterOptions) == 0 {
		return sdkdiag.AppendErrorf(diags, "no Redshift Orderable Cluster Options found matching criteria; try different search")
	}

	var orderableClusterOption *redshift.OrderableClusterOption
	preferredNodeTypes := d.Get("preferred_node_types").([]interface{})
	if len(preferredNodeTypes) > 0 {
		for _, preferredNodeTypeRaw := range preferredNodeTypes {
			preferredNodeType, ok := preferredNodeTypeRaw.(string)

			if !ok {
				continue
			}

			for _, option := range orderableClusterOptions {
				if preferredNodeType == aws.StringValue(option.NodeType) {
					orderableClusterOption = option
					break
				}
			}

			if orderableClusterOption != nil {
				break
			}
		}
	}

	if orderableClusterOption == nil && len(orderableClusterOptions) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple Redshift Orderable Cluster Options (%v) match the criteria; try a different search", orderableClusterOptions)
	}

	if orderableClusterOption == nil && len(orderableClusterOptions) == 1 {
		orderableClusterOption = orderableClusterOptions[0]
	}

	if orderableClusterOption == nil {
		return sdkdiag.AppendErrorf(diags, "no Redshift Orderable Cluster Options match the criteria; try a different search")
	}

	d.SetId(aws.StringValue(orderableClusterOption.NodeType))

	var availabilityZones []string
	for _, az := range orderableClusterOption.AvailabilityZones {
		availabilityZones = append(availabilityZones, aws.StringValue(az.Name))
	}
	d.Set(names.AttrAvailabilityZones, availabilityZones)

	d.Set("cluster_type", orderableClusterOption.ClusterType)
	d.Set("cluster_version", orderableClusterOption.ClusterVersion)
	d.Set("node_type", orderableClusterOption.NodeType)

	return diags
}
