// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_dx_gateway_association_proposal")
func ResourceGatewayAssociationProposal() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGatewayAssociationProposalCreate,
		ReadWithoutTimeout:   resourceGatewayAssociationProposalRead,
		DeleteWithoutTimeout: resourceGatewayAssociationProposalDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceGatewayAssociationProposalImport,
		},

		CustomizeDiff: customdiff.Sequence(
			// Accepting the proposal with overridden prefixes changes the returned RequestedAllowedPrefixesToDirectConnectGateway value (allowed_prefixes attribute).
			// We only want to force a new resource if this value changes and the current proposal state is "requested".
			customdiff.ForceNewIf("allowed_prefixes", func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

				log.Printf("[DEBUG] CustomizeDiff for Direct Connect Gateway Association Proposal (%s) allowed_prefixes", d.Id())

				output, err := FindGatewayAssociationProposalByID(ctx, conn, d.Id())

				if tfresource.NotFound(err) {
					// Proposal may be end-of-life and removed by AWS.
					return false
				}

				if err != nil {
					log.Printf("[ERROR] Error reading Direct Connect Gateway Association Proposal (%s): %s", d.Id(), err)
					return false
				}

				return aws.StringValue(output.ProposalState) == directconnect.GatewayAssociationProposalStateRequested
			}),
		),

		Schema: map[string]*schema.Schema{
			"allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"associated_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"associated_gateway_owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"associated_gateway_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"dx_gateway_owner_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
		},
	}
}

func resourceGatewayAssociationProposalCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	directConnectGatewayID := d.Get("dx_gateway_id").(string)
	associatedGatewayID := d.Get("associated_gateway_id").(string)
	input := &directconnect.CreateDirectConnectGatewayAssociationProposalInput{
		DirectConnectGatewayId:           aws.String(directConnectGatewayID),
		DirectConnectGatewayOwnerAccount: aws.String(d.Get("dx_gateway_owner_account_id").(string)),
		GatewayId:                        aws.String(associatedGatewayID),
	}

	if v, ok := d.GetOk("allowed_prefixes"); ok && v.(*schema.Set).Len() > 0 {
		input.AddAllowedPrefixesToDirectConnectGateway = expandRouteFilterPrefixes(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Creating Direct Connect Gateway Association Proposal: %s", input)
	output, err := conn.CreateDirectConnectGatewayAssociationProposalWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Gateway Association Proposal (%s/%s): %s", directConnectGatewayID, associatedGatewayID, err)
	}

	d.SetId(aws.StringValue(output.DirectConnectGatewayAssociationProposal.ProposalId))

	return append(diags, resourceGatewayAssociationProposalRead(ctx, d, meta)...)
}

func resourceGatewayAssociationProposalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	// First attempt to find by proposal ID.
	output, err := FindGatewayAssociationProposalByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		// Attempt to find an existing association.
		directConnectGatewayID := d.Get("dx_gateway_id").(string)
		associatedGatewayID := d.Get("associated_gateway_id").(string)

		output, err := FindGatewayAssociationByGatewayIDAndAssociatedGatewayID(ctx, conn, directConnectGatewayID, associatedGatewayID)

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Direct Connect Gateway Association Proposal (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Direct Connect Gateway Association (%s/%s): %s", directConnectGatewayID, associatedGatewayID, err)
		}

		// Once accepted, AWS will delete the proposal after after some time (days?).
		// In this case we don't need to create a new proposal, use metadata from the association
		// to artificially populate the missing proposal in state as if it was still there.
		log.Printf("[INFO] Direct Connect Gateway Association Proposal (%s) has reached end-of-life and has been removed by AWS", d.Id())

		if err := d.Set("allowed_prefixes", flattenRouteFilterPrefixes(output.AllowedPrefixesToDirectConnectGateway)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting allowed_prefixes: %s", err)
		}

		d.Set("associated_gateway_id", output.AssociatedGateway.Id)
		d.Set("associated_gateway_owner_account_id", output.AssociatedGateway.OwnerAccount)
		d.Set("associated_gateway_type", output.AssociatedGateway.Type)
		d.Set("dx_gateway_id", output.DirectConnectGatewayId)
		d.Set("dx_gateway_owner_account_id", output.DirectConnectGatewayOwnerAccount)
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Gateway Association Proposal (%s): %s", d.Id(), err)
	} else {
		if err := d.Set("allowed_prefixes", flattenRouteFilterPrefixes(output.RequestedAllowedPrefixesToDirectConnectGateway)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting allowed_prefixes: %s", err)
		}

		d.Set("associated_gateway_id", output.AssociatedGateway.Id)
		d.Set("associated_gateway_owner_account_id", output.AssociatedGateway.OwnerAccount)
		d.Set("associated_gateway_type", output.AssociatedGateway.Type)
		d.Set("dx_gateway_id", output.DirectConnectGatewayId)
		d.Set("dx_gateway_owner_account_id", output.DirectConnectGatewayOwnerAccount)
	}

	return diags
}

func resourceGatewayAssociationProposalDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	log.Printf("[DEBUG] Deleting Direct Connect Gateway Association Proposal: %s", d.Id())
	_, err := conn.DeleteDirectConnectGatewayAssociationProposalWithContext(ctx, &directconnect.DeleteDirectConnectGatewayAssociationProposalInput{
		ProposalId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "is not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect Gateway Association Proposal (%s): %s", d.Id(), err)
	}

	return diags
}

func expandRouteFilterPrefixes(tfList []interface{}) []*directconnect.RouteFilterPrefix {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*directconnect.RouteFilterPrefix

	for _, tfStringRaw := range tfList {
		tfString, ok := tfStringRaw.(string)

		if !ok {
			continue
		}

		apiObject := &directconnect.RouteFilterPrefix{
			Cidr: aws.String(tfString),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenRouteFilterPrefixes(apiObjects []*directconnect.RouteFilterPrefix) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, aws.StringValue(apiObject.Cidr))
	}

	return tfList
}

func resourceGatewayAssociationProposalImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	switch parts := strings.Split(strings.ToLower(d.Id()), "/"); len(parts) {
	case 1:
		break

	case 3:
		proposalID := parts[0]
		directConnectGatewayID := parts[1]
		associatedGatewayID := parts[2]

		if proposalID == "" || directConnectGatewayID == "" || associatedGatewayID == "" {
			return nil, fmt.Errorf("Incorrect resource ID format: %q. PROPOSALID, DXGATEWAYID and ASSOCIATEDGATEWAYID must not be empty strings", d.Id())
		}

		// Use pseudo-proposal ID and actual DirectConnectGatewayId and AssociatedGatewayId.
		d.SetId(proposalID)
		d.Set("associated_gateway_id", associatedGatewayID)
		d.Set("dx_gateway_id", directConnectGatewayID)

	default:
		return nil, fmt.Errorf("Incorrect resource ID format: %q. Expected PROPOSALID or PROPOSALID/DXGATEWAYID/ASSOCIATEDGATEWAYID", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}
