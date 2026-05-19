// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_endpoint_route_table_association", name="VPC Endpoint Route Table Association")
// @IdentityAttribute("vpc_endpoint_id")
// @IdentityAttribute("route_table_id")
// @ImportIDHandler("vpcEndpointRouteTableAssociationImportID")
// @Testing(preIdentityVersion="v6.43.0")
// @Testing(importStateIdFunc="testAccVPCEndpointRouteTableAssociationImportStateIdFunc")
func resourceVPCEndpointRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointRouteTableAssociationCreate,
		ReadWithoutTimeout:   resourceVPCEndpointRouteTableAssociationRead,
		DeleteWithoutTimeout: resourceVPCEndpointRouteTableAssociationDelete,

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrVPCEndpointID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCEndpointRouteTableAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get(names.AttrVPCEndpointID).(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:    aws.String(endpointID),
		AddRouteTableIds: []string{routeTableID},
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Route Table Association: %v", input)
	_, err := conn.ModifyVpcEndpoint(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPC Endpoint Route Table Association (%s): %s", id, err)
	}

	d.SetId(vpcEndpointRouteTableAssociationCreateID(endpointID, routeTableID))

	err = waitVPCEndpointRouteTableAssociationReady(ctx, conn, endpointID, routeTableID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint Route Table Association (%s) to become available: %s", id, err)
	}

	return append(diags, resourceVPCEndpointRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceVPCEndpointRouteTableAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get(names.AttrVPCEndpointID).(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	_, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func(ctx context.Context) (any, error) {
		return nil, findVPCEndpointRouteTableAssociationExists(ctx, conn, endpointID, routeTableID)
	}, d.IsNewResource())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Route Table Association (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint Route Table Association (%s): %s", id, err)
	}

	return diags
}

func resourceVPCEndpointRouteTableAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	endpointID := d.Get(names.AttrVPCEndpointID).(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:       aws.String(endpointID),
		RemoveRouteTableIds: []string{routeTableID},
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint Route Table Association: %s", id)
	_, err := conn.ModifyVpcEndpoint(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound) || tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIdNotFound) || tfawserr.ErrCodeEquals(err, errCodeInvalidParameter) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPC Endpoint Route Table Association (%s): %s", id, err)
	}

	err = waitVPCEndpointRouteTableAssociationDeleted(ctx, conn, endpointID, routeTableID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint Route Table Association (%s) to delete: %s", id, err)
	}

	return diags
}

var _ inttypes.SDKv2ImportID = vpcEndpointRouteTableAssociationImportID{}

type vpcEndpointRouteTableAssociationImportID struct{}

func (vpcEndpointRouteTableAssociationImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("wrong format of import ID (%s), use: 'vpc-endpoint-id/route-table-id'", id)
	}

	endpointID := parts[0]
	routeTableID := parts[1]

	result := map[string]any{
		names.AttrVPCEndpointID: endpointID,
		"route_table_id":        routeTableID,
	}

	return vpcEndpointRouteTableAssociationCreateID(endpointID, routeTableID), result, nil
}

func (vpcEndpointRouteTableAssociationImportID) Create(d *schema.ResourceData) string {
	return vpcEndpointRouteTableAssociationCreateID(d.Get(names.AttrVPCEndpointID).(string), d.Get("route_table_id").(string))
}
