// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_vpn_gateway_route_propagation", name="VPN Gateway Route Propagation")
func resourceVPNGatewayRoutePropagation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPNGatewayRoutePropagationEnable,
		ReadWithoutTimeout:   resourceVPNGatewayRoutePropagationRead,
		DeleteWithoutTimeout: resourceVPNGatewayRoutePropagationDisable,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpn_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPNGatewayRoutePropagationEnable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	gatewayID := d.Get("vpn_gateway_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	err := routeTableEnableVGWRoutePropagationV2(ctx, conn, routeTableID, gatewayID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(VPNGatewayRoutePropagationCreateID(routeTableID, gatewayID))

	return append(diags, resourceVPNGatewayRoutePropagationRead(ctx, d, meta)...)
}

func resourceVPNGatewayRoutePropagationDisable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	routeTableID, gatewayID, err := VPNGatewayRoutePropagationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = routeTableDisableVGWRoutePropagationV2(ctx, conn, routeTableID, gatewayID)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return diags
	}

	return sdkdiag.AppendFromErr(diags, err)
}

func resourceVPNGatewayRoutePropagationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	routeTableID, gatewayID, err := VPNGatewayRoutePropagationParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = FindVPNGatewayRoutePropagationExists(ctx, conn, routeTableID, gatewayID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route Table (%s) VPN Gateway (%s) route propagation not found, removing from state", routeTableID, gatewayID)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route Table (%s) VPN Gateway (%s) route propagation: %s", routeTableID, gatewayID, err)
	}

	return diags
}

// routeTableDisableVGWRoutePropagation attempts to disable VGW route propagation.
// Any error is returned.
func routeTableDisableVGWRoutePropagationV2(ctx context.Context, conn *ec2.Client, routeTableID, gatewayID string) error {
	input := &ec2.DisableVgwRoutePropagationInput{
		GatewayId:    aws.String(gatewayID),
		RouteTableId: aws.String(routeTableID),
	}

	_, err := conn.DisableVgwRoutePropagation(ctx, input)

	if err != nil {
		return fmt.Errorf("disabling Route Table (%s) VPN Gateway (%s) route propagation: %w", routeTableID, gatewayID, err)
	}

	return nil
}

// routeTableEnableVGWRoutePropagation attempts to enable VGW route propagation.
// The specified eventual consistency timeout is respected.
// Any error is returned.
func routeTableEnableVGWRoutePropagationV2(ctx context.Context, conn *ec2.Client, routeTableID, gatewayID string, timeout time.Duration) error {
	input := &ec2.EnableVgwRoutePropagationInput{
		GatewayId:    aws.String(gatewayID),
		RouteTableId: aws.String(routeTableID),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout,
		func() (interface{}, error) {
			return conn.EnableVgwRoutePropagation(ctx, input)
		},
		errCodeGatewayNotAttached,
	)

	if err != nil {
		return fmt.Errorf("enabling Route Table (%s) VPN Gateway (%s) route propagation: %w", routeTableID, gatewayID, err)
	}

	return nil
}
