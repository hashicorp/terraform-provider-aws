// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_wafv2_regex_pattern_set")
func DataSourceRegexPatternSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRegexPatternSetRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"description": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
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
				"scope": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(wafv2.Scope_Values(), false),
				},
			}
		},
	}
}

func dataSourceRegexPatternSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)
	name := d.Get("name").(string)

	var foundRegexPatternSet *wafv2.RegexPatternSetSummary
	input := &wafv2.ListRegexPatternSetsInput{
		Scope: aws.String(d.Get("scope").(string)),
		Limit: aws.Int64(100),
	}

	for {
		resp, err := conn.ListRegexPatternSetsWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 RegexPatternSets: %s", err)
		}

		if resp == nil || resp.RegexPatternSets == nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 RegexPatternSets")
		}

		for _, regexPatternSet := range resp.RegexPatternSets {
			if regexPatternSet != nil && aws.StringValue(regexPatternSet.Name) == name {
				foundRegexPatternSet = regexPatternSet
				break
			}
		}

		if resp.NextMarker == nil || foundRegexPatternSet != nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundRegexPatternSet == nil {
		return sdkdiag.AppendErrorf(diags, "WAFv2 RegexPatternSet not found for name: %s", name)
	}

	resp, err := conn.GetRegexPatternSetWithContext(ctx, &wafv2.GetRegexPatternSetInput{
		Id:    foundRegexPatternSet.Id,
		Name:  foundRegexPatternSet.Name,
		Scope: aws.String(d.Get("scope").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 RegexPatternSet: %s", err)
	}

	if resp == nil || resp.RegexPatternSet == nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 RegexPatternSet")
	}

	d.SetId(aws.StringValue(resp.RegexPatternSet.Id))
	d.Set("arn", resp.RegexPatternSet.ARN)
	d.Set("description", resp.RegexPatternSet.Description)

	if err := d.Set("regular_expression", flattenRegexPatternSet(resp.RegexPatternSet.RegularExpressionList)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting regular_expression: %s", err)
	}

	return diags
}
