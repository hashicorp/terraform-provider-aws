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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	localGatewayRouteEventualConsistencyTimeout = 1 * time.Minute
)

// @SDKResource("aws_ec2_local_gateway_route")
func ResourceLocalGatewayRoute() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocalGatewayRouteCreate,
		ReadWithoutTimeout:   resourceLocalGatewayRouteRead,
		DeleteWithoutTimeout: resourceLocalGatewayRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidCIDRNetworkAddress,
			},
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"local_gateway_virtual_interface_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLocalGatewayRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	destination := d.Get("destination_cidr_block").(string)
	localGatewayRouteTableID := d.Get("local_gateway_route_table_id").(string)

	input := &ec2.CreateLocalGatewayRouteInput{
		DestinationCidrBlock:                aws.String(destination),
		LocalGatewayRouteTableId:            aws.String(localGatewayRouteTableID),
		LocalGatewayVirtualInterfaceGroupId: aws.String(d.Get("local_gateway_virtual_interface_group_id").(string)),
	}

	_, err := conn.CreateLocalGatewayRouteWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Local Gateway Route: %s", err)
	}

	d.SetId(fmt.Sprintf("%s_%s", localGatewayRouteTableID, destination))

	return append(diags, resourceLocalGatewayRouteRead(ctx, d, meta)...)
}

func resourceLocalGatewayRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	localGatewayRouteTableID, destination, err := DecodeLocalGatewayRouteID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Local Gateway Route (%s): %s", d.Id(), err)
	}

	var localGatewayRoute *ec2.LocalGatewayRoute
	err = retry.RetryContext(ctx, localGatewayRouteEventualConsistencyTimeout, func() *retry.RetryError {
		var err error
		localGatewayRoute, err = GetLocalGatewayRoute(ctx, conn, localGatewayRouteTableID, destination)

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if d.IsNewResource() && localGatewayRoute == nil {
			return retry.RetryableError(&retry.NotFoundError{})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		localGatewayRoute, err = GetLocalGatewayRoute(ctx, conn, localGatewayRouteTableID, destination)
	}

	if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
		log.Printf("[WARN] EC2 Local Gateway Route Table (%s) not found, removing from state", localGatewayRouteTableID)
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Local Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Local Gateway Route (%s): %s", d.Id(), err)
	}

	if localGatewayRoute == nil {
		log.Printf("[WARN] EC2 Local Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	state := aws.StringValue(localGatewayRoute.State)
	if state == ec2.LocalGatewayRouteStateDeleted || state == ec2.LocalGatewayRouteStateDeleting {
		log.Printf("[WARN] EC2 Local Gateway Route (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("destination_cidr_block", localGatewayRoute.DestinationCidrBlock)
	d.Set("local_gateway_virtual_interface_group_id", localGatewayRoute.LocalGatewayVirtualInterfaceGroupId)
	d.Set("local_gateway_route_table_id", localGatewayRoute.LocalGatewayRouteTableId)

	return diags
}

func resourceLocalGatewayRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	localGatewayRouteTableID, destination, err := DecodeLocalGatewayRouteID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Local Gateway Route (%s): %s", d.Id(), err)
	}

	input := &ec2.DeleteLocalGatewayRouteInput{
		DestinationCidrBlock:     aws.String(destination),
		LocalGatewayRouteTableId: aws.String(localGatewayRouteTableID),
	}

	log.Printf("[DEBUG] Deleting EC2 Local Gateway Route (%s): %s", d.Id(), input)
	_, err = conn.DeleteLocalGatewayRouteWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, "InvalidRoute.NotFound") || tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Local Gateway Route (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeLocalGatewayRouteID(id string) (string, string, error) {
	parts := strings.Split(id, "_")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected tgw-rtb-ID_DESTINATION", id)
	}

	return parts[0], parts[1], nil
}

func GetLocalGatewayRoute(ctx context.Context, conn *ec2.EC2, localGatewayRouteTableID, destination string) (*ec2.LocalGatewayRoute, error) {
	input := &ec2.SearchLocalGatewayRoutesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("type"),
				Values: aws.StringSlice([]string{"static"}),
			},
		},
		LocalGatewayRouteTableId: aws.String(localGatewayRouteTableID),
	}

	output, err := conn.SearchLocalGatewayRoutesWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Routes) == 0 {
		return nil, nil
	}

	for _, route := range output.Routes {
		if route == nil {
			continue
		}

		if aws.StringValue(route.DestinationCidrBlock) == destination {
			return route, nil
		}
	}

	return nil, nil
}
