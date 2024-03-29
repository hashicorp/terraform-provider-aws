// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_iam_account_alias", name="Account Alias")
func dataSourceAccountAlias() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAccountAliasRead,

		Schema: map[string]*schema.Schema{
			"account_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAccountAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	log.Printf("[DEBUG] Reading IAM Account Aliases.")

	req := &iam.ListAccountAliasesInput{}
	resp, err := conn.ListAccountAliases(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Account Alias: %s", err)
	}

	// 'AccountAliases': [] if there is no alias.
	if resp == nil || len(resp.AccountAliases) == 0 {
		return sdkdiag.AppendErrorf(diags, "reading IAM Account Alias: empty result")
	}

	alias := resp.AccountAliases[0]
	d.SetId(alias)
	d.Set("account_alias", alias)

	return diags
}
