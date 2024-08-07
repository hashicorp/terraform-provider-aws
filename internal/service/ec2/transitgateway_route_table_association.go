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

// @SDKResource("aws_ec2_transit_gateway_route_table_association", name="Transit Gateway Route Table Association")
func resourceTransitGatewayRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayRouteTableAssociationCreate,
		ReadWithoutTimeout:   resourceTransitGatewayRouteTableAssociationRead,
		UpdateWithoutTimeout: schema.NoopContext,
		DeleteWithoutTimeout: resourceTransitGatewayRouteTableAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"replace_existing_association": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
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

func resourceTransitGatewayRouteTableAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayAttachmentID := d.Get(names.AttrTransitGatewayAttachmentID).(string)
	transitGatewayRouteTableID := d.Get("transit_gateway_route_table_id").(string)
	id := transitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID)

	if d.Get("replace_existing_association").(bool) {
		transitGatewayAttachment, err := findTransitGatewayAttachmentByID(ctx, conn, transitGatewayAttachmentID)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		// If no Association object was found then Gateway Attachment is not linked to a Route Table.
		if transitGatewayAttachment.Association != nil {
			transitGatewayRouteTableID := aws.ToString(transitGatewayAttachment.Association.TransitGatewayRouteTableId)

			if state := transitGatewayAttachment.Association.State; state != awstypes.TransitGatewayAssociationStateAssociated {
				return sdkdiag.AppendErrorf(diags, "existing EC2 Transit Gateway Route Table (%s) Association (%s) in unexpected state: %s", transitGatewayRouteTableID, transitGatewayAttachmentID, state)
			}

			if err := disassociateTransitGatewayRouteTable(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	input := &ec2.AssociateTransitGatewayRouteTableInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	_, err := conn.AssociateTransitGatewayRouteTable(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Route Table Association (%s): %s", id, err)
	}

	d.SetId(id)

	if err := waitTransitGatewayRouteTableAssociationCreated(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Route Table Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceTransitGatewayRouteTableAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := transitGatewayRouteTableAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	transitGatewayRouteTableAssociation, err := findTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Route Table Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route Table Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrResourceID, transitGatewayRouteTableAssociation.ResourceId)
	d.Set(names.AttrResourceType, transitGatewayRouteTableAssociation.ResourceType)
	d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayRouteTableAssociation.TransitGatewayAttachmentId)
	d.Set("transit_gateway_route_table_id", transitGatewayRouteTableID)

	return diags
}

func resourceTransitGatewayRouteTableAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := transitGatewayRouteTableAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Route Table Association: %s", d.Id())
	err = disassociateTransitGatewayRouteTable(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAssociationNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

// transitGatewayRouteTableAssociationUpdate is used by Transit Gateway attachment resources to modify their route table associations.
// The route table ID may be empty (e.g. when the Transit Gateway itself has default route table association disabled).
func transitGatewayRouteTableAssociationUpdate(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, transitGatewayAttachmentID string, associate bool) error {
	if transitGatewayRouteTableID == "" {
		// Do nothing if no route table was specified.
		return nil
	}

	id := transitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID)
	_, err := findTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

	if tfresource.NotFound(err) {
		if associate {
			input := &ec2.AssociateTransitGatewayRouteTableInput{
				TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
				TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
			}

			_, err := conn.AssociateTransitGatewayRouteTable(ctx, input)

			if err != nil {
				return fmt.Errorf("creating EC2 Transit Gateway Route Table Association (%s): %w", id, err)
			}

			if err := waitTransitGatewayRouteTableAssociationCreated(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
				return fmt.Errorf("waiting for EC2 Transit Gateway Route Table Association (%s) create: %w", id, err)
			}
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway Route Table Association (%s): %w", id, err)
	}

	if !associate {
		// Disassociation must be done only on already associated state.
		if err := waitTransitGatewayRouteTableAssociationCreated(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
			return fmt.Errorf("waiting for EC2 Transit Gateway Route Table Association (%s) create: %w", id, err)
		}

		if err := disassociateTransitGatewayRouteTable(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
			return err
		}
	}

	return nil
}

func disassociateTransitGatewayRouteTable(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, transitGatewayAttachmentID string) error {
	input := &ec2.DisassociateTransitGatewayRouteTableInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	_, err := conn.DisassociateTransitGatewayRouteTable(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway Route Table (%s) Association (%s): %w", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	if err := waitTransitGatewayRouteTableAssociationDeleted(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Route Table (%s) Association (%s) delete: %w", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	return nil
}

const transitGatewayRouteTableAssociationIDSeparator = "_"

func transitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID string) string {
	parts := []string{transitGatewayRouteTableID, transitGatewayAttachmentID}
	id := strings.Join(parts, transitGatewayRouteTableAssociationIDSeparator)

	return id
}

func transitGatewayRouteTableAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayRouteTableAssociationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-ROUTE-TABLE-ID%[2]sTRANSIT-GATEWAY-ATTACHMENT-ID", id, transitGatewayRouteTableAssociationIDSeparator)
}
