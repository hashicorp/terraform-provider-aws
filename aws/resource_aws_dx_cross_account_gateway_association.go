package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxCrossAccountGatewayAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxCrossAccountGatewayAssociationCreate,
		Read:   resourceAwsDxCrossAccountGatewayAssociationRead,
		Update: resourceAwsDxCrossAccountGatewayAssociationUpdate,
		Delete: resourceAwsDxCrossAccountGatewayAssociationDelete,

		Schema: map[string]*schema.Schema{
			"allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"proposal_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpn_gateway_owner_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},

			"dx_gateway_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsDxCrossAccountGatewayAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxgwId := d.Get("dx_gateway_id").(string)
	req := &directconnect.AcceptDirectConnectGatewayAssociationProposalInput{
		AssociatedGatewayOwnerAccount:                 aws.String(d.Get("vpn_gateway_owner_account_id").(string)),
		DirectConnectGatewayId:                        aws.String(dxgwId),
		OverrideAllowedPrefixesToDirectConnectGateway: expandDxRouteFilterPrefixes(d.Get("allowed_prefixes").(*schema.Set)),
		ProposalId: aws.String(d.Get("proposal_id").(string)),
	}

	log.Printf("[DEBUG] Accepting Direct Connect gateway association proposal: %#v", req)
	resp, err := conn.AcceptDirectConnectGatewayAssociationProposal(req)
	if err != nil {
		return fmt.Errorf("error accepting Direct Connect gateway association proposal: %s", err)
	}

	vgwId := aws.StringValue(resp.DirectConnectGatewayAssociation.VirtualGatewayId)
	d.SetId(dxGatewayAssociationId(dxgwId, vgwId))
	d.Set("vpn_gateway_id", vgwId)

	if err := waitForDirectConnectGatewayAssociationAvailabilityOnCreate(conn, dxgwId, vgwId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for cross-account Direct Connect gateway association (%s) to become available: %s", d.Id(), err)
	}

	return resourceAwsDxCrossAccountGatewayAssociationRead(d, meta)
}

func resourceAwsDxCrossAccountGatewayAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	assocRaw, state, err := dxGatewayAssociationStateRefresh(conn, d.Get("dx_gateway_id").(string), d.Get("vpn_gateway_id").(string))()
	if err != nil {
		return fmt.Errorf("error reading cross-account Direct Connect gateway association (%s): %s", d.Id(), err)
	}
	if state == gatewayAssociationStateDeleted {
		log.Printf("[WARN] Cross-account Direct Connect gateway association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	assoc := assocRaw.(*directconnect.GatewayAssociation)
	d.Set("dx_gateway_id", assoc.DirectConnectGatewayId)
	d.Set("vpn_gateway_id", assoc.VirtualGatewayId)
	d.Set("vpn_gateway_owner_account_id", assoc.VirtualGatewayOwnerAccount)
	d.Set("dx_gateway_association_id", assoc.AssociationId)
	err = d.Set("allowed_prefixes", flattenDxRouteFilterPrefixes(assoc.AllowedPrefixesToDirectConnectGateway))
	if err != nil {
		return fmt.Errorf("error setting allowed_prefixes: %s", err)
	}

	return nil
}

func resourceAwsDxCrossAccountGatewayAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxgwId := d.Get("dx_gateway_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)

	if d.HasChange("allowed_prefixes") {
		oraw, nraw := d.GetChange("allowed_prefixes")
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)
		del := o.Difference(n)
		add := n.Difference(o)

		req := &directconnect.UpdateDirectConnectGatewayAssociationInput{
			AddAllowedPrefixesToDirectConnectGateway:    expandDxRouteFilterPrefixes(add),
			AssociationId:                               aws.String(d.Get("dx_gateway_association_id").(string)),
			RemoveAllowedPrefixesToDirectConnectGateway: expandDxRouteFilterPrefixes(del),
		}

		log.Printf("[DEBUG] Updating cross-account Direct Connect gateway association: %#v", req)
		_, err := conn.UpdateDirectConnectGatewayAssociation(req)
		if err != nil {
			return fmt.Errorf("error updating cross-account Direct Connect gateway association (%s): %s", d.Id(), err)
		}

		if err := waitForDirectConnectGatewayAssociationAvailabilityOnUpdate(conn, dxgwId, vgwId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for cross-account Direct Connect gateway association (%s) to become available: %s", d.Id(), err)
		}
	}

	return resourceAwsDxCrossAccountGatewayAssociationRead(d, meta)
}

func resourceAwsDxCrossAccountGatewayAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxgwId := d.Get("dx_gateway_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)

	log.Printf("[DEBUG] Deleting cross-account Direct Connect gateway association: %s", d.Id())
	_, err := conn.DeleteDirectConnectGatewayAssociation(&directconnect.DeleteDirectConnectGatewayAssociationInput{
		DirectConnectGatewayId: aws.String(dxgwId),
		VirtualGatewayId:       aws.String(vgwId),
	})
	if isAWSErr(err, directconnect.ErrCodeClientException, "No association exists") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting cross-account Direct Connect gateway association: %s", err)
	}

	if err := waitForDirectConnectGatewayAssociationDeletion(conn, dxgwId, vgwId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for cross-account Direct Connect gateway association (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}
