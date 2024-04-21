// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/fis"
	"github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_fis_experiment_templates")
func DataSourceExperimentTemplates() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceExperimentTemplatesRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceExperimentTemplatesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).FISClient(ctx)

	input := &fis.ListExperimentTemplatesInput{}

	var inputTags map[string]string

	if tags, tagsOk := d.GetOk("tags"); tagsOk && len(tags.(map[string]interface{})) > 0 {
		inputTags = Tags(tftags.New(ctx, tags.(map[string]interface{})))
	}

	var output []types.ExperimentTemplateSummary

	pages := fis.NewListExperimentTemplatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading FIS Experiment Templates: %s", err)
		}

		for _, result := range page.ExperimentTemplates {
			if len(inputTags) > 0 {
				if IsSubset(inputTags, result.Tags) {
					output = append(output, result)
				}
			} else {
				output = append(output, result)
			}
		}
	}

	var expIds []string

	for _, exp := range output {
		expIds = append(expIds, aws.StringValue(exp.Id))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", expIds)

	return diags
}

func IsSubset(subset map[string]string, superset map[string]string) bool {
	if len(subset) > len(superset) {
		return false
	}

	for k, v := range subset {
		if supersetValue, ok := superset[k]; !ok || supersetValue != v {
			return false
		}
	}
	return true
}
