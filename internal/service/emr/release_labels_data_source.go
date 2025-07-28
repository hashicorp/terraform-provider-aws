// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_emr_release_labels", name="Release Labels")
func dataSourceReleaseLabels() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReleaseLabelsRead,

		Schema: map[string]*schema.Schema{
			"filters": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrPrefix: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"release_labels": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func dataSourceReleaseLabelsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	input := &emr.ListReleaseLabelsInput{}

	if v, ok := d.GetOk("filters"); ok && len(v.([]any)) > 0 {
		input.Filters = expandReleaseLabelsFilters(v.([]any))
	}

	releaseLabels, err := findReleaseLabels(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Release Labels: %s", err)
	}

	if len(releaseLabels) == 0 {
		d.SetId(",")
	} else {
		d.SetId(strings.Join(releaseLabels, ","))
	}
	d.Set("release_labels", releaseLabels)

	return diags
}

func expandReleaseLabelsFilters(filters []any) *awstypes.ReleaseLabelFilter {
	if len(filters) == 0 || filters[0] == nil {
		return nil
	}

	m := filters[0].(map[string]any)
	app := &awstypes.ReleaseLabelFilter{}

	if v, ok := m["application"].(string); ok && v != "" {
		app.Application = aws.String(v)
	}

	if v, ok := m[names.AttrPrefix].(string); ok && v != "" {
		app.Prefix = aws.String(v)
	}

	return app
}

func findReleaseLabels(ctx context.Context, conn *emr.Client, input *emr.ListReleaseLabelsInput) ([]string, error) {
	var output []string

	pages := emr.NewListReleaseLabelsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ReleaseLabels...)
	}

	return output, nil
}
