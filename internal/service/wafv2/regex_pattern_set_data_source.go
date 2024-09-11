// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_wafv2_regex_pattern_set", name="Regex Pattern Set")
func dataSourceRegexPatternSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRegexPatternSetRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				"regular_expression": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"regex_string": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				names.AttrScope: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.Scope](),
				},
			}
		},
	}
}

func dataSourceRegexPatternSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)
	name := d.Get(names.AttrName).(string)

	var foundRegexPatternSet awstypes.RegexPatternSetSummary
	input := &wafv2.ListRegexPatternSetsInput{
		Scope: awstypes.Scope(d.Get(names.AttrScope).(string)),
		Limit: aws.Int32(100),
	}

	for {
		resp, err := conn.ListRegexPatternSets(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 RegexPatternSets: %s", err)
		}

		if resp == nil || resp.RegexPatternSets == nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 RegexPatternSets")
		}

		for _, regexPatternSet := range resp.RegexPatternSets {
			if aws.ToString(regexPatternSet.Name) == name {
				foundRegexPatternSet = regexPatternSet
				break
			}
		}

		if resp.NextMarker == nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundRegexPatternSet.Id == nil {
		return sdkdiag.AppendErrorf(diags, "WAFv2 RegexPatternSet not found for name: %s", name)
	}

	resp, err := conn.GetRegexPatternSet(ctx, &wafv2.GetRegexPatternSetInput{
		Id:    foundRegexPatternSet.Id,
		Name:  foundRegexPatternSet.Name,
		Scope: awstypes.Scope(d.Get(names.AttrScope).(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 RegexPatternSet: %s", err)
	}

	if resp == nil || resp.RegexPatternSet == nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 RegexPatternSet")
	}

	d.SetId(aws.ToString(resp.RegexPatternSet.Id))
	d.Set(names.AttrARN, resp.RegexPatternSet.ARN)
	d.Set(names.AttrDescription, resp.RegexPatternSet.Description)

	if err := d.Set("regular_expression", flattenRegexPatternSet(resp.RegexPatternSet.RegularExpressionList)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting regular_expression: %s", err)
	}

	return diags
}
