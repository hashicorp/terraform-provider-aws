// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lbs")
func DataSourceLoadBalancers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLoadBalancersRead,
		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchema(),
		},
	}
}

const (
	DSNameLoadBalancers = "Load Balancers Data Source"
)

func dataSourceLoadBalancersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	results, err := FindLoadBalancers(ctx, conn, &elbv2.DescribeLoadBalancersInput{})

	if err != nil {
		return create.DiagError(names.ELBV2, create.ErrActionReading, DSNameLoadBalancers, "", err)
	}

	tagsToMatch := tftags.New(ctx, d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	if len(tagsToMatch) > 0 {
		var loadBalancers []*elbv2.LoadBalancer

		for _, loadBalancer := range results {
			arn := aws.StringValue(loadBalancer.LoadBalancerArn)
			tags, err := listTags(ctx, conn, arn)

			if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeLoadBalancerNotFoundException) {
				continue
			}

			if err != nil {
				return create.DiagError(names.ELBV2, "listing tags", DSNameLoadBalancers, arn, err)
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
	d.Set("arns", loadBalancerARNs)

	return nil
}
