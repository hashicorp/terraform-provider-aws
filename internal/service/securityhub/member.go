// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Associated is the member status naming for regions that do not support Organizations
	memberStatusAssociated = "Associated"
	memberStatusInvited    = "Invited"
	memberStatusEnabled    = "Enabled"
	memberStatusResigned   = "Resigned"
)

// @SDKResource("aws_securityhub_member")
func ResourceMember() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMemberCreate,
		ReadWithoutTimeout:   resourceMemberRead,
		DeleteWithoutTimeout: resourceMemberDelete,

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
			"email": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"invite": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"master_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"member_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	accountID := d.Get("account_id").(string)
	input := &securityhub.CreateMembersInput{
		AccountDetails: []*securityhub.AccountDetails{{
			AccountId: aws.String(accountID),
		}},
	}

	if v, ok := d.GetOk("email"); ok {
		input.AccountDetails[0].Email = aws.String(v.(string))
	}

	output, err := conn.CreateMembersWithContext(ctx, input)

	if err == nil && output != nil {
		err = unprocessedAccountsError(output.UnprocessedAccounts)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Member (%s): %s", accountID, err)
	}

	d.SetId(accountID)

	if d.Get("invite").(bool) {
		input := &securityhub.InviteMembersInput{
			AccountIds: aws.StringSlice([]string{d.Id()}),
		}

		output, err := conn.InviteMembersWithContext(ctx, input)

		if err == nil && output != nil {
			err = unprocessedAccountsError(output.UnprocessedAccounts)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "inviting Security Hub Member (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	member, err := FindMemberByAccountID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Member (%s): %s", d.Id(), err)
	}

	d.Set("account_id", member.AccountId)
	d.Set("email", member.Email)
	d.Set("master_id", member.MasterId)
	status := aws.StringValue(member.MemberStatus)
	d.Set("member_status", status)
	invited := status == memberStatusInvited || status == memberStatusEnabled || status == memberStatusAssociated || status == memberStatusResigned
	d.Set("invite", invited)

	return diags
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	_, err := conn.DisassociateMembersWithContext(ctx, &securityhub.DisassociateMembersInput{
		AccountIds: aws.StringSlice([]string{d.Id()}),
	})

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Security Hub Member (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Security Hub Member: %s", d.Id())
	output, err := conn.DeleteMembersWithContext(ctx, &securityhub.DeleteMembersInput{
		AccountIds: aws.StringSlice([]string{d.Id()}),
	})

	if err == nil && output != nil {
		err = unprocessedAccountsError(output.UnprocessedAccounts)
	}

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Member (%s): %s", d.Id(), err)
	}

	return diags
}

const (
	errCodeBadRequestException = "BadRequestException"
)

func FindMemberByAccountID(ctx context.Context, conn *securityhub.SecurityHub, accountID string) (*securityhub.Member, error) {
	input := &securityhub.GetMembersInput{
		AccountIds: aws.StringSlice([]string{accountID}),
	}

	output, err := conn.GetMembersWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if tfawserr.ErrMessageContains(err, errCodeBadRequestException, "no such resource found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Members) == 0 || output.Members[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Members); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Members[0], nil
}

func unprocessedAccountError(apiObject *securityhub.Result) error {
	if apiObject == nil || apiObject.ProcessingResult == nil {
		return nil
	}

	return errors.New(aws.StringValue(apiObject.ProcessingResult))
}

func unprocessedAccountsError(apiObjects []*securityhub.Result) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if err := unprocessedAccountError(apiObject); err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", aws.StringValue(apiObject.AccountId), err))
		}
	}

	return errors.ErrorOrNil()
}
