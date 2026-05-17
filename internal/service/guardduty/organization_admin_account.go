// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package guardduty

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_guardduty_organization_admin_account", name="Organization Admin Account")
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
	input := guardduty.EnableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(adminAccountID),
	}
	_, err := conn.EnableOrganizationAdminAccount(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling GuardDuty Organization Admin Account (%s): %s", adminAccountID, err)
	}

	d.SetId(adminAccountID)

	if _, err := waitAdminAccountCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Organization Admin Account (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationAdminAccountRead(ctx, d, meta)...)
}

func resourceOrganizationAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	adminAccount, err := findOrganizationAdminAccountByID(ctx, conn, d.Id())
	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] GuardDuty Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Organization Admin Account (%s): %s", d.Id(), err)
	}

	d.Set("admin_account_id", adminAccount.AdminAccountId)

	return diags
}

func resourceOrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	input := guardduty.DisableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(d.Id()),
	}
	_, err := conn.DisableOrganizationAdminAccount(ctx, &input)
	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request failed because the delegated administrator account has already been disabled and/or GuardDuty protection has been disabled.") {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling GuardDuty Organization Admin Account (%s): %s", d.Id(), err)
	}

	if _, err := waitAdminAccountDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Organization Admin Account (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findOrganizationAdminAccountByID(ctx context.Context, conn *guardduty.Client, adminAccountID string) (*awstypes.AdminAccount, error) {
	var input guardduty.ListOrganizationAdminAccountsInput
	output, err := findOrganizationAdminAccounts(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output, func(v awstypes.AdminAccount) bool {
		return aws.ToString(v.AdminAccountId) == adminAccountID
	}))
}

func findOrganizationAdminAccounts(ctx context.Context, conn *guardduty.Client, input *guardduty.ListOrganizationAdminAccountsInput) ([]awstypes.AdminAccount, error) {
	var output []awstypes.AdminAccount

	pages := guardduty.NewListOrganizationAdminAccountsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.AdminAccounts...)
	}

	return output, nil
}

func statusAdminAccount(conn *guardduty.Client, adminAccountID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findOrganizationAdminAccountByID(ctx, conn, adminAccountID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AdminStatus), nil
	}
}

func waitAdminAccountCreated(ctx context.Context, conn *guardduty.Client, adminAccountID string) (*awstypes.AdminAccount, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.AdminStatusEnabled),
		Refresh: statusAdminAccount(conn, adminAccountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

func waitAdminAccountDeleted(ctx context.Context, conn *guardduty.Client, adminAccountID string) (*awstypes.AdminAccount, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AdminStatusDisableInProgress),
		Target:  []string{},
		Refresh: statusAdminAccount(conn, adminAccountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.AdminAccount); ok {
		return output, err
	}

	return nil, err
}
