// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_fis_experiment_templates", name="Experiment Templates")
func DataSourceExperimentTemplates() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceExperimentTemplatesRead,

		Schema: map[string]*schema.Schema{
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchema(),
		},
	}
}

func dataSourceExperimentTemplatesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FISClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	input := &fis.ListExperimentTemplatesInput{}

	filter := tfslices.PredicateTrue[*awstypes.ExperimentTemplateSummary]()
	if tagsToMatch := tftags.New(ctx, d.Get(names.AttrTags).(map[string]any)).IgnoreAWS().IgnoreConfig(ignoreTagsConfig); len(tagsToMatch) > 0 {
		filter = func(v *awstypes.ExperimentTemplateSummary) bool {
			return keyValueTags(ctx, v.Tags).ContainsAll(tagsToMatch)
		}
	}

	output, err := findExperimentTemplates(ctx, conn, input, filter)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FIS Experiment Templates: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrIDs, tfslices.ApplyToAll(output, func(v awstypes.ExperimentTemplateSummary) string {
		return aws.ToString(v.Id)
	}))

	return diags
}

func findExperimentTemplates(ctx context.Context, conn *fis.Client, input *fis.ListExperimentTemplatesInput, filter tfslices.Predicate[*awstypes.ExperimentTemplateSummary]) ([]awstypes.ExperimentTemplateSummary, error) {
	var output []awstypes.ExperimentTemplateSummary

	pages := fis.NewListExperimentTemplatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ExperimentTemplates {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
