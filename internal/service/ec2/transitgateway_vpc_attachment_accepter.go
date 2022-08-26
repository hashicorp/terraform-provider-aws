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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

	transitGatewayAttachmentID := d.Get("transit_gateway_attachment_id").(string)
	input := &ec2.AcceptTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
	}

	log.Printf("[DEBUG] Accepting EC2 Transit Gateway VPC Attachment: %s", input)
	output, err := conn.AcceptTransitGatewayVpcAttachment(input)

	if err != nil {
		return fmt.Errorf("accepting EC2 Transit Gateway VPC Attachment (%s): %w", transitGatewayAttachmentID, err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))
	transitGatewayID := aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayId)

	if _, err := WaitTransitGatewayVPCAttachmentAccepted(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway VPC Attachment (%s) update: %w", d.Id(), err)
	}

	if len(tags) > 0 {
		if err := CreateTags(conn, d.Id(), tags); err != nil {
			return fmt.Errorf("updating EC2 Transit Gateway VPC Attachment (%s) tags: %w", d.Id(), err)
		}
	}

	transitGateway, err := FindTransitGatewayByID(conn, transitGatewayID)

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway (%s): %w", transitGatewayID, err)
	}

	if err := transitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
		return err
	}

	if err := transitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
		return err
	}

	return resourceTransitGatewayVPCAttachmentAccepterRead(d, meta)
}

func resourceTransitGatewayVPCAttachmentAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayVPCAttachment, err := FindTransitGatewayVPCAttachmentByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway VPC Attachment (%s): %w", d.Id(), err)
	}

	transitGatewayID := aws.StringValue(transitGatewayVPCAttachment.TransitGatewayId)
	transitGateway, err := FindTransitGatewayByID(conn, transitGatewayID)

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway (%s): %w", transitGatewayID, err)
	}

	transitGatewayDefaultRouteTableAssociation := true
	transitGatewayDefaultRouteTablePropagation := true

	if transitGatewayRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId); transitGatewayRouteTableID != "" {
		_, err := FindTransitGatewayRouteTableAssociationByTwoPartKey(conn, transitGatewayRouteTableID, d.Id())

		if tfresource.NotFound(err) {
			transitGatewayDefaultRouteTableAssociation = false
		} else if err != nil {
			return fmt.Errorf("reading EC2 Transit Gateway Route Table Association (%s): %w", TransitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, d.Id()), err)
		}
	} else {
		transitGatewayDefaultRouteTableAssociation = false
	}

	if transitGatewayRouteTableID := aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId); transitGatewayRouteTableID != "" {
		_, err := FindTransitGatewayRouteTablePropagationByTwoPartKey(conn, transitGatewayRouteTableID, d.Id())

		if tfresource.NotFound(err) {
			transitGatewayDefaultRouteTablePropagation = false
		} else if err != nil {
			return fmt.Errorf("reading EC2 Transit Gateway Route Table Propagation (%s): %w", TransitGatewayRouteTablePropagationCreateResourceID(transitGatewayRouteTableID, d.Id()), err)
		}
	} else {
		transitGatewayDefaultRouteTablePropagation = false
	}

	d.Set("appliance_mode_support", transitGatewayVPCAttachment.Options.ApplianceModeSupport)
	d.Set("dns_support", transitGatewayVPCAttachment.Options.DnsSupport)
	d.Set("ipv6_support", transitGatewayVPCAttachment.Options.Ipv6Support)
	d.Set("subnet_ids", aws.StringValueSlice(transitGatewayVPCAttachment.SubnetIds))
	d.Set("transit_gateway_attachment_id", transitGatewayVPCAttachment.TransitGatewayAttachmentId)
	d.Set("transit_gateway_default_route_table_association", transitGatewayDefaultRouteTableAssociation)
	d.Set("transit_gateway_default_route_table_propagation", transitGatewayDefaultRouteTablePropagation)
	d.Set("transit_gateway_id", transitGatewayVPCAttachment.TransitGatewayId)
	d.Set("vpc_id", transitGatewayVPCAttachment.VpcId)
	d.Set("vpc_owner_id", transitGatewayVPCAttachment.VpcOwnerId)

	tags := KeyValueTags(transitGatewayVPCAttachment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceTransitGatewayVPCAttachmentAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChanges("transit_gateway_default_route_table_association", "transit_gateway_default_route_table_propagation") {
		transitGatewayID := d.Get("transit_gateway_id").(string)
		transitGateway, err := FindTransitGatewayByID(conn, transitGatewayID)

		if err != nil {
			return fmt.Errorf("reading EC2 Transit Gateway (%s): %w", transitGatewayID, err)
		}

		if d.HasChange("transit_gateway_default_route_table_association") {
			if err := transitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
				return err
			}
		}

		if d.HasChange("transit_gateway_default_route_table_propagation") {
			if err := transitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
				return err
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating EC2 Transit Gateway VPC Attachment (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceTransitGatewayVPCAttachmentAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway VPC Attachment: %s", d.Id())
	_, err := conn.DeleteTransitGatewayVpcAttachment(&ec2.DeleteTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway VPC Attachment (%s): %w", d.Id(), err)
	}

	if _, err := WaitTransitGatewayVPCAttachmentDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway VPC Attachment (%s) delete: %w", d.Id(), err)
	}

	return nil
}
