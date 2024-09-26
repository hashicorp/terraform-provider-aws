// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_securityhub_member", name="Member")
func resourceMember() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMemberCreate,
		ReadWithoutTimeout:   resourceMemberRead,
		DeleteWithoutTimeout: resourceMemberDelete,

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
			names.AttrEmail: {
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
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	input := &securityhub.CreateMembersInput{
		AccountDetails: []types.AccountDetails{{
			AccountId: aws.String(accountID),
		}},
	}

	if v, ok := d.GetOk(names.AttrEmail); ok {
		input.AccountDetails[0].Email = aws.String(v.(string))
	}

	output, err := conn.CreateMembers(ctx, input)

	if err == nil && output != nil {
		err = unprocessedAccountsError(output.UnprocessedAccounts)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Member (%s): %s", accountID, err)
	}

	d.SetId(accountID)

	if d.Get("invite").(bool) {
		input := &securityhub.InviteMembersInput{
			AccountIds: []string{d.Id()},
		}

		output, err := conn.InviteMembers(ctx, input)

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
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	member, err := findMemberByAccountID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Member (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, member.AccountId)
	d.Set(names.AttrEmail, member.Email)
	status := aws.ToString(member.MemberStatus)
	const (
		// Associated is the member status naming for Regions that do not support Organizations.
		memberStatusAssociated = "Associated"
		memberStatusInvited    = "Invited"
		memberStatusEnabled    = "Enabled"
		memberStatusResigned   = "Resigned"
	)
	invited := status == memberStatusInvited || status == memberStatusEnabled || status == memberStatusAssociated || status == memberStatusResigned
	d.Set("invite", invited)
	d.Set("master_id", member.MasterId)
	d.Set("member_status", status)

	return diags
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	_, err := conn.DisassociateMembers(ctx, &securityhub.DisassociateMembersInput{
		AccountIds: []string{d.Id()},
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Security Hub Member (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Security Hub Member: %s", d.Id())
	output, err := conn.DeleteMembers(ctx, &securityhub.DeleteMembersInput{
		AccountIds: []string{d.Id()},
	})

	if err == nil && output != nil {
		err = unprocessedAccountsError(output.UnprocessedAccounts)
	}

	if tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Insight (%s) not found, removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Member (%s): %s", d.Id(), err)
	}

	return diags
}

func findMemberByAccountID(ctx context.Context, conn *securityhub.Client, accountID string) (*types.Member, error) {
	input := &securityhub.GetMembersInput{
		AccountIds: []string{accountID},
	}

	return findMember(ctx, conn, input)
}

func findMember(ctx context.Context, conn *securityhub.Client, input *securityhub.GetMembersInput) (*types.Member, error) {
	output, err := findMembers(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findMembers(ctx context.Context, conn *securityhub.Client, input *securityhub.GetMembersInput) ([]types.Member, error) {
	output, err := conn.GetMembers(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "The request is rejected since no such resource found") || tfawserr.ErrMessageContains(err, errCodeBadRequestException, "The request is rejected since no such resource found") {
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

	return output.Members, nil
}

func unprocessedAccountError(apiObject types.Result) error {
	if apiObject.ProcessingResult == nil {
		return nil
	}

	return errors.New(aws.ToString(apiObject.ProcessingResult))
}

func unprocessedAccountsError(apiObjects []types.Result) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := unprocessedAccountError(apiObject); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws.ToString(apiObject.AccountId), err))
		}
	}

	return errors.Join(errs...)
}
