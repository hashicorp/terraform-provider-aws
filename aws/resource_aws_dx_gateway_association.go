package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	gatewayAssociationStateDeleted = "deleted"
)

func resourceAwsDxGatewayAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxGatewayAssociationCreate,
		Read:   resourceAwsDxGatewayAssociationRead,
		Update: resourceAwsDxGatewayAssociationUpdate,
		Delete: resourceAwsDxGatewayAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDxGatewayAssociationImport,
		},

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

			"vpn_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"dx_gateway_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
	}
}

func resourceAwsDxGatewayAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxgwId := d.Get("dx_gateway_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)
	req := &directconnect.CreateDirectConnectGatewayAssociationInput{
		AddAllowedPrefixesToDirectConnectGateway: expandDxRouteFilterPrefixes(d.Get("allowed_prefixes").(*schema.Set)),
		DirectConnectGatewayId:                   aws.String(dxgwId),
		VirtualGatewayId:                         aws.String(vgwId),
	}

	log.Printf("[DEBUG] Creating Direct Connect gateway association: %#v", req)
	_, err := conn.CreateDirectConnectGatewayAssociation(req)
	if err != nil {
		return fmt.Errorf("error creating Direct Connect gateway association: %s", err)
	}

	d.SetId(dxGatewayAssociationId(dxgwId, vgwId))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.GatewayAssociationStateAssociating},
		Target:     []string{directconnect.GatewayAssociationStateAssociated},
		Refresh:    dxGatewayAssociationStateRefresh(conn, dxgwId, vgwId),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for Direct Connect gateway association (%s) to become available: %s", d.Id(), err)
	}

	return resourceAwsDxGatewayAssociationRead(d, meta)
}

func resourceAwsDxGatewayAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxgwId := d.Get("dx_gateway_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)
	assocRaw, state, err := dxGatewayAssociationStateRefresh(conn, dxgwId, vgwId)()
	if err != nil {
		return fmt.Errorf("error reading Direct Connect gateway association: %s", err)
	}
	if state == gatewayAssociationStateDeleted {
		log.Printf("[WARN] Direct Connect gateway association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	assoc := assocRaw.(*directconnect.GatewayAssociation)
	d.Set("dx_gateway_id", assoc.DirectConnectGatewayId)
	d.Set("vpn_gateway_id", assoc.VirtualGatewayId)
	d.Set("dx_gateway_association_id", assoc.AssociationId)
	err = d.Set("allowed_prefixes", flattenDxRouteFilterPrefixes(assoc.AllowedPrefixesToDirectConnectGateway))
	if err != nil {
		return fmt.Errorf("error setting allowed_prefixes: %s", err)
	}

	return nil
}

func resourceAwsDxGatewayAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
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

		log.Printf("[DEBUG] Direct Connect gateway association: %#v", req)
		_, err := conn.UpdateDirectConnectGatewayAssociation(req)
		if err != nil {
			return fmt.Errorf("error updating Direct Connect gateway association (%s): %s", d.Id(), err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{directconnect.GatewayAssociationStateUpdating},
			Target:     []string{directconnect.GatewayAssociationStateAssociated},
			Refresh:    dxGatewayAssociationStateRefresh(conn, dxgwId, vgwId),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      10 * time.Second,
			MinTimeout: 5 * time.Second,
		}
		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for Direct Connect gateway association (%s) to become available: %s", d.Id(), err)
		}
	}

	return resourceAwsDxGatewayAssociationRead(d, meta)
}

func resourceAwsDxGatewayAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxgwId := d.Get("dx_gateway_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)

	log.Printf("[DEBUG] Deleting Direct Connect gateway association: %s", d.Id())
	_, err := conn.DeleteDirectConnectGatewayAssociation(&directconnect.DeleteDirectConnectGatewayAssociationInput{
		DirectConnectGatewayId: aws.String(dxgwId),
		VirtualGatewayId:       aws.String(vgwId),
	})
	if isAWSErr(err, directconnect.ErrCodeClientException, "No association exists") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Direct Connect gateway association: %s", err)
	}

	if err := waitForDirectConnectGatewayAssociationDeletion(conn, dxgwId, vgwId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Direct Connect gateway association (%s) to be deleted: %s", d.Id(), err.Error())
	}

	return nil
}

func resourceAwsDxGatewayAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'dx-gw-id/vgw-id'", d.Id())
	}

	dxgwId := parts[0]
	vgwId := parts[1]
	log.Printf("[DEBUG] Importing Direct Connect gateway association %s/%s", dxgwId, vgwId)

	d.SetId(dxGatewayAssociationId(dxgwId, vgwId))
	d.Set("dx_gateway_id", dxgwId)
	d.Set("vpn_gateway_id", vgwId)

	return []*schema.ResourceData{d}, nil
}

func dxGatewayAssociationStateRefresh(conn *directconnect.DirectConnect, dxgwId, vgwId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeDirectConnectGatewayAssociations(&directconnect.DescribeDirectConnectGatewayAssociationsInput{
			DirectConnectGatewayId: aws.String(dxgwId),
			VirtualGatewayId:       aws.String(vgwId),
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
			return assoc, aws.StringValue(assoc.AssociationState), nil

		default:
			return nil, "", fmt.Errorf("Found %d Direct Connect gateway associations for %s, expected 1", n, dxGatewayAssociationId(dxgwId, vgwId))
		}
	}
}

func dxGatewayAssociationId(dxgwId, vgwId string) string {
	return fmt.Sprintf("ga-%s%s", dxgwId, vgwId)
}

func waitForDirectConnectGatewayAssociationDeletion(conn *directconnect.DirectConnect, directConnectGatewayID, virtualGatewayID string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.GatewayAssociationStateDisassociating},
		Target:     []string{directconnect.GatewayAssociationStateDisassociated, gatewayAssociationStateDeleted},
		Refresh:    dxGatewayAssociationStateRefresh(conn, directConnectGatewayID, virtualGatewayID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
