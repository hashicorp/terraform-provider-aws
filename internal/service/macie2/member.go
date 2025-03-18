// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_member", name="Member")
// @Tags(identifierAttribute="arn")
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
			"administrator_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEmail: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"invitation_disable_email_notification": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"invitation_message": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"invite": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"invited_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"relationship_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MacieStatus](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(4 * time.Minute),
			Update: schema.DefaultTimeout(4 * time.Minute),
		},
	}
}

func resourceMemberCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	input := macie2.CreateMemberInput{
		Account: &awstypes.AccountDetail{
			AccountId: aws.String(accountID),
			Email:     aws.String(d.Get(names.AttrEmail).(string)),
		},
		Tags: getTagsIn(ctx),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return conn.CreateMember(ctx, &input)
	}, errCodeClientError)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie Member (%s): %s", accountID, err)
	}

	d.SetId(accountID)

	if d.Get("invite").(bool) {
		if err := inviteMember(ctx, conn, d, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
	relationshipStatus := output.RelationshipStatus
	switch relationshipStatus {
	case awstypes.RelationshipStatusEnabled, awstypes.RelationshipStatusInvited, awstypes.RelationshipStatusEmailVerificationInProgress, awstypes.RelationshipStatusPaused:
		d.Set("invite", true)
	case awstypes.RelationshipStatusRemoved:
		d.Set("invite", false)
	}
	d.Set("invited_at", aws.ToTime(output.InvitedAt).Format(time.RFC3339))
	d.Set("master_account_id", output.MasterAccountId)
	d.Set("relationship_status", relationshipStatus)
	// To fake a result for status in order to avoid an error related to difference for ImportVerifyState.
	// It sets to MacieStatusPaused because it can only be changed to PAUSED, normally when it's accepted its status is ENABLED.
	macieStatus := awstypes.MacieStatusEnabled
	if relationshipStatus == awstypes.RelationshipStatusPaused {
		macieStatus = awstypes.MacieStatusPaused
	}
	d.Set(names.AttrStatus, macieStatus)
	d.Set("updated_at", aws.ToTime(output.UpdatedAt).Format(time.RFC3339))

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceMemberUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	if d.HasChange("invite") {
		if d.Get("invite").(bool) {
			if err := inviteMember(ctx, conn, d, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else {
			input := macie2.DisassociateMemberInput{
				Id: aws.String(d.Id()),
			}

			_, err := conn.DisassociateMember(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating Macie Member invite (%s): %s", d.Id(), err)
			}
		}
	}

	if d.HasChange(names.AttrStatus) {
		input := macie2.UpdateMemberSessionInput{
			Id:     aws.String(d.Id()),
			Status: awstypes.MacieStatus(d.Get(names.AttrStatus).(string)),
		}

		_, err := conn.UpdateMemberSession(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Macie Member (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	_, err := conn.DeleteMember(ctx, &macie2.DeleteMemberInput{
		Id: aws.String(d.Id()),
	})

	if isMemberNotFoundError(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Macie Member (%s): %s", d.Id(), err)
	}

	return diags
}

func inviteMember(ctx context.Context, conn *macie2.Client, d *schema.ResourceData, timeout time.Duration) error {
	input := macie2.CreateInvitationsInput{
		AccountIds: []string{d.Id()},
	}

	if v, ok := d.GetOk("invitation_disable_email_notification"); ok {
		input.DisableEmailNotification = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("invitation_message"); ok {
		input.Message = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (any, error) {
		return conn.CreateInvitations(ctx, &input)
	}, errCodeClientError)

	if err == nil {
		if output := outputRaw.(*macie2.CreateInvitationsOutput); output != nil {
			err = unprocessedAccountsError(output.UnprocessedAccounts)
		}
	}

	if err != nil {
		return fmt.Errorf("inviting Macie Member (%s): %w", d.Id(), err)
	}

	if _, err := waitMemberInvited(ctx, conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for Macie Member (%s) invite: %s", d.Id(), err)
	}

	return nil
}

func findMemberByID(ctx context.Context, conn *macie2.Client, id string) (*macie2.GetMemberOutput, error) {
	input := macie2.GetMemberInput{
		Id: aws.String(id),
	}

	output, err := conn.GetMember(ctx, &input)

	if isMemberNotFoundError(err) {
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

func findMemberNotAssociated(ctx context.Context, conn *macie2.Client, accountID string) (*awstypes.Member, error) {
	input := macie2.ListMembersInput{
		OnlyAssociated: aws.String("false"),
	}

	return findMember(ctx, conn, &input, func(v *awstypes.Member) bool {
		return aws.ToString(v.AccountId) == accountID
	})
}

func findMember(ctx context.Context, conn *macie2.Client, input *macie2.ListMembersInput, filter tfslices.Predicate[*awstypes.Member]) (*awstypes.Member, error) {
	output, err := findMembers(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findMembers(ctx context.Context, conn *macie2.Client, input *macie2.ListMembersInput, filter tfslices.Predicate[*awstypes.Member]) ([]awstypes.Member, error) {
	var output []awstypes.Member

	pages := macie2.NewListMembersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Members {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusMemberRelationship(ctx context.Context, conn *macie2.Client, adminAccountID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findMemberNotAssociated(ctx, conn, adminAccountID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.RelationshipStatus), nil
	}
}

func waitMemberInvited(ctx context.Context, conn *macie2.Client, adminAccountID string) (*awstypes.Member, error) { //nolint:unparam
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RelationshipStatusCreated, awstypes.RelationshipStatusEmailVerificationInProgress),
		Target:  enum.Slice(awstypes.RelationshipStatusInvited, awstypes.RelationshipStatusEnabled, awstypes.RelationshipStatusPaused),
		Refresh: statusMemberRelationship(ctx, conn, adminAccountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Member); ok {
		return output, err
	}

	return nil, err
}

func isMemberNotFoundError(err error) bool {
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

func unprocessedAccountError(apiObject awstypes.UnprocessedAccount) error {
	return fmt.Errorf("%s: %s", apiObject.ErrorCode, aws.ToString(apiObject.ErrorMessage))
}

func unprocessedAccountsError(apiObjects []awstypes.UnprocessedAccount) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := unprocessedAccountError(apiObject); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws.ToString(apiObject.AccountId), err))
		}
	}

	return errors.Join(errs...)
}
