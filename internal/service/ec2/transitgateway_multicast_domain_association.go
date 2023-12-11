// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_ec2_transit_gateway_multicast_domain_association")
func ResourceTransitGatewayMulticastDomainAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayMulticastDomainAssociationCreate,
		ReadWithoutTimeout:   resourceTransitGatewayMulticastDomainAssociationRead,
		DeleteWithoutTimeout: resourceTransitGatewayMulticastDomainAssociationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_multicast_domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTransitGatewayMulticastDomainAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	multicastDomainID := d.Get("transit_gateway_multicast_domain_id").(string)
	attachmentID := d.Get("transit_gateway_attachment_id").(string)
	subnetID := d.Get("subnet_id").(string)
	id := TransitGatewayMulticastDomainAssociationCreateResourceID(multicastDomainID, attachmentID, subnetID)
	input := &ec2.AssociateTransitGatewayMulticastDomainInput{
		SubnetIds:                       aws.StringSlice([]string{subnetID}),
		TransitGatewayAttachmentId:      aws.String(attachmentID),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Multicast Domain Association: %s", input)
	_, err := conn.AssociateTransitGatewayMulticastDomainWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Multicast Domain Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := WaitTransitGatewayMulticastDomainAssociationCreated(ctx, conn, multicastDomainID, attachmentID, subnetID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Multicast Domain Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayMulticastDomainAssociationRead(ctx, d, meta)...)
}

func resourceTransitGatewayMulticastDomainAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	multicastDomainID, attachmentID, subnetID, err := TransitGatewayMulticastDomainAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	multicastDomainAssociation, err := FindTransitGatewayMulticastDomainAssociationByThreePartKey(ctx, conn, multicastDomainID, attachmentID, subnetID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Multicast Domain Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Multicast Domain Association (%s): %s", d.Id(), err)
	}

	d.Set("subnet_id", multicastDomainAssociation.Subnet.SubnetId)
	d.Set("transit_gateway_attachment_id", multicastDomainAssociation.TransitGatewayAttachmentId)
	d.Set("transit_gateway_multicast_domain_id", multicastDomainID)

	return diags
}

func resourceTransitGatewayMulticastDomainAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	multicastDomainID, attachmentID, subnetID, err := TransitGatewayMulticastDomainAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = disassociateTransitGatewayMulticastDomain(ctx, conn, multicastDomainID, attachmentID, subnetID, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func disassociateTransitGatewayMulticastDomain(ctx context.Context, conn *ec2.EC2, multicastDomainID, attachmentID, subnetID string, timeout time.Duration) error {
	id := TransitGatewayMulticastDomainAssociationCreateResourceID(multicastDomainID, attachmentID, subnetID)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Domain Association: %s", id)
	_, err := conn.DisassociateTransitGatewayMulticastDomainWithContext(ctx, &ec2.DisassociateTransitGatewayMulticastDomainInput{
		SubnetIds:                       aws.StringSlice([]string{subnetID}),
		TransitGatewayAttachmentId:      aws.String(attachmentID),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway Multicast Domain Association (%s): %w", id, err)
	}

	if _, err := WaitTransitGatewayMulticastDomainAssociationDeleted(ctx, conn, multicastDomainID, attachmentID, subnetID, timeout); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Multicast Domain Association (%s) delete: %w", id, err)
	}

	return nil
}

const transitGatewayMulticastDomainAssociationIDSeparator = "/"

func TransitGatewayMulticastDomainAssociationCreateResourceID(multicastDomainID, attachmentID, subnetID string) string {
	parts := []string{multicastDomainID, attachmentID, subnetID}
	id := strings.Join(parts, transitGatewayMulticastDomainAssociationIDSeparator)

	return id
}

func TransitGatewayMulticastDomainAssociationParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, transitGatewayMulticastDomainAssociationIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected MULTICAST-DOMAIN-ID%[2]sATTACHMENT-ID%[2]sSUBNET-ID", id, transitGatewayMulticastDomainAssociationIDSeparator)
}
