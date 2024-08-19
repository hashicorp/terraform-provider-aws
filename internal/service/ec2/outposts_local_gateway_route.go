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
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ec2_local_gateway_route")
func resourceLocalGatewayRoute() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	destinationCIDRBlock := d.Get("destination_cidr_block").(string)
	localGatewayRouteTableID := d.Get("local_gateway_route_table_id").(string)
	id := localGatewayRouteCreateResourceID(localGatewayRouteTableID, destinationCIDRBlock)
	input := &ec2.CreateLocalGatewayRouteInput{
		DestinationCidrBlock:                aws.String(destinationCIDRBlock),
		LocalGatewayRouteTableId:            aws.String(localGatewayRouteTableID),
		LocalGatewayVirtualInterfaceGroupId: aws.String(d.Get("local_gateway_virtual_interface_group_id").(string)),
	}

	_, err := conn.CreateLocalGatewayRoute(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Local Gateway Route (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceLocalGatewayRouteRead(ctx, d, meta)...)
}

func resourceLocalGatewayRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	localGatewayRouteTableID, destination, err := localGatewayRouteParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	const (
		timeout = 1 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, timeout, func() (interface{}, error) {
		return findLocalGatewayRouteByTwoPartKey(ctx, conn, localGatewayRouteTableID, destination)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Local Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Local Gateway Route (%s): %s", d.Id(), err)
	}

	localGatewayRoute := outputRaw.(*awstypes.LocalGatewayRoute)

	d.Set("destination_cidr_block", localGatewayRoute.DestinationCidrBlock)
	d.Set("local_gateway_virtual_interface_group_id", localGatewayRoute.LocalGatewayVirtualInterfaceGroupId)
	d.Set("local_gateway_route_table_id", localGatewayRoute.LocalGatewayRouteTableId)

	return diags
}

func resourceLocalGatewayRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	localGatewayRouteTableID, destination, err := localGatewayRouteParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EC2 Local Gateway Route: %s", d.Id())
	_, err = conn.DeleteLocalGatewayRoute(ctx, &ec2.DeleteLocalGatewayRouteInput{
		DestinationCidrBlock:     aws.String(destination),
		LocalGatewayRouteTableId: aws.String(localGatewayRouteTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteNotFound, errCodeInvalidLocalGatewayRouteTableIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Local Gateway Route (%s): %s", d.Id(), err)
	}

	if _, err := waitLocalGatewayRouteDeleted(ctx, conn, localGatewayRouteTableID, destination); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Local Gateway Route (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const localGatewayRouteResourceIDSeparator = "_"

func localGatewayRouteCreateResourceID(localGatewayRouteTableID, destinationCIDRBlock string) string {
	parts := []string{localGatewayRouteTableID, destinationCIDRBlock}
	id := strings.Join(parts, localGatewayRouteResourceIDSeparator)

	return id
}

func localGatewayRouteParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, localGatewayRouteResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected LOCAL-GATEWAY-ROUTE-TABLE-ID%[2]sDESTINATION", id, localGatewayRouteResourceIDSeparator)
}
