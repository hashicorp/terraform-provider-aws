package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsEc2TransitGatewayPeeringAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayPeeringAttachmentCreate,
		Read:   resourceAwsEc2TransitGatewayPeeringAttachmentRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"peer_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"peer_transit_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tagsSchema(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsEc2TransitGatewayPeeringAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	transitGatewayID := d.Get("transit_gateway_id").(string)

	input := &ec2.CreateTransitGatewayPeeringAttachmentInput{
		PeerAccountId:        aws.String(d.Get("peer_account_id").(string)),
		PeerRegion:           aws.String(d.Get("peer_region").(string)),
		PeerTransitGatewayId: aws.String(d.Get("peer_transit_gateway_id").(string)),
		TagSpecifications:    expandEc2TransitGatewayAttachmentTagSpecifications(d.Get("tags").(map[string]interface{})),
		TransitGatewayId:     aws.String(transitGatewayID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Peering Attachment: %s", input)
	output, err := conn.CreateTransitGatewayPeeringAttachment(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Peering Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayPeeringAttachment.TransitGatewayAttachmentId))

	if err := waitForEc2TransitGatewayPeeringAttachmentCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Peering Attachment (%s) availability: %s", d.Id(), err)
	}

	transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	return resourceAwsEc2TransitGatewayPeeringAttachmentRead(d, meta)
}

func resourceAwsEc2TransitGatewayPeeringAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	transitGatewayPeeringAttachment, err := ec2DescribeTransitGatewayPeeringAttachment(conn, d.Id())

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Peering Attachment: %s", err)
	}

	if transitGatewayPeeringAttachment == nil {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGatewayPeeringAttachment.State) == ec2.TransitGatewayAttachmentStateDeleting || aws.StringValue(transitGatewayPeeringAttachment.State) == ec2.TransitGatewayAttachmentStateDeleted {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(transitGatewayPeeringAttachment.State))
		d.SetId("")
		return nil
	}

	transitGatewayID := aws.StringValue(transitGatewayPeeringAttachment.RequesterTgwInfo.TransitGatewayId)
	transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	if err := d.Set("tags", tagsToMap(transitGatewayPeeringAttachment.Tags)); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("peer_account_id", (transitGatewayPeeringAttachment.AccepterTgwInfo.OwnerId != nil))
	d.Set("peer_region", (transitGatewayPeeringAttachment.AccepterTgwInfo.Region != nil))
	d.Set("peer_transit_gateway_id", (transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId != nil))
	d.Set("tags", (transitGatewayPeeringAttachment.Tags))
	d.Set("transit_gateway_id", (transitGatewayPeeringAttachment.RequesterTgwInfo.TransitGatewayId))

	return nil
}
