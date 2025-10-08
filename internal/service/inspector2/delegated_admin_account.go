// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_inspector2_delegated_admin_account", name="Delegated Admin Account")
func resourceDelegatedAdminAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDelegatedAdminAccountCreate,
		ReadWithoutTimeout:   resourceDelegatedAdminAccountRead,
		DeleteWithoutTimeout: resourceDelegatedAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Second),
			Delete: schema.DefaultTimeout(15 * time.Second),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"relationship_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDelegatedAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	input := &inspector2.EnableDelegatedAdminAccountInput{
		DelegatedAdminAccountId: aws.String(accountID),
		ClientToken:             aws.String(id.UniqueId()),
	}

	_, err := conn.EnableDelegatedAdminAccount(ctx, input)

	if err != nil && !errs.IsAErrorMessageContains[*awstypes.ConflictException](err, fmt.Sprintf("Delegated administrator %s is already enabled for the organization", accountID)) {
		return sdkdiag.AppendErrorf(diags, "enabling Inspector2 Delegated Admin Account (%s): %s", accountID, err)
	}

	d.SetId(accountID)

	if _, err := waitDelegatedAdminAccountEnabled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Inspector2 Delegated Admin Account (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDelegatedAdminAccountRead(ctx, d, meta)...)
}

func resourceDelegatedAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	output, err := findDelegatedAdminAccountByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 Delegated Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector2 Delegated Admin Account (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, output.AccountId)
	d.Set("relationship_status", output.Status)

	return diags
}

func resourceDelegatedAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	log.Printf("[INFO] Deleting Inspector2 Delegated Admin Account: %s", d.Id())
	_, err := conn.DisableDelegatedAdminAccount(ctx, &inspector2.DisableDelegatedAdminAccountInput{
		DelegatedAdminAccountId: aws.String(d.Get(names.AttrAccountID).(string)),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Inspector2 Delegated Admin Account (%s): %s", d.Id(), err)
	}

	if _, err := waitDelegatedAdminAccountDisabled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Inspector2 Delegated Admin Account (%s) delete: %s", d.Id(), err)
	}

	return diags
}

var (
	errIsDelegatedAdmin = errors.New("is the delegated administrator")
)

func findDelegatedAdminAccountByID(ctx context.Context, conn *inspector2.Client, accountID string) (*awstypes.DelegatedAdminAccount, error) {
	input := &inspector2.ListDelegatedAdminAccountsInput{}
	output, err := findDelegatedAdminAccount(ctx, conn, input, func(v *awstypes.DelegatedAdminAccount) bool {
		return aws.ToString(v.AccountId) == accountID
	})

	if errors.Is(err, errIsDelegatedAdmin) {
		return &awstypes.DelegatedAdminAccount{
			AccountId: aws.String(accountID),
			Status:    awstypes.DelegatedAdminStatusEnabled,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findDelegatedAdminAccount(ctx context.Context, conn *inspector2.Client, input *inspector2.ListDelegatedAdminAccountsInput, filter tfslices.Predicate[*awstypes.DelegatedAdminAccount]) (*awstypes.DelegatedAdminAccount, error) {
	output, err := findDelegatedAdminAccounts(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDelegatedAdminAccounts(ctx context.Context, conn *inspector2.Client, input *inspector2.ListDelegatedAdminAccountsInput, filter tfslices.Predicate[*awstypes.DelegatedAdminAccount]) ([]awstypes.DelegatedAdminAccount, error) {
	var output []awstypes.DelegatedAdminAccount

	pages := inspector2.NewListDelegatedAdminAccountsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "is the delegated admin") {
			return nil, errIsDelegatedAdmin
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DelegatedAdminAccounts {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDelegatedAdminAccount(ctx context.Context, conn *inspector2.Client, accountID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDelegatedAdminAccountByID(ctx, conn, accountID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

type delegatedAdminStatus string

const (
	delegatedAdminStatusDisableInProgress delegatedAdminStatus = delegatedAdminStatus(awstypes.DelegatedAdminStatusDisableInProgress)
	delegatedAdminStatusEnabling          delegatedAdminStatus = "ENABLING"
	delegatedAdminStatusEnableInProgress  delegatedAdminStatus = "ENABLE_IN_PROGRESS"
	delegatedAdminStatusEnabled           delegatedAdminStatus = delegatedAdminStatus(awstypes.DelegatedAdminStatusEnabled)
	delegatedAdminStatusCreated           delegatedAdminStatus = "CREATED"
)

func waitDelegatedAdminAccountEnabled(ctx context.Context, conn *inspector2.Client, accountID string, timeout time.Duration) (*awstypes.DelegatedAdminAccount, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(delegatedAdminStatusDisableInProgress, delegatedAdminStatusEnableInProgress, delegatedAdminStatusEnabling),
		Target:  enum.Slice(delegatedAdminStatusEnabled),
		Refresh: statusDelegatedAdminAccount(ctx, conn, accountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DelegatedAdminAccount); ok {
		return output, err
	}

	return nil, err
}

func waitDelegatedAdminAccountDisabled(ctx context.Context, conn *inspector2.Client, accountID string, timeout time.Duration) (*awstypes.DelegatedAdminAccount, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(delegatedAdminStatusDisableInProgress, delegatedAdminStatusCreated, delegatedAdminStatusEnabled),
		Target:  []string{},
		Refresh: statusDelegatedAdminAccount(ctx, conn, accountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DelegatedAdminAccount); ok {
		return output, err
	}

	return nil, err
}
