// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/detective"
	awstypes "github.com/aws/aws-sdk-go-v2/service/detective/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			names.AttrAccountID: {
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

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	input := &detective.EnableOrganizationAdminAccountInput{
		AccountId: aws.String(accountID),
	}

	_, err := conn.EnableOrganizationAdminAccount(ctx, input)

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

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	administrator, err := FindOrganizationAdminAccountByAccountID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Detective Organization Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Detective Organization Admin Account (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, administrator.AccountId)

	return diags
}

func resourceOrganizationAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	_, err := conn.DisableOrganizationAdminAccount(ctx, &detective.DisableOrganizationAdminAccountInput{})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func FindOrganizationAdminAccountByAccountID(ctx context.Context, conn *detective.Client, accountID string) (*awstypes.Administrator, error) {
	input := &detective.ListOrganizationAdminAccountsInput{}

	return findOrganizationAdminAccount(ctx, conn, input, func(v awstypes.Administrator) bool {
		return aws.ToString(v.AccountId) == accountID
	})
}

func findOrganizationAdminAccount(ctx context.Context, conn *detective.Client, input *detective.ListOrganizationAdminAccountsInput, filter tfslices.Predicate[awstypes.Administrator]) (*awstypes.Administrator, error) {
	output, err := findOrganizationAdminAccounts(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOrganizationAdminAccounts(ctx context.Context, conn *detective.Client, input *detective.ListOrganizationAdminAccountsInput, filter tfslices.Predicate[awstypes.Administrator]) ([]awstypes.Administrator, error) {
	var output []awstypes.Administrator

	pages := detective.NewListOrganizationAdminAccountsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "account is not a member of an organization") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Administrators {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
