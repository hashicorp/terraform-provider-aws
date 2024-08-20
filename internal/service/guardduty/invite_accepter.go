// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_guardduty_invite_accepter")
func ResourceInviteAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInviteAccepterCreate,
		ReadWithoutTimeout:   resourceInviteAccepterRead,
		DeleteWithoutTimeout: resourceInviteAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"master_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
		},
	}
}

func resourceInviteAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID := d.Get("detector_id").(string)
	invitationID := ""
	masterAccountID := d.Get("master_account_id").(string)

	listInvitationsInput := &guardduty.ListInvitationsInput{}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		log.Printf("[DEBUG] Listing GuardDuty Invitations: %+v", listInvitationsInput)
		pages := guardduty.NewListInvitationsPaginator(conn, listInvitationsInput)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil {
				return retry.NonRetryableError(err)
			}

			if invitationID == "" {
				return retry.RetryableError(fmt.Errorf("unable to find pending GuardDuty Invitation for detector ID (%s) from master account ID (%s)", detectorID, masterAccountID))
			}

			for _, invitation := range page.Invitations {
				if aws.ToString(invitation.AccountId) == masterAccountID {
					invitationID = aws.ToString(invitation.InvitationId)
					break
				}
			}
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		pages := guardduty.NewListInvitationsPaginator(conn, listInvitationsInput)

		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "listing GuardDuty Invitations: %s", err)
			}

			for _, invitation := range page.Invitations {
				if aws.ToString(invitation.AccountId) == masterAccountID {
					invitationID = aws.ToString(invitation.InvitationId)
					break
				}
			}
		}
	}

	acceptInvitationInput := &guardduty.AcceptInvitationInput{
		DetectorId:   aws.String(detectorID),
		InvitationId: aws.String(invitationID),
		MasterId:     aws.String(masterAccountID),
	}

	log.Printf("[DEBUG] Accepting GuardDuty Invitation: %+v", acceptInvitationInput)
	_, err = conn.AcceptInvitation(ctx, acceptInvitationInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting GuardDuty Invitation (%s): %s", invitationID, err)
	}

	d.SetId(detectorID)

	return append(diags, resourceInviteAccepterRead(ctx, d, meta)...)
}

func resourceInviteAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	input := &guardduty.GetMasterAccountInput{
		DetectorId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading GuardDuty Master Account: %+v", input)
	output, err := conn.GetMasterAccount(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
		log.Printf("[WARN] GuardDuty Detector %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Detector (%s) GuardDuty Master Account: %s", d.Id(), err)
	}

	if output == nil || output.Master == nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Detector (%s) GuardDuty Master Account: empty response", d.Id())
	}

	d.Set("detector_id", d.Id())
	d.Set("master_account_id", output.Master.AccountId)

	return diags
}

func resourceInviteAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	input := &guardduty.DisassociateFromMasterAccountInput{
		DetectorId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Disassociating GuardDuty Detector (%s) from GuardDuty Master Account: %+v", d.Id(), input)
	_, err := conn.DisassociateFromMasterAccount(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating GuardDuty Member Detector (%s) from GuardDuty Master Account: %s", d.Id(), err)
	}

	return diags
}
