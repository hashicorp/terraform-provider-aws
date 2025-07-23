// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceGatewayAssociationResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"associated_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"associated_gateway_owner_account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"associated_gateway_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dx_gateway_association_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dx_gateway_owner_account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"proposal_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vpn_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceGatewayAssociationResourceV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"associated_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"associated_gateway_owner_account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"associated_gateway_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dx_gateway_association_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dx_gateway_owner_account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"proposal_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func gatewayAssociationStateUpgradeV0(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	log.Println("[INFO] Found Direct Connect Gateway Association state v0; migrating to v1")

	// dx_gateway_association_id was introduced in v2.8.0. Handle the case where it's not yet present.
	if v, ok := rawState["dx_gateway_association_id"]; !ok || v == nil {
		output, err := findGatewayAssociationByGatewayIDAndVirtualGatewayID(ctx, conn, rawState["dx_gateway_id"].(string), rawState["vpn_gateway_id"].(string))

		if err != nil {
			return nil, err
		}

		rawState["dx_gateway_association_id"] = aws.ToString(output.AssociationId)
	}

	return rawState, nil
}

func gatewayAssociationStateUpgradeV1(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Println("[INFO] Found Direct Connect Gateway Association state v1; migrating to v2")

	// transit_gateway_attachment_id was introduced in v6.5.0, handle the case where it's not yet present.
	if rawState["associated_gateway_type"].(string) == string(awstypes.GatewayTypeTransitGateway) {
		if v, ok := rawState[names.AttrTransitGatewayAttachmentID]; !ok || v == nil {
			output, err := tfec2.FindTransitGatewayAttachmentByTransitGatewayIDAndDirectConnectGatewayID(ctx, conn, rawState["associated_gateway_id"].(string), rawState["dx_gateway_id"].(string))

			switch {
			case tfawserr.ErrCodeEquals(err, "UnauthorizedOperation"):
				rawState[names.AttrTransitGatewayAttachmentID] = nil
			case err != nil:
				return nil, err
			default:
				rawState[names.AttrTransitGatewayAttachmentID] = aws.ToString(output.TransitGatewayAttachmentId)
			}
		}
	}

	return rawState, nil
}
