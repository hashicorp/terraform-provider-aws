// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_securityhub_organization_admin_account")
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

func resourceOrganizationAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	adminAccountID := d.Get("admin_account_id").(string)

	input := &securityhub.EnableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(adminAccountID),
	}

	_, err := conn.EnableOrganizationAdminAccountWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling Security Hub Organization Admin Account (%s): %s", adminAccountID, err)
	}

	d.SetId(adminAccountID)

	if _, err := waitAdminAccountEnabled(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Organization Admin Account (%s) to enable: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationAdminAccountRead(ctx, d, meta)...)
}

func resourceOrganizationAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	adminAccount, err := FindAdminAccount(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Security Hub Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Organization Admin Account (%s): %s", d.Id(), err)
	}

	if adminAccount == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading Security Hub Organization Admin Account (%s): %s", d.Id(), err)
		}

		log.Printf("[WARN] Security Hub Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("admin_account_id", adminAccount.AccountId)

	return diags
}

func resourceOrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	input := &securityhub.DisableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(d.Id()),
	}

	_, err := conn.DisableOrganizationAdminAccountWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Security Hub Organization Admin Account (%s): %s", d.Id(), err)
	}

	if _, err := waitAdminAccountNotFound(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Organization Admin Account (%s) to disable: %s", d.Id(), err)
	}

	return diags
}
