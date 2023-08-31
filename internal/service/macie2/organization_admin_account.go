// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_macie2_organization_admin_account")
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceOrganizationAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)
	adminAccountID := d.Get("admin_account_id").(string)
	input := &macie2.EnableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(adminAccountID),
		ClientToken:    aws.String(id.UniqueId()),
	}

	var err error
	err = retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		_, err := conn.EnableOrganizationAdminAccountWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.EnableOrganizationAdminAccountWithContext(ctx, input)
	}

	if err != nil {
		return diag.Errorf("creating Macie OrganizationAdminAccount: %s", err)
	}

	d.SetId(adminAccountID)

	return resourceOrganizationAdminAccountRead(ctx, d, meta)
}

func resourceOrganizationAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	var err error

	res, err := GetOrganizationAdminAccount(ctx, conn, d.Id())

	if !d.IsNewResource() && (tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled")) {
		log.Printf("[WARN] Macie OrganizationAdminAccount (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Macie OrganizationAdminAccount (%s): %s", d.Id(), err)
	}

	if res == nil {
		if !d.IsNewResource() {
			log.Printf("[WARN] Macie OrganizationAdminAccount (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return diag.FromErr(&retry.NotFoundError{})
	}

	d.Set("admin_account_id", res.AccountId)

	return nil
}

func resourceOrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.DisableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(d.Id()),
	}

	_, err := conn.DisableOrganizationAdminAccountWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			return nil
		}
		return diag.Errorf("deleting Macie OrganizationAdminAccount (%s): %s", d.Id(), err)
	}
	return nil
}

func GetOrganizationAdminAccount(ctx context.Context, conn *macie2.Macie2, adminAccountID string) (*macie2.AdminAccount, error) {
	var res *macie2.AdminAccount

	err := conn.ListOrganizationAdminAccountsPagesWithContext(ctx, &macie2.ListOrganizationAdminAccountsInput{}, func(page *macie2.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, adminAccount := range page.AdminAccounts {
			if adminAccount == nil {
				continue
			}

			if aws.StringValue(adminAccount.AccountId) == adminAccountID {
				res = adminAccount
				return false
			}
		}

		return !lastPage
	})

	return res, err
}
