// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package guardduty

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_guardduty_member", name="Member")
func resourceMember() *schema.Resource {
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
			"disable_email_notification": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			names.AttrEmail: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"invitation_message": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"invite": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"relationship_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Second),
			Update: schema.DefaultTimeout(60 * time.Second),
		},
	}
}

func resourceMemberCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID, accountID := d.Get("detector_id").(string), d.Get(names.AttrAccountID).(string)
	email := d.Get(names.AttrEmail).(string)
	input := guardduty.CreateMembersInput{
		AccountDetails: []awstypes.AccountDetail{{
			AccountId: aws.String(accountID),
			Email:     aws.String(email),
		}},
		DetectorId: aws.String(detectorID),
	}
	output, err := conn.CreateMembers(ctx, &input)
	if err == nil {
		err = unprocessedAccountsError(output.UnprocessedAccounts)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GuardDuty Member (%s): %s", email, err)
	}

	d.SetId(memberCreateResourceID(detectorID, accountID))

	if d.Get("invite").(bool) {
		input := guardduty.InviteMembersInput{
			AccountIds:               []string{accountID},
			DetectorId:               aws.String(detectorID),
			DisableEmailNotification: aws.Bool(d.Get("disable_email_notification").(bool)),
			Message:                  aws.String(d.Get("invitation_message").(string)),
		}
		output, err := conn.InviteMembers(ctx, &input)
		if err == nil {
			err = unprocessedAccountsError(output.UnprocessedAccounts)
		}
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "inviting GuardDuty Member (%s): %s", d.Id(), err)
		}

		if _, err := waitMemberInvited(ctx, conn, detectorID, accountID, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Member (%s) invite: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID, accountID, err := memberParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	member, err := findMemberByTwoPartKey(ctx, conn, detectorID, accountID)
	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] GuardDuty Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GuardDuty Member (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	d.Set("detector_id", detectorID)
	d.Set(names.AttrEmail, member.Email)
	// https://docs.aws.amazon.com/guardduty/latest/ug/list-members.html
	status := aws.ToString(member.RelationshipStatus)
	switch status {
	case memberRelationshipStatusDisabled, memberRelationshipStatusEnabled, memberRelationshipStatusInvited, memberRelationshipStatusEmailVerificationInProgress:
		d.Set("invite", true)
	default:
		d.Set("invite", false)
	}
	d.Set("relationship_status", status)

	return diags
}

func resourceMemberUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	if d.HasChange("invite") {
		detectorID, accountID, err := memberParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if d.Get("invite").(bool) {
			input := guardduty.InviteMembersInput{
				AccountIds:               []string{accountID},
				DetectorId:               aws.String(detectorID),
				DisableEmailNotification: aws.Bool(d.Get("disable_email_notification").(bool)),
				Message:                  aws.String(d.Get("invitation_message").(string)),
			}
			output, err := conn.InviteMembers(ctx, &input)
			if err == nil {
				err = unprocessedAccountsError(output.UnprocessedAccounts)
			}
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "inviting GuardDuty Member (%s): %s", d.Id(), err)
			}

			if _, err := waitMemberInvited(ctx, conn, detectorID, accountID, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for GuardDuty Member (%s) invite: %s", d.Id(), err)
			}
		} else {
			input := guardduty.DisassociateMembersInput{
				AccountIds: []string{accountID},
				DetectorId: aws.String(detectorID),
			}
			output, err := conn.DisassociateMembers(ctx, &input)
			if err == nil {
				err = unprocessedAccountsError(output.UnprocessedAccounts)
			}
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating GuardDuty Member (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GuardDutyClient(ctx)

	detectorID, accountID, err := memberParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting GuardDuty Member: %s", d.Id())
	input := guardduty.DeleteMembersInput{
		AccountIds: []string{accountID},
		DetectorId: aws.String(detectorID),
	}
	output, err := conn.DeleteMembers(ctx, &input)
	if err == nil {
		err = unprocessedAccountsError(output.UnprocessedAccounts)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GuardDuty Member (%s): %s", d.Id(), err)
	}

	return diags
}

const memberResourceIDSeparator = ":"

func memberCreateResourceID(detectorID, accountID string) string {
	parts := []string{detectorID, accountID}
	id := strings.Join(parts, memberResourceIDSeparator)

	return id
}

func memberParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, memberResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected <Detector ID>%[2]s<Member AWS Account ID>", id, memberResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findMember(ctx context.Context, conn *guardduty.Client, input *guardduty.GetMembersInput) (*awstypes.Member, error) {
	output, err := findMembers(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findMembers(ctx context.Context, conn *guardduty.Client, input *guardduty.GetMembersInput) ([]awstypes.Member, error) {
	output, err := conn.GetMembers(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Members, nil
}

func findMemberByTwoPartKey(ctx context.Context, conn *guardduty.Client, detectorID, accountID string) (*awstypes.Member, error) {
	input := guardduty.GetMembersInput{
		AccountIds: []string{accountID},
		DetectorId: aws.String(detectorID),
	}

	return findMember(ctx, conn, &input)
}

func statusMember(conn *guardduty.Client, detectorID, accountID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findMemberByTwoPartKey(ctx, conn, detectorID, accountID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.RelationshipStatus), nil
	}
}

func waitMemberInvited(ctx context.Context, conn *guardduty.Client, detectorID, accountID string, timeout time.Duration) (*awstypes.Member, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{memberRelationshipStatusCreated, memberRelationshipStatusEmailVerificationInProgress},
		Target:  []string{memberRelationshipStatusDisabled, memberRelationshipStatusEnabled, memberRelationshipStatusInvited},
		Refresh: statusMember(conn, detectorID, accountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Member); ok {
		return output, err
	}

	return nil, err
}

const (
	memberRelationshipStatusCreated                     = "Created"
	memberRelationshipStatusDisabled                    = "Disabled"
	memberRelationshipStatusEnabled                     = "Enabled"
	memberRelationshipStatusEmailVerificationInProgress = "EmailVerificationInProgress"
	memberRelationshipStatusInvited                     = "Invited"
)

func unprocessedAccountError(apiObject awstypes.UnprocessedAccount) error {
	return fmt.Errorf("%s: %s", aws.ToString(apiObject.AccountId), aws.ToString(apiObject.Result))
}

func unprocessedAccountsError(apiObjects []awstypes.UnprocessedAccount) error {
	var errs []error

	for _, apiObject := range apiObjects {
		errs = append(errs, unprocessedAccountError(apiObject))
	}

	return errors.Join(errs...)
}
