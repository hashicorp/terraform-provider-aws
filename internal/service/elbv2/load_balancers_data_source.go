// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lbs", name="Load Balancers")
func dataSourceLoadBalancers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLoadBalancersRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchema(),
		},
	}
}

const (
	DSNameLoadBalancers = "Load Balancers Data Source"
)

func dataSourceLoadBalancersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	results, err := findLoadBalancers(ctx, conn, &elasticloadbalancingv2.DescribeLoadBalancersInput{})

	if err != nil {
		return create.AppendDiagError(diags, names.ELBV2, create.ErrActionReading, DSNameLoadBalancers, "", err)
	}

	tagsToMatch := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	if len(tagsToMatch) > 0 {
		var loadBalancers []awstypes.LoadBalancer

		for _, loadBalancer := range results {
			arn := aws.StringValue(loadBalancer.LoadBalancerArn)
			tags, err := listTags(ctx, conn, arn)

			if errs.IsA[*awstypes.LoadBalancerNotFoundException](err) {
				continue
			}

			if err != nil {
				return create.AppendDiagError(diags, names.ELBV2, "listing tags", DSNameLoadBalancers, arn, err)
			}
			if !tags.ContainsAll(tagsToMatch) {
				continue
			}

			loadBalancers = append(loadBalancers, loadBalancer)
		}

		results = loadBalancers
	}

	var loadBalancerARNs []string
	for _, lb := range results {
		loadBalancerARNs = append(loadBalancerARNs, aws.StringValue(lb.LoadBalancerArn))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrARNs, loadBalancerARNs)

	return diags
}
