// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_macie2_organization_admin_account", name="Organization Admin Account")
func resourceOrganizationAdminAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationAdminAccountCreate,
		ReadWithoutTimeout:   resourceOrganizationAdminAccountRead,
		DeleteWithoutTimeout: resourceOrganizationAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"admin_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceOrganizationAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)
	adminAccountID := d.Get("admin_account_id").(string)
	input := &macie2.EnableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(adminAccountID),
		ClientToken:    aws.String(id.UniqueId()),
	}

	var err error
	err = retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		_, err := conn.EnableOrganizationAdminAccount(ctx, input)

		if tfawserr.ErrCodeEquals(err, string(awstypes.ErrorCodeClientError)) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.EnableOrganizationAdminAccount(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie OrganizationAdminAccount: %s", err)
	}

	d.SetId(adminAccountID)

	return append(diags, resourceOrganizationAdminAccountRead(ctx, d, meta)...)
}

func resourceOrganizationAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	res, err := GetOrganizationAdminAccount(ctx, conn, d.Id())

	if !d.IsNewResource() && (errs.IsA[*awstypes.ResourceNotFoundException](err) ||
		errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled")) {
		log.Printf("[WARN] Macie OrganizationAdminAccount (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie OrganizationAdminAccount (%s): %s", d.Id(), err)
	}

	if res == nil {
		if !d.IsNewResource() {
			log.Printf("[WARN] Macie OrganizationAdminAccount (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendFromErr(diags, &retry.NotFoundError{})
	}

	d.Set("admin_account_id", res.AccountId)

	return diags
}

func resourceOrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := &macie2.DisableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(d.Id()),
	}

	_, err := conn.DisableOrganizationAdminAccount(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
			errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Macie OrganizationAdminAccount (%s): %s", d.Id(), err)
	}
	return diags
}

func GetOrganizationAdminAccount(ctx context.Context, conn *macie2.Client, adminAccountID string) (*awstypes.AdminAccount, error) {
	input := &macie2.ListOrganizationAdminAccountsInput{}

	pages := macie2.NewListOrganizationAdminAccountsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, adminAccount := range page.AdminAccounts {
			if aws.ToString(adminAccount.AccountId) == adminAccountID {
				return &adminAccount, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}
