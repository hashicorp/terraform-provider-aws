package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceMainRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMainRouteTableAssociationCreate,
		ReadWithoutTimeout:   resourceMainRouteTableAssociationRead,
		UpdateWithoutTimeout: resourceMainRouteTableAssociationUpdate,
		DeleteWithoutTimeout: resourceMainRouteTableAssociationDelete,

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

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceMainRouteTableAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	vpcID := d.Get("vpc_id").(string)

	association, err := FindMainRouteTableAssociationByVPCID(ctx, conn, vpcID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Main Route Table Association (%s): %s", vpcID, err)
	}

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: association.RouteTableAssociationId,
		RouteTableId:  aws.String(routeTableID),
	}

	log.Printf("[DEBUG] Creating Main Route Table Association: %s", input)
	output, err := conn.ReplaceRouteTableAssociationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Main Route Table Association (%s): %s", routeTableID, err)
	}

	d.SetId(aws.StringValue(output.NewAssociationId))

	log.Printf("[DEBUG] Waiting for Main Route Table Association (%s) creation", d.Id())
	if _, err := WaitRouteTableAssociationUpdated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Main Route Table Association (%s) create: %s", d.Id(), err)
	}

	d.Set("original_route_table_id", association.RouteTableId)

	return append(diags, resourceMainRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceMainRouteTableAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	_, err := FindMainRouteTableAssociationByID(ctx, conn, d.Id())

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
	conn := meta.(*conns.AWSClient).EC2Conn()

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(routeTableID),
	}

	log.Printf("[DEBUG] Updating Main Route Table Association: %s", input)
	output, err := conn.ReplaceRouteTableAssociationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Main Route Table Association (%s): %s", routeTableID, err)
	}

	// This whole thing with the resource ID being changed on update seems unsustainable.
	// Keeping it here for backwards compatibility...
	d.SetId(aws.StringValue(output.NewAssociationId))

	log.Printf("[DEBUG] Waiting for Main Route Table Association (%s) update", d.Id())
	if _, err := WaitRouteTableAssociationUpdated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Main Route Table Association (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceMainRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceMainRouteTableAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(d.Get("original_route_table_id").(string)),
	}

	log.Printf("[DEBUG] Deleting Main Route Table Association: %s", input)
	output, err := conn.ReplaceRouteTableAssociationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Main Route Table Association (%s): %s", d.Get("route_table_id").(string), err)
	}

	log.Printf("[DEBUG] Waiting for Main Route Table Association (%s) deletion", d.Id())
	if _, err := WaitRouteTableAssociationUpdated(ctx, conn, aws.StringValue(output.NewAssociationId)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Main Route Table Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}
