// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

func resourceAccountAliasCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	accountAlias := d.Get("account_alias").(string)
	input := iam.CreateAccountAliasInput{
		AccountAlias: aws.String(accountAlias),
	}

	_, err := conn.CreateAccountAlias(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Account Alias (%s): %s", accountAlias, err)
	}

	d.SetId(accountAlias)

	return diags
}

func resourceAccountAliasRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	var input iam.ListAccountAliasesInput
	output, err := findAccountAlias(ctx, conn, &input)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IAM Account Alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Account Alias (%s): %s", d.Id(), err)
	}

	d.Set("account_alias", output)

	return diags
}

func resourceAccountAliasDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	log.Printf("[DEBUG] Deleting IAM Account Alias: %s", d.Id())
	input := iam.DeleteAccountAliasInput{
		AccountAlias: aws.String(d.Id()),
	}

	_, err := conn.DeleteAccountAlias(ctx, &input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Account Alias (%s): %s", d.Id(), err)
	}

	return diags
}

func findAccountAlias(ctx context.Context, conn *iam.Client, input *iam.ListAccountAliasesInput) (*string, error) {
	output, err := findAccountAliases(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAccountAliases(ctx context.Context, conn *iam.Client, input *iam.ListAccountAliasesInput) ([]string, error) {
	var output []string

	pages := iam.NewListAccountAliasesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.AccountAliases...)
	}

	return output, nil
}
