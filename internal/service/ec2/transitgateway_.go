// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway", name="Transit Gateway")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTransitGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayCreate,
		ReadWithoutTimeout:   resourceTransitGatewayRead,
		UpdateWithoutTimeout: resourceTransitGatewayUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("default_route_table_association", func(_ context.Context, old, new, meta interface{}) bool {
				// Only changes from disable to enable for feature_set should force a new resource.
				return old.(string) == string(awstypes.DefaultRouteTableAssociationValueDisable) && new.(string) == string(awstypes.DefaultRouteTableAssociationValueEnable)
			}),
			customdiff.ForceNewIfChange("default_route_table_propagation", func(_ context.Context, old, new, meta interface{}) bool {
				// Only changes from disable to enable for feature_set should force a new resource.
				return old.(string) == string(awstypes.DefaultRouteTablePropagationValueDisable) && new.(string) == string(awstypes.DefaultRouteTablePropagationValueEnable)
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  64512,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"association_default_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_accept_shared_attachments": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.AutoAcceptSharedAttachmentsValueDisable,
				ValidateDiagFunc: enum.Validate[awstypes.AutoAcceptSharedAttachmentsValue](),
			},
			"default_route_table_association": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.DefaultRouteTableAssociationValueEnable,
				ValidateDiagFunc: enum.Validate[awstypes.DefaultRouteTableAssociationValue](),
			},
			"default_route_table_propagation": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.DefaultRouteTablePropagationValueEnable,
				ValidateDiagFunc: enum.Validate[awstypes.DefaultRouteTablePropagationValue](),
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dns_support": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.DnsSupportValueEnable,
				ValidateDiagFunc: enum.Validate[awstypes.DnsSupportValue](),
			},
			"multicast_support": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.MulticastSupportValueDisable,
				ValidateDiagFunc: enum.Validate[awstypes.MulticastSupportValue](),
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"propagation_default_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"transit_gateway_cidr_blocks": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: verify.IsIPv4CIDRBlockOrIPv6CIDRBlock(
						validation.All(
							validation.IsCIDRNetwork(0, 24),
							validation.StringDoesNotMatch(regexache.MustCompile(`^169\.254\.`), "must not be from range 169.254.0.0/16"),
						),
						validation.IsCIDRNetwork(0, 64),
					),
				},
			},
			"vpn_ecmp_support": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.VpnEcmpSupportValueEnable,
				ValidateDiagFunc: enum.Validate[awstypes.VpnEcmpSupportValue](),
			},
		},
	}
}

func resourceTransitGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateTransitGatewayInput{
		Options: &awstypes.TransitGatewayRequestOptions{
			AutoAcceptSharedAttachments:  awstypes.AutoAcceptSharedAttachmentsValue(d.Get("auto_accept_shared_attachments").(string)),
			DefaultRouteTableAssociation: awstypes.DefaultRouteTableAssociationValue(d.Get("default_route_table_association").(string)),
			DefaultRouteTablePropagation: awstypes.DefaultRouteTablePropagationValue(d.Get("default_route_table_propagation").(string)),
			DnsSupport:                   awstypes.DnsSupportValue(d.Get("dns_support").(string)),
			MulticastSupport:             awstypes.MulticastSupportValue(d.Get("multicast_support").(string)),
			VpnEcmpSupport:               awstypes.VpnEcmpSupportValue(d.Get("vpn_ecmp_support").(string)),
		},
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeTransitGateway),
	}

	if v, ok := d.GetOk("amazon_side_asn"); ok {
		input.Options.AmazonSideAsn = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_cidr_blocks"); ok && v.(*schema.Set).Len() > 0 {
		input.Options.TransitGatewayCidrBlocks = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway: %+v", input)
	output, err := conn.CreateTransitGateway(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway: %s", err)
	}

	d.SetId(aws.ToString(output.TransitGateway.TransitGatewayId))

	if _, err := waitTransitGatewayCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayRead(ctx, d, meta)...)
}

func resourceTransitGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGateway, err := findTransitGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", d.Id(), err)
	}

	d.Set("amazon_side_asn", transitGateway.Options.AmazonSideAsn)
	d.Set(names.AttrARN, transitGateway.TransitGatewayArn)
	d.Set("association_default_route_table_id", transitGateway.Options.AssociationDefaultRouteTableId)
	d.Set("auto_accept_shared_attachments", transitGateway.Options.AutoAcceptSharedAttachments)
	d.Set("default_route_table_association", transitGateway.Options.DefaultRouteTableAssociation)
	d.Set("default_route_table_propagation", transitGateway.Options.DefaultRouteTablePropagation)
	d.Set(names.AttrDescription, transitGateway.Description)
	d.Set("dns_support", transitGateway.Options.DnsSupport)
	d.Set("multicast_support", transitGateway.Options.MulticastSupport)
	d.Set(names.AttrOwnerID, transitGateway.OwnerId)
	d.Set("propagation_default_route_table_id", transitGateway.Options.PropagationDefaultRouteTableId)
	d.Set("transit_gateway_cidr_blocks", transitGateway.Options.TransitGatewayCidrBlocks)
	d.Set("vpn_ecmp_support", transitGateway.Options.VpnEcmpSupport)

	setTagsOut(ctx, transitGateway.Tags)

	return diags
}

func resourceTransitGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ec2.ModifyTransitGatewayInput{
			Options:          &awstypes.ModifyTransitGatewayOptions{},
			TransitGatewayId: aws.String(d.Id()),
		}

		if d.HasChange("amazon_side_asn") {
			input.Options.AmazonSideAsn = aws.Int64(int64(d.Get("amazon_side_asn").(int)))
		}

		if d.HasChange("auto_accept_shared_attachments") {
			input.Options.AutoAcceptSharedAttachments = awstypes.AutoAcceptSharedAttachmentsValue(d.Get("auto_accept_shared_attachments").(string))
		}

		if d.HasChange("default_route_table_association") {
			input.Options.DefaultRouteTableAssociation = awstypes.DefaultRouteTableAssociationValue(d.Get("default_route_table_association").(string))
		}

		if d.HasChange("default_route_table_propagation") {
			input.Options.DefaultRouteTablePropagation = awstypes.DefaultRouteTablePropagationValue(d.Get("default_route_table_propagation").(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("dns_support") {
			input.Options.DnsSupport = awstypes.DnsSupportValue(d.Get("dns_support").(string))
		}

		if d.HasChange("transit_gateway_cidr_blocks") {
			oRaw, nRaw := d.GetChange("transit_gateway_cidr_blocks")
			o, n := oRaw.(*schema.Set), nRaw.(*schema.Set)

			if add := n.Difference(o); add.Len() > 0 {
				input.Options.AddTransitGatewayCidrBlocks = flex.ExpandStringValueSet(add)
			}

			if del := o.Difference(n); del.Len() > 0 {
				input.Options.RemoveTransitGatewayCidrBlocks = flex.ExpandStringValueSet(del)
			}
		}

		if d.HasChange("vpn_ecmp_support") {
			input.Options.VpnEcmpSupport = awstypes.VpnEcmpSupportValue(d.Get("vpn_ecmp_support").(string))
		}

		if _, err := conn.ModifyTransitGateway(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway (%s): %s", d.Id(), err)
		}

		if _, err := waitTransitGatewayUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway (%s) update: %s", d.Id(), err)
		}
	}

	return diags
}

func resourceTransitGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway: %s", d.Id())
	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.DeleteTransitGateway(ctx, &ec2.DeleteTransitGatewayInput{
			TransitGatewayId: aws.String(d.Id()),
		})
	}, errCodeIncorrectState)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway (%s): %s", d.Id(), err)
	}

	if _, err := waitTransitGatewayDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway (%s) delete: %s", d.Id(), err)
	}

	return diags
}
