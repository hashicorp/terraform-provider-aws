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

func ResourceTransitGatewayMulticastGroupSource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayMulticastGroupSourceCreate,
		ReadWithoutTimeout:   resourceTransitGatewayMulticastGroupSourceRead,
		DeleteWithoutTimeout: resourceTransitGatewayMulticastGroupSourceDelete,

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

func resourceTransitGatewayMulticastGroupSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	multicastDomainID := d.Get("transit_gateway_multicast_domain_id").(string)
	groupIPAddress := d.Get("group_ip_address").(string)
	eniID := d.Get("network_interface_id").(string)
	id := TransitGatewayMulticastGroupSourceCreateResourceID(multicastDomainID, groupIPAddress, eniID)
	input := &ec2.RegisterTransitGatewayMulticastGroupSourcesInput{
		GroupIpAddress:                  aws.String(groupIPAddress),
		NetworkInterfaceIds:             aws.StringSlice([]string{eniID}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Multicast Group Source: %s", input)
	_, err := conn.RegisterTransitGatewayMulticastGroupSourcesWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating EC2 Transit Gateway Multicast Group Source (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceTransitGatewayMulticastGroupSourceRead(ctx, d, meta)
}

func resourceTransitGatewayMulticastGroupSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	multicastDomainID, groupIPAddress, eniID, err := TransitGatewayMulticastGroupSourceParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFoundContext(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTransitGatewayMulticastGroupSourceByThreePartKey(conn, multicastDomainID, groupIPAddress, eniID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Multicast Group Source %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading EC2 Transit Gateway Multicast Group Source (%s): %s", d.Id(), err)
	}

	multicastGroup := outputRaw.(*ec2.TransitGatewayMulticastGroup)

	d.Set("group_ip_address", multicastGroup.GroupIpAddress)
	d.Set("network_interface_id", multicastGroup.NetworkInterfaceId)
	d.Set("transit_gateway_multicast_domain_id", multicastDomainID)

	return nil
}

func resourceTransitGatewayMulticastGroupSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	multicastDomainID, groupIPAddress, eniID, err := TransitGatewayMulticastGroupSourceParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	err = deregisterTransitGatewayMulticastGroupSource(ctx, conn, multicastDomainID, groupIPAddress, eniID)

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func deregisterTransitGatewayMulticastGroupSource(ctx context.Context, conn *ec2.EC2, multicastDomainID, groupIPAddress, eniID string) error {
	id := TransitGatewayMulticastGroupSourceCreateResourceID(multicastDomainID, groupIPAddress, eniID)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Group Source: %s", id)
	_, err := conn.DeregisterTransitGatewayMulticastGroupSourcesWithContext(ctx, &ec2.DeregisterTransitGatewayMulticastGroupSourcesInput{
		GroupIpAddress:                  aws.String(groupIPAddress),
		NetworkInterfaceIds:             aws.StringSlice([]string{eniID}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Multicast Group Source (%s): %w", id, err)
	}

	_, err = tfresource.RetryUntilNotFoundContext(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTransitGatewayMulticastGroupSourceByThreePartKey(conn, multicastDomainID, groupIPAddress, eniID)
	})

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Multicast Group Source (%s) delete: %w", id, err)
	}

	return nil
}

const transitGatewayMulticastGroupSourceIDSeparator = "/"

func TransitGatewayMulticastGroupSourceCreateResourceID(multicastDomainID, groupIPAddress, eniID string) string {
	parts := []string{multicastDomainID, groupIPAddress, eniID}
	id := strings.Join(parts, transitGatewayMulticastGroupSourceIDSeparator)

	return id
}

func TransitGatewayMulticastGroupSourceParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, transitGatewayMulticastGroupSourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected MULTICAST-DOMAIN-ID%[2]sGROUP-IP-ADDRESS%[2]sENI-ID", id, transitGatewayMulticastGroupSourceIDSeparator)
}
