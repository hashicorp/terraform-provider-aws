package directconnect

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGatewayAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceGatewayAssociationCreate,
		Read:   resourceGatewayAssociationRead,
		Update: resourceGatewayAssociationUpdate,
		Delete: resourceGatewayAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: resourceGatewayAssociationImport,
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

func resourceGatewayAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

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
		output, err := conn.AcceptDirectConnectGatewayAssociationProposal(input)

		if err != nil {
			return fmt.Errorf("error accepting Direct Connect Gateway Association Proposal (%s): %w", proposalID, err)
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
		output, err := conn.CreateDirectConnectGatewayAssociation(input)

		if err != nil {
			return fmt.Errorf("error creating Direct Connect Gateway Association (%s/%s): %w", directConnectGatewayID, associatedGatewayID, err)
		}

		// For historical reasons the resource ID isn't set to the association ID returned from the API.
		associationID = aws.StringValue(output.DirectConnectGatewayAssociation.AssociationId)
		d.SetId(GatewayAssociationCreateResourceID(directConnectGatewayID, associatedGatewayID))
	}

	d.Set("dx_gateway_association_id", associationID)

	if _, err := waitGatewayAssociationCreated(conn, associationID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Direct Connect Gateway Association (%s) to create: %w", d.Id(), err)
	}

	return resourceGatewayAssociationRead(d, meta)
}

func resourceGatewayAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	associationID := d.Get("dx_gateway_association_id").(string)

	output, err := FindGatewayAssociationByID(conn, associationID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Gateway Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Gateway Association (%s): %w", d.Id(), err)
	}

	if err := d.Set("allowed_prefixes", flattenRouteFilterPrefixes(output.AllowedPrefixesToDirectConnectGateway)); err != nil {
		return fmt.Errorf("error setting allowed_prefixes: %w", err)
	}

	d.Set("associated_gateway_id", output.AssociatedGateway.Id)
	d.Set("associated_gateway_owner_account_id", output.AssociatedGateway.OwnerAccount)
	d.Set("associated_gateway_type", output.AssociatedGateway.Type)
	d.Set("dx_gateway_association_id", output.AssociationId)
	d.Set("dx_gateway_id", output.DirectConnectGatewayId)
	d.Set("dx_gateway_owner_account_id", output.DirectConnectGatewayOwnerAccount)

	return nil
}

func resourceGatewayAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

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
	_, err := conn.UpdateDirectConnectGatewayAssociation(input)

	if err != nil {
		return fmt.Errorf("error updating Direct Connect Gateway Association (%s): %w", d.Id(), err)
	}

	if _, err := waitGatewayAssociationUpdated(conn, associationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("error waiting for Direct Connect Gateway Association (%s) to update: %w", d.Id(), err)
	}

	return resourceGatewayAssociationRead(d, meta)
}

func resourceGatewayAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	associationID := d.Get("dx_gateway_association_id").(string)

	log.Printf("[DEBUG] Deleting Direct Connect Gateway Association: %s", d.Id())
	_, err := conn.DeleteDirectConnectGatewayAssociation(&directconnect.DeleteDirectConnectGatewayAssociationInput{
		AssociationId: aws.String(associationID),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "does not exist") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Direct Connect Gateway Association (%s): %w", d.Id(), err)
	}

	if _, err := waitGatewayAssociationDeleted(conn, associationID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Direct Connect Gateway Association (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func resourceGatewayAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	parts := strings.Split(d.Id(), "/")

	if len(parts) != 2 {
		return nil, fmt.Errorf("Incorrect resource ID format: %q. Expected DXGATEWAYID/ASSOCIATEDGATEWAYID", d.Id())
	}

	directConnectGatewayID := parts[0]
	associatedGatewayID := parts[1]

	output, err := FindGatewayAssociationByGatewayIDAndAssociatedGatewayID(conn, directConnectGatewayID, associatedGatewayID)

	if err != nil {
		return nil, err
	}

	d.SetId(GatewayAssociationCreateResourceID(directConnectGatewayID, associatedGatewayID))
	d.Set("dx_gateway_id", output.DirectConnectGatewayId)
	d.Set("dx_gateway_association_id", output.AssociationId)

	return []*schema.ResourceData{d}, nil
}
