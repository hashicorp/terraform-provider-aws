// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_main_route_table_association", name="Main Route Table Association")
func resourceMainRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMainRouteTableAssociationCreate,
		ReadWithoutTimeout:   resourceMainRouteTableAssociationRead,
		UpdateWithoutTimeout: resourceMainRouteTableAssociationUpdate,
		DeleteWithoutTimeout: resourceMainRouteTableAssociationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// We use this field to record the main route table that is automatically
			// created when the VPC is created. We need this to be able to "destroy"
			// our main route table association, which we do by returning this route
			// table to its original place as the Main Route Table for the VPC.
			"original_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceMainRouteTableAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcID := d.Get(names.AttrVPCID).(string)
	association, err := findMainRouteTableAssociationByVPCID(ctx, conn, vpcID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Main Route Table Association (%s): %s", vpcID, err)
	}

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: association.RouteTableAssociationId,
		RouteTableId:  aws.String(routeTableID),
	}

	output, err := conn.ReplaceRouteTableAssociation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Main Route Table Association (%s): %s", routeTableID, err)
	}

	d.SetId(aws.ToString(output.NewAssociationId))

	if _, err := waitRouteTableAssociationUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Main Route Table Association (%s) create: %s", d.Id(), err)
	}

	d.Set("original_route_table_id", association.RouteTableId)

	return append(diags, resourceMainRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceMainRouteTableAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	_, err := findMainRouteTableAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Main Route Table Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Main Route Table Association (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceMainRouteTableAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(routeTableID),
	}

	output, err := conn.ReplaceRouteTableAssociation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Main Route Table Association (%s): %s", routeTableID, err)
	}

	// This whole thing with the resource ID being changed on update seems unsustainable.
	// Keeping it here for backwards compatibility...
	d.SetId(aws.ToString(output.NewAssociationId))

	log.Printf("[DEBUG] Waiting for Main Route Table Association (%s) update", d.Id())
	if _, err := waitRouteTableAssociationUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Main Route Table Association (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceMainRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceMainRouteTableAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting Main Route Table Association: %s", d.Id())
	output, err := conn.ReplaceRouteTableAssociation(ctx, &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(d.Get("original_route_table_id").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Main Route Table Association (%s): %s", d.Get("route_table_id").(string), err)
	}

	if _, err := waitRouteTableAssociationUpdated(ctx, conn, aws.ToString(output.NewAssociationId), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Main Route Table Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}
