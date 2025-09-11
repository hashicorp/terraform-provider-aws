// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssm_parameters", name="Parameters")
func dataSourceParameters() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceParametersRead,

		Schema: map[string]*schema.Schema{
			"parameter_filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						"option": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"shared": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrARNs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceParametersRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	input := &ssm.DescribeParametersInput{}

	if v, ok := d.GetOk("parameter_filter"); ok {
		input.ParameterFilters = expandParameterStringFilterFilters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("shared"); ok {
		input.Shared = aws.Bool(v.(bool))
	}

	var output []awstypes.ParameterMetadata

	pages := ssm.NewDescribeParametersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "describing SSM Parameters: %s", err)
		}

		output = append(output, page.Parameters...)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrARNs, tfslices.ApplyToAll(output, func(v awstypes.ParameterMetadata) string {
		return aws.ToString(v.ARN)
	}))

	return diags
}

func expandParameterStringFilterFilters(tfList []any) []awstypes.ParameterStringFilter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ParameterStringFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandParameterStringFilterFilter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandParameterStringFilterFilter(tfMap map[string]any) *awstypes.ParameterStringFilter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ParameterStringFilter{}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["option"].(string); ok && v != "" {
		apiObject.Option = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandStringValueList(v)
	}

	return apiObject
}
