package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayVPCAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayVPCAttachmentAccepterCreate,
		ReadWithoutTimeout:   resourceTransitGatewayVPCAttachmentAccepterRead,
		UpdateWithoutTimeout: resourceTransitGatewayVPCAttachmentAccepterUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayVPCAttachmentAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceTransitGatewayVPCAttachmentAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	transitGatewayAttachmentID := d.Get("transit_gateway_attachment_id").(string)
	input := &ec2.AcceptTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
	}

	log.Printf("[DEBUG] Accepting EC2 Transit Gateway VPC Attachment: %s", input)
	output, err := conn.AcceptTransitGatewayVpcAttachmentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway VPC Attachment (%s): %s", transitGatewayAttachmentID, err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))
	transitGatewayID := aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayId)

	if _, err := WaitTransitGatewayVPCAttachmentAccepted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway VPC Attachment (%s): waiting for completion: %s", transitGatewayAttachmentID, err)
	}

	if len(tags) > 0 {
		if err := CreateTags(ctx, conn, d.Id(), tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway VPC Attachment (%s): setting tags: %s", transitGatewayAttachmentID, err)
		}
	}

	transitGateway, err := FindTransitGatewayByID(ctx, conn, transitGatewayID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if err := transitGatewayRouteTableAssociationUpdate(ctx, conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway VPC Attachment (%s): %s", transitGatewayAttachmentID, err)
	}

	if err := transitGatewayRouteTablePropagationUpdate(ctx, conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway VPC Attachment (%s): %s", transitGatewayAttachmentID, err)
	}

	return append(diags, resourceTransitGatewayVPCAttachmentAccepterRead(ctx, d, meta)...)
}

func resourceTransitGatewayVPCAttachmentAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayVPCAttachment, err := FindTransitGatewayVPCAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
	}

	transitGatewayID := aws.StringValue(transitGatewayVPCAttachment.TransitGatewayId)
	transitGateway, err := FindTransitGatewayByID(ctx, conn, transitGatewayID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	transitGatewayDefaultRouteTableAssociation := true
	transitGatewayDefaultRouteTablePropagation := true

	if transitGatewayRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId); transitGatewayRouteTableID != "" {
		_, err := FindTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, d.Id())

		if tfresource.NotFound(err) {
			transitGatewayDefaultRouteTableAssociation = false
		} else if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route Table Association (%s): %s", TransitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, d.Id()), err)
		}
	} else {
		transitGatewayDefaultRouteTableAssociation = false
	}

	if transitGatewayRouteTableID := aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId); transitGatewayRouteTableID != "" {
		_, err := FindTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, d.Id())

		if tfresource.NotFound(err) {
			transitGatewayDefaultRouteTablePropagation = false
		} else if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route Table Propagation (%s): %s", TransitGatewayRouteTablePropagationCreateResourceID(transitGatewayRouteTableID, d.Id()), err)
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
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceTransitGatewayVPCAttachmentAccepterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChanges("transit_gateway_default_route_table_association", "transit_gateway_default_route_table_propagation") {
		transitGatewayID := d.Get("transit_gateway_id").(string)
		transitGateway, err := FindTransitGatewayByID(ctx, conn, transitGatewayID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
		}

		if d.HasChange("transit_gateway_default_route_table_association") {
			if err := transitGatewayRouteTableAssociationUpdate(ctx, conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
			}
		}

		if d.HasChange("transit_gateway_default_route_table_propagation") {
			if err := transitGatewayRouteTablePropagationUpdate(ctx, conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway VPC Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return diags
}

func resourceTransitGatewayVPCAttachmentAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway VPC Attachment: %s", d.Id())
	_, err := conn.DeleteTransitGatewayVpcAttachmentWithContext(ctx, &ec2.DeleteTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
	}

	if _, err := WaitTransitGatewayVPCAttachmentDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway VPC Attachment (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}
