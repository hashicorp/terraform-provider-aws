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

// @SDKResource("aws_ec2_transit_gateway_policy_table_association")
func ResourceTransitGatewayPolicyTableAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayPolicyTableAssociationCreate,
		ReadWithoutTimeout:   resourceTransitGatewayPolicyTableAssociationRead,
		DeleteWithoutTimeout: resourceTransitGatewayPolicyTableAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_attachment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"transit_gateway_policy_table_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceTransitGatewayPolicyTableAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	// If the TGW attachment is already associated with a TGW route table, disassociate it to prevent errors like
	// "IncorrectState: Cannot have both PolicyTableAssociation and RouteTableAssociation on the same TransitGateway Attachment".
	transitGatewayAttachmentID := d.Get("transit_gateway_attachment_id").(string)
	transitGatewayAttachment, err := FindTransitGatewayAttachmentByID(ctx, conn, transitGatewayAttachmentID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Attachment (%s): %s", transitGatewayAttachmentID, err)
	}

	if v := transitGatewayAttachment.Association; v != nil {
		if transitGatewayRouteTableID := aws.StringValue(v.TransitGatewayRouteTableId); transitGatewayRouteTableID != "" && aws.StringValue(v.State) == ec2.TransitGatewayAssociationStateAssociated {
			id := TransitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID)
			input := &ec2.DisassociateTransitGatewayRouteTableInput{
				TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
				TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
			}

			if _, err := conn.DisassociateTransitGatewayRouteTableWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Route Table Association (%s): %s", id, err)
			}

			if _, err := WaitTransitGatewayRouteTableAssociationDeleted(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Route Table Association (%s) delete: %s", id, err)
			}
		}
	}

	transitGatewayPolicyTableID := d.Get("transit_gateway_policy_table_id").(string)
	id := TransitGatewayPolicyTableAssociationCreateResourceID(transitGatewayPolicyTableID, transitGatewayAttachmentID)
	input := &ec2.AssociateTransitGatewayPolicyTableInput{
		TransitGatewayAttachmentId:  aws.String(transitGatewayAttachmentID),
		TransitGatewayPolicyTableId: aws.String(transitGatewayPolicyTableID),
	}

	_, err = conn.AssociateTransitGatewayPolicyTableWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Policy Table Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := WaitTransitGatewayPolicyTableAssociationCreated(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Policy Table Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPolicyTableAssociationRead(ctx, d, meta)...)
}

func resourceTransitGatewayPolicyTableAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	transitGatewayPolicyTableID, transitGatewayAttachmentID, err := TransitGatewayPolicyTableAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Policy Table Association (%s): %s", d.Id(), err)
	}

	transitGatewayPolicyTableAssociation, err := FindTransitGatewayPolicyTableAssociationByTwoPartKey(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Policy Table Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Policy Table Association (%s): %s", d.Id(), err)
	}

	d.Set("resource_id", transitGatewayPolicyTableAssociation.ResourceId)
	d.Set("resource_type", transitGatewayPolicyTableAssociation.ResourceType)
	d.Set("transit_gateway_attachment_id", transitGatewayPolicyTableAssociation.TransitGatewayAttachmentId)
	d.Set("transit_gateway_policy_table_id", transitGatewayPolicyTableAssociation.TransitGatewayPolicyTableId)

	return diags
}

func resourceTransitGatewayPolicyTableAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	transitGatewayPolicyTableID, transitGatewayAttachmentID, err := TransitGatewayPolicyTableAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Policy Table Association (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Policy Table Association: %s", d.Id())
	_, err = conn.DisassociateTransitGatewayPolicyTableWithContext(ctx, &ec2.DisassociateTransitGatewayPolicyTableInput{
		TransitGatewayAttachmentId:  aws.String(transitGatewayAttachmentID),
		TransitGatewayPolicyTableId: aws.String(transitGatewayPolicyTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Policy Table Association (%s): %s", d.Id(), err)
	}

	if _, err := WaitTransitGatewayPolicyTableAssociationDeleted(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Policy Table Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const transitGatewayPolicyTableAssociationIDSeparator = "_"

func TransitGatewayPolicyTableAssociationCreateResourceID(transitGatewayPolicyTableID, transitGatewayAttachmentID string) string {
	parts := []string{transitGatewayPolicyTableID, transitGatewayAttachmentID}
	id := strings.Join(parts, transitGatewayPolicyTableAssociationIDSeparator)

	return id
}

func TransitGatewayPolicyTableAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayPolicyTableAssociationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-POLICY-TABLE-ID%[2]sTRANSIT-GATEWAY-ATTACHMENT-ID", id, transitGatewayPolicyTableAssociationIDSeparator)
}
