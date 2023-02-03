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

func ResourceRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRouteTableAssociationCreate,
		ReadWithoutTimeout:   resourceRouteTableAssociationRead,
		UpdateWithoutTimeout: resourceRouteTableAssociationUpdate,
		DeleteWithoutTimeout: resourceRouteTableAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRouteTableAssociationImport,
		},

		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"subnet_id", "gateway_id"},
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"subnet_id", "gateway_id"},
			},
		},
	}
}

func resourceRouteTableAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(routeTableID),
	}

	if v, ok := d.GetOk("gateway_id"); ok {
		input.GatewayId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Route Table Association: %s", input)
	output, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, RouteTableAssociationPropagationTimeout,
		func() (interface{}, error) {
			return conn.AssociateRouteTableWithContext(ctx, input)
		},
		errCodeInvalidRouteTableIDNotFound,
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route Table (%s) Association: %s", routeTableID, err)
	}

	d.SetId(aws.StringValue(output.(*ec2.AssociateRouteTableOutput).AssociationId))

	log.Printf("[DEBUG] Waiting for Route Table Association (%s) creation", d.Id())
	if _, err := WaitRouteTableAssociationCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route Table Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceRouteTableAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindRouteTableAssociationByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route Table Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route Table Association (%s): %s", d.Id(), err)
	}

	association := outputRaw.(*ec2.RouteTableAssociation)

	d.Set("gateway_id", association.GatewayId)
	d.Set("route_table_id", association.RouteTableId)
	d.Set("subnet_id", association.SubnetId)

	return diags
}

func resourceRouteTableAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(d.Get("route_table_id").(string)),
	}

	log.Printf("[DEBUG] Updating Route Table Association: %s", input)
	output, err := conn.ReplaceRouteTableAssociationWithContext(ctx, input)

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

	d.SetId(aws.StringValue(output.NewAssociationId))

	log.Printf("[DEBUG] Waiting for Route Table Association (%s) update", d.Id())
	if _, err := WaitRouteTableAssociationUpdated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route Table Association (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceRouteTableAssociationRead(ctx, d, meta)...)
}

func resourceRouteTableAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if err := routeTableAssociationDelete(ctx, conn, d.Id()); err != nil {
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

	conn := meta.(*conns.AWSClient).EC2Conn()

	routeTable, err := FindRouteTableByID(ctx, conn, routeTableID)

	if err != nil {
		return nil, err
	}

	var associationID string

	for _, association := range routeTable.Associations {
		if aws.StringValue(association.SubnetId) == targetID {
			d.Set("subnet_id", targetID)
			associationID = aws.StringValue(association.RouteTableAssociationId)

			break
		}

		if aws.StringValue(association.GatewayId) == targetID {
			d.Set("gateway_id", targetID)
			associationID = aws.StringValue(association.RouteTableAssociationId)

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
func routeTableAssociationDelete(ctx context.Context, conn *ec2.EC2, associationID string) error {
	log.Printf("[INFO] Deleting Route Table Association: %s", associationID)
	_, err := conn.DisassociateRouteTableWithContext(ctx, &ec2.DisassociateRouteTableInput{
		AssociationId: aws.String(associationID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAssociationIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Route Table Association (%s): %w", associationID, err)
	}

	log.Printf("[DEBUG] Waiting for Route Table Association (%s) deletion", associationID)
	if _, err := WaitRouteTableAssociationDeleted(ctx, conn, associationID); err != nil {
		return fmt.Errorf("deleting Route Table Association (%s): waiting for completion: %w", associationID, err)
	}

	return nil
}
