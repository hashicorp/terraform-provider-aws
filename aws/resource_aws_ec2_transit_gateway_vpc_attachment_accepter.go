package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsEc2TransitGatewayVpcAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayVpcAttachmentAccepterCreate,
		Read:   resourceAwsEc2TransitGatewayVpcAttachmentAccepterRead,
		Update: resourceAwsEc2TransitGatewayVpcAttachmentAccepterUpdate,
		Delete: resourceAwsEc2TransitGatewayVpcAttachmentAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"dns_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
			"transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2TransitGatewayVpcAttachmentAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.AcceptTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Get("transit_gateway_attachment_id").(string)),
	}

	log.Printf("[DEBUG] Accepting EC2 Transit Gateway VPC Attachment: %s", input)
	output, err := conn.AcceptTransitGatewayVpcAttachment(input)
	if err != nil {
		return fmt.Errorf("error accepting EC2 Transit Gateway VPC Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))

	if err := waitForEc2TransitGatewayVpcAttachmentAcceptance(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway VPC Attachment (%s) availability: %s", d.Id(), err)
	}

	if err := setTags(conn, d); err != nil {
		return fmt.Errorf("error updating EC2 Transit Gateway VPC Attachment (%s) tags: %s", d.Id(), err)
	}

	return resourceAwsEc2TransitGatewayVpcAttachmentAccepterRead(d, meta)
}

func resourceAwsEc2TransitGatewayVpcAttachmentAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	transitGatewayVpcAttachment, err := ec2DescribeTransitGatewayVpcAttachment(conn, d.Id())

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachment: %s", err)
	}

	if transitGatewayVpcAttachment == nil {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGatewayVpcAttachment.State) == ec2.TransitGatewayAttachmentStateDeleting || aws.StringValue(transitGatewayVpcAttachment.State) == ec2.TransitGatewayAttachmentStateDeleted {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(transitGatewayVpcAttachment.State))
		d.SetId("")
		return nil
	}

	if transitGatewayVpcAttachment.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachment (%s): missing options", d.Id())
	}

	d.Set("dns_support", transitGatewayVpcAttachment.Options.DnsSupport)
	d.Set("ipv6_support", transitGatewayVpcAttachment.Options.Ipv6Support)

	if err := d.Set("subnet_ids", aws.StringValueSlice(transitGatewayVpcAttachment.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}

	if err := d.Set("tags", tagsToMap(transitGatewayVpcAttachment.Tags)); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("transit_gateway_attachment_id", aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId))
	d.Set("transit_gateway_id", aws.StringValue(transitGatewayVpcAttachment.TransitGatewayId))
	d.Set("vpc_id", aws.StringValue(transitGatewayVpcAttachment.VpcId))
	d.Set("vpc_owner_id", aws.StringValue(transitGatewayVpcAttachment.VpcOwnerId))

	return nil
}

func resourceAwsEc2TransitGatewayVpcAttachmentAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		if err := setTags(conn, d); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway VPC Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEc2TransitGatewayVpcAttachmentAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete EC2 Transit Gateway VPC Attachment. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
