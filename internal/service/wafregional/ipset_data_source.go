// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_wafregional_ipset", name="IPSet")
func dataSourceIPSet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceIPSetRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	name := d.Get(names.AttrName).(string)

	ipsets := make([]awstypes.IPSetSummary, 0)
	// ListIPSetsInput does not have a name parameter for filtering or a paginator
	input := &wafregional.ListIPSetsInput{}
	for {
		output, err := conn.ListIPSets(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAF Regional IP sets: %s", err)
		}
		for _, ipset := range output.IPSets {
			if aws.ToString(ipset.Name) == name {
				ipsets = append(ipsets, ipset)
			}
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if len(ipsets) == 0 {
		return sdkdiag.AppendErrorf(diags, "WAF Regional IP Set not found for name: %s", name)
	}
	if len(ipsets) > 1 {
		return sdkdiag.AppendErrorf(diags, "Multiple WAF Regional IP Sets found for name: %s", name)
	}

	ipset := ipsets[0]
	d.SetId(aws.ToString(ipset.IPSetId))

	return diags
}
