// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_detective_organization_admin_account")
func ResourceOrganizationAdminAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationAdminAccountCreate,
		ReadWithoutTimeout:   resourceOrganizationAdminAccountRead,
		DeleteWithoutTimeout: resourceOrganizationAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
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

	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	accountID := d.Get("account_id").(string)
	input := &detective.EnableOrganizationAdminAccountInput{
		AccountId: aws.String(accountID),
	}

	_, err := conn.EnableOrganizationAdminAccountWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling Detective Organization Admin Account (%s): %s", accountID, err)
	}

	d.SetId(accountID)

	_, err = tfresource.RetryWhenNotFound(ctx, 5*time.Minute, func() (interface{}, error) {
		return FindOrganizationAdminAccountByAccountID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Detective Organization Admin Account (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOrganizationAdminAccountRead(ctx, d, meta)...)
}

func resourceOrganizationAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	administrator, err := FindOrganizationAdminAccountByAccountID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Detective Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Detective Organization Admin Account (%s): %s", d.Id(), err)
	}

	d.Set("account_id", administrator.AccountId)

	return diags
}

func resourceOrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	_, err := conn.DisableOrganizationAdminAccountWithContext(ctx, &detective.DisableOrganizationAdminAccountInput{})

	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Detective Organization Admin Account (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 5*time.Minute, func() (interface{}, error) {
		return FindOrganizationAdminAccountByAccountID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Detective Organization Admin Account (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindOrganizationAdminAccountByAccountID(ctx context.Context, conn *detective.Detective, accountID string) (*detective.Administrator, error) {
	input := &detective.ListOrganizationAdminAccountsInput{}

	return findOrganizationAdminAccount(ctx, conn, input, func(v *detective.Administrator) bool {
		return aws.StringValue(v.AccountId) == accountID
	})
}

func findOrganizationAdminAccount(ctx context.Context, conn *detective.Detective, input *detective.ListOrganizationAdminAccountsInput, filter tfslices.Predicate[*detective.Administrator]) (*detective.Administrator, error) {
	output, err := findOrganizationAdminAccounts(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findOrganizationAdminAccounts(ctx context.Context, conn *detective.Detective, input *detective.ListOrganizationAdminAccountsInput, filter tfslices.Predicate[*detective.Administrator]) ([]*detective.Administrator, error) {
	var output []*detective.Administrator

	err := conn.ListOrganizationAdminAccountsPagesWithContext(ctx, input, func(page *detective.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Administrators {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, detective.ErrCodeValidationException, "account is not a member of an organization") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
