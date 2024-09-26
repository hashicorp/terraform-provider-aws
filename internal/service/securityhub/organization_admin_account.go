// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_securityhub_organization_admin_account", name="Organization Admin Account")
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

func resourceOrganizationAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	adminAccountID := d.Get("admin_account_id").(string)
	input := &securityhub.EnableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(adminAccountID),
	}

	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.EnableOrganizationAdminAccount(ctx, input)
	}, errCodeResourceConflictException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling Security Hub Organization Admin Account (%s): %s", adminAccountID, err)
	}

	d.SetId(adminAccountID)

	if _, err := waitAdminAccountCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Organization Admin Account (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationAdminAccountRead(ctx, d, meta)...)
}

func resourceOrganizationAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	adminAccount, err := findAdminAccountByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Organization Admin Account (%s): %s", d.Id(), err)
	}

	d.Set("admin_account_id", adminAccount.AccountId)

	return diags
}

func resourceOrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.DisableOrganizationAdminAccountInput{
		AdminAccountId: aws.String(d.Id()),
	}

	_, err := conn.DisableOrganizationAdminAccount(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Security Hub Organization Admin Account (%s): %s", d.Id(), err)
	}

	if _, err := waitAdminAccountDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Organization Admin Account (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findAdminAccountByID(ctx context.Context, conn *securityhub.Client, adminAccountID string) (*types.AdminAccount, error) {
	input := &securityhub.ListOrganizationAdminAccountsInput{}

	return findAdminAccount(ctx, conn, input, func(v *types.AdminAccount) bool {
		return aws.ToString(v.AccountId) == adminAccountID
	})
}

func findAdminAccount(ctx context.Context, conn *securityhub.Client, input *securityhub.ListOrganizationAdminAccountsInput, filter tfslices.Predicate[*types.AdminAccount]) (*types.AdminAccount, error) {
	output, err := findAdminAccounts(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAdminAccounts(ctx context.Context, conn *securityhub.Client, input *securityhub.ListOrganizationAdminAccountsInput, filter tfslices.Predicate[*types.AdminAccount]) ([]types.AdminAccount, error) {
	var output []types.AdminAccount

	pages := securityhub.NewListOrganizationAdminAccountsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "Your account is not a member of an organization") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AdminAccounts {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusAdminAccount(ctx context.Context, conn *securityhub.Client, adminAccountID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAdminAccountByID(ctx, conn, adminAccountID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

const (
	adminAccountDeletedTimeout = 5 * time.Minute
)

func waitAdminAccountCreated(ctx context.Context, conn *securityhub.Client, adminAccountID string) (*types.AdminAccount, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(types.AdminStatusEnabled),
		Refresh: statusAdminAccount(ctx, conn, adminAccountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

func waitAdminAccountDeleted(ctx context.Context, conn *securityhub.Client, adminAccountID string) (*types.AdminAccount, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AdminStatusDisableInProgress),
		Target:  []string{},
		Refresh: statusAdminAccount(ctx, conn, adminAccountID),
		Timeout: adminAccountDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AdminAccount); ok {
		return output, err
	}

	return nil, err
}
