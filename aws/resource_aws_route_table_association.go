package aws

import (
	"fmt"
	"log"
	"strings"
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
		Importer: &schema.ResourceImporter{
			State: resourceAwsRouteTableAssociationImport,
		},

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
		},
	}
}

func resourceAwsRouteTableAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idData := strings.Split(d.Id(), "/")
	if len(idData) != 2 {
		return nil, fmt.Errorf("ID needs to be in the form of <route-table-id>/<subnet-id>")
	}

	conn := meta.(*AWSClient).ec2conn
	rtRaw, _, err := resourceAwsRouteTableStateRefreshFunc(
		conn, idData[0])()
	if err != nil {
		return nil, err
	}
	if rtRaw == nil {
		return nil, nil
	}
	rt := rtRaw.(*ec2.RouteTable)

	found := false
	for _, a := range rt.Associations {
		if *a.SubnetId == idData[1] {
			found = true
			if err = d.Set("subnet_id", idData[1]); err != nil {
				return nil, err
			}
			if err = d.Set("route_table_id", idData[0]); err != nil {
				return nil, err
			}
			d.SetId(*a.RouteTableAssociationId)
		}
	}

	if !found {
		return nil, fmt.Errorf("association with subnet %s not found", idData[1])
	}

	return []*schema.ResourceData{d}, nil
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

	var resp *ec2.AssociateRouteTableOutput
	var err error
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err = conn.AssociateRouteTable(&associationOpts)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidRouteTableID.NotFound" {
					return resource.RetryableError(awsErr)
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Set the ID and return
	d.SetId(*resp.AssociationId)
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
