// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	tagRootElementSchemaLevel = 2
)

// @SDKDataSource("aws_ce_tags", name="Tags")
func dataSourceTags() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTagsRead,

		Schema: map[string]*schema.Schema{
			names.AttrFilter: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     expressionElem(tagRootElementSchemaLevel),
			},
			"search_string": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringLenBetween(1, 1024),
				ConflictsWith: []string{"sort_by"},
			},
			"sort_by": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"search_string"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.Metric](),
						},
						"sort_order": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SortOrder](),
						},
					},
				},
			},
			"tag_key": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrTags: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"time_period": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"end": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 40),
						},
						"start": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 40),
						},
					},
				},
			},
		},
	}
}

func dataSourceTagsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	input := &costexplorer.GetTagsInput{
		TimePeriod: expandTagsTimePeriod(d.Get("time_period").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.GetOk(names.AttrFilter); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Filter = expandExpression(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("search_string"); ok {
		input.SearchString = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sort_by"); ok {
		input.SortBy = expandTagsSortBys(v.([]interface{}))
	}

	if v, ok := d.GetOk("tag_key"); ok {
		input.TagKey = aws.String(v.(string))
	}

	output, err := conn.GetTags(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cost Explorer Tags: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)
	d.Set(names.AttrTags, output.Tags)

	return diags
}

func expandTagsSortBys(tfList []interface{}) []awstypes.SortDefinition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.SortDefinition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandTagsSortBy(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandTagsSortBy(tfMap map[string]interface{}) awstypes.SortDefinition {
	apiObject := awstypes.SortDefinition{}
	apiObject.Key = aws.String(tfMap[names.AttrKey].(string))
	if v, ok := tfMap["sort_order"]; ok {
		apiObject.SortOrder = awstypes.SortOrder(v.(string))
	}

	return apiObject
}

func expandTagsTimePeriod(tfMap map[string]interface{}) *awstypes.DateInterval {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DateInterval{}
	apiObject.Start = aws.String(tfMap["start"].(string))
	apiObject.End = aws.String(tfMap["end"].(string))

	return apiObject
}
