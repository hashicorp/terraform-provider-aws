// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_iam_account_alias")
func ResourceAccountAlias() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	account_alias := d.Get("account_alias").(string)

	params := &iam.CreateAccountAliasInput{
		AccountAlias: aws.String(account_alias),
	}

	_, err := conn.CreateAccountAliasWithContext(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating account alias with name '%s': %s", account_alias, err)
	}

	d.SetId(account_alias)

	return diags
}

func resourceAccountAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	params := &iam.ListAccountAliasesInput{}

	resp, err := conn.ListAccountAliasesWithContext(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing account aliases: %s", err)
	}

	if !d.IsNewResource() && (resp == nil || len(resp.AccountAliases) == 0) {
		d.SetId("")
		return diags
	}

	account_alias := aws.StringValue(resp.AccountAliases[0])

	d.SetId(account_alias)
	d.Set("account_alias", account_alias)

	return diags
}

func resourceAccountAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	account_alias := d.Get("account_alias").(string)

	params := &iam.DeleteAccountAliasInput{
		AccountAlias: aws.String(account_alias),
	}

	_, err := conn.DeleteAccountAliasWithContext(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting account alias with name '%s': %s", account_alias, err)
	}

	return diags
}
