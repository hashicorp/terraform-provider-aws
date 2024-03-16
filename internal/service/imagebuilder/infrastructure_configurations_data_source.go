// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfiltersv2"
)

// @SDKDataSource("aws_imagebuilder_infrastructure_configurations")
func DataSourceInfrastructureConfigurations() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInfrastructureConfigurationsRead,
		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": namevaluesfiltersv2.Schema(),
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInfrastructureConfigurationsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.ListInfrastructureConfigurationsInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = namevaluesfiltersv2.New(v.(*schema.Set)).ImagebuilderFilters()
	}

	out, err := conn.ListInfrastructureConfigurations(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Infrastructure Configurations: %s", err)
	}

	var arns, names []string

	for _, r := range out.InfrastructureConfigurationSummaryList {
		arns = append(arns, aws.ToString(r.Arn))
		names = append(names, aws.ToString(r.Name))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("arns", arns)
	d.Set("names", names)

	return diags
}
