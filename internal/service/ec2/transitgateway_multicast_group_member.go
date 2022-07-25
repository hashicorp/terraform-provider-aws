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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayMulticastGroupMember() *schema.Resource {
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
			"network_interface_id": {
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
	conn := meta.(*conns.AWSClient).EC2Conn

	multicastDomainID := d.Get("transit_gateway_multicast_domain_id").(string)
	groupIPAddress := d.Get("group_ip_address").(string)
	eniID := d.Get("network_interface_id").(string)
	id := TransitGatewayMulticastGroupMemberCreateResourceID(multicastDomainID, groupIPAddress, eniID)
	input := &ec2.RegisterTransitGatewayMulticastGroupMembersInput{
		GroupIpAddress:                  aws.String(groupIPAddress),
		NetworkInterfaceIds:             aws.StringSlice([]string{eniID}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Multicast Group Member: %s", input)
	_, err := conn.RegisterTransitGatewayMulticastGroupMembersWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating EC2 Transit Gateway Multicast Group Member (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceTransitGatewayMulticastGroupMemberRead(ctx, d, meta)
}

func resourceTransitGatewayMulticastGroupMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	multicastDomainID, groupIPAddress, eniID, err := TransitGatewayMulticastGroupMemberParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFoundContext(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTransitGatewayMulticastGroupMemberByThreePartKey(conn, multicastDomainID, groupIPAddress, eniID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Multicast Group Member %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading EC2 Transit Gateway Multicast Group Member (%s): %s", d.Id(), err)
	}

	multicastGroup := outputRaw.(*ec2.TransitGatewayMulticastGroup)

	d.Set("group_ip_address", multicastGroup.GroupIpAddress)
	d.Set("network_interface_id", multicastGroup.NetworkInterfaceId)
	d.Set("transit_gateway_multicast_domain_id", multicastDomainID)

	return nil
}

func resourceTransitGatewayMulticastGroupMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	multicastDomainID, groupIPAddress, eniID, err := TransitGatewayMulticastGroupMemberParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	err = deregisterTransitGatewayMulticastGroupMember(ctx, conn, multicastDomainID, groupIPAddress, eniID)

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func deregisterTransitGatewayMulticastGroupMember(ctx context.Context, conn *ec2.EC2, multicastDomainID, groupIPAddress, eniID string) error {
	id := TransitGatewayMulticastGroupMemberCreateResourceID(multicastDomainID, groupIPAddress, eniID)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Group Member: %s", id)
	_, err := conn.DeregisterTransitGatewayMulticastGroupMembersWithContext(ctx, &ec2.DeregisterTransitGatewayMulticastGroupMembersInput{
		GroupIpAddress:                  aws.String(groupIPAddress),
		NetworkInterfaceIds:             aws.StringSlice([]string{eniID}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Multicast Group Member (%s): %w", id, err)
	}

	_, err = tfresource.RetryUntilNotFoundContext(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTransitGatewayMulticastGroupMemberByThreePartKey(conn, multicastDomainID, groupIPAddress, eniID)
	})

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Multicast Group Member (%s) delete: %w", id, err)
	}

	return nil
}

const transitGatewayMulticastGroupMemberIDSeparator = "/"

func TransitGatewayMulticastGroupMemberCreateResourceID(multicastDomainID, groupIPAddress, eniID string) string {
	parts := []string{multicastDomainID, groupIPAddress, eniID}
	id := strings.Join(parts, transitGatewayMulticastGroupMemberIDSeparator)

	return id
}

func TransitGatewayMulticastGroupMemberParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, transitGatewayMulticastGroupMemberIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected MULTICAST-DOMAIN-ID%[2]sGROUP-IP-ADDRESS%[2]sENI-ID", id, transitGatewayMulticastGroupMemberIDSeparator)
}
