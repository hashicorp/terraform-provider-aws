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

// @SDKResource("aws_iam_account_alias", name="Account Alias")
func resourceAccountAlias() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountAliasCreate,
		ReadWithoutTimeout:   resourceAccountAliasRead,
		DeleteWithoutTimeout: resourceAccountAliasDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_alias": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validAccountAlias,
			},
		},
	}
}

func resourceAccountAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	accountAlias := d.Get("account_alias").(string)

	params := &iam.CreateAccountAliasInput{
		AccountAlias: aws.String(accountAlias),
	}

	_, err := conn.CreateAccountAlias(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating account alias with name '%s': %s", accountAlias, err)
	}

	d.SetId(accountAlias)

	return diags
}

func resourceAccountAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	params := &iam.ListAccountAliasesInput{}

	resp, err := conn.ListAccountAliases(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing account aliases: %s", err)
	}

	if !d.IsNewResource() && (resp == nil || len(resp.AccountAliases) == 0) {
		d.SetId("")
		return diags
	}

	accountAlias := resp.AccountAliases[0]

	d.SetId(accountAlias)
	d.Set("account_alias", accountAlias)

	return diags
}

func resourceAccountAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	accountAlias := d.Get("account_alias").(string)

	params := &iam.DeleteAccountAliasInput{
		AccountAlias: aws.String(accountAlias),
	}

	_, err := conn.DeleteAccountAlias(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting account alias with name '%s': %s", accountAlias, err)
	}

	return diags
}
