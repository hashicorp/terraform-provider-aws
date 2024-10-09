// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fms/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fms_admin_account", name="Admin Account")
func resourceAdminAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAdminAccountCreate,
		ReadWithoutTimeout:   resourceAdminAccountRead,
		DeleteWithoutTimeout: resourceAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
	}
}

func resourceAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	// Ensure there is not an existing FMS Admin Account.
	output, err := findAdminAccount(ctx, conn)

	if !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "FMS Admin Account (%s) already associated: import this Terraform resource to manage", aws.ToString(output.AdminAccount))
	}

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}

	if _, err := waitAdminAccountCreated(ctx, conn, accountID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FMS Admin Account (%s) create: %s", d.Id(), err)
	}

	d.SetId(accountID)

	return append(diags, resourceAdminAccountRead(ctx, d, meta)...)
}

func resourceAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	output, err := findAdminAccount(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FMS Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FMS Admin Account (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, output.AdminAccount)

	return diags
}

func resourceAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	_, err := conn.DisassociateAdminAccount(ctx, &fms.DisassociateAdminAccountInput{})

	if tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "No default admin could be found for account") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating FMS Admin Account (%s): %s", d.Id(), err)
	}

	if _, err := waitAdminAccountDeleted(ctx, conn, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FMS Admin Account (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findAdminAccount(ctx context.Context, conn *fms.Client) (*fms.GetAdminAccountOutput, error) {
	input := &fms.GetAdminAccountInput{}

	output, err := conn.GetAdminAccount(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.RoleStatus; status == awstypes.AccountRoleStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func statusAssociateAdminAccount(ctx context.Context, conn *fms.Client, accountID string) retry.StateRefreshFunc {
	// This is all wrapped in a StateRefreshFunc since AssociateAdminAccount returns
	// success even though it failed if called too quickly after creating an Organization.
	return func() (interface{}, string, error) {
		input := &fms.AssociateAdminAccountInput{
			AdminAccount: aws.String(accountID),
		}

		_, err := conn.AssociateAdminAccount(ctx, input)

		if err != nil {
			return nil, "", err
		}

		output, err := conn.GetAdminAccount(ctx, &fms.GetAdminAccountInput{})

		// FMS returns an AccessDeniedException if no account is associated,
		// but does not define this in its error codes.
		if tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "is not currently delegated by AWS FM") {
			return nil, "", nil
		}

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if aws.ToString(output.AdminAccount) != accountID {
			return nil, "", nil
		}

		return output, string(output.RoleStatus), err
	}
}

func statusAdminAccount(ctx context.Context, conn *fms.Client) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAdminAccount(ctx, conn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.RoleStatus), nil
	}
}

func waitAdminAccountCreated(ctx context.Context, conn *fms.Client, accountID string, timeout time.Duration) (*fms.GetAdminAccountOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.AccountRoleStatusDeleted, // Recreating association can return this status.
			awstypes.AccountRoleStatusCreating,
		),
		Target:  enum.Slice(awstypes.AccountRoleStatusReady),
		Refresh: statusAssociateAdminAccount(ctx, conn, accountID),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fms.GetAdminAccountOutput); ok {
		return output, err
	}

	return nil, err
}

func waitAdminAccountDeleted(ctx context.Context, conn *fms.Client, timeout time.Duration) (*fms.GetAdminAccountOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.AccountRoleStatusDeleting,
			awstypes.AccountRoleStatusPendingDeletion,
			awstypes.AccountRoleStatusReady,
		),
		Target:  []string{},
		Refresh: statusAdminAccount(ctx, conn),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fms.GetAdminAccountOutput); ok {
		return output, err
	}

	return nil, err
}
