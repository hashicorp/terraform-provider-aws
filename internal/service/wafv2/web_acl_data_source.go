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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
					Type:         schema.TypeString,
					Optional:     true,
					ExactlyOneOf: []string{names.AttrName, "resource"},
				},
				"resource": {
					Type:         schema.TypeString,
					Optional:     true,
					ExactlyOneOf: []string{names.AttrName, "resource"},
					ValidateFunc: verify.ValidARN,
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

func dataSourceWebACLRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	name := d.Get(names.AttrName).(string)
	resourceArn := d.Get("resource").(string)
	scope := awstypes.Scope(d.Get(names.AttrScope).(string))

	var webACL *awstypes.WebACL
	var err error

	if resourceArn != "" {
		// Use GetWebACLForResource API
		webACL, err = findWebACLByResourceARN(ctx, conn, resourceArn)
		if err != nil {
			if tfresource.NotFound(err) {
				return sdkdiag.AppendErrorf(diags, "WAFv2 WebACL not found for resource: %s", resourceArn)
			}
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL for resource (%s): %s", resourceArn, err)
		}
	} else {
		// Use existing ListWebACLs + filter by name logic
		var foundWebACL awstypes.WebACLSummary
		input := &wafv2.ListWebACLsInput{
			Scope: scope,
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

			for _, acl := range resp.WebACLs {
				if aws.ToString(acl.Name) == name {
					foundWebACL = acl
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

		// Get full WebACL details using GetWebACL
		getInput := &wafv2.GetWebACLInput{
			Id:    foundWebACL.Id,
			Name:  foundWebACL.Name,
			Scope: scope,
		}

		getResp, err := conn.GetWebACL(ctx, getInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL (%s): %s", aws.ToString(foundWebACL.Id), err)
		}

		webACL = getResp.WebACL
	}

	if webACL == nil {
		return sdkdiag.AppendErrorf(diags, "WAFv2 WebACL not found")
	}

	d.SetId(aws.ToString(webACL.Id))
	d.Set(names.AttrARN, webACL.ARN)
	d.Set(names.AttrDescription, webACL.Description)
	d.Set(names.AttrName, webACL.Name)

	return diags
}
