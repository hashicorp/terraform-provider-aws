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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_multicast_group_member", name="Transit Gateway Multicast Group Member")
func resourceTransitGatewayMulticastGroupMember() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayMulticastGroupMemberCreate,
		ReadWithoutTimeout:   resourceTransitGatewayMulticastGroupMemberRead,
		DeleteWithoutTimeout: resourceTransitGatewayMulticastGroupMemberDelete,

		Schema: map[string]*schema.Schema{
			"group_ip_address": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidMulticastIPAddress,
			},
			names.AttrNetworkInterfaceID: {
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

func resourceTransitGatewayMulticastGroupMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	multicastDomainID := d.Get("transit_gateway_multicast_domain_id").(string)
	groupIPAddress := d.Get("group_ip_address").(string)
	eniID := d.Get(names.AttrNetworkInterfaceID).(string)
	id := transitGatewayMulticastGroupMemberCreateResourceID(multicastDomainID, groupIPAddress, eniID)
	input := &ec2.RegisterTransitGatewayMulticastGroupMembersInput{
		GroupIpAddress:                  aws.String(groupIPAddress),
		NetworkInterfaceIds:             []string{eniID},
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Multicast Group Member: %+v", input)
	_, err := conn.RegisterTransitGatewayMulticastGroupMembers(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway Multicast Group Member (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceTransitGatewayMulticastGroupMemberRead(ctx, d, meta)...)
}

func resourceTransitGatewayMulticastGroupMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	multicastDomainID, groupIPAddress, eniID, err := transitGatewayMulticastGroupMemberParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findTransitGatewayMulticastGroupMemberByThreePartKey(ctx, conn, multicastDomainID, groupIPAddress, eniID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Multicast Group Member %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Multicast Group Member (%s): %s", d.Id(), err)
	}

	multicastGroup := outputRaw.(*awstypes.TransitGatewayMulticastGroup)

	d.Set("group_ip_address", multicastGroup.GroupIpAddress)
	d.Set(names.AttrNetworkInterfaceID, multicastGroup.NetworkInterfaceId)
	d.Set("transit_gateway_multicast_domain_id", multicastDomainID)

	return diags
}

func resourceTransitGatewayMulticastGroupMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	multicastDomainID, groupIPAddress, eniID, err := transitGatewayMulticastGroupMemberParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = deregisterTransitGatewayMulticastGroupMember(ctx, conn, multicastDomainID, groupIPAddress, eniID)

	if tfawserr.ErrCodeEquals(err, errCodeTransitGatewayMulticastGroupMemberNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func deregisterTransitGatewayMulticastGroupMember(ctx context.Context, conn *ec2.Client, multicastDomainID, groupIPAddress, eniID string) error {
	id := transitGatewayMulticastGroupMemberCreateResourceID(multicastDomainID, groupIPAddress, eniID)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Group Member: %s", id)
	_, err := conn.DeregisterTransitGatewayMulticastGroupMembers(ctx, &ec2.DeregisterTransitGatewayMulticastGroupMembersInput{
		GroupIpAddress:                  aws.String(groupIPAddress),
		NetworkInterfaceIds:             []string{eniID},
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway Multicast Group Member (%s): %w", id, err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findTransitGatewayMulticastGroupMemberByThreePartKey(ctx, conn, multicastDomainID, groupIPAddress, eniID)
	})

	if err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Multicast Group Member (%s) delete: %w", id, err)
	}

	return nil
}

const transitGatewayMulticastGroupMemberIDSeparator = "/"

func transitGatewayMulticastGroupMemberCreateResourceID(multicastDomainID, groupIPAddress, eniID string) string {
	parts := []string{multicastDomainID, groupIPAddress, eniID}
	id := strings.Join(parts, transitGatewayMulticastGroupMemberIDSeparator)

	return id
}

func transitGatewayMulticastGroupMemberParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, transitGatewayMulticastGroupMemberIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected MULTICAST-DOMAIN-ID%[2]sGROUP-IP-ADDRESS%[2]sENI-ID", id, transitGatewayMulticastGroupMemberIDSeparator)
}
