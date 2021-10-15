package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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

	vpcId := d.Get("vpc_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)

	input := &ec2.AttachVpnGatewayInput{
		VpcId:        aws.String(vpcId),
		VpnGatewayId: aws.String(vgwId),
	}

	log.Printf("[DEBUG] Creating VPN Gateway Attachment: %s", input)
	_, err := conn.AttachVpnGateway(input)

	if err != nil {
		return fmt.Errorf("error creating VPN Gateway (%s) Attachment (%s): %w", vgwId, vpcId, err)
	}

	d.SetId(VPNGatewayVPCAttachmentCreateID(vgwId, vpcId))

	_, err = WaitVPNGatewayVPCAttachmentAttached(conn, vgwId, vpcId)

	if err != nil {
		return fmt.Errorf("error waiting for VPN Gateway (%s) Attachment (%s) to become attached: %w", vgwId, vpcId, err)
	}

	return resourceVPNGatewayAttachmentRead(d, meta)
}

func resourceVPNGatewayAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcId := d.Get("vpc_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)

	vpcAttachment, err := FindVPNGatewayVPCAttachment(conn, vgwId, vpcId)

	if tfawserr.ErrMessageContains(err, InvalidVPNGatewayIDNotFound, "") {
		log.Printf("[WARN] VPN Gateway (%s) Attachment (%s) not found, removing from state", vgwId, vpcId)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPN Gateway (%s) Attachment (%s): %w", vgwId, vpcId, err)
	}

	if vpcAttachment == nil || aws.StringValue(vpcAttachment.State) == ec2.AttachmentStatusDetached {
		log.Printf("[WARN] VPN Gateway (%s) Attachment (%s) not found, removing from state", vgwId, vpcId)
		d.SetId("")
		return nil
	}

	return nil
}

func resourceVPNGatewayAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcId := d.Get("vpc_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)

	log.Printf("[INFO] Deleting VPN Gateway (%s) Attachment (%s)", vgwId, vpcId)
	_, err := conn.DetachVpnGateway(&ec2.DetachVpnGatewayInput{
		VpcId:        aws.String(vpcId),
		VpnGatewayId: aws.String(vgwId),
	})

	if tfawserr.ErrMessageContains(err, InvalidVPNGatewayAttachmentNotFound, "") || tfawserr.ErrMessageContains(err, InvalidVPNGatewayIDNotFound, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPN Gateway (%s) Attachment (%s): %w", vgwId, vpcId, err)
	}

	_, err = WaitVPNGatewayVPCAttachmentDetached(conn, vgwId, vpcId)

	if err != nil {
		return fmt.Errorf("error waiting for VPN Gateway (%s) Attachment (%s) to become detached: %w", vgwId, vpcId, err)
	}

	return nil
}
