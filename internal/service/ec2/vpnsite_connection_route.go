package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPNConnectionRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPNConnectionRouteCreate,
		Read:   resourceVPNConnectionRouteRead,
		Delete: resourceVPNConnectionRouteDelete,

		Schema: map[string]*schema.Schema{
			"destination_cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpn_connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPNConnectionRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	cidrBlock := d.Get("destination_cidr_block").(string)
	vpnConnectionID := d.Get("vpn_connection_id").(string)
	id := VPNConnectionRouteCreateResourceID(cidrBlock, vpnConnectionID)
	input := &ec2.CreateVpnConnectionRouteInput{
		DestinationCidrBlock: aws.String(cidrBlock),
		VpnConnectionId:      aws.String(vpnConnectionID),
	}

	log.Printf("[DEBUG] Creating EC2 VPN Connection Route: %s", input)
	_, err := conn.CreateVpnConnectionRoute(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 VPN Connection Route (%s): %w", id, err)
	}

	d.SetId(id)

	if _, err := WaitVPNConnectionRouteCreated(conn, vpnConnectionID, cidrBlock); err != nil {
		return fmt.Errorf("error waiting for EC2 VPN Connection Route (%s) create: %w", d.Id(), err)
	}

	return resourceVPNConnectionRouteRead(d, meta)
}

func resourceVPNConnectionRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	cidrBlock, vpnConnectionID, err := VPNConnectionRouteParseResourceID(d.Id())

	if err != nil {
		return err
	}

	_, err = FindVPNConnectionRouteByVPNConnectionIDAndCIDR(conn, vpnConnectionID, cidrBlock)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPN Connection Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPN Connection Route (%s): %w", d.Id(), err)
	}

	d.Set("destination_cidr_block", cidrBlock)
	d.Set("vpn_connection_id", vpnConnectionID)

	return nil
}

func resourceVPNConnectionRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	cidrBlock, vpnConnectionID, err := VPNConnectionRouteParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[INFO] Deleting EC2 VPN Connection Route: %s", d.Id())
	_, err = conn.DeleteVpnConnectionRoute(&ec2.DeleteVpnConnectionRouteInput{
		DestinationCidrBlock: aws.String(cidrBlock),
		VpnConnectionId:      aws.String(vpnConnectionID),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpnConnectionIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 VPN Connection Route (%s): %w", d.Id(), err)
	}

	if _, err := WaitVPNConnectionRouteDeleted(conn, vpnConnectionID, cidrBlock); err != nil {
		return fmt.Errorf("error waiting for EC2 VPN Connection Route (%s) delete: %w", d.Id(), err)
	}

	return nil
}

const vpnConnectionRouteResourceIDSeparator = ":"

func VPNConnectionRouteCreateResourceID(cidrBlock, vpcConnectionID string) string {
	parts := []string{cidrBlock, vpcConnectionID}
	id := strings.Join(parts, vpnConnectionRouteResourceIDSeparator)

	return id
}

func VPNConnectionRouteParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, vpnConnectionRouteResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DestinationCIDRBlock%[2]sVPNConnectionID", id, vpnConnectionRouteResourceIDSeparator)
}
