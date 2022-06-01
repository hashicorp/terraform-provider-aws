package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayConnect() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayConnectCreate,
		ReadWithoutTimeout:   resourceTransitGatewayConnectRead,
		UpdateWithoutTimeout: resourceTransitGatewayConnectUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayConnectDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.ProtocolValueGre,
				ValidateFunc: validation.StringInSlice(ec2.ProtocolValue_Values(), false),
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
			"transport_attachment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceTransitGatewayConnectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	transportAttachmentID := d.Get("transport_attachment_id").(string)

	input := &ec2.CreateTransitGatewayConnectInput{
		Options: &ec2.CreateTransitGatewayConnectRequestOptions{
			Protocol: aws.String(d.Get("protocol").(string)),
		},
		TagSpecifications:                   tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayAttachment),
		TransportTransitGatewayAttachmentId: aws.String(transportAttachmentID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Connect Attachment: %s", input)
	output, err := conn.CreateTransitGatewayConnectWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("error creating EC2 Transit Gateway Connect Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayConnect.TransitGatewayAttachmentId))

	if _, err := WaitTransitGatewayConnectCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for EC2 Transit Gateway Connect (%s) create: %s", d.Id(), err)
	}

	transportAttachment, err := FindTransitGatewayAttachmentByID(conn, transportAttachmentID)

	if err != nil {
		return diag.Errorf("error reading EC2 Transit Gateway Attachment (%s): %s", transportAttachmentID, err)
	}

	transitGatewayID := aws.StringValue(transportAttachment.TransitGatewayId)
	transitGateway, err := FindTransitGatewayByID(conn, transitGatewayID)

	if err != nil {
		return diag.Errorf("error reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	// We cannot modify Transit Gateway Route Tables for Resource Access Manager shared Transit Gateways
	if aws.StringValue(transitGateway.OwnerId) == aws.StringValue(transportAttachment.ResourceOwnerId) {
		if err := transitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
			return diag.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %s", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
		}

		if err := transitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
			return diag.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) propagation: %s", d.Id(), aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), err)
		}
	}

	return resourceTransitGatewayConnectRead(ctx, d, meta)
}

func resourceTransitGatewayConnectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayConnect, err := FindTransitGatewayConnectByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Connect %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading EC2 Transit Gateway Connect (%s): %s", d.Id(), err)
	}

	transitGatewayID := aws.StringValue(transitGatewayConnect.TransitGatewayId)
	transitGateway, err := FindTransitGatewayByID(conn, transitGatewayID)

	if err != nil {
		return diag.Errorf("error reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	transitGatewayAttachment, err := FindTransitGatewayAttachmentByID(conn, d.Id())

	if err != nil {
		return diag.Errorf("error reading EC2 Transit Gateway Attachment (%s): %s", d.Id(), err)
	}

	// We cannot read Transit Gateway Route Tables for Resource Access Manager shared Transit Gateways
	// Default these to a non-nil value so we can match the existing schema of Default: true
	transitGatewayDefaultRouteTableAssociation := &ec2.TransitGatewayRouteTableAssociation{}
	transitGatewayDefaultRouteTablePropagation := &ec2.TransitGatewayRouteTablePropagation{}
	if aws.StringValue(transitGateway.OwnerId) == aws.StringValue(transitGatewayAttachment.ResourceOwnerId) {
		transitGatewayAssociationDefaultRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		transitGatewayDefaultRouteTableAssociation, err = DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayAssociationDefaultRouteTableID, d.Id())
		if err != nil {
			return diag.Errorf("error determining EC2 Transit Gateway Attachment (%s) association to Route Table (%s): %s", d.Id(), transitGatewayAssociationDefaultRouteTableID, err)
		}

		transitGatewayPropagationDefaultRouteTableID := aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId)
		transitGatewayDefaultRouteTablePropagation, err = FindTransitGatewayRouteTablePropagation(conn, transitGatewayPropagationDefaultRouteTableID, d.Id())
		if err != nil {
			return diag.Errorf("error determining EC2 Transit Gateway Attachment (%s) propagation to Route Table (%s): %s", d.Id(), transitGatewayPropagationDefaultRouteTableID, err)
		}
	}

	d.Set("protocol", transitGatewayConnect.Options.Protocol)
	d.Set("transit_gateway_default_route_table_association", (transitGatewayDefaultRouteTableAssociation != nil))
	d.Set("transit_gateway_default_route_table_propagation", (transitGatewayDefaultRouteTablePropagation != nil))
	d.Set("transit_gateway_id", transitGatewayConnect.TransitGatewayId)
	d.Set("transport_attachment_id", transitGatewayConnect.TransportTransitGatewayAttachmentId)

	tags := KeyValueTags(transitGatewayConnect.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceTransitGatewayConnectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChanges("transit_gateway_default_route_table_association", "transit_gateway_default_route_table_propagation") {
		transitGatewayID := d.Get("transit_gateway_id").(string)
		transitGateway, err := FindTransitGatewayByID(conn, transitGatewayID)

		if err != nil {
			return diag.Errorf("error reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
		}

		if d.HasChange("transit_gateway_default_route_table_association") {
			if err := transitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
				return diag.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %s", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
			}
		}

		if d.HasChange("transit_gateway_default_route_table_propagation") {
			if err := transitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
				return diag.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) propagation: %s", d.Id(), aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return diag.Errorf("error updating EC2 Transit Gateway Connect Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceTransitGatewayConnectRead(ctx, d, meta)
}

func resourceTransitGatewayConnectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Connect: %s", d.Id())
	_, err := conn.DeleteTransitGatewayConnectWithContext(ctx, &ec2.DeleteTransitGatewayConnectInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting EC2 Transit Gateway Connect (%s): %s", d.Id(), err)
	}

	if _, err := WaitTransitGatewayConnectDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("error waiting for EC2 Transit Gateway Connect (%s) delete: %s", d.Id(), err)
	}

	return nil
}
