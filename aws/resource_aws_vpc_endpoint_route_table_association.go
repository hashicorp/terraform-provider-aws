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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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

	endpointId := d.Get("vpc_endpoint_id").(string)
	rtId := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointId, rtId)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:    aws.String(endpointId),
		AddRouteTableIds: aws.StringSlice([]string{rtId}),
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Route Table Association: %s", input)

	_, err := conn.ModifyVpcEndpoint(input)

	if err != nil {
		return fmt.Errorf("error creating VPC Endpoint Route Table Association (%s): %w", id, err)
	}

	d.SetId(tfec2.VpcEndpointRouteTableAssociationCreateID(endpointId, rtId))

	return resourceAwsVpcEndpointRouteTableAssociationRead(d, meta)
}

func resourceAwsVpcEndpointRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointId := d.Get("vpc_endpoint_id").(string)
	rtId := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointId, rtId)

	_, err := tfresource.RetryUntilFound(waiter.PropagationTimeout, d.IsNewResource(), func() (interface{}, error) {
		err := finder.VpcEndpointRouteTableAssociationExists(conn, endpointId, rtId)

		if err != nil {
			return nil, err
		}

		return struct{}{}, nil
	})

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

func resourceAwsVpcEndpointRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointId := d.Get("vpc_endpoint_id").(string)
	rtId := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointId, rtId)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:       aws.String(endpointId),
		RemoveRouteTableIds: aws.StringSlice([]string{rtId}),
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint Route Table Association: %s", input)

	_, err := conn.ModifyVpcEndpoint(input)

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidVpcEndpointIdNotFound) || tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteTableIdNotFound) || tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidParameter) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPC Endpoint Route Table Association (%s): %w", id, err)
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
	log.Printf("[DEBUG] Importing VPC Endpoint (%s) Route Table (%s) Association", vpceId, rtId)

	d.SetId(tfec2.VpcEndpointRouteTableAssociationCreateID(vpceId, rtId))
	d.Set("vpc_endpoint_id", vpceId)
	d.Set("route_table_id", rtId)

	return []*schema.ResourceData{d}, nil
}
