// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_guardduty_organization_admin_account", name="Organization Admin Account")
func ResourceOrganizationAdminAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationAdminAccountCreate,
		ReadWithoutTimeout:   resourceOrganizationAdminAccountRead,
		DeleteWithoutTimeout: resourceOrganizationAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"admin_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
	}
}

func resourceOrganizationAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	adminAccountID := d.Get("admin_account_id").(string)

	input := &guardduty.EnableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(adminAccountID),
	}

	_, err := conn.EnableOrganizationAdminAccount(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling GuardDuty Organization Admin Account (%s): %s", adminAccountID, err)
	}

	d.SetId(adminAccountID)

	if _, err := waitAdminAccountEnabled(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Organization Admin Account (%s) to enable: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationAdminAccountRead(ctx, d, meta)...)
}

func resourceOrganizationAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	adminAccount, err := GetOrganizationAdminAccount(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Organization Admin Account (%s): %s", d.Id(), err)
	}

	if adminAccount == nil {
		log.Printf("[WARN] GuardDuty Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("admin_account_id", adminAccount.AdminAccountId)

	return diags
}

func resourceOrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	input := &guardduty.DisableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(d.Id()),
	}

	_, err := conn.DisableOrganizationAdminAccount(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling GuardDuty Organization Admin Account (%s): %s", d.Id(), err)
	}

	if _, err := waitAdminAccountNotFound(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Organization Admin Account (%s) to disable: %s", d.Id(), err)
	}

	return diags
}

func GetOrganizationAdminAccount(ctx context.Context, conn *guardduty.Client, adminAccountID string) (*awstypes.AdminAccount, error) {
	input := &guardduty.ListOrganizationAdminAccountsInput{}
	result := awstypes.AdminAccount{}

	pages := guardduty.NewListOrganizationAdminAccountsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return &result, err
		}
		for _, adminAccount := range page.AdminAccounts {
			if aws.ToString(adminAccount.AdminAccountId) == adminAccountID {
				result = adminAccount
				break
			}
		}
	}

	return &result, nil
}
