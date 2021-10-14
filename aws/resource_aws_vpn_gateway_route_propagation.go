package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/hashicorp/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsVpnGatewayRoutePropagation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpnGatewayRoutePropagationEnable,
		Read:   resourceAwsVpnGatewayRoutePropagationRead,
		Delete: resourceAwsVpnGatewayRoutePropagationDisable,

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

func resourceAwsVpnGatewayRoutePropagationEnable(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	gatewayID := d.Get("vpn_gateway_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	err := ec2RouteTableEnableVgwRoutePropagation(conn, routeTableID, gatewayID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return err
	}

	d.SetId(tfec2.VpnGatewayRoutePropagationCreateID(routeTableID, gatewayID))

	return resourceAwsVpnGatewayRoutePropagationRead(d, meta)
}

func resourceAwsVpnGatewayRoutePropagationDisable(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	routeTableID, gatewayID, err := tfec2.VpnGatewayRoutePropagationParseID(d.Id())

	if err != nil {
		return err
	}

	err = ec2RouteTableDisableVgwRoutePropagation(conn, routeTableID, gatewayID)

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsVpnGatewayRoutePropagationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	routeTableID, gatewayID, err := tfec2.VpnGatewayRoutePropagationParseID(d.Id())

	if err != nil {
		return err
	}

	err = finder.VpnGatewayRoutePropagationExists(conn, routeTableID, gatewayID)

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
