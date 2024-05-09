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
	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_inspector2_delegated_admin_account")
func ResourceDelegatedAdminAccount() *schema.Resource {
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

const (
	ResNameDelegatedAdminAccount = "Delegated Admin Account"
)

func resourceDelegatedAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	in := &inspector2.EnableDelegatedAdminAccountInput{
		DelegatedAdminAccountId: aws.String(d.Get(names.AttrAccountID).(string)),
		ClientToken:             aws.String(id.UniqueId()),
	}

	out, err := conn.EnableDelegatedAdminAccount(ctx, in)

	if err != nil && !errs.MessageContains(err, "ConflictException", "already enabled") {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionCreating, ResNameDelegatedAdminAccount, d.Get(names.AttrAccountID).(string), err)
	}

	if err == nil && (out == nil || out.DelegatedAdminAccountId == nil) {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionCreating, ResNameDelegatedAdminAccount, d.Get(names.AttrAccountID).(string), errors.New("empty output"))
	}

	d.SetId(d.Get(names.AttrAccountID).(string))

	if err := WaitDelegatedAdminAccountEnabled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionWaitingForCreation, ResNameDelegatedAdminAccount, d.Id(), err)
	}

	return append(diags, resourceDelegatedAdminAccountRead(ctx, d, meta)...)
}

func resourceDelegatedAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	st, ai, err := FindDelegatedAdminAccountStatusID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector Delegated Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionReading, ResNameDelegatedAdminAccount, d.Id(), err)
	}

	d.Set(names.AttrAccountID, ai)
	d.Set("relationship_status", st)

	return diags
}

func resourceDelegatedAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	log.Printf("[INFO] Deleting Inspector DelegatedAdminAccount %s", d.Id())

	_, err := conn.DisableDelegatedAdminAccount(ctx, &inspector2.DisableDelegatedAdminAccountInput{
		DelegatedAdminAccountId: aws.String(d.Get(names.AttrAccountID).(string)),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionDeleting, ResNameDelegatedAdminAccount, d.Id(), err)
	}

	if err := WaitDelegatedAdminAccountDisabled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionWaitingForDeletion, ResNameDelegatedAdminAccount, d.Id(), err)
	}

	return diags
}

type DelegatedAccountStatus string

// Enum values for DelegatedAccountStatus
const (
	DelegatedAccountStatusDisableInProgress DelegatedAccountStatus = "DISABLE_IN_PROGRESS"
	DelegatedAccountStatusEnableInProgress  DelegatedAccountStatus = "ENABLE_IN_PROGRESS"
	DelegatedAccountStatusEnabling          DelegatedAccountStatus = "ENABLING"
)

// Values returns all known values for DelegatedAccountStatus.
func (DelegatedAccountStatus) Values() []DelegatedAccountStatus {
	return []DelegatedAccountStatus{
		DelegatedAccountStatusDisableInProgress,
		DelegatedAccountStatusEnableInProgress,
		DelegatedAccountStatusEnabling,
		DelegatedAccountStatus(types.RelationshipStatusAccountSuspended),
		DelegatedAccountStatus(types.RelationshipStatusCannotCreateDetectorInOrgMaster),
		DelegatedAccountStatus(types.RelationshipStatusCreated),
		DelegatedAccountStatus(types.RelationshipStatusDeleted),
		DelegatedAccountStatus(types.RelationshipStatusDisabled),
		DelegatedAccountStatus(types.RelationshipStatusEmailVerificationFailed),
		DelegatedAccountStatus(types.RelationshipStatusEmailVerificationInProgress),
		DelegatedAccountStatus(types.RelationshipStatusEnabled),
		DelegatedAccountStatus(types.RelationshipStatusInvited),
		DelegatedAccountStatus(types.RelationshipStatusRegionDisabled),
		DelegatedAccountStatus(types.RelationshipStatusRemoved),
		DelegatedAccountStatus(types.RelationshipStatusResigned),
	}
}

func WaitDelegatedAdminAccountEnabled(ctx context.Context, conn *inspector2.Client, accountID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(DelegatedAccountStatusDisableInProgress, DelegatedAccountStatusEnableInProgress, DelegatedAccountStatusEnabling),
		Target:  enum.Slice(types.RelationshipStatusEnabled),
		Refresh: statusDelegatedAdminAccount(ctx, conn, accountID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitDelegatedAdminAccountDisabled(ctx context.Context, conn *inspector2.Client, accountID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(DelegatedAccountStatusDisableInProgress, DelegatedAccountStatus(types.RelationshipStatusCreated), DelegatedAccountStatus(types.RelationshipStatusEnabled)),
		Target:  []string{},
		Refresh: statusDelegatedAdminAccount(ctx, conn, accountID),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func statusDelegatedAdminAccount(ctx context.Context, conn *inspector2.Client, accountID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		s, _, err := FindDelegatedAdminAccountStatusID(ctx, conn, accountID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return "", s, nil
	}
}

func FindDelegatedAdminAccountStatusID(ctx context.Context, conn *inspector2.Client, accountID string) (string, string, error) {
	pages := inspector2.NewListDelegatedAdminAccountsPaginator(conn, &inspector2.ListDelegatedAdminAccountsInput{})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsAErrorMessageContains[*types.ValidationException](err, "is the delegated admin") {
			return string(types.RelationshipStatusEnabled), accountID, nil
		}

		if err != nil {
			return "", "", err
		}

		for _, account := range page.DelegatedAdminAccounts {
			if aws.ToString(account.AccountId) == accountID {
				return string(account.Status), aws.ToString(account.AccountId), nil
			}
		}
	}

	return "", "", &retry.NotFoundError{
		Message: fmt.Sprintf("delegated admin account not found for %s", accountID),
	}
}
