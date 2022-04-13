package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPNGatewayRoutePropagation() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPNGatewayRoutePropagationEnable,
		Read:   resourceVPNGatewayRoutePropagationRead,
		Delete: resourceVPNGatewayRoutePropagationDisable,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpn_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPNGatewayRoutePropagationEnable(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	gatewayID := d.Get("vpn_gateway_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	err := ec2RouteTableEnableVgwRoutePropagation(conn, routeTableID, gatewayID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return err
	}

	d.SetId(VPNGatewayRoutePropagationCreateID(routeTableID, gatewayID))

	return resourceVPNGatewayRoutePropagationRead(d, meta)
}

func resourceVPNGatewayRoutePropagationDisable(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	routeTableID, gatewayID, err := VPNGatewayRoutePropagationParseID(d.Id())

	if err != nil {
		return err
	}

	err = ec2RouteTableDisableVgwRoutePropagation(conn, routeTableID, gatewayID)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func resourceVPNGatewayRoutePropagationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	routeTableID, gatewayID, err := VPNGatewayRoutePropagationParseID(d.Id())

	if err != nil {
		return err
	}

	err = FindVPNGatewayRoutePropagationExists(conn, routeTableID, gatewayID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route Table (%s) VPN Gateway (%s) route propagation not found, removing from state", routeTableID, gatewayID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route Table (%s) VPN Gateway (%s) route propagation: %w", routeTableID, gatewayID, err)
	}

	return nil
}
