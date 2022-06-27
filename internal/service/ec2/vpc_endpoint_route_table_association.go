package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPCEndpointRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCEndpointRouteTableAssociationCreate,
		Read:   resourceVPCEndpointRouteTableAssociationRead,
		Delete: resourceVPCEndpointRouteTableAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVPCEndpointRouteTableAssociationImport,
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

func resourceVPCEndpointRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID := d.Get("vpc_endpoint_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:    aws.String(endpointID),
		AddRouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Route Table Association: %s", input)
	_, err := conn.ModifyVpcEndpoint(input)

	if err != nil {
		return fmt.Errorf("error creating VPC Endpoint Route Table Association (%s): %w", id, err)
	}

	d.SetId(VPCEndpointRouteTableAssociationCreateID(endpointID, routeTableID))

	err = WaitVPCEndpointRouteTableAssociationReady(conn, endpointID, routeTableID)

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint Route Table Association (%s) to become available: %w", id, err)
	}

	return resourceVPCEndpointRouteTableAssociationRead(d, meta)
}

func resourceVPCEndpointRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID := d.Get("vpc_endpoint_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	err := FindVPCEndpointRouteTableAssociationExists(conn, endpointID, routeTableID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Route Table Association (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint Route Table Association (%s): %w", id, err)
	}

	return nil
}

func resourceVPCEndpointRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID := d.Get("vpc_endpoint_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:       aws.String(endpointID),
		RemoveRouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint Route Table Association: %s", id)
	_, err := conn.ModifyVpcEndpoint(input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound) || tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIdNotFound) || tfawserr.ErrCodeEquals(err, errCodeInvalidParameter) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPC Endpoint Route Table Association (%s): %w", id, err)
	}

	err = WaitVPCEndpointRouteTableAssociationDeleted(conn, endpointID, routeTableID)

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint Route Table Association (%s) to delete: %w", id, err)
	}

	return nil
}

func resourceVPCEndpointRouteTableAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Wrong format of resource: %s. Please follow 'vpc-endpoint-id/route-table-id'", d.Id())
	}

	endpointID := parts[0]
	routeTableID := parts[1]
	log.Printf("[DEBUG] Importing VPC Endpoint (%s) Route Table (%s) Association", endpointID, routeTableID)

	d.SetId(VPCEndpointRouteTableAssociationCreateID(endpointID, routeTableID))
	d.Set("vpc_endpoint_id", endpointID)
	d.Set("route_table_id", routeTableID)

	return []*schema.ResourceData{d}, nil
}
