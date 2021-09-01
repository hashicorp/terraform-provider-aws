package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TransitGatewayConnect() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayConnectCreate,
		Read:   resourceAwsEc2TransitGatewayConnectRead,
		Update: resourceAwsEc2TransitGatewayConnectUpdate,
		Delete: resourceAwsEc2TransitGatewayConnectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"transport_transit_gateway_attachment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.ProtocolValueGre,
				ValidateFunc: validation.StringInSlice(ec2.ProtocolValue_Values(), false),
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
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
		},
	}
}

func resourceAwsEc2TransitGatewayConnectCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	transportTransitGatewayAttachmentId := d.Get("transport_transit_gateway_attachment_id").(string)

	input := &ec2.CreateTransitGatewayConnectInput{
		Options: &ec2.CreateTransitGatewayConnectRequestOptions{
			Protocol: aws.String(d.Get("protocol").(string)),
		},
		TransportTransitGatewayAttachmentId: aws.String(transportTransitGatewayAttachmentId),
		TagSpecifications:                   ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayAttachment),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Connect: %s", input)
	output, err := conn.CreateTransitGatewayConnect(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Connect: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayConnect.TransitGatewayAttachmentId))

	if err := waitForEc2TransitGatewayConnectCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Connect (%s) availability: %s", d.Id(), err)
	}

	transportTransitGatewayAttachment, err := ec2DescribeTransitGatewayVpcAttachment(conn, transportTransitGatewayAttachmentId)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit VPC Attachment (%s): %s", transportTransitGatewayAttachmentId, err)
	}

	transitGatewayID := *output.TransitGatewayConnect.TransitGatewayId
	transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	// We cannot modify Transit Gateway Route Tables for Resource Access Manager shared Transit Gateways
	if aws.StringValue(transitGateway.OwnerId) == aws.StringValue(transportTransitGatewayAttachment.VpcOwnerId) {
		if err := ec2TransitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %s", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
		}

		if err := ec2TransitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) propagation: %s", d.Id(), aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), err)
		}
	}

	return resourceAwsEc2TransitGatewayConnectRead(d, meta)
}

func resourceAwsEc2TransitGatewayConnectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	transitGatewayConnect, err := ec2DescribeTransitGatewayConnect(conn, d.Id())

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway Connect (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect: %s", err)
	}

	if transitGatewayConnect == nil {
		log.Printf("[WARN] EC2 Transit Gateway Connect (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGatewayConnect.State) == ec2.TransitGatewayAttachmentStateDeleting || aws.StringValue(transitGatewayConnect.State) == ec2.TransitGatewayAttachmentStateDeleted {
		log.Printf("[WARN] EC2 Transit Gateway Connect (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(transitGatewayConnect.State))
		d.SetId("")
		return nil
	}

	transportTransitGatewayAttachmentId := *transitGatewayConnect.TransportTransitGatewayAttachmentId
	transportTransitGatewayAttachment, err := ec2DescribeTransitGatewayVpcAttachment(conn, transportTransitGatewayAttachmentId)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit VPC Attachment (%s): %s", transportTransitGatewayAttachmentId, err)
	}

	transitGatewayID := aws.StringValue(transitGatewayConnect.TransitGatewayId)
	transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	// We cannot read Transit Gateway Route Tables for Resource Access Manager shared Transit Gateways
	// Default these to a non-nil value so we can match the existing schema of Default: true
	transitGatewayDefaultRouteTableAssociation := &ec2.TransitGatewayRouteTableAssociation{}
	transitGatewayDefaultRouteTablePropagation := &ec2.TransitGatewayRouteTablePropagation{}
	if aws.StringValue(transitGateway.OwnerId) == aws.StringValue(transportTransitGatewayAttachment.VpcOwnerId) {
		transitGatewayAssociationDefaultRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		transitGatewayDefaultRouteTableAssociation, err = ec2DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayAssociationDefaultRouteTableID, d.Id())
		if err != nil {
			return fmt.Errorf("error determining EC2 Transit Gateway Attachment (%s) association to Route Table (%s): %s", d.Id(), transitGatewayAssociationDefaultRouteTableID, err)
		}

		transitGatewayPropagationDefaultRouteTableID := aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId)
		transitGatewayDefaultRouteTablePropagation, err = ec2DescribeTransitGatewayRouteTablePropagation(conn, transitGatewayPropagationDefaultRouteTableID, d.Id())
		if err != nil {
			return fmt.Errorf("error determining EC2 Transit Gateway Attachment (%s) propagation to Route Table (%s): %s", d.Id(), transitGatewayPropagationDefaultRouteTableID, err)
		}
	}

	if transitGatewayConnect.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect (%s): missing options", d.Id())
	}

	tags := keyvaluetags.Ec2KeyValueTags(transitGatewayConnect.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("transit_gateway_default_route_table_association", (transitGatewayDefaultRouteTableAssociation != nil))
	d.Set("transit_gateway_default_route_table_propagation", (transitGatewayDefaultRouteTablePropagation != nil))
	d.Set("transport_transit_gateway_attachment_id", aws.StringValue(transitGatewayConnect.TransportTransitGatewayAttachmentId))
	d.Set("protocol", aws.StringValue(transitGatewayConnect.Options.Protocol))
	d.Set("transit_gateway_id", aws.StringValue(transitGatewayConnect.TransitGatewayId))

	return nil
}

func resourceAwsEc2TransitGatewayConnectUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChanges("transit_gateway_default_route_table_association", "transit_gateway_default_route_table_propagation") {
		transportTransitGatewayAttachmentId := d.Get("transport_transit_gateway_attachment_id").(string)
		transportTransitGatewayAttachment, err := ec2DescribeTransitGatewayVpcAttachment(conn, transportTransitGatewayAttachmentId)
		if err != nil {
			return fmt.Errorf("error describing EC2 Transit VPC Attachment (%s): %s", transportTransitGatewayAttachmentId, err)
		}

		transitGatewayID := *transportTransitGatewayAttachment.TransitGatewayId
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

		if d.HasChange("transit_gateway_default_route_table_propagation") {
			if err := ec2TransitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
				return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) propagation: %s", d.Id(), aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Connect (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEc2TransitGatewayConnectDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteTransitGatewayConnectInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Connect (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayConnect(input)

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Connect: %s", err)
	}

	if err := waitForEc2TransitGatewayConnectDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Connect (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
