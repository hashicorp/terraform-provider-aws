package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayVPCAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayVPCAttachmentCreate,
		Read:   resourceTransitGatewayVPCAttachmentRead,
		Update: resourceTransitGatewayVPCAttachmentUpdate,
		Delete: resourceTransitGatewayVPCAttachmentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"appliance_mode_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.ApplianceModeSupportValueDisable,
				ValidateFunc: validation.StringInSlice(ec2.ApplianceModeSupportValue_Values(), false),
			},
			"dns_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.DnsSupportValueEnable,
				ValidateFunc: validation.StringInSlice(ec2.DnsSupportValue_Values(), false),
			},
			"ipv6_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.Ipv6SupportValueDisable,
				ValidateFunc: validation.StringInSlice(ec2.Ipv6SupportValue_Values(), false),
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"vpc_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"vpc_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTransitGatewayVPCAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	transitGatewayID := d.Get("transit_gateway_id").(string)
	input := &ec2.CreateTransitGatewayVpcAttachmentInput{
		Options: &ec2.CreateTransitGatewayVpcAttachmentRequestOptions{
			ApplianceModeSupport: aws.String(d.Get("appliance_mode_support").(string)),
			DnsSupport:           aws.String(d.Get("dns_support").(string)),
			Ipv6Support:          aws.String(d.Get("ipv6_support").(string)),
		},
		SubnetIds:         flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		TransitGatewayId:  aws.String(transitGatewayID),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayAttachment),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway VPC Attachment: %s", input)
	output, err := conn.CreateTransitGatewayVpcAttachment(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Transit Gateway VPC Attachment: %w", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))

	if err := waitForTransitGatewayAttachmentCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway VPC Attachment (%s) create: %w", d.Id(), err)
	}

	transitGateway, err := FindTransitGatewayByID(conn, transitGatewayID)

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway (%s): %w", transitGatewayID, err)
	}

	// We cannot modify Transit Gateway Route Tables for Resource Access Manager shared Transit Gateways.
	if aws.StringValue(transitGateway.OwnerId) == aws.StringValue(output.TransitGatewayVpcAttachment.VpcOwnerId) {
		if err := transitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
			return fmt.Errorf("updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %w", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
		}

		if err := transitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
			return fmt.Errorf("updating EC2 Transit Gateway Attachment (%s) Route Table (%s) propagation: %w", d.Id(), aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), err)
		}
	}

	return resourceTransitGatewayVPCAttachmentRead(d, meta)
}

func resourceTransitGatewayVPCAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayVpcAttachment, err := FindTransitGatewayVPCAttachmentByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway VPC Attachment (%s): %w", d.Id(), err)
	}

	transitGatewayID := aws.StringValue(transitGatewayVpcAttachment.TransitGatewayId)
	transitGateway, err := FindTransitGatewayByID(conn, transitGatewayID)

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway (%s): %w", transitGatewayID, err)
	}

	// We cannot read Transit Gateway Route Tables for Resource Access Manager shared Transit Gateways.
	// Default these to a non-nil value so we can match the existing schema of Default: true.
	transitGatewayDefaultRouteTableAssociation := &ec2.TransitGatewayRouteTableAssociation{}
	transitGatewayDefaultRouteTablePropagation := &ec2.TransitGatewayRouteTablePropagation{}
	if aws.StringValue(transitGateway.OwnerId) == aws.StringValue(transitGatewayVpcAttachment.VpcOwnerId) {
		transitGatewayAssociationDefaultRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		transitGatewayDefaultRouteTableAssociation, err = DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayAssociationDefaultRouteTableID, d.Id())
		if err != nil {
			return fmt.Errorf("determining EC2 Transit Gateway Attachment (%s) association to Route Table (%s): %w", d.Id(), transitGatewayAssociationDefaultRouteTableID, err)
		}

		transitGatewayPropagationDefaultRouteTableID := aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId)
		transitGatewayDefaultRouteTablePropagation, err = FindTransitGatewayRouteTablePropagation(conn, transitGatewayPropagationDefaultRouteTableID, d.Id())
		if err != nil {
			return fmt.Errorf("determining EC2 Transit Gateway Attachment (%s) propagation to Route Table (%s): %w", d.Id(), transitGatewayPropagationDefaultRouteTableID, err)
		}
	}

	if transitGatewayVpcAttachment.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachment (%s): missing options", d.Id())
	}

	d.Set("appliance_mode_support", transitGatewayVpcAttachment.Options.ApplianceModeSupport)
	d.Set("dns_support", transitGatewayVpcAttachment.Options.DnsSupport)
	d.Set("ipv6_support", transitGatewayVpcAttachment.Options.Ipv6Support)
	d.Set("subnet_ids", aws.StringValueSlice(transitGatewayVpcAttachment.SubnetIds))
	d.Set("transit_gateway_default_route_table_association", (transitGatewayDefaultRouteTableAssociation != nil))
	d.Set("transit_gateway_default_route_table_propagation", (transitGatewayDefaultRouteTablePropagation != nil))
	d.Set("transit_gateway_id", transitGatewayVpcAttachment.TransitGatewayId)
	d.Set("vpc_id", transitGatewayVpcAttachment.VpcId)
	d.Set("vpc_owner_id", transitGatewayVpcAttachment.VpcOwnerId)

	tags := KeyValueTags(transitGatewayVpcAttachment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceTransitGatewayVPCAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChanges("appliance_mode_support", "dns_support", "ipv6_support", "subnet_ids") {
		input := &ec2.ModifyTransitGatewayVpcAttachmentInput{
			Options: &ec2.ModifyTransitGatewayVpcAttachmentRequestOptions{
				ApplianceModeSupport: aws.String(d.Get("appliance_mode_support").(string)),
				DnsSupport:           aws.String(d.Get("dns_support").(string)),
				Ipv6Support:          aws.String(d.Get("ipv6_support").(string)),
			},
			TransitGatewayAttachmentId: aws.String(d.Id()),
		}

		oldRaw, newRaw := d.GetChange("subnet_ids")
		oldSet := oldRaw.(*schema.Set)
		newSet := newRaw.(*schema.Set)

		if added := newSet.Difference(oldSet); added.Len() > 0 {
			input.AddSubnetIds = flex.ExpandStringSet(added)
		}

		if removed := oldSet.Difference(newSet); removed.Len() > 0 {
			input.RemoveSubnetIds = flex.ExpandStringSet(removed)
		}

		if _, err := conn.ModifyTransitGatewayVpcAttachment(input); err != nil {
			return fmt.Errorf("updating EC2 Transit Gateway VPC Attachment (%s): %w", d.Id(), err)
		}

		if err := waitForTransitGatewayAttachmentUpdate(conn, d.Id()); err != nil {
			return fmt.Errorf("waiting for EC2 Transit Gateway VPC Attachment (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChanges("transit_gateway_default_route_table_association", "transit_gateway_default_route_table_propagation") {
		transitGatewayID := d.Get("transit_gateway_id").(string)
		transitGateway, err := FindTransitGatewayByID(conn, transitGatewayID)

		if err != nil {
			return fmt.Errorf("reading EC2 Transit Gateway (%s): %w", transitGatewayID, err)
		}

		if d.HasChange("transit_gateway_default_route_table_association") {
			if err := transitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
				return fmt.Errorf("updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %w", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
			}
		}

		if d.HasChange("transit_gateway_default_route_table_propagation") {
			if err := transitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
				return fmt.Errorf("updating EC2 Transit Gateway Attachment (%s) Route Table (%s) propagation: %w", d.Id(), aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), err)
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

func resourceTransitGatewayVPCAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway VPC Attachment: %s", d.Id())
	_, err := conn.DeleteTransitGatewayVpcAttachment(&ec2.DeleteTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway VPC Attachment: %w", err)
	}

	if err := WaitForTransitGatewayAttachmentDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway VPC Attachment (%s) delete: %w", d.Id(), err)
	}

	return nil
}
