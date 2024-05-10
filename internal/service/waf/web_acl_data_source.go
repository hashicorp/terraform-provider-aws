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

// @SDKDataSource("aws_waf_web_acl", name="Web ACL")
func dataSourceWebACL() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWebACLRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceWebACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)
	name := d.Get(names.AttrName).(string)

	acls := make([]awstypes.WebACLSummary, 0)
	// ListWebACLsInput does not have a name parameter for filtering
	input := &waf.ListWebACLsInput{}
	for {
		output, err := conn.ListWebACLs(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading web ACLs: %s", err)
		}
		for _, acl := range output.WebACLs {
			if aws.ToString(acl.Name) == name {
				acls = append(acls, acl)
			}
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if len(acls) == 0 {
		return sdkdiag.AppendErrorf(diags, "web ACLs not found for name: %s", name)
	}

	if len(acls) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple web ACLs found for name: %s", name)
	}

	acl := acls[0]

	d.SetId(aws.ToString(acl.WebACLId))

	return diags
}
