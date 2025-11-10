// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_imagebuilder_distribution_configurations", name="Distribution Configurations")
func dataSourceDistributionConfigurations() *schema.Resource {
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

func dataSourceDistributionConfigurationsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.ListDistributionConfigurationsInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).ImageBuilderFilters()
	}

	distributionConfigurations, err := findDistributionConfigurations(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Distribution Configurations: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrARNs, tfslices.ApplyToAll(distributionConfigurations, func(v awstypes.DistributionConfigurationSummary) string {
		return aws.ToString(v.Arn)
	}))
	d.Set(names.AttrNames, tfslices.ApplyToAll(distributionConfigurations, func(v awstypes.DistributionConfigurationSummary) string {
		return aws.ToString(v.Name)
	}))

	return diags
}

func findDistributionConfigurations(ctx context.Context, conn *imagebuilder.Client, input *imagebuilder.ListDistributionConfigurationsInput) ([]awstypes.DistributionConfigurationSummary, error) {
	var output []awstypes.DistributionConfigurationSummary

	pages := imagebuilder.NewListDistributionConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.DistributionConfigurationSummaryList...)
	}

	return output, nil
}
