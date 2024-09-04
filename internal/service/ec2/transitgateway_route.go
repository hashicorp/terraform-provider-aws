// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_route", name="Transit Gateway Route")
func resourceTransitGatewayRoute() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayRouteCreate,
		ReadWithoutTimeout:   resourceTransitGatewayRouteRead,
		DeleteWithoutTimeout: resourceTransitGatewayRouteDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"blackhole": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"destination_cidr_block": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     verify.ValidCIDRNetworkAddress,
				DiffSuppressFunc: suppressEqualCIDRBlockDiffs,
			},
			names.AttrTransitGatewayAttachmentID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"transit_gateway_route_table_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceTransitGatewayRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	destination := d.Get("destination_cidr_block").(string)
	transitGatewayRouteTableID := d.Get("transit_gateway_route_table_id").(string)
	id := transitGatewayRouteCreateResourceID(transitGatewayRouteTableID, destination)
	input := &ec2.CreateTransitGatewayRouteInput{
		Blackhole:                  aws.Bool(d.Get("blackhole").(bool)),
		DestinationCidrBlock:       aws.String(destination),
		TransitGatewayAttachmentId: aws.String(d.Get(names.AttrTransitGatewayAttachmentID).(string)),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	_, err := conn.CreateTransitGatewayRoute(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Route (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTransitGatewayRouteCreated(ctx, conn, transitGatewayRouteTableID, destination); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Route (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayRouteRead(ctx, d, meta)...)
}

func resourceTransitGatewayRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayRouteTableID, destination, err := transitGatewayRouteParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findTransitGatewayStaticRoute(ctx, conn, transitGatewayRouteTableID, destination)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Route %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route (%s): %s", d.Id(), err)
	}

	transitGatewayRoute := outputRaw.(*awstypes.TransitGatewayRoute)

	d.Set("destination_cidr_block", transitGatewayRoute.DestinationCidrBlock)
	if len(transitGatewayRoute.TransitGatewayAttachments) > 0 {
		d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayRoute.TransitGatewayAttachments[0].TransitGatewayAttachmentId)
		d.Set("blackhole", false)
	} else {
		d.Set(names.AttrTransitGatewayAttachmentID, "")
		d.Set("blackhole", true)
	}
	d.Set("transit_gateway_route_table_id", transitGatewayRouteTableID)

	return diags
}

func resourceTransitGatewayRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayRouteTableID, destination, err := transitGatewayRouteParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Route: %s", d.Id())
	_, err = conn.DeleteTransitGatewayRoute(ctx, &ec2.DeleteTransitGatewayRouteInput{
		DestinationCidrBlock:       aws.String(destination),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteNotFound, errCodeInvalidRouteTableIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Route (%s): %s", d.Id(), err)
	}

	if _, err := waitTransitGatewayRouteDeleted(ctx, conn, transitGatewayRouteTableID, destination); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Route (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const transitGatewayRouteIDSeparator = "_"

func transitGatewayRouteCreateResourceID(transitGatewayRouteTableID, destination string) string {
	parts := []string{transitGatewayRouteTableID, destination}
	id := strings.Join(parts, transitGatewayRouteIDSeparator)

	return id
}

func transitGatewayRouteParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayRouteIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-ROUTE-TABLE-ID%[2]sDESTINATION", id, transitGatewayRouteIDSeparator)
}
