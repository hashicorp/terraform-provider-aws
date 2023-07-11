// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	detectorID := d.Get("detector_id").(string)
	invitationID := ""
	masterAccountID := d.Get("master_account_id").(string)

	listInvitationsInput := &guardduty.ListInvitationsInput{}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		log.Printf("[DEBUG] Listing GuardDuty Invitations: %s", listInvitationsInput)
		err := conn.ListInvitationsPagesWithContext(ctx, listInvitationsInput, func(page *guardduty.ListInvitationsOutput, lastPage bool) bool {
			for _, invitation := range page.Invitations {
				if aws.StringValue(invitation.AccountId) == masterAccountID {
					invitationID = aws.StringValue(invitation.InvitationId)
					return false
				}
			}
			return !lastPage
		})

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if invitationID == "" {
			return retry.RetryableError(fmt.Errorf("unable to find pending GuardDuty Invitation for detector ID (%s) from master account ID (%s)", detectorID, masterAccountID))
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		err = conn.ListInvitationsPagesWithContext(ctx, listInvitationsInput, func(page *guardduty.ListInvitationsOutput, lastPage bool) bool {
			for _, invitation := range page.Invitations {
				if aws.StringValue(invitation.AccountId) == masterAccountID {
					invitationID = aws.StringValue(invitation.InvitationId)
					return false
				}
			}
			return !lastPage
		})
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing GuardDuty Invitations: %s", err)
	}

	acceptInvitationInput := &guardduty.AcceptInvitationInput{
		DetectorId:   aws.String(detectorID),
		InvitationId: aws.String(invitationID),
		MasterId:     aws.String(masterAccountID),
	}

	log.Printf("[DEBUG] Accepting GuardDuty Invitation: %s", acceptInvitationInput)
	_, err = conn.AcceptInvitationWithContext(ctx, acceptInvitationInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting GuardDuty Invitation (%s): %s", invitationID, err)
	}

	d.SetId(detectorID)

	return append(diags, resourceInviteAccepterRead(ctx, d, meta)...)
}

func resourceInviteAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	input := &guardduty.GetMasterAccountInput{
		DetectorId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading GuardDuty Master Account: %s", input)
	output, err := conn.GetMasterAccountWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
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
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	input := &guardduty.DisassociateFromMasterAccountInput{
		DetectorId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Disassociating GuardDuty Detector (%s) from GuardDuty Master Account: %s", d.Id(), input)
	_, err := conn.DisassociateFromMasterAccountWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating GuardDuty Member Detector (%s) from GuardDuty Master Account: %s", d.Id(), err)
	}

	return diags
}
