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

// @SDKResource("aws_ec2_transit_gateway_route_table_propagation", name="Transit Gateway Route Table Propagation")
func resourceTransitGatewayRouteTablePropagation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayRouteTablePropagationCreate,
		ReadWithoutTimeout:   resourceTransitGatewayRouteTablePropagationRead,
		DeleteWithoutTimeout: resourceTransitGatewayRouteTablePropagationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTransitGatewayAttachmentID: {
				Type:         schema.TypeString,
				Required:     true,
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

func resourceTransitGatewayRouteTablePropagationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayAttachmentID := d.Get(names.AttrTransitGatewayAttachmentID).(string)
	transitGatewayRouteTableID := d.Get("transit_gateway_route_table_id").(string)
	id := transitGatewayRouteTablePropagationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID)
	input := &ec2.EnableTransitGatewayRouteTablePropagationInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	_, err := conn.EnableTransitGatewayRouteTablePropagation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Route Table Propagation (%s): %s", id, err)
	}

	d.SetId(id)

	if err := waitTransitGatewayRouteTablePropagationCreated(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Route Table Propagation (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayRouteTablePropagationRead(ctx, d, meta)...)
}

func resourceTransitGatewayRouteTablePropagationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := transitGatewayRouteTablePropagationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	transitGatewayPropagation, err := findTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Route Table Propagation %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route Table Propagation (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrResourceID, transitGatewayPropagation.ResourceId)
	d.Set(names.AttrResourceType, transitGatewayPropagation.ResourceType)
	d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayPropagation.TransitGatewayAttachmentId)
	d.Set("transit_gateway_route_table_id", transitGatewayRouteTableID)

	return diags
}

func resourceTransitGatewayRouteTablePropagationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := transitGatewayRouteTablePropagationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Route Table Propagation: %s", d.Id())
	_, err = conn.DisableTransitGatewayRouteTablePropagation(ctx, &ec2.DisableTransitGatewayRouteTablePropagationInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound, errCodeTransitGatewayRouteTablePropagationNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Route Table Propagation (%s): %s", d.Id(), err)
	}

	if err := waitTransitGatewayRouteTablePropagationDeleted(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Route Table Propagation (%s) delete: %s", d.Id(), err)
	}

	return diags
}

// transitGatewayRouteTablePropagationUpdate is used by Transit Gateway attachment resources to modify their route table propagations.
// The route table ID may be empty (e.g. when the Transit Gateway itself has default route table propagation disabled).
func transitGatewayRouteTablePropagationUpdate(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, transitGatewayAttachmentID string, enable bool) error {
	if transitGatewayRouteTableID == "" {
		// Do nothing if no route table was specified.
		return nil
	}

	id := transitGatewayRouteTablePropagationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID)
	_, err := findTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

	if tfresource.NotFound(err) {
		if enable {
			input := &ec2.EnableTransitGatewayRouteTablePropagationInput{
				TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
				TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
			}

			if _, err := conn.EnableTransitGatewayRouteTablePropagation(ctx, input); err != nil {
				return fmt.Errorf("creating EC2 Transit Gateway Route Table Propagation (%s): %w", id, err)
			}

			if err := waitTransitGatewayRouteTablePropagationCreated(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
				return fmt.Errorf("waiting for EC2 Transit Gateway Route Table Propagation (%s) create: %w", id, err)
			}
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway Route Table Propagation (%s): %w", id, err)
	}

	if !enable {
		// Disabling must be done only on already enabled state.
		if err := waitTransitGatewayRouteTablePropagationCreated(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
			return fmt.Errorf("waiting for EC2 Transit Gateway Route Table Propagation (%s) create: %w", id, err)
		}

		input := &ec2.DisableTransitGatewayRouteTablePropagationInput{
			TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
			TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
		}

		if _, err := conn.DisableTransitGatewayRouteTablePropagation(ctx, input); err != nil {
			return fmt.Errorf("deleting EC2 Transit Gateway Route Table Propagation (%s): %w", id, err)
		}

		if err := waitTransitGatewayRouteTablePropagationDeleted(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
			return fmt.Errorf("waiting for EC2 Transit Gateway Route Table Propagation (%s) delete: %w", id, err)
		}
	}

	return nil
}

const transitGatewayRouteTablePropagationIDSeparator = "_"

func transitGatewayRouteTablePropagationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID string) string {
	parts := []string{transitGatewayRouteTableID, transitGatewayAttachmentID}
	id := strings.Join(parts, transitGatewayRouteTablePropagationIDSeparator)

	return id
}

func transitGatewayRouteTablePropagationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayRouteTablePropagationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-ROUTE-TABLE-ID%[2]sTRANSIT-GATEWAY-ATTACHMENT-ID", id, transitGatewayRouteTablePropagationIDSeparator)
}
