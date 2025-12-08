// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func dataSourceAccountAliasRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	var input iam.ListAccountAliasesInput
	output, err := findAccountAlias(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Account Alias: %s", err)
	}

	d.SetId(aws.ToString(output))
	d.Set("account_alias", output)

	return diags
}
