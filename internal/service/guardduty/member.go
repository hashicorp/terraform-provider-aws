// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_member")
func ResourceMember() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMemberCreate,
		ReadWithoutTimeout:   resourceMemberRead,
		UpdateWithoutTimeout: resourceMemberUpdate,
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
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrEmail: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"relationship_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invite": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"disable_email_notification": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"invitation_message": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Second),
			Update: schema.DefaultTimeout(60 * time.Second),
		},
	}
}

func resourceMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)
	accountID := d.Get(names.AttrAccountID).(string)
	detectorID := d.Get("detector_id").(string)

	input := guardduty.CreateMembersInput{
		AccountDetails: []*guardduty.AccountDetail{{
			AccountId: aws.String(accountID),
			Email:     aws.String(d.Get(names.AttrEmail).(string)),
		}},
		DetectorId: aws.String(detectorID),
	}

	log.Printf("[DEBUG] Creating GuardDuty Member: %s", input)
	_, err := conn.CreateMembersWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating GuardDuty Member failed: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", detectorID, accountID))

	if !d.Get("invite").(bool) {
		return append(diags, resourceMemberRead(ctx, d, meta)...)
	}

	imi := &guardduty.InviteMembersInput{
		DetectorId:               aws.String(detectorID),
		AccountIds:               []*string{aws.String(accountID)},
		DisableEmailNotification: aws.Bool(d.Get("disable_email_notification").(bool)),
		Message:                  aws.String(d.Get("invitation_message").(string)),
	}

	log.Printf("[INFO] Inviting GuardDuty Member: %s", input)
	_, err = conn.InviteMembersWithContext(ctx, imi)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "inviting GuardDuty Member %q: %s", d.Id(), err)
	}

	err = inviteMemberWaiter(ctx, accountID, detectorID, d.Timeout(schema.TimeoutUpdate), conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Member %q invite: %s", d.Id(), err)
	}

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	accountID, detectorID, err := DecodeMemberID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Member (%s): %s", d.Id(), err)
	}

	input := guardduty.GetMembersInput{
		AccountIds: []*string{aws.String(accountID)},
		DetectorId: aws.String(detectorID),
	}

	log.Printf("[DEBUG] Reading GuardDuty Member: %s", input)
	gmo, err := conn.GetMembersWithContext(ctx, &input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty detector %q not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Member (%s): %s", d.Id(), err)
	}

	if gmo.Members == nil || (len(gmo.Members) < 1) {
		log.Printf("[WARN] GuardDuty Member %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	member := gmo.Members[0]
	d.Set(names.AttrAccountID, member.AccountId)
	d.Set("detector_id", detectorID)
	d.Set(names.AttrEmail, member.Email)

	status := aws.StringValue(member.RelationshipStatus)
	d.Set("relationship_status", status)

	// https://docs.aws.amazon.com/guardduty/latest/ug/list-members.html
	d.Set("invite", false)
	if status == "Disabled" || status == "Enabled" || status == "Invited" || status == "EmailVerificationInProgress" {
		d.Set("invite", true)
	}

	return diags
}

func resourceMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	accountID, detectorID, err := DecodeMemberID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GuardDuty Member (%s): %s", d.Id(), err)
	}

	if d.HasChange("invite") {
		if d.Get("invite").(bool) {
			input := &guardduty.InviteMembersInput{
				DetectorId:               aws.String(detectorID),
				AccountIds:               []*string{aws.String(accountID)},
				DisableEmailNotification: aws.Bool(d.Get("disable_email_notification").(bool)),
				Message:                  aws.String(d.Get("invitation_message").(string)),
			}

			log.Printf("[INFO] Inviting GuardDuty Member: %s", input)
			output, err := conn.InviteMembersWithContext(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "inviting GuardDuty Member %q: %s", d.Id(), err)
			}

			// {"unprocessedAccounts":[{"result":"The request is rejected because the current account has already invited or is already the GuardDuty master of the given member account ID.","accountId":"067819342479"}]}
			if len(output.UnprocessedAccounts) > 0 {
				return sdkdiag.AppendErrorf(diags, "inviting GuardDuty Member %q: %s", d.Id(), aws.StringValue(output.UnprocessedAccounts[0].Result))
			}

			err = inviteMemberWaiter(ctx, accountID, detectorID, d.Timeout(schema.TimeoutUpdate), conn)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Member %q invite: %s", d.Id(), err)
			}
		} else {
			input := &guardduty.DisassociateMembersInput{
				AccountIds: []*string{aws.String(accountID)},
				DetectorId: aws.String(detectorID),
			}
			log.Printf("[INFO] Disassociating GuardDuty Member: %s", input)
			_, err := conn.DisassociateMembersWithContext(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating GuardDuty Member %q: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyConn(ctx)

	accountID, detectorID, err := DecodeMemberID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Member (%s): %s", d.Id(), err)
	}

	input := guardduty.DeleteMembersInput{
		AccountIds: []*string{aws.String(accountID)},
		DetectorId: aws.String(detectorID),
	}

	log.Printf("[DEBUG] Delete GuardDuty Member: %s", input)
	_, err = conn.DeleteMembersWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Member (%s): %s", d.Id(), err)
	}
	return diags
}

func inviteMemberWaiter(ctx context.Context, accountID, detectorID string, timeout time.Duration, conn *guardduty.GuardDuty) error {
	input := guardduty.GetMembersInput{
		DetectorId: aws.String(detectorID),
		AccountIds: []*string{aws.String(accountID)},
	}

	// wait until e-mail verification finishes
	var out *guardduty.GetMembersOutput
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		log.Printf("[DEBUG] Reading GuardDuty Member: %s", input)
		var err error
		out, err = conn.GetMembersWithContext(ctx, &input)

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("reading GuardDuty Member %q: %s", accountID, err))
		}

		retryable, err := memberInvited(out, accountID)
		if err != nil {
			if retryable {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.GetMembersWithContext(ctx, &input)

		if err != nil {
			return fmt.Errorf("reading GuardDuty member: %w", err)
		}
		_, err = memberInvited(out, accountID)
		return err
	}
	if err != nil {
		return fmt.Errorf("waiting for GuardDuty email verification: %w", err)
	}
	return nil
}

func memberInvited(out *guardduty.GetMembersOutput, accountID string) (bool, error) {
	if out == nil || len(out.Members) == 0 {
		return true, fmt.Errorf("reading GuardDuty Member %q: member missing from response", accountID)
	}

	member := out.Members[0]
	status := aws.StringValue(member.RelationshipStatus)

	if status == "Disabled" || status == "Enabled" || status == "Invited" {
		return false, nil
	}

	if status == "Created" || status == "EmailVerificationInProgress" {
		return true, fmt.Errorf("Expected member to be invited but was in state: %s", status)
	}

	return false, fmt.Errorf("inviting GuardDuty Member %q: invalid status: %s", accountID, status)
}

func DecodeMemberID(id string) (accountID, detectorID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("GuardDuty Member ID must be of the form <Detector ID>:<Member AWS Account ID>, was provided: %s", id)
		return
	}
	accountID = parts[1]
	detectorID = parts[0]
	return
}
