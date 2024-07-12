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

// @SDKDataSource("aws_wafv2_web_acl", name="Web ACL")
func dataSourceWebACL() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWebACLRead,

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
				names.AttrScope: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.Scope](),
				},
			}
		},
	}
}

func dataSourceWebACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)
	name := d.Get(names.AttrName).(string)

	var foundWebACL awstypes.WebACLSummary
	input := &wafv2.ListWebACLsInput{
		Scope: awstypes.Scope(d.Get(names.AttrScope).(string)),
		Limit: aws.Int32(100),
	}

	for {
		resp, err := conn.ListWebACLs(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACLs: %s", err)
		}

		if resp == nil || resp.WebACLs == nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACLs")
		}

		for _, webACL := range resp.WebACLs {
			if aws.ToString(webACL.Name) == name {
				foundWebACL = webACL
				break
			}
		}

		if resp.NextMarker == nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundWebACL.Id == nil {
		return sdkdiag.AppendErrorf(diags, "WAFv2 WebACL not found for name: %s", name)
	}

	d.SetId(aws.ToString(foundWebACL.Id))
	d.Set(names.AttrARN, foundWebACL.ARN)
	d.Set(names.AttrDescription, foundWebACL.Description)

	return diags
}
