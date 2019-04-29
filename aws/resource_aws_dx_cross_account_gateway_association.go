package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/resource"
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

			"associated_gateway_owner_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
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

			"associated_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"associated_gateway_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dx_gateway_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dx_gateway_owner_account_id": {
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
		AssociatedGatewayOwnerAccount:                 aws.String(d.Get("associated_gateway_owner_account_id").(string)),
		DirectConnectGatewayId:                        aws.String(dxgwId),
		OverrideAllowedPrefixesToDirectConnectGateway: expandDxRouteFilterPrefixes(d.Get("allowed_prefixes").(*schema.Set)),
		ProposalId: aws.String(d.Get("proposal_id").(string)),
	}

	log.Printf("[DEBUG] Accepting Direct Connect gateway association proposal: %#v", req)
	resp, err := conn.AcceptDirectConnectGatewayAssociationProposal(req)
	if err != nil {
		return fmt.Errorf("error accepting Direct Connect gateway association proposal: %s", err)
	}

	// For historical reasons the resource ID isn't set to the association ID returned from the API.
	associationId := aws.StringValue(resp.DirectConnectGatewayAssociation.AssociationId)
	d.SetId(dxGatewayAssociationId(dxgwId, aws.StringValue(resp.DirectConnectGatewayAssociation.AssociatedGateway.Id)))
	d.Set("dx_gateway_association_id", associationId)

	if err := waitForDirectConnectGatewayAssociationAvailabilityOnCreate(conn, associationId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for cross-account Direct Connect gateway association (%s) to become available: %s", d.Id(), err)
	}

	return resourceAwsDxCrossAccountGatewayAssociationRead(d, meta)
}

func resourceAwsDxCrossAccountGatewayAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	assocRaw, state, err := my_dxGatewayAssociationStateRefresh(conn, d.Get("dx_gateway_association_id").(string))()
	if err != nil {
		return fmt.Errorf("error reading cross-account Direct Connect gateway association (%s): %s", d.Id(), err)
	}
	if state == gatewayAssociationStateDeleted {
		log.Printf("[WARN] Cross-account Direct Connect gateway association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	assoc := assocRaw.(*directconnect.GatewayAssociation)
	d.Set("associated_gateway_id", assoc.AssociatedGateway.Id)
	d.Set("associated_gateway_owner_account_id", assoc.VirtualGatewayOwnerAccount)
	d.Set("associated_gateway_type", assoc.AssociatedGateway.Type)
	d.Set("dx_gateway_association_id", assoc.AssociationId)
	d.Set("dx_gateway_id", assoc.DirectConnectGatewayId)
	d.Set("dx_gateway_owner_account_id", assoc.DirectConnectGatewayOwnerAccount)
	err = d.Set("allowed_prefixes", flattenDxRouteFilterPrefixes(assoc.AllowedPrefixesToDirectConnectGateway))
	if err != nil {
		return fmt.Errorf("error setting allowed_prefixes: %s", err)
	}

	return nil
}

func resourceAwsDxCrossAccountGatewayAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	if d.HasChange("allowed_prefixes") {
		associationId := d.Get("dx_gateway_association_id").(string)

		oraw, nraw := d.GetChange("allowed_prefixes")
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)
		del := o.Difference(n)
		add := n.Difference(o)

		req := &directconnect.UpdateDirectConnectGatewayAssociationInput{
			AddAllowedPrefixesToDirectConnectGateway:    expandDxRouteFilterPrefixes(add),
			AssociationId:                               aws.String(associationId),
			RemoveAllowedPrefixesToDirectConnectGateway: expandDxRouteFilterPrefixes(del),
		}

		log.Printf("[DEBUG] Updating cross-account Direct Connect gateway association: %#v", req)
		_, err := conn.UpdateDirectConnectGatewayAssociation(req)
		if err != nil {
			return fmt.Errorf("error updating cross-account Direct Connect gateway association (%s): %s", d.Id(), err)
		}

		if err := waitForDirectConnectGatewayAssociationAvailabilityOnUpdate(conn, associationId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for cross-account Direct Connect gateway association (%s) to become available: %s", d.Id(), err)
		}
	}

	return resourceAwsDxCrossAccountGatewayAssociationRead(d, meta)
}

func resourceAwsDxCrossAccountGatewayAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	associationId := d.Get("dx_gateway_association_id").(string)

	log.Printf("[DEBUG] Deleting cross-account Direct Connect gateway association: %s", d.Id())
	_, err := conn.DeleteDirectConnectGatewayAssociation(&directconnect.DeleteDirectConnectGatewayAssociationInput{
		AssociationId: aws.String(associationId),
	})
	if isAWSErr(err, directconnect.ErrCodeClientException, "No association exists") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting cross-account Direct Connect gateway association: %s", err)
	}

	if err := my_waitForDirectConnectGatewayAssociationDeletion(conn, associationId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for cross-account Direct Connect gateway association (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

// TODO
// TODO Remove these four once https://github.com/terraform-providers/terraform-provider-aws/pull/8528 is merged.
// TODO

func my_dxGatewayAssociationStateRefresh(conn *directconnect.DirectConnect, associationId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeDirectConnectGatewayAssociations(&directconnect.DescribeDirectConnectGatewayAssociationsInput{
			AssociationId: aws.String(associationId),
		})
		if err != nil {
			return nil, "", err
		}

		n := len(resp.DirectConnectGatewayAssociations)
		switch n {
		case 0:
			return "", gatewayAssociationStateDeleted, nil

		case 1:
			assoc := resp.DirectConnectGatewayAssociations[0]

			if stateChangeError := aws.StringValue(assoc.StateChangeError); stateChangeError != "" {
				id := dxGatewayAssociationId(
					aws.StringValue(resp.DirectConnectGatewayAssociations[0].DirectConnectGatewayId),
					aws.StringValue(resp.DirectConnectGatewayAssociations[0].AssociatedGateway.Id))
				log.Printf("[INFO] Direct Connect gateway association (%s) state change error: %s", id, stateChangeError)
			}

			return assoc, aws.StringValue(assoc.AssociationState), nil

		default:
			return nil, "", fmt.Errorf("Found %d Direct Connect gateway associations for %s, expected 1", n, associationId)
		}
	}
}

func waitForDirectConnectGatewayAssociationAvailabilityOnCreate(conn *directconnect.DirectConnect, associationId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.GatewayAssociationStateAssociating},
		Target:     []string{directconnect.GatewayAssociationStateAssociated},
		Refresh:    my_dxGatewayAssociationStateRefresh(conn, associationId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForDirectConnectGatewayAssociationAvailabilityOnUpdate(conn *directconnect.DirectConnect, associationId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.GatewayAssociationStateUpdating},
		Target:     []string{directconnect.GatewayAssociationStateAssociated},
		Refresh:    my_dxGatewayAssociationStateRefresh(conn, associationId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func my_waitForDirectConnectGatewayAssociationDeletion(conn *directconnect.DirectConnect, associationId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.GatewayAssociationStateDisassociating},
		Target:     []string{directconnect.GatewayAssociationStateDisassociated, gatewayAssociationStateDeleted},
		Refresh:    my_dxGatewayAssociationStateRefresh(conn, associationId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
