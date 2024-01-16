// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_ec2_transit_gateway_prefix_list_reference")
func ResourceTransitGatewayPrefixListReference() *schema.Resource {
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
			"transit_gateway_attachment_id": {
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
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateTransitGatewayPrefixListReferenceInput{}

	if v, ok := d.GetOk("blackhole"); ok {
		input.Blackhole = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_attachment_id"); ok {
		input.TransitGatewayAttachmentId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_route_table_id"); ok {
		input.TransitGatewayRouteTableId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Prefix List Reference: %s", input)
	output, err := conn.CreateTransitGatewayPrefixListReferenceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Prefix List Reference: %s", err)
	}

	d.SetId(TransitGatewayPrefixListReferenceCreateResourceID(aws.StringValue(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.StringValue(output.TransitGatewayPrefixListReference.PrefixListId)))

	if _, err := WaitTransitGatewayPrefixListReferenceStateCreated(ctx, conn, aws.StringValue(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.StringValue(output.TransitGatewayPrefixListReference.PrefixListId)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Prefix List Reference (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPrefixListReferenceRead(ctx, d, meta)...)
}

func resourceTransitGatewayPrefixListReferenceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	transitGatewayRouteTableID, prefixListID, err := TransitGatewayPrefixListReferenceParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Prefix List Reference (%s): %s", d.Id(), err)
	}

	transitGatewayPrefixListReference, err := FindTransitGatewayPrefixListReferenceByTwoPartKey(ctx, conn, transitGatewayRouteTableID, prefixListID)

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
		d.Set("transit_gateway_attachment_id", nil)
	} else {
		d.Set("transit_gateway_attachment_id", transitGatewayPrefixListReference.TransitGatewayAttachment.TransitGatewayAttachmentId)
	}
	d.Set("transit_gateway_route_table_id", transitGatewayPrefixListReference.TransitGatewayRouteTableId)

	return diags
}

func resourceTransitGatewayPrefixListReferenceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.ModifyTransitGatewayPrefixListReferenceInput{}

	if v, ok := d.GetOk("blackhole"); ok {
		input.Blackhole = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_attachment_id"); ok {
		input.TransitGatewayAttachmentId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_route_table_id"); ok {
		input.TransitGatewayRouteTableId = aws.String(v.(string))
	}

	output, err := conn.ModifyTransitGatewayPrefixListReferenceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway Prefix List Reference (%s): %s", d.Id(), err)
	}

	if _, err := WaitTransitGatewayPrefixListReferenceStateUpdated(ctx, conn, aws.StringValue(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.StringValue(output.TransitGatewayPrefixListReference.PrefixListId)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Prefix List Reference (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPrefixListReferenceRead(ctx, d, meta)...)
}

func resourceTransitGatewayPrefixListReferenceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	transitGatewayRouteTableID, prefixListID, err := TransitGatewayPrefixListReferenceParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Prefix List Reference (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Prefix List Reference: %s", d.Id())
	_, err = conn.DeleteTransitGatewayPrefixListReferenceWithContext(ctx, &ec2.DeleteTransitGatewayPrefixListReferenceInput{
		PrefixListId:               aws.String(prefixListID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Prefix List Reference (%s): %s", d.Id(), err)
	}

	if _, err := WaitTransitGatewayPrefixListReferenceStateDeleted(ctx, conn, transitGatewayRouteTableID, prefixListID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Prefix List Reference (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const transitGatewayPrefixListReferenceIDSeparator = "_"

func TransitGatewayPrefixListReferenceCreateResourceID(transitGatewayRouteTableID string, prefixListID string) string {
	parts := []string{transitGatewayRouteTableID, prefixListID}
	id := strings.Join(parts, transitGatewayPrefixListReferenceIDSeparator)

	return id
}

func TransitGatewayPrefixListReferenceParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayPrefixListReferenceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-ROUTE-TABLE-ID%[2]sPREFIX-LIST-ID", id, transitGatewayPrefixListReferenceIDSeparator)
}
