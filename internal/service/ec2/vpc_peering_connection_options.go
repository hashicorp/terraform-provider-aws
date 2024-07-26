// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_vpc_peering_connection_options", name="VPC Peering Connection Options")
func resourceVPCPeeringConnectionOptions() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCPeeringConnectionOptionsCreate,
		ReadWithoutTimeout:   resourceVPCPeeringConnectionOptionsRead,
		UpdateWithoutTimeout: resourceVPCPeeringConnectionOptionsUpdate,
		DeleteWithoutTimeout: resourceVPCPeeringConnectionOptionsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"accepter":  vpcPeeringConnectionOptionsSchema,
			"requester": vpcPeeringConnectionOptionsSchema,
			"vpc_peering_connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCPeeringConnectionOptionsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcPeeringConnectionID := d.Get("vpc_peering_connection_id").(string)
	vpcPeeringConnection, err := findVPCPeeringConnectionByID(ctx, conn, vpcPeeringConnectionID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Peering Connection (%s): %s", vpcPeeringConnectionID, err)
	}

	d.SetId(vpcPeeringConnectionID)

	if err := modifyVPCPeeringConnectionOptions(ctx, conn, d, vpcPeeringConnection, false); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceVPCPeeringConnectionOptionsRead(ctx, d, meta)...)
}

func resourceVPCPeeringConnectionOptionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcPeeringConnection, err := findVPCPeeringConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC Peering Connection Options %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Peering Connection Options (%s): %s", d.Id(), err)
	}

	d.Set("vpc_peering_connection_id", vpcPeeringConnection.VpcPeeringConnectionId)

	if vpcPeeringConnection.AccepterVpcInfo.PeeringOptions != nil {
		if err := d.Set("accepter", []interface{}{flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.AccepterVpcInfo.PeeringOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting accepter: %s", err)
		}
	} else {
		d.Set("accepter", nil)
	}

	if vpcPeeringConnection.RequesterVpcInfo.PeeringOptions != nil {
		if err := d.Set("requester", []interface{}{flattenVPCPeeringConnectionOptionsDescription(vpcPeeringConnection.RequesterVpcInfo.PeeringOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting requester: %s", err)
		}
	} else {
		d.Set("requester", nil)
	}

	return diags
}

func resourceVPCPeeringConnectionOptionsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcPeeringConnection, err := findVPCPeeringConnectionByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Peering Connection (%s): %s", d.Id(), err)
	}

	if err := modifyVPCPeeringConnectionOptions(ctx, conn, d, vpcPeeringConnection, false); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceVPCPeeringConnectionOptionsRead(ctx, d, meta)...)
}

func resourceVPCPeeringConnectionOptionsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var
	// Don't do anything with the underlying VPC Peering Connection.
	diags diag.Diagnostics

	return diags
}
