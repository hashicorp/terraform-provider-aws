package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRouteTableAssociationCreate,
		Read:   resourceAwsRouteTableAssociationRead,
		Update: resourceAwsRouteTableAssociationUpdate,
		Delete: resourceAwsRouteTableAssociationDelete,

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"force_replace": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceAwsRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf(
		"[INFO] Creating route table association: %s => %s",
		d.Get("subnet_id").(string),
		d.Get("route_table_id").(string))

	associationOpts := ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(d.Get("route_table_id").(string)),
		SubnetId:     aws.String(d.Get("subnet_id").(string)),
	}

	var associationID string
	var err error
	err = resource.Retry(30*time.Second, func() *resource.RetryError {
		resp, err := conn.AssociateRouteTable(&associationOpts)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidRouteTableID.NotFound" {
					return resource.RetryableError(awsErr)
				}
				if awsErr.Code() == "Resource.AlreadyAssociated" && d.Get(
					"force_replace").(bool) {

					associationID, err = findExistingSubnetAssociationID(
						conn,
						d.Get("subnet_id").(string),
					)
					if err != nil {
						return resource.NonRetryableError(err)
					}

					log.Print("[INFO] Replacing route table association")
					input := &ec2.ReplaceRouteTableAssociationInput{
						AssociationId: aws.String(associationID),
						RouteTableId:  aws.String(d.Get("route_table_id").(string)),
					}
					_, err = conn.ReplaceRouteTableAssociation(input)
					if err == nil {
						return nil
					}
				}
			}
			return resource.NonRetryableError(err)
		} else {
			associationID = *resp.AssociationId
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Set the ID and return
	d.SetId(associationID)
	log.Printf("[INFO] Association ID: %s", d.Id())

	return nil
}

func resourceAwsRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// Get the routing table that this association belongs to
	rtRaw, _, err := resourceAwsRouteTableStateRefreshFunc(
		conn, d.Get("route_table_id").(string))()
	if err != nil {
		return err
	}
	if rtRaw == nil {
		return nil
	}
	rt := rtRaw.(*ec2.RouteTable)

	// Inspect that the association exists
	found := false
	for _, a := range rt.Associations {
		if *a.RouteTableAssociationId == d.Id() {
			found = true
			d.Set("subnet_id", *a.SubnetId)
			break
		}
	}

	if !found {
		// It seems it doesn't exist anymore, so clear the ID
		d.SetId("")
	}

	return nil
}

func resourceAwsRouteTableAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf(
		"[INFO] Creating route table association: %s => %s",
		d.Get("subnet_id").(string),
		d.Get("route_table_id").(string))

	req := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(d.Get("route_table_id").(string)),
	}
	resp, err := conn.ReplaceRouteTableAssociation(req)

	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && ec2err.Code() == "InvalidAssociationID.NotFound" {
			// Not found, so just create a new one
			return resourceAwsRouteTableAssociationCreate(d, meta)
		}

		return err
	}

	// Update the ID
	d.SetId(*resp.NewAssociationId)
	log.Printf("[INFO] Association ID: %s", d.Id())

	return nil
}

func resourceAwsRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[INFO] Deleting route table association: %s", d.Id())
	_, err := conn.DisassociateRouteTable(&ec2.DisassociateRouteTableInput{
		AssociationId: aws.String(d.Id()),
	})
	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && ec2err.Code() == "InvalidAssociationID.NotFound" {
			return nil
		}

		return fmt.Errorf("Error deleting route table association: %s", err)
	}

	return nil
}

func findExistingSubnetAssociationID(conn *ec2.EC2, subnetID string) (string, error) {
	// only way to get association id is through the route table

	input := &ec2.DescribeRouteTablesInput{}
	input.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"association.subnet-id": subnetID,
		},
	)

	output, err := conn.DescribeRouteTables(input)
	if err != nil || len(output.RouteTables) == 0 {
		return "", fmt.Errorf("Error finding route table: %v", err)
	}

	rt := output.RouteTables[0]

	var associationID string
	for _, a := range rt.Associations {
		if *a.SubnetId == subnetID {
			associationID = *a.RouteTableAssociationId
			break
		}
	}
	if associationID == "" {
		return "", fmt.Errorf("Error finding route table, ID: %v", *rt.RouteTableId)
	}

	return associationID, nil
}
