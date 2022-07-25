package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayVPCAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayVPCAttachmentAccepterCreate,
		Read:   resourceTransitGatewayVPCAttachmentAccepterRead,
		Update: resourceTransitGatewayVPCAttachmentAccepterUpdate,
		Delete: resourceTransitGatewayVPCAttachmentAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"appliance_mode_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
			"transit_gateway_default_route_table_propagation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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

func resourceTransitGatewayVPCAttachmentAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.AcceptTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Get("transit_gateway_attachment_id").(string)),
	}

	log.Printf("[DEBUG] Accepting EC2 Transit Gateway VPC Attachment: %s", input)
	output, err := conn.AcceptTransitGatewayVpcAttachment(input)
	if err != nil {
		return fmt.Errorf("error accepting EC2 Transit Gateway VPC Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))
	transitGatewayID := aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayId)

	if err := waitForTransitGatewayVPCAttachmentAcceptance(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway VPC Attachment (%s) availability: %s", d.Id(), err)
	}

	if len(tags) > 0 {
		if err := CreateTags(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway VPC Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	transitGateway, err := DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	if err := transitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
		return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %s", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
	}

	if err := transitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
		return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) propagation: %s", d.Id(), aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), err)
	}

	return resourceTransitGatewayVPCAttachmentAccepterRead(d, meta)
}

func resourceTransitGatewayVPCAttachmentAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayVpcAttachment, err := DescribeTransitGatewayVPCAttachment(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, "InvalidTransitGatewayAttachmentID.NotFound") {
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

	transitGatewayID := aws.StringValue(transitGatewayVpcAttachment.TransitGatewayId)
	transitGateway, err := DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	transitGatewayAssociationDefaultRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
	transitGatewayDefaultRouteTableAssociation, err := DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayAssociationDefaultRouteTableID, d.Id())
	if err != nil {
		return fmt.Errorf("error determining EC2 Transit Gateway Attachment (%s) association to Route Table (%s): %s", d.Id(), transitGatewayAssociationDefaultRouteTableID, err)
	}

	transitGatewayPropagationDefaultRouteTableID := aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId)
	transitGatewayDefaultRouteTablePropagation, err := FindTransitGatewayRouteTablePropagation(conn, transitGatewayPropagationDefaultRouteTableID, d.Id())
	if err != nil {
		return fmt.Errorf("error determining EC2 Transit Gateway Attachment (%s) propagation to Route Table (%s): %s", d.Id(), transitGatewayPropagationDefaultRouteTableID, err)
	}

	if transitGatewayVpcAttachment.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachment (%s): missing options", d.Id())
	}

	d.Set("appliance_mode_support", transitGatewayVpcAttachment.Options.ApplianceModeSupport)
	d.Set("dns_support", transitGatewayVpcAttachment.Options.DnsSupport)
	d.Set("ipv6_support", transitGatewayVpcAttachment.Options.Ipv6Support)

	if err := d.Set("subnet_ids", aws.StringValueSlice(transitGatewayVpcAttachment.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}

	tags := KeyValueTags(transitGatewayVpcAttachment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("transit_gateway_attachment_id", transitGatewayVpcAttachment.TransitGatewayAttachmentId)
	d.Set("transit_gateway_default_route_table_association", (transitGatewayDefaultRouteTableAssociation != nil))
	d.Set("transit_gateway_default_route_table_propagation", (transitGatewayDefaultRouteTablePropagation != nil))
	d.Set("transit_gateway_id", transitGatewayVpcAttachment.TransitGatewayId)
	d.Set("vpc_id", transitGatewayVpcAttachment.VpcId)
	d.Set("vpc_owner_id", transitGatewayVpcAttachment.VpcOwnerId)

	return nil
}

func resourceTransitGatewayVPCAttachmentAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChanges("transit_gateway_default_route_table_association", "transit_gateway_default_route_table_propagation") {
		transitGatewayID := d.Get("transit_gateway_id").(string)

		transitGateway, err := DescribeTransitGateway(conn, transitGatewayID)
		if err != nil {
			return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
		}

		if transitGateway.Options == nil {
			return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
		}

		if d.HasChange("transit_gateway_default_route_table_association") {
			if err := transitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
				return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %s", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
			}
		}

		if d.HasChange("transit_gateway_default_route_table_propagation") {
			if err := transitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
				return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) propagation: %s", d.Id(), aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway VPC Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceTransitGatewayVPCAttachmentAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DeleteTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayVpcAttachment(input)

	if tfawserr.ErrCodeEquals(err, "InvalidTransitGatewayAttachmentID.NotFound") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway VPC Attachment: %s", err)
	}

	if err := WaitForTransitGatewayAttachmentDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway VPC Attachment (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
