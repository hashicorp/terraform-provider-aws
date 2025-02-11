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
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_prefix_list_reference", name="Transit Gateway Prefix List Reference")
func resourceTransitGatewayPrefixListReference() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayPrefixListReferenceCreate,
		ReadWithoutTimeout:   resourceTransitGatewayPrefixListReferenceRead,
		UpdateWithoutTimeout: resourceTransitGatewayPrefixListReferenceUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayPrefixListReferenceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"blackhole": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"prefix_list_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTransitGatewayAttachmentID: {
				Type:         schema.TypeString,
				Optional:     true,
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

func resourceTransitGatewayPrefixListReferenceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateTransitGatewayPrefixListReferenceInput{}

	if v, ok := d.GetOk("blackhole"); ok {
		input.Blackhole = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrTransitGatewayAttachmentID); ok {
		input.TransitGatewayAttachmentId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_route_table_id"); ok {
		input.TransitGatewayRouteTableId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Prefix List Reference: %+v", input)
	output, err := conn.CreateTransitGatewayPrefixListReference(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Prefix List Reference: %s", err)
	}

	d.SetId(transitGatewayPrefixListReferenceCreateResourceID(aws.ToString(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.ToString(output.TransitGatewayPrefixListReference.PrefixListId)))

	if _, err := waitTransitGatewayPrefixListReferenceStateCreated(ctx, conn, aws.ToString(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.ToString(output.TransitGatewayPrefixListReference.PrefixListId)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Prefix List Reference (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPrefixListReferenceRead(ctx, d, meta)...)
}

func resourceTransitGatewayPrefixListReferenceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayRouteTableID, prefixListID, err := transitGatewayPrefixListReferenceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	transitGatewayPrefixListReference, err := findTransitGatewayPrefixListReferenceByTwoPartKey(ctx, conn, transitGatewayRouteTableID, prefixListID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Prefix List Reference (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Prefix List Reference (%s): %s", d.Id(), err)
	}

	d.Set("blackhole", transitGatewayPrefixListReference.Blackhole)
	d.Set("prefix_list_id", transitGatewayPrefixListReference.PrefixListId)
	d.Set("prefix_list_owner_id", transitGatewayPrefixListReference.PrefixListOwnerId)
	if transitGatewayPrefixListReference.TransitGatewayAttachment == nil {
		d.Set(names.AttrTransitGatewayAttachmentID, nil)
	} else {
		d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayPrefixListReference.TransitGatewayAttachment.TransitGatewayAttachmentId)
	}
	d.Set("transit_gateway_route_table_id", transitGatewayPrefixListReference.TransitGatewayRouteTableId)

	return diags
}

func resourceTransitGatewayPrefixListReferenceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.ModifyTransitGatewayPrefixListReferenceInput{}

	if v, ok := d.GetOk("blackhole"); ok {
		input.Blackhole = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrTransitGatewayAttachmentID); ok {
		input.TransitGatewayAttachmentId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_route_table_id"); ok {
		input.TransitGatewayRouteTableId = aws.String(v.(string))
	}

	output, err := conn.ModifyTransitGatewayPrefixListReference(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway Prefix List Reference (%s): %s", d.Id(), err)
	}

	if _, err := waitTransitGatewayPrefixListReferenceStateUpdated(ctx, conn, aws.ToString(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.ToString(output.TransitGatewayPrefixListReference.PrefixListId)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Prefix List Reference (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPrefixListReferenceRead(ctx, d, meta)...)
}

func resourceTransitGatewayPrefixListReferenceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayRouteTableID, prefixListID, err := transitGatewayPrefixListReferenceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Prefix List Reference: %s", d.Id())
	_, err = conn.DeleteTransitGatewayPrefixListReference(ctx, &ec2.DeleteTransitGatewayPrefixListReferenceInput{
		PrefixListId:               aws.String(prefixListID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound, errCodeInvalidPrefixListIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Prefix List Reference (%s): %s", d.Id(), err)
	}

	if _, err := waitTransitGatewayPrefixListReferenceStateDeleted(ctx, conn, transitGatewayRouteTableID, prefixListID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Prefix List Reference (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const transitGatewayPrefixListReferenceIDSeparator = "_"

func transitGatewayPrefixListReferenceCreateResourceID(transitGatewayRouteTableID string, prefixListID string) string {
	parts := []string{transitGatewayRouteTableID, prefixListID}
	id := strings.Join(parts, transitGatewayPrefixListReferenceIDSeparator)

	return id
}

func transitGatewayPrefixListReferenceParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayPrefixListReferenceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-ROUTE-TABLE-ID%[2]sPREFIX-LIST-ID", id, transitGatewayPrefixListReferenceIDSeparator)
}
