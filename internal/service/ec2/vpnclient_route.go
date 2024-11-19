// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_client_vpn_route", name="Client VPN Route")
func resourceClientVPNRoute() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClientVPNRouteCreate,
		ReadWithoutTimeout:   resourceClientVPNRouteRead,
		DeleteWithoutTimeout: resourceClientVPNRouteDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(4 * time.Minute),
			Delete: schema.DefaultTimeout(4 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
			},
			"origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target_vpc_subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceClientVPNRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	targetSubnetID := d.Get("target_vpc_subnet_id").(string)
	destinationCIDR := d.Get("destination_cidr_block").(string)
	id := clientVPNRouteCreateResourceID(endpointID, targetSubnetID, destinationCIDR)
	input := &ec2.CreateClientVpnRouteInput{
		ClientToken:          aws.String(sdkid.UniqueId()),
		ClientVpnEndpointId:  aws.String(endpointID),
		DestinationCidrBlock: aws.String(destinationCIDR),
		TargetVpcSubnetId:    aws.String(targetSubnetID),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return conn.CreateClientVpnRoute(ctx, input)
	}, errCodeInvalidClientVPNActiveAssociationNotFound)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Client VPN Route (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitClientVPNRouteCreated(ctx, conn, endpointID, targetSubnetID, destinationCIDR, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Client VPN Route (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClientVPNRouteRead(ctx, d, meta)...)
}

func resourceClientVPNRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID, targetSubnetID, destinationCIDR, err := clientVPNRouteParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	route, err := findClientVPNRouteByThreePartKey(ctx, conn, endpointID, targetSubnetID, destinationCIDR)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Client VPN Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Client VPN Route (%s): %s", d.Id(), err)
	}

	d.Set("client_vpn_endpoint_id", route.ClientVpnEndpointId)
	d.Set(names.AttrDescription, route.Description)
	d.Set("destination_cidr_block", route.DestinationCidr)
	d.Set("origin", route.Origin)
	d.Set("target_vpc_subnet_id", route.TargetSubnet)
	d.Set(names.AttrType, route.Type)

	return diags
}

func resourceClientVPNRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID, targetSubnetID, destinationCIDR, err := clientVPNRouteParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EC2 Client VPN Route: %s", d.Id())
	_, err = conn.DeleteClientVpnRoute(ctx, &ec2.DeleteClientVpnRouteInput{
		ClientVpnEndpointId:  aws.String(endpointID),
		DestinationCidrBlock: aws.String(destinationCIDR),
		TargetVpcSubnetId:    aws.String(targetSubnetID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound, errCodeInvalidClientVPNRouteNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Client VPN Route (%s): %s", d.Id(), err)
	}

	if _, err := waitClientVPNRouteDeleted(ctx, conn, endpointID, targetSubnetID, destinationCIDR, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Client VPN Route (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

const clientVPNRouteIDSeparator = ","

func clientVPNRouteCreateResourceID(endpointID, targetSubnetID, destinationCIDR string) string {
	parts := []string{endpointID, targetSubnetID, destinationCIDR}
	id := strings.Join(parts, clientVPNRouteIDSeparator)

	return id
}

func clientVPNRouteParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, clientVPNRouteIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected EndpointID%[2]sTargetSubnetID%[2]sDestinationCIDRBlock", id, clientVPNRouteIDSeparator)
}
