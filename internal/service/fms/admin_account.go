// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_fms_admin_account")
func ResourceAdminAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAdminAccountCreate,
		ReadWithoutTimeout:   resourceAdminAccountRead,
		DeleteWithoutTimeout: resourceAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
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
	conn := meta.(*conns.AWSClient).FMSConn(ctx)

	// Ensure there is not an existing FMS Admin Account
	output, err := conn.GetAdminAccountWithContext(ctx, &fms.GetAdminAccountInput{})

	if err != nil && !tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
		return sdkdiag.AppendErrorf(diags, "getting FMS Admin Account: %s", err)
	}

	if output != nil && output.AdminAccount != nil && aws.StringValue(output.RoleStatus) == fms.AccountRoleStatusReady {
		return sdkdiag.AppendErrorf(diags, "FMS Admin Account (%s) already associated: import this Terraform resource to manage", aws.StringValue(output.AdminAccount))
	}

	accountID := meta.(*conns.AWSClient).AccountID

	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{
			fms.AccountRoleStatusDeleted, // Recreating association can return this status
			fms.AccountRoleStatusCreating,
		},
		Target:  []string{fms.AccountRoleStatusReady},
		Refresh: associateAdminAccountRefreshFunc(ctx, conn, accountID),
		Timeout: 30 * time.Minute,
		Delay:   10 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FMS Admin Account (%s) association: %s", accountID, err)
	}

	d.SetId(accountID)

	return append(diags, resourceAdminAccountRead(ctx, d, meta)...)
}

func associateAdminAccountRefreshFunc(ctx context.Context, conn *fms.FMS, accountId string) retry.StateRefreshFunc {
	// This is all wrapped in a refresh func since AssociateAdminAccount returns
	// success even though it failed if called too quickly after creating an organization
	return func() (interface{}, string, error) {
		req := &fms.AssociateAdminAccountInput{
			AdminAccount: aws.String(accountId),
		}

		_, aserr := conn.AssociateAdminAccountWithContext(ctx, req)
		if aserr != nil {
			return nil, "", aserr
		}

		res, err := conn.GetAdminAccountWithContext(ctx, &fms.GetAdminAccountInput{})
		if err != nil {
			// FMS returns an AccessDeniedException if no account is associated,
			// but does not define this in its error codes
			if tfawserr.ErrMessageContains(err, "AccessDeniedException", "is not currently delegated by AWS FM") {
				return nil, "", nil
			}
			if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
				return nil, "", nil
			}
			return nil, "", err
		}

		if aws.StringValue(res.AdminAccount) != accountId {
			return nil, "", nil
		}

		return res, aws.StringValue(res.RoleStatus), err
	}
}

func resourceAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSConn(ctx)

	output, err := conn.GetAdminAccountWithContext(ctx, &fms.GetAdminAccountInput{})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] FMS Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting FMS Admin Account (%s): %s", d.Id(), err)
	}

	if aws.StringValue(output.RoleStatus) == fms.AccountRoleStatusDeleted {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "getting FMS Admin Account (%s): %s after creation", d.Id(), aws.StringValue(output.RoleStatus))
		}

		log.Printf("[WARN] FMS Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("account_id", output.AdminAccount)

	return diags
}

func resourceAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSConn(ctx)

	_, err := conn.DisassociateAdminAccountWithContext(ctx, &fms.DisassociateAdminAccountInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating FMS Admin Account (%s): %s", d.Id(), err)
	}

	if err := waitForAdminAccountDeletion(ctx, conn); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FMS Admin Account (%s) disassociation: %s", d.Id(), err)
	}

	return diags
}

func waitForAdminAccountDeletion(ctx context.Context, conn *fms.FMS) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			fms.AccountRoleStatusDeleting,
			fms.AccountRoleStatusPendingDeletion,
			fms.AccountRoleStatusReady,
		},
		Target: []string{fms.AccountRoleStatusDeleted},
		Refresh: func() (interface{}, string, error) {
			output, err := conn.GetAdminAccountWithContext(ctx, &fms.GetAdminAccountInput{})

			if tfawserr.ErrCodeEquals(err, fms.ErrCodeResourceNotFoundException) {
				return output, fms.AccountRoleStatusDeleted, nil
			}

			if err != nil {
				return nil, "", err
			}

			return output, aws.StringValue(output.RoleStatus), nil
		},
		Timeout: 10 * time.Minute,
		Delay:   10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
