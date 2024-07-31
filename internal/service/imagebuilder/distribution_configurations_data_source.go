// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_imagebuilder_distribution_configurations")
func DataSourceDistributionConfigurations() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDistributionConfigurationsRead,
		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrFilter: namevaluesfilters.Schema(),
			names.AttrNames: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDistributionConfigurationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.ListDistributionConfigurationsInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).ImagebuilderFilters()
	}

	var results []*imagebuilder.DistributionConfigurationSummary

	err := conn.ListDistributionConfigurationsPagesWithContext(ctx, input, func(page *imagebuilder.ListDistributionConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, distributionConfigurationSummary := range page.DistributionConfigurationSummaryList {
			if distributionConfigurationSummary == nil {
				continue
			}

			results = append(results, distributionConfigurationSummary)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Distribution Configurations: %s", err)
	}

	var arns, nms []string

	for _, r := range results {
		arns = append(arns, aws.StringValue(r.Arn))
		nms = append(nms, aws.StringValue(r.Name))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrARNs, arns)
	d.Set(names.AttrNames, nms)

	return diags
}
