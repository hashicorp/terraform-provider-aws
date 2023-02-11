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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPCEndpointRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointRouteTableAssociationCreate,
		ReadWithoutTimeout:   resourceVPCEndpointRouteTableAssociationRead,
		DeleteWithoutTimeout: resourceVPCEndpointRouteTableAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCEndpointRouteTableAssociationImport,
		},

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCEndpointRouteTableAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	endpointID := d.Get("vpc_endpoint_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:    aws.String(endpointID),
		AddRouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Route Table Association: %s", input)
	_, err := conn.ModifyVpcEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPC Endpoint Route Table Association (%s): %s", id, err)
	}

	d.SetId(VPCEndpointRouteTableAssociationCreateID(endpointID, routeTableID))

	err = WaitVPCEndpointRouteTableAssociationReady(ctx, conn, endpointID, routeTableID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint Route Table Association (%s) to become available: %s", id, err)
	}

	return append(diags, resourceVPCEndpointRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceVPCEndpointRouteTableAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	endpointID := d.Get("vpc_endpoint_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	err := FindVPCEndpointRouteTableAssociationExists(ctx, conn, endpointID, routeTableID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Route Table Association (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint Route Table Association (%s): %s", id, err)
	}

	return diags
}

func resourceVPCEndpointRouteTableAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	endpointID := d.Get("vpc_endpoint_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:       aws.String(endpointID),
		RemoveRouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint Route Table Association: %s", id)
	_, err := conn.ModifyVpcEndpointWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound) || tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIdNotFound) || tfawserr.ErrCodeEquals(err, errCodeInvalidParameter) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPC Endpoint Route Table Association (%s): %s", id, err)
	}

	err = WaitVPCEndpointRouteTableAssociationDeleted(ctx, conn, endpointID, routeTableID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint Route Table Association (%s) to delete: %s", id, err)
	}

	return diags
}

func resourceVPCEndpointRouteTableAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("wrong format of import ID (%s), use: 'vpc-endpoint-id/route-table-id'", d.Id())
	}

	endpointID := parts[0]
	routeTableID := parts[1]
	log.Printf("[DEBUG] Importing VPC Endpoint (%s) Route Table (%s) Association", endpointID, routeTableID)

	d.SetId(VPCEndpointRouteTableAssociationCreateID(endpointID, routeTableID))
	d.Set("vpc_endpoint_id", endpointID)
	d.Set("route_table_id", routeTableID)

	return []*schema.ResourceData{d}, nil
}
