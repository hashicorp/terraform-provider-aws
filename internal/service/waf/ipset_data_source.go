// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_waf_ipset", name="IPSet")
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
	conn := meta.(*conns.AWSClient).WAFClient(ctx)
	name := d.Get(names.AttrName).(string)

	ipsets := make([]awstypes.IPSetSummary, 0)
	// ListIPSetsInput does not have a name parameter for filtering or a paginator
	input := &waf.ListIPSetsInput{}
	for {
		output, err := conn.ListIPSets(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAF IP sets: %s", err)
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
		return sdkdiag.AppendErrorf(diags, "WAF IP Set not found for name: %s", name)
	}
	if len(ipsets) > 1 {
		return sdkdiag.AppendErrorf(diags, "Multiple WAF IP Sets found for name: %s", name)
	}

	ipset := ipsets[0]
	d.SetId(aws.ToString(ipset.IPSetId))

	return diags
}
