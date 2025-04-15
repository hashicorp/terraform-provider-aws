// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_inspector2_member_association", name="Member Association")
func resourceMemberAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMemberAssociationCreate,
		ReadWithoutTimeout:   resourceMemberAssociationRead,
		DeleteWithoutTimeout: resourceMemberAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"delegated_admin_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"relationship_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMemberAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	input := &inspector2.AssociateMemberInput{
		AccountId: aws.String(accountID),
	}

	_, err := conn.AssociateMember(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector2 Member Association (%s): %s", accountID, err)
	}

	d.SetId(accountID)

	if _, err := waitMemberAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Inspector2 Member Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceMemberAssociationRead(ctx, d, meta)...)
}

func resourceMemberAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	member, err := findMemberByAccountID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 Member Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector2 Member Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, member.AccountId)
	d.Set("delegated_admin_account_id", member.DelegatedAdminAccountId)
	d.Set("relationship_status", member.RelationshipStatus)
	d.Set("updated_at", aws.ToTime(member.UpdatedAt).Format(time.RFC3339))

	return diags
}

func resourceMemberAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	log.Printf("[DEBUG] Deleting Inspector2 Member Association: %s", d.Id())
	_, err := conn.DisassociateMember(ctx, &inspector2.DisassociateMemberInput{
		AccountId: aws.String(d.Id()),
	})

	// An error occurred (ValidationException) when calling the DisassociateMember operation: The request is rejected because the current account cannot disassociate the given member account ID since the latter is not yet associated to it.
	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "is not yet associated to it") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Inspector2 Member Association (%s): %s", d.Id(), err)
	}

	if _, err := waitMemberAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Inspector2 Member Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findMemberByAccountID(ctx context.Context, conn *inspector2.Client, id string) (*awstypes.Member, error) {
	input := &inspector2.GetMemberInput{
		AccountId: aws.String(id),
	}
	output, err := findMember(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.RelationshipStatus; status == awstypes.RelationshipStatusRemoved {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findMember(ctx context.Context, conn *inspector2.Client, input *inspector2.GetMemberInput) (*awstypes.Member, error) {
	output, err := conn.GetMember(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Invoking account does not have access to get member account") || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Member == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Member, nil
}

func statusMemberAssociation(ctx context.Context, conn *inspector2.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findMemberByAccountID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, string(output.RelationshipStatus), nil
	}
}

func waitMemberAssociationCreated(ctx context.Context, conn *inspector2.Client, id string, timeout time.Duration) (*awstypes.Member, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RelationshipStatusCreated),
		Target:  enum.Slice(awstypes.RelationshipStatusEnabled),
		Refresh: statusMemberAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Member); ok {
		return output, err
	}

	return nil, err
}

func waitMemberAssociationDeleted(ctx context.Context, conn *inspector2.Client, id string, timeout time.Duration) (*awstypes.Member, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RelationshipStatusCreated, awstypes.RelationshipStatusEnabled),
		Target:  []string{},
		Refresh: statusMemberAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Member); ok {
		return output, err
	}

	return nil, err
}
