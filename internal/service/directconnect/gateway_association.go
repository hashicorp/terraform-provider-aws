// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_dx_gateway_association")
func ResourceGatewayAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGatewayAssociationCreate,
		ReadWithoutTimeout:   resourceGatewayAssociationRead,
		UpdateWithoutTimeout: resourceGatewayAssociationUpdate,
		DeleteWithoutTimeout: resourceGatewayAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceGatewayAssociationImport,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceGatewayAssociationResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: GatewayAssociationStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"associated_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"associated_gateway_owner_account_id", "proposal_id"},
				AtLeastOneOf:  []string{"associated_gateway_id", "associated_gateway_owner_account_id", "proposal_id"},
			},

			"associated_gateway_owner_account_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  verify.ValidAccountID,
				ConflictsWith: []string{"associated_gateway_id"},
				RequiredWith:  []string{"proposal_id"},
				AtLeastOneOf:  []string{"associated_gateway_id", "associated_gateway_owner_account_id", "proposal_id"},
			},

			"associated_gateway_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dx_gateway_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"dx_gateway_owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"proposal_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"associated_gateway_id", "vpn_gateway_id"},
				AtLeastOneOf:  []string{"associated_gateway_id", "associated_gateway_owner_account_id", "proposal_id"},
			},

			"vpn_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"associated_gateway_id", "associated_gateway_owner_account_id", "proposal_id"},
				Deprecated:    "use 'associated_gateway_id' argument instead",
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceGatewayAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	var associationID string
	directConnectGatewayID := d.Get("dx_gateway_id").(string)

	if associatedGatewayOwnerAccount := d.Get("associated_gateway_owner_account_id").(string); associatedGatewayOwnerAccount != "" {
		proposalID := d.Get("proposal_id").(string)
		input := &directconnect.AcceptDirectConnectGatewayAssociationProposalInput{
			AssociatedGatewayOwnerAccount: aws.String(associatedGatewayOwnerAccount),
			DirectConnectGatewayId:        aws.String(directConnectGatewayID),
			ProposalId:                    aws.String(proposalID),
		}

		if v, ok := d.GetOk("allowed_prefixes"); ok && v.(*schema.Set).Len() > 0 {
			input.OverrideAllowedPrefixesToDirectConnectGateway = expandRouteFilterPrefixes(v.(*schema.Set).List())
		}

		log.Printf("[DEBUG] Accepting Direct Connect Gateway Association Proposal: %s", input)
		output, err := conn.AcceptDirectConnectGatewayAssociationProposalWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "accepting Direct Connect Gateway Association Proposal (%s): %s", proposalID, err)
		}

		// For historical reasons the resource ID isn't set to the association ID returned from the API.
		associationID = aws.StringValue(output.DirectConnectGatewayAssociation.AssociationId)
		d.SetId(GatewayAssociationCreateResourceID(directConnectGatewayID, aws.StringValue(output.DirectConnectGatewayAssociation.AssociatedGateway.Id)))
	} else {
		associatedGatewayID := d.Get("associated_gateway_id").(string)
		input := &directconnect.CreateDirectConnectGatewayAssociationInput{
			DirectConnectGatewayId: aws.String(directConnectGatewayID),
			GatewayId:              aws.String(associatedGatewayID),
		}

		if v, ok := d.GetOk("allowed_prefixes"); ok && v.(*schema.Set).Len() > 0 {
			input.AddAllowedPrefixesToDirectConnectGateway = expandRouteFilterPrefixes(v.(*schema.Set).List())
		}

		log.Printf("[DEBUG] Creating Direct Connect Gateway Association: %s", input)
		output, err := conn.CreateDirectConnectGatewayAssociationWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Direct Connect Gateway Association (%s/%s): %s", directConnectGatewayID, associatedGatewayID, err)
		}

		// For historical reasons the resource ID isn't set to the association ID returned from the API.
		associationID = aws.StringValue(output.DirectConnectGatewayAssociation.AssociationId)
		d.SetId(GatewayAssociationCreateResourceID(directConnectGatewayID, associatedGatewayID))
	}

	d.Set("dx_gateway_association_id", associationID)

	if _, err := waitGatewayAssociationCreated(ctx, conn, associationID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway Association (%s) to create: %s", d.Id(), err)
	}

	return append(diags, resourceGatewayAssociationRead(ctx, d, meta)...)
}

func resourceGatewayAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	associationID := d.Get("dx_gateway_association_id").(string)

	output, err := FindGatewayAssociationByID(ctx, conn, associationID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Gateway Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Gateway Association (%s): %s", d.Id(), err)
	}

	if err := d.Set("allowed_prefixes", flattenRouteFilterPrefixes(output.AllowedPrefixesToDirectConnectGateway)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting allowed_prefixes: %s", err)
	}

	d.Set("associated_gateway_id", output.AssociatedGateway.Id)
	d.Set("associated_gateway_owner_account_id", output.AssociatedGateway.OwnerAccount)
	d.Set("associated_gateway_type", output.AssociatedGateway.Type)
	d.Set("dx_gateway_association_id", output.AssociationId)
	d.Set("dx_gateway_id", output.DirectConnectGatewayId)
	d.Set("dx_gateway_owner_account_id", output.DirectConnectGatewayOwnerAccount)

	return diags
}

func resourceGatewayAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	associationID := d.Get("dx_gateway_association_id").(string)
	input := &directconnect.UpdateDirectConnectGatewayAssociationInput{
		AssociationId: aws.String(associationID),
	}

	oraw, nraw := d.GetChange("allowed_prefixes")
	o, n := oraw.(*schema.Set), nraw.(*schema.Set)

	if add := n.Difference(o); add.Len() > 0 {
		input.AddAllowedPrefixesToDirectConnectGateway = expandRouteFilterPrefixes(add.List())
	}

	if del := o.Difference(n); del.Len() > 0 {
		input.RemoveAllowedPrefixesToDirectConnectGateway = expandRouteFilterPrefixes(del.List())
	}

	log.Printf("[DEBUG] Updating Direct Connect Gateway Association: %s", input)
	_, err := conn.UpdateDirectConnectGatewayAssociationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Direct Connect Gateway Association (%s): %s", d.Id(), err)
	}

	if _, err := waitGatewayAssociationUpdated(ctx, conn, associationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway Association (%s) to update: %s", d.Id(), err)
	}

	return append(diags, resourceGatewayAssociationRead(ctx, d, meta)...)
}

func resourceGatewayAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	associationID := d.Get("dx_gateway_association_id").(string)

	log.Printf("[DEBUG] Deleting Direct Connect Gateway Association: %s", d.Id())
	_, err := conn.DeleteDirectConnectGatewayAssociationWithContext(ctx, &directconnect.DeleteDirectConnectGatewayAssociationInput{
		AssociationId: aws.String(associationID),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect Gateway Association (%s): %s", d.Id(), err)
	}

	if _, err := waitGatewayAssociationDeleted(ctx, conn, associationID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway Association (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func resourceGatewayAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	parts := strings.Split(d.Id(), "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("Incorrect resource ID format: %q. Expected DXGATEWAYID/ASSOCIATEDGATEWAYID", d.Id())
	}

	directConnectGatewayID := parts[0]
	associatedGatewayID := parts[1]

	output, err := FindGatewayAssociationByGatewayIDAndAssociatedGatewayID(ctx, conn, directConnectGatewayID, associatedGatewayID)

	if err != nil {
		return nil, err
	}

	d.SetId(GatewayAssociationCreateResourceID(directConnectGatewayID, associatedGatewayID))
	d.Set("dx_gateway_id", output.DirectConnectGatewayId)
	d.Set("dx_gateway_association_id", output.AssociationId)

	return []*schema.ResourceData{d}, nil
}
