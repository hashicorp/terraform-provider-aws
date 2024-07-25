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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route_table_association", name="Route Table Association")
func resourceRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRouteTableAssociationCreate,
		ReadWithoutTimeout:   resourceRouteTableAssociationRead,
		UpdateWithoutTimeout: resourceRouteTableAssociationUpdate,
		DeleteWithoutTimeout: resourceRouteTableAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRouteTableAssociationImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrSubnetID, "gateway_id"},
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrSubnetID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrSubnetID, "gateway_id"},
			},
		},
	}
}

func resourceRouteTableAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(routeTableID),
	}

	if v, ok := d.GetOk("gateway_id"); ok {
		input.GatewayId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSubnetID); ok {
		input.SubnetId = aws.String(v.(string))
	}

	output, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, ec2PropagationTimeout,
		func() (interface{}, error) {
			return conn.AssociateRouteTable(ctx, input)
		},
		errCodeInvalidRouteTableIDNotFound,
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route Table (%s) Association: %s", routeTableID, err)
	}

	d.SetId(aws.ToString(output.(*ec2.AssociateRouteTableOutput).AssociationId))

	if _, err := waitRouteTableAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route Table Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceRouteTableAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findRouteTableAssociationByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route Table Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route Table Association (%s): %s", d.Id(), err)
	}

	association := outputRaw.(*awstypes.RouteTableAssociation)

	d.Set("gateway_id", association.GatewayId)
	d.Set("route_table_id", association.RouteTableId)
	d.Set(names.AttrSubnetID, association.SubnetId)

	return diags
}

func resourceRouteTableAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(d.Get("route_table_id").(string)),
	}

	log.Printf("[DEBUG] Updating Route Table Association: %v", input)
	output, err := conn.ReplaceRouteTableAssociation(ctx, input)

	// This whole thing with the resource ID being changed on update seems unsustainable.
	// Keeping it here for backwards compatibility...

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAssociationIDNotFound) {
		// Not found, so just create a new one
		return append(diags, resourceRouteTableAssociationCreate(ctx, d, meta)...)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route Table Association (%s): %s", d.Id(), err)
	}

	// I don't think we'll ever reach this code for a subnet/gateway route table association.
	// It would only come in to play for a VPC main route table association.

	d.SetId(aws.ToString(output.NewAssociationId))

	if _, err := waitRouteTableAssociationUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route Table Association (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceRouteTableAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if err := routeTableAssociationDelete(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	return diags
}

func resourceRouteTableAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Unexpected format for import: %s. Use 'subnet ID/route table ID' or 'gateway ID/route table ID", d.Id())
	}

	targetID := parts[0]
	routeTableID := parts[1]

	log.Printf("[DEBUG] Importing route table association, target: %s, route table: %s", targetID, routeTableID)

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	routeTable, err := findRouteTableByID(ctx, conn, routeTableID)

	if err != nil {
		return nil, err
	}

	var associationID string

	for _, association := range routeTable.Associations {
		if aws.ToString(association.SubnetId) == targetID {
			d.Set(names.AttrSubnetID, targetID)
			associationID = aws.ToString(association.RouteTableAssociationId)

			break
		}

		if aws.ToString(association.GatewayId) == targetID {
			d.Set("gateway_id", targetID)
			associationID = aws.ToString(association.RouteTableAssociationId)

			break
		}
	}

	if associationID == "" {
		return nil, fmt.Errorf("No association found between route table ID %s and target ID %s", routeTableID, targetID)
	}

	d.SetId(associationID)
	d.Set("route_table_id", routeTableID)

	return []*schema.ResourceData{d}, nil
}

// routeTableAssociationDelete attempts to delete a route table association.
func routeTableAssociationDelete(ctx context.Context, conn *ec2.Client, associationID string, timeout time.Duration) error {
	log.Printf("[INFO] Deleting Route Table Association: %s", associationID)
	_, err := conn.DisassociateRouteTable(ctx, &ec2.DisassociateRouteTableInput{
		AssociationId: aws.String(associationID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAssociationIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Route Table Association (%s): %w", associationID, err)
	}

	if _, err := waitRouteTableAssociationDeleted(ctx, conn, associationID, timeout); err != nil {
		return fmt.Errorf("deleting Route Table Association (%s): waiting for completion: %w", associationID, err)
	}

	return nil
}
