// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_securityhub_invite_accepter", name="Invite Accepter")
func resourceInviteAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInviteAccepterCreate,
		ReadWithoutTimeout:   resourceInviteAccepterRead,
		DeleteWithoutTimeout: resourceInviteAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"invitation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceInviteAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	masterID := d.Get("master_id").(string)
	invitation, err := findInvitation(ctx, conn, &securityhub.ListInvitationsInput{}, func(v *types.Invitation) bool {
		return aws.ToString(v.AccountId) == masterID
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Invitation (%s): %s", masterID, err)
	}

	invitationID := aws.ToString(invitation.InvitationId)
	input := &securityhub.AcceptInvitationInput{
		InvitationId: aws.String(invitationID),
		MasterId:     aws.String(masterID),
	}

	_, err = conn.AcceptInvitation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Security Hub Invitation (%s): %s", invitationID, err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return append(diags, resourceInviteAccepterRead(ctx, d, meta)...)
}

func resourceInviteAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	master, err := findMasterAccount(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Master Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Master Account (%s): %s", d.Id(), err)
	}

	d.Set("invitation_id", master.InvitationId)
	d.Set("master_id", master.AccountId)

	return diags
}

func resourceInviteAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Dissasociating from Security Hub Master Account: %s", d.Id())
	_, err := conn.DisassociateFromMasterAccount(ctx, &securityhub.DisassociateFromMasterAccountInput{})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "The request is rejected since no such resource found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating from Security Hub Master Account (%s): %s", d.Id(), err)
	}

	return diags
}

func findMasterAccount(ctx context.Context, conn *securityhub.Client) (*types.Invitation, error) {
	input := &securityhub.GetMasterAccountInput{}

	output, err := conn.GetMasterAccount(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "The request is rejected since no such resource found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Master == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Master, nil
}

func findInvitation(ctx context.Context, conn *securityhub.Client, input *securityhub.ListInvitationsInput, filter tfslices.Predicate[*types.Invitation]) (*types.Invitation, error) {
	output, err := findInvitations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInvitations(ctx context.Context, conn *securityhub.Client, input *securityhub.ListInvitationsInput, filter tfslices.Predicate[*types.Invitation]) ([]types.Invitation, error) {
	var output []types.Invitation

	pages := securityhub.NewListInvitationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "The request is rejected since no such resource found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Invitations {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
