// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_member", name="Member")
// @Tags
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrEmail: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"relationship_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"administrator_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invited_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MacieStatus](),
			},
			"invite": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"invitation_disable_email_notification": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"invitation_message": {
				Type:     schema.TypeString,
				Optional: true,
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

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	accountId := d.Get(names.AttrAccountID).(string)
	input := &macie2.CreateMemberInput{
		Account: &awstypes.AccountDetail{
			AccountId: aws.String(accountId),
			Email:     aws.String(d.Get(names.AttrEmail).(string)),
		},
		Tags: getTagsIn(ctx),
	}

	var err error
	err = retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		_, err := conn.CreateMember(ctx, input)

		if tfawserr.ErrCodeEquals(err, string(awstypes.ErrorCodeClientError)) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateMember(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie Member: %s", err)
	}

	d.SetId(accountId)

	if !d.Get("invite").(bool) {
		return append(diags, resourceMemberRead(ctx, d, meta)...)
	}

	// Invitation workflow

	inputInvite := &macie2.CreateInvitationsInput{
		AccountIds: []string{d.Id()},
	}

	if v, ok := d.GetOk("invitation_disable_email_notification"); ok {
		inputInvite.DisableEmailNotification = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("invitation_message"); ok {
		inputInvite.Message = aws.String(v.(string))
	}

	log.Printf("[INFO] Inviting Macie2 Member: %+v", inputInvite)

	var output *macie2.CreateInvitationsOutput
	err = retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		output, err = conn.CreateInvitations(ctx, inputInvite)

		if tfawserr.ErrCodeEquals(err, string(awstypes.ErrorCodeClientError)) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateInvitations(ctx, inputInvite)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "inviting Macie Member: %s", err)
	}

	if len(output.UnprocessedAccounts) != 0 {
		return sdkdiag.AppendErrorf(diags, "inviting Macie Member: %s: %s", output.UnprocessedAccounts[0].ErrorCode, aws.ToString(output.UnprocessedAccounts[0].ErrorMessage))
	}

	if _, err = waitMemberInvited(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Macie Member (%s) invitation: %s", d.Id(), err)
	}

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	output, err := findMemberByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Macie Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie Member (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, output.AccountId)
	d.Set("administrator_account_id", output.AdministratorAccountId)
	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrEmail, output.Email)
	d.Set("invited_at", aws.ToTime(output.InvitedAt).Format(time.RFC3339))
	d.Set("master_account_id", output.MasterAccountId)
	d.Set("relationship_status", output.RelationshipStatus)
	d.Set("updated_at", aws.ToTime(output.UpdatedAt).Format(time.RFC3339))

	setTagsOut(ctx, output.Tags)

	relationshipStatus := output.RelationshipStatus
	if relationshipStatus == awstypes.RelationshipStatusEnabled ||
		relationshipStatus == awstypes.RelationshipStatusInvited ||
		relationshipStatus == awstypes.RelationshipStatusEmailVerificationInProgress ||
		relationshipStatus == awstypes.RelationshipStatusPaused {
		d.Set("invite", true)
	}

	if relationshipStatus == awstypes.RelationshipStatusRemoved {
		d.Set("invite", false)
	}

	// To fake a result for status in order to avoid an error related to difference for ImportVerifyState
	// It sets to MacieStatusPaused because it can only be changed to PAUSED, normally when it's accepted its status is ENABLED
	macieStatus := awstypes.MacieStatusEnabled
	if relationshipStatus == awstypes.RelationshipStatusPaused {
		macieStatus = awstypes.MacieStatusPaused
	}
	d.Set(names.AttrStatus, macieStatus)

	return diags
}

func resourceMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	// Invitation workflow

	if d.HasChange("invite") {
		if d.Get("invite").(bool) {
			inputInvite := &macie2.CreateInvitationsInput{
				AccountIds: []string{d.Id()},
			}

			if v, ok := d.GetOk("invitation_disable_email_notification"); ok {
				inputInvite.DisableEmailNotification = aws.Bool(v.(bool))
			}
			if v, ok := d.GetOk("invitation_message"); ok {
				inputInvite.Message = aws.String(v.(string))
			}

			log.Printf("[INFO] Inviting Macie2 Member: %+v", inputInvite)
			var output *macie2.CreateInvitationsOutput
			var err error
			err = retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
				output, err = conn.CreateInvitations(ctx, inputInvite)

				if tfawserr.ErrCodeEquals(err, string(awstypes.ErrorCodeClientError)) {
					return retry.RetryableError(err)
				}

				if err != nil {
					return retry.NonRetryableError(err)
				}

				return nil
			})

			if tfresource.TimedOut(err) {
				output, err = conn.CreateInvitations(ctx, inputInvite)
			}

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "inviting Macie Member: %s", err)
			}

			if len(output.UnprocessedAccounts) != 0 {
				return sdkdiag.AppendErrorf(diags, "inviting Macie Member: %s: %s", output.UnprocessedAccounts[0].ErrorCode, aws.ToString(output.UnprocessedAccounts[0].ErrorMessage))
			}

			if _, err = waitMemberInvited(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Macie Member (%s) invitation: %s", d.Id(), err)
			}
		} else {
			input := &macie2.DisassociateMemberInput{
				Id: aws.String(d.Id()),
			}

			_, err := conn.DisassociateMember(ctx, input)
			if err != nil {
				if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
					errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
					return diags
				}
				return sdkdiag.AppendErrorf(diags, "disassociating Macie Member invite (%s): %s", d.Id(), err)
			}
		}
	}

	// End Invitation workflow

	if d.HasChange(names.AttrStatus) {
		input := &macie2.UpdateMemberSessionInput{
			Id:     aws.String(d.Id()),
			Status: awstypes.MacieStatus(d.Get(names.AttrStatus).(string)),
		}

		_, err := conn.UpdateMemberSession(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Macie Member (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	_, err := conn.DeleteMember(ctx, &macie2.DeleteMemberInput{
		Id: aws.String(d.Id()),
	})

	if memberNotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Macie Member (%s): %s", d.Id(), err)
	}

	return diags
}

func findMemberByID(ctx context.Context, conn *macie2.Client, id string) (*macie2.GetMemberOutput, error) {
	input := &macie2.GetMemberInput{
		Id: aws.String(id),
	}

	return findMember(ctx, conn, input)
}

func findMember(ctx context.Context, conn *macie2.Client, input *macie2.GetMemberInput) (*macie2.GetMemberOutput, error) {
	output, err := conn.GetMember(ctx, input)

	if memberNotFound(err) {
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

	return output, nil
}

func memberNotFound(err error) bool {
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return true
	}
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
		return true
	}
	if errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "member accounts are associated with your account") {
		return true
	}
	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "account is not associated with your account") {
		return true
	}
	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "not an associated member of account") {
		return true
	}

	return false
}
