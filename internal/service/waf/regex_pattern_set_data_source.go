// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_waf_regex_pattern_set")
func DataSourceRegexPatternSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRegexPatternSetRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regex_pattern_strings": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceRegexPatternSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn(ctx)
	name := d.Get("name").(string)

	var foundRegexPatternSet *waf.RegexPatternSetSummary
	input := &waf.ListRegexPatternSetsInput{
		Limit: aws.Int64(100),
	}

	for {
		resp, err := conn.ListRegexPatternSetsWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAF RegexPatternSets: %s", err)
		}

		if resp == nil || resp.RegexPatternSets == nil {
			return sdkdiag.AppendErrorf(diags, "reading WAF RegexPatternSets")
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
		return sdkdiag.AppendErrorf(diags, "WAF RegexPatternSet not found for name: %s", name)
	}

	resp, err := conn.GetRegexPatternSetWithContext(ctx, &waf.GetRegexPatternSetInput{
		RegexPatternSetId: foundRegexPatternSet.RegexPatternSetId,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF RegexPatternSet: %s", err)
	}

	if resp == nil || resp.RegexPatternSet == nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF RegexPatternSet")
	}

	d.SetId(aws.StringValue(resp.RegexPatternSet.RegexPatternSetId))
	d.Set("name", resp.RegexPatternSet.Name)
	d.Set("regex_pattern_strings", aws.StringValueSlice(resp.RegexPatternSet.RegexPatternStrings))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("regexpatternset/%s", d.Id()),
	}
	d.Set("arn", arn.String())

	return diags
}
