// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_macie2_invitation_accepter", name="Invitation Accepter")
func resourceInvitationAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInvitationAccepterCreate,
		ReadWithoutTimeout:   resourceInvitationAccepterRead,
		DeleteWithoutTimeout: resourceInvitationAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"administrator_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"invitation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
		},
	}
}

func resourceInvitationAccepterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	adminAccountID := d.Get("administrator_account_id").(string)

	var invitationID string
	var err error
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		invitationID, err = findInvitationByAccount(ctx, conn, adminAccountID)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, string(awstypes.ErrorCodeClientError)) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		if invitationID == "" {
			return retry.RetryableError(fmt.Errorf("unable to find pending Macie Invitation for administrator account ID (%s)", adminAccountID))
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		invitationID, err = findInvitationByAccount(ctx, conn, adminAccountID)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Macie InvitationAccepter (%s): %s", d.Id(), err)
	}

	acceptInvitationInput := &macie2.AcceptInvitationInput{
		InvitationId:           aws.String(invitationID),
		AdministratorAccountId: aws.String(adminAccountID),
	}

	_, err = conn.AcceptInvitation(ctx, acceptInvitationInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Macie InvitationAccepter (%s): %s", d.Id(), err)
	}

	d.SetId(adminAccountID)

	return append(diags, resourceInvitationAccepterRead(ctx, d, meta)...)
}

func resourceInvitationAccepterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	var err error

	input := &macie2.GetAdministratorAccountInput{}

	output, err := conn.GetAdministratorAccount(ctx, input)

	if !d.IsNewResource() && (errs.IsA[*awstypes.ResourceNotFoundException](err) ||
		errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled")) {
		log.Printf("[WARN] Macie InvitationAccepter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie InvitationAccepter (%s): %s", d.Id(), err)
	}

	if output == nil || output.Administrator == nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie InvitationAccepter (%s): %s", d.Id(), err)
	}

	d.Set("administrator_account_id", output.Administrator.AccountId)
	d.Set("invitation_id", output.Administrator.InvitationId)
	return diags
}

func resourceInvitationAccepterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := &macie2.DisassociateFromAdministratorAccountInput{}

	_, err := conn.DisassociateFromAdministratorAccount(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
			errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "disassociating Macie InvitationAccepter (%s): %s", d.Id(), err)
	}
	return diags
}
