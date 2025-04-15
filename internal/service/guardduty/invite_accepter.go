// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_guardduty_invite_accepter", name="Invite Accepter")
func resourceInviteAccepter() *schema.Resource {
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

func resourceInviteAccepterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID := d.Get("detector_id").(string)
	masterAccountID := d.Get("master_account_id").(string)

	inputLI := &guardduty.ListInvitationsInput{}
	outputRaw, err := tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return findInvitation(ctx, conn, inputLI, func(v *awstypes.Invitation) bool {
			return aws.ToString(v.AccountId) == masterAccountID
		})
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Invitation (%s): %s", masterAccountID, err)
	}

	invitationID := aws.ToString(outputRaw.(*awstypes.Invitation).InvitationId)
	inputAI := &guardduty.AcceptInvitationInput{
		DetectorId:   aws.String(detectorID),
		InvitationId: aws.String(invitationID),
		MasterId:     aws.String(masterAccountID),
	}

	_, err = conn.AcceptInvitation(ctx, inputAI)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting GuardDuty Invitation (%s): %s", invitationID, err)
	}

	d.SetId(detectorID)

	return append(diags, resourceInviteAccepterRead(ctx, d, meta)...)
}

func resourceInviteAccepterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	master, err := findMasterAccountByDetectorID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GuardDuty Detector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Detector (%s): %s", d.Id(), err)
	}

	d.Set("detector_id", d.Id())
	d.Set("master_account_id", master.AccountId)

	return diags
}

func resourceInviteAccepterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	_, err := conn.DisassociateFromMasterAccount(ctx, &guardduty.DisassociateFromMasterAccountInput{
		DetectorId: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating GuardDuty Member Detector (%s) from GuardDuty Master Account: %s", d.Id(), err)
	}

	return diags
}

func findInvitation(ctx context.Context, conn *guardduty.Client, input *guardduty.ListInvitationsInput, filter tfslices.Predicate[*awstypes.Invitation]) (*awstypes.Invitation, error) {
	output, err := findInvitations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInvitations(ctx context.Context, conn *guardduty.Client, input *guardduty.ListInvitationsInput, filter tfslices.Predicate[*awstypes.Invitation]) ([]awstypes.Invitation, error) {
	var output []awstypes.Invitation

	pages := guardduty.NewListInvitationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

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

func findMasterAccountByDetectorID(ctx context.Context, conn *guardduty.Client, id string) (*awstypes.Master, error) {
	input := &guardduty.GetMasterAccountInput{
		DetectorId: aws.String(id),
	}

	return findMasterAccount(ctx, conn, input)
}

func findMasterAccount(ctx context.Context, conn *guardduty.Client, input *guardduty.GetMasterAccountInput) (*awstypes.Master, error) {
	output, err := conn.GetMasterAccount(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
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
