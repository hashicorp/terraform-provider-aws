package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsVpcEndpointSubnetAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcEndpointSubnetAssociationCreate,
		Read:   resourceAwsVpcEndpointSubnetAssociationRead,
		Delete: resourceAwsVpcEndpointSubnetAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsVpcEndpointSubnetAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointId := d.Get("vpc_endpoint_id").(string)
	snId := d.Get("subnet_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointId, snId)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(endpointId),
		AddSubnetIds:  aws.StringSlice([]string{snId}),
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Subnet Association: %s", input)

	// See https://github.com/hashicorp/terraform-provider-aws/issues/3382.
	// Prevent concurrent subnet association requests and delay between requests.
	mk := "vpc_endpoint_subnet_association_" + endpointId
	awsMutexKV.Lock(mk)
	defer awsMutexKV.Unlock(mk)

	c := &resource.StateChangeConf{
		Delay:   1 * time.Minute,
		Timeout: 3 * time.Minute,
		Target:  []string{"ok"},
		Refresh: func() (interface{}, string, error) {
			output, err := conn.ModifyVpcEndpoint(input)

			return output, "ok", err
		},
	}
	_, err := c.WaitForState()

	if err != nil {
		return fmt.Errorf("error creating VPC Endpoint Subnet Association (%s): %w", id, err)
	}

	d.SetId(tfec2.VpcEndpointSubnetAssociationCreateID(endpointId, snId))

	_, err = waiter.VpcEndpointAvailable(conn, endpointId, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint (%s) to become available: %w", endpointId, err)
	}

	return resourceAwsVpcEndpointSubnetAssociationRead(d, meta)
}

func resourceAwsVpcEndpointSubnetAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointId := d.Get("vpc_endpoint_id").(string)
	snId := d.Get("subnet_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointId, snId)

	err := finder.VpcEndpointSubnetAssociationExists(conn, endpointId, snId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Subnet Association (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint Subnet Association (%s): %w", id, err)
	}

	return nil
}

func resourceAwsVpcEndpointSubnetAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointId := d.Get("vpc_endpoint_id").(string)
	snId := d.Get("subnet_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointId, snId)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:   aws.String(endpointId),
		RemoveSubnetIds: aws.StringSlice([]string{snId}),
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint Subnet Association: %s", input)

	_, err := conn.ModifyVpcEndpoint(input)

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidVpcEndpointIdNotFound) || tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidSubnetIdNotFound) || tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidParameter) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPC Endpoint Subnet Association (%s): %w", id, err)
	}

	_, err = waiter.VpcEndpointAvailable(conn, endpointId, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint (%s) to become available: %w", endpointId, err)
	}

	return nil
}
