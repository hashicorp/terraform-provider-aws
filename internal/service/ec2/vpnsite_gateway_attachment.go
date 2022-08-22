package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPNGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPNGatewayAttachmentCreate,
		Read:   resourceVPNGatewayAttachmentRead,
		Delete: resourceVPNGatewayAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
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

func resourceVPNGatewayAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcID := d.Get("vpc_id").(string)
	vpnGatewayID := d.Get("vpn_gateway_id").(string)
	input := &ec2.AttachVpnGatewayInput{
		VpcId:        aws.String(vpcID),
		VpnGatewayId: aws.String(vpnGatewayID),
	}

	log.Printf("[DEBUG] Creating EC2 VPN Gateway Attachment: %s", input)
	_, err := conn.AttachVpnGateway(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 VPN Gateway (%s) Attachment (%s): %w", vpnGatewayID, vpcID, err)
	}

	d.SetId(VPNGatewayVPCAttachmentCreateID(vpnGatewayID, vpcID))

	_, err = WaitVPNGatewayVPCAttachmentAttached(conn, vpnGatewayID, vpcID)

	if err != nil {
		return fmt.Errorf("error waiting for EC2 VPN Gateway (%s) Attachment (%s) to become attached: %w", vpnGatewayID, vpcID, err)
	}

	return resourceVPNGatewayAttachmentRead(d, meta)
}

func resourceVPNGatewayAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcID := d.Get("vpc_id").(string)
	vpnGatewayID := d.Get("vpn_gateway_id").(string)

	_, err := FindVPNGatewayVPCAttachment(conn, vpnGatewayID, vpcID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPN Gateway (%s) Attachment (%s) not found, removing from state", vpnGatewayID, vpcID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPN Gateway (%s) Attachment (%s): %w", vpnGatewayID, vpcID, err)
	}

	return nil
}

func resourceVPNGatewayAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcID := d.Get("vpc_id").(string)
	vpnGatewayID := d.Get("vpn_gateway_id").(string)

	log.Printf("[INFO] Deleting EC2 VPN Gateway (%s) Attachment (%s)", vpnGatewayID, vpcID)
	_, err := conn.DetachVpnGateway(&ec2.DetachVpnGatewayInput{
		VpcId:        aws.String(vpcID),
		VpnGatewayId: aws.String(vpnGatewayID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNGatewayAttachmentNotFound, errCodeInvalidVPNGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 VPN Gateway (%s) Attachment (%s): %w", vpnGatewayID, vpcID, err)
	}

	_, err = WaitVPNGatewayVPCAttachmentDetached(conn, vpnGatewayID, vpcID)

	if err != nil {
		return fmt.Errorf("error waiting for EC2 VPN Gateway (%s) Attachment (%s) to become detached: %w", vpnGatewayID, vpcID, err)
	}

	return nil
}
