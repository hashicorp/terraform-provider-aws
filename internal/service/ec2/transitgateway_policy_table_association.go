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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_policy_table_association", name="Transit Gateway Policy Table Association")
func resourceTransitGatewayPolicyTableAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayPolicyTableAssociationCreate,
		ReadWithoutTimeout:   resourceTransitGatewayPolicyTableAssociationRead,
		DeleteWithoutTimeout: resourceTransitGatewayPolicyTableAssociationDelete,

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
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// If the TGW attachment is already associated with a TGW route table, disassociate it to prevent errors like
	// "IncorrectState: Cannot have both PolicyTableAssociation and RouteTableAssociation on the same TransitGateway Attachment".
	transitGatewayAttachmentID := d.Get(names.AttrTransitGatewayAttachmentID).(string)
	transitGatewayAttachment, err := findTransitGatewayAttachmentByID(ctx, conn, transitGatewayAttachmentID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Attachment (%s): %s", transitGatewayAttachmentID, err)
	}

	if v := transitGatewayAttachment.Association; v != nil {
		if transitGatewayRouteTableID := aws.ToString(v.TransitGatewayRouteTableId); transitGatewayRouteTableID != "" && v.State == awstypes.TransitGatewayAssociationStateAssociated {
			id := transitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID)
			input := &ec2.DisassociateTransitGatewayRouteTableInput{
				TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
				TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
			}

			if _, err := conn.DisassociateTransitGatewayRouteTable(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Route Table Association (%s): %s", id, err)
			}

			if err := waitTransitGatewayRouteTableAssociationDeleted(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Route Table Association (%s) delete: %s", id, err)
			}
		}
	}

	transitGatewayPolicyTableID := d.Get("transit_gateway_policy_table_id").(string)
	id := transitGatewayPolicyTableAssociationCreateResourceID(transitGatewayPolicyTableID, transitGatewayAttachmentID)
	input := &ec2.AssociateTransitGatewayPolicyTableInput{
		TransitGatewayAttachmentId:  aws.String(transitGatewayAttachmentID),
		TransitGatewayPolicyTableId: aws.String(transitGatewayPolicyTableID),
	}

	_, err = conn.AssociateTransitGatewayPolicyTable(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Policy Table Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTransitGatewayPolicyTableAssociationCreated(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Policy Table Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPolicyTableAssociationRead(ctx, d, meta)...)
}

func resourceTransitGatewayPolicyTableAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayPolicyTableID, transitGatewayAttachmentID, err := transitGatewayPolicyTableAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	transitGatewayPolicyTableAssociation, err := findTransitGatewayPolicyTableAssociationByTwoPartKey(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Policy Table Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Policy Table Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrResourceID, transitGatewayPolicyTableAssociation.ResourceId)
	d.Set(names.AttrResourceType, transitGatewayPolicyTableAssociation.ResourceType)
	d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayPolicyTableAssociation.TransitGatewayAttachmentId)
	d.Set("transit_gateway_policy_table_id", transitGatewayPolicyTableAssociation.TransitGatewayPolicyTableId)

	return diags
}

func resourceTransitGatewayPolicyTableAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayPolicyTableID, transitGatewayAttachmentID, err := transitGatewayPolicyTableAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Policy Table Association: %s", d.Id())
	_, err = conn.DisassociateTransitGatewayPolicyTable(ctx, &ec2.DisassociateTransitGatewayPolicyTableInput{
		TransitGatewayAttachmentId:  aws.String(transitGatewayAttachmentID),
		TransitGatewayPolicyTableId: aws.String(transitGatewayPolicyTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound, errCodeInvalidTransitGatewayPolicyTableAssociationNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Policy Table Association (%s): %s", d.Id(), err)
	}

	if _, err := waitTransitGatewayPolicyTableAssociationDeleted(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Policy Table Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const transitGatewayPolicyTableAssociationIDSeparator = "_"

func transitGatewayPolicyTableAssociationCreateResourceID(transitGatewayPolicyTableID, transitGatewayAttachmentID string) string {
	parts := []string{transitGatewayPolicyTableID, transitGatewayAttachmentID}
	id := strings.Join(parts, transitGatewayPolicyTableAssociationIDSeparator)

	return id
}

func transitGatewayPolicyTableAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayPolicyTableAssociationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-POLICY-TABLE-ID%[2]sTRANSIT-GATEWAY-ATTACHMENT-ID", id, transitGatewayPolicyTableAssociationIDSeparator)
}
