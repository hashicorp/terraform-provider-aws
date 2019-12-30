package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsEc2TransitGatewayPeeringAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayPeeringAttachmentAccepterCreate,
		Read:   resourceAwsEc2TransitGatewayPeeringAttachmentAccepterRead,
		Update: resourceAwsEc2TransitGatewayPeeringAttachmentAccepterUpdate,
		Delete: resourceAwsEc2TransitGatewayPeeringAttachmentAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"peer_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_default_route_table_association": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2TransitGatewayPeeringAttachmentAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.AcceptTransitGatewayPeeringAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Get("transit_gateway_attachment_id").(string)),
	}

	log.Printf("[DEBUG] Accepting EC2 Transit Gateway Peering Attachment: %s", input)
	output, err := conn.AcceptTransitGatewayPeeringAttachment(input)
	if err != nil {
		return fmt.Errorf("error accepting EC2 Transit Gateway Peering Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayPeeringAttachment.TransitGatewayAttachmentId))
	transitGatewayID := aws.StringValue(output.TransitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)

	if err := waitForEc2TransitGatewayPeeringAttachmentAcceptance(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Peering Attachment (%s) availability: %s", d.Id(), err)
	}

	if err := setTags(conn, d); err != nil {
		return fmt.Errorf("error updating EC2 Transit Gateway Peering Attachment (%s) tags: %s", d.Id(), err)
	}

	transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	if err := ec2TransitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
		return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %s", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
	}

	return resourceAwsEc2TransitGatewayPeeringAttachmentAccepterRead(d, meta)
}

func resourceAwsEc2TransitGatewayPeeringAttachmentAccepterRead(d *schema.ResourceData, meta interface{}) error {
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

	transitGatewayID := aws.StringValue(transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)
	transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	transitGatewayAssociationDefaultRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
	transitGatewayDefaultRouteTableAssociation, err := ec2DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayAssociationDefaultRouteTableID, d.Id())
	if err != nil {
		return fmt.Errorf("error determining EC2 Transit Gateway Attachment (%s) association to Route Table (%s): %s", d.Id(), transitGatewayAssociationDefaultRouteTableID, err)
	}

	if err := d.Set("tags", tagsToMap(transitGatewayPeeringAttachment.Tags)); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("transit_gateway_default_route_table_association", (transitGatewayDefaultRouteTableAssociation != nil))
	d.Set("peer_account_id", (transitGatewayPeeringAttachment.AccepterTgwInfo.OwnerId != nil))
	d.Set("peer_region", (transitGatewayPeeringAttachment.AccepterTgwInfo.Region != nil))
	d.Set("peer_transit_gateway_id", (transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId != nil))
	d.Set("tags", (transitGatewayPeeringAttachment.Tags))
	d.Set("transit_gateway_id", (transitGatewayPeeringAttachment.RequesterTgwInfo.TransitGatewayId))

	return nil
}

func resourceAwsEc2TransitGatewayPeeringAttachmentAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("transit_gateway_default_route_table_association") {
		transitGatewayID := d.Get("transit_gateway_id").(string)

		transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
		if err != nil {
			return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
		}

		if transitGateway.Options == nil {
			return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
		}

		if d.HasChange("transit_gateway_default_route_table_association") {
			if err := ec2TransitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
				return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %s", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
			}
		}
	}

	if d.HasChange("tags") {
		if err := setTags(conn, d); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Peering Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEc2TransitGatewayPeeringAttachmentAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteTransitGatewayPeeringAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Peering Attachment (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayPeeringAttachment(input)

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Peering Attachment: %s", err)
	}

	if err := waitForEc2TransitGatewayPeeringAttachmentDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Peering Attachment (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
