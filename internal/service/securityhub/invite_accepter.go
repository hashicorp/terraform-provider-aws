// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_securityhub_invite_accepter")
func ResourceInviteAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInviteAccepterCreate,
		ReadWithoutTimeout:   resourceInviteAccepterRead,
		DeleteWithoutTimeout: resourceInviteAccepterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"master_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"invitation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceInviteAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)
	log.Print("[DEBUG] Accepting Security Hub invitation")

	invitationId, err := resourceInviteAccepterGetInvitationID(ctx, conn, d.Get("master_id").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Security Hub invitation: %s", err)
	}

	_, err = conn.AcceptInvitationWithContext(ctx, &securityhub.AcceptInvitationInput{
		InvitationId: aws.String(invitationId),
		MasterId:     aws.String(d.Get("master_id").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Security Hub invitation: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return append(diags, resourceInviteAccepterRead(ctx, d, meta)...)
}

func resourceInviteAccepterGetInvitationID(ctx context.Context, conn *securityhub.SecurityHub, masterId string) (string, error) {
	log.Printf("[DEBUG] Getting InvitationId for MasterId %s", masterId)

	resp, err := conn.ListInvitationsWithContext(ctx, &securityhub.ListInvitationsInput{})

	if err != nil {
		return "", fmt.Errorf("listing Security Hub invitations: %w", err)
	}

	for _, invitation := range resp.Invitations {
		log.Printf("[DEBUG] Invitation: %s", invitation)
		if aws.StringValue(invitation.AccountId) == masterId {
			return *invitation.InvitationId, nil
		}
	}

	return "", fmt.Errorf("Cannot find InvitationId for MasterId %s", masterId)
}

func resourceInviteAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)
	log.Print("[DEBUG] Reading Security Hub master account")

	resp, err := conn.GetMasterAccountWithContext(ctx, &securityhub.GetMasterAccountInput{})
	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		log.Print("[WARN] Security Hub master account not found, removing from state")
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "retrieving Security Hub master account: %s", err)
	}

	master := resp.Master

	if master == nil {
		log.Print("[WARN] Security Hub master account not found, removing from state")
		d.SetId("")
		return diags
	}

	d.Set("invitation_id", master.InvitationId)
	d.Set("master_id", master.AccountId)

	return diags
}

func resourceInviteAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)
	log.Print("[DEBUG] Disassociating from Security Hub master account")

	_, err := conn.DisassociateFromMasterAccountWithContext(ctx, &securityhub.DisassociateFromMasterAccountInput{})

	if tfawserr.ErrMessageContains(err, "BadRequestException", "The request is rejected because the current account is not associated to a master account") {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating from Security Hub master account: %s", err)
	}

	return diags
}
