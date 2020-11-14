package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsVpcEndpointRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcEndpointRouteTableAssociationCreate,
		Read:   resourceAwsVpcEndpointRouteTableAssociationRead,
		Delete: resourceAwsVpcEndpointRouteTableAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsVpcEndpointRouteTableAssociationImport,
		},

		Schema: map[string]*schema.Schema{
			"vpc_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsVpcEndpointRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcEndpointID := d.Get("vpc_endpoint_id").(string)
	routeTableID := d.Get("route_table_id").(string)

	_, err := finder.VpcEndpointByID(conn, vpcEndpointID)

	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint (%s): %w", vpcEndpointID, err)
	}

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:    aws.String(vpcEndpointID),
		AddRouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	log.Printf("[DEBUG] Creating VPC Endpoint/Route Table association: %s", input)
	_, err = conn.ModifyVpcEndpoint(input)

	if err != nil {
		return fmt.Errorf("error creating VPC Endpoint/Route Table association: %w", err)
	}

	d.SetId(tfec2.VpcEndpointRouteTableAssociationCreateID(vpcEndpointID, routeTableID))

	err = waiter.VpcEndpointRouteTableAssociationCreated(conn, vpcEndpointID, routeTableID)

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint/Route Table association (%s) to be created: %w", d.Id(), err)
	}

	return resourceAwsVpcEndpointRouteTableAssociationRead(d, meta)
}

func resourceAwsVpcEndpointRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	associated, err := finder.VpcEndpointRouteTableAssociation(conn, d.Get("vpc_endpoint_id").(string), d.Get("route_table_id").(string))

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidVpcEndpointIdNotFound) {
		log.Printf("[WARN] VPC Endpoint not found, removing VPC Endpoint/Route Table association (%s) from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint/Route Table association (%s): %w", d.Id(), err)
	}

	if !associated {
		log.Printf("[WARN] VPC Endpoint/Route Table association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsVpcEndpointRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[INFO] Deleting VPC Endpoint/Route Table association (%s)", d.Id())
	_, err := conn.ModifyVpcEndpoint(&ec2.ModifyVpcEndpointInput{
		VpcEndpointId:       aws.String(d.Get("vpc_endpoint_id").(string)),
		RemoveRouteTableIds: aws.StringSlice([]string{d.Get("route_table_id").(string)}),
	})

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidVpcEndpointIdNotFound) ||
		tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteTableIdNotFound) ||
		tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidParameter) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPC Endpoint/Route Table association (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsVpcEndpointRouteTableAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Wrong format of resource: %s. Please follow 'vpc-endpoint-id/route-table-id'", d.Id())
	}

	vpceId := parts[0]
	rtId := parts[1]
	log.Printf("[DEBUG] Importing VPC Endpoint (%s) Route Table (%s) association", vpceId, rtId)

	d.SetId(tfec2.VpcEndpointRouteTableAssociationCreateID(vpceId, rtId))
	d.Set("vpc_endpoint_id", vpceId)
	d.Set("route_table_id", rtId)

	return []*schema.ResourceData{d}, nil
}
