// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_connect_peer", name="Transit Gateway Connect Peer")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTransitGatewayConnectPeer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayConnectPeerCreate,
		ReadWithoutTimeout:   resourceTransitGatewayConnectPeerRead,
		UpdateWithoutTimeout: resourceTransitGatewayConnectPeerUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayConnectPeerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_asn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.Valid4ByteASN,
			},
			"bgp_peer_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_transit_gateway_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"inside_cidr_blocks": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 2,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: verify.IsIPv4CIDRBlockOrIPv6CIDRBlock(
						validation.All(
							validation.IsCIDRNetwork(29, 29),
							validation.StringMatch(regexache.MustCompile(`^169\.254\.`), "IPv4 range must be from range 169.254.0.0/16"),
							validation.StringDoesNotMatch(regexache.MustCompile(`^169\.254\.([0-5]\.0|169\.248)/29`), "IPv4 range must not be 169.254.([0-5].0|169.248)/29"),
						),
						validation.All(
							validation.IsCIDRNetwork(125, 125),
							validation.StringMatch(regexache.MustCompile(`^[fF][dD]`), "IPv6 range must be from fd00::/8"),
						),
					),
				},
			},
			"peer_address": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPAddress,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"transit_gateway_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPAddress,
			},
			names.AttrTransitGatewayAttachmentID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTransitGatewayConnectPeerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateTransitGatewayConnectPeerInput{
		InsideCidrBlocks:           flex.ExpandStringValueSet(d.Get("inside_cidr_blocks").(*schema.Set)),
		PeerAddress:                aws.String(d.Get("peer_address").(string)),
		TagSpecifications:          getTagSpecificationsIn(ctx, awstypes.ResourceTypeTransitGatewayConnectPeer),
		TransitGatewayAttachmentId: aws.String(d.Get(names.AttrTransitGatewayAttachmentID).(string)),
	}

	if v, ok := d.GetOk("bgp_asn"); ok {
		v, err := strconv.ParseInt(v.(string), 10, 64)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.BgpOptions = &awstypes.TransitGatewayConnectRequestBgpOptions{
			PeerAsn: aws.Int64(v),
		}
	}

	if v, ok := d.GetOk("transit_gateway_address"); ok {
		input.TransitGatewayAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Connect Peer: %+v", input)
	output, err := conn.CreateTransitGatewayConnectPeer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Connect Peer: %s", err)
	}

	d.SetId(aws.ToString(output.TransitGatewayConnectPeer.TransitGatewayConnectPeerId))

	if _, err := waitTransitGatewayConnectPeerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Connect Peer (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayConnectPeerRead(ctx, d, meta)...)
}

func resourceTransitGatewayConnectPeerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayConnectPeer, err := findTransitGatewayConnectPeerByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Connect Peer %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Connect Peer (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("transit-gateway-connect-peer/%s", d.Id()),
	}.String()
	bgpConfigurations := transitGatewayConnectPeer.ConnectPeerConfiguration.BgpConfigurations
	d.Set(names.AttrARN, arn)
	d.Set("bgp_asn", strconv.FormatInt(aws.ToInt64(bgpConfigurations[0].PeerAsn), 10))
	d.Set("bgp_peer_address", bgpConfigurations[0].PeerAddress)
	d.Set("bgp_transit_gateway_addresses", slices.ApplyToAll(bgpConfigurations, func(v awstypes.TransitGatewayAttachmentBgpConfiguration) string {
		return aws.ToString(v.TransitGatewayAddress)
	}))
	d.Set("inside_cidr_blocks", transitGatewayConnectPeer.ConnectPeerConfiguration.InsideCidrBlocks)
	d.Set("peer_address", transitGatewayConnectPeer.ConnectPeerConfiguration.PeerAddress)
	d.Set("transit_gateway_address", transitGatewayConnectPeer.ConnectPeerConfiguration.TransitGatewayAddress)
	d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayConnectPeer.TransitGatewayAttachmentId)

	setTagsOut(ctx, transitGatewayConnectPeer.Tags)

	return diags
}

func resourceTransitGatewayConnectPeerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceTransitGatewayConnectPeerRead(ctx, d, meta)
}

func resourceTransitGatewayConnectPeerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Connect Peer: %s", d.Id())
	_, err := conn.DeleteTransitGatewayConnectPeer(ctx, &ec2.DeleteTransitGatewayConnectPeerInput{
		TransitGatewayConnectPeerId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayConnectPeerIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Connect Peer: %s", err)
	}

	if _, err := waitTransitGatewayConnectPeerDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Connect Peer (%s) delete: %s", d.Id(), err)
	}

	return diags
}
