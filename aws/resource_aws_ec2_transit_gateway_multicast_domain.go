package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TransitGatewayMulticastDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayMulticastDomainCreate,
		Read:   resourceAwsEc2TransitGatewayMulticastDomainRead,
		Update: resourceAwsEc2TransitGatewayMulticastDomainUpdate,
		Delete: resourceAwsEc2TransitGatewayMulticastDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"transit_gateway_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEc2TransitGatewayMulticastDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.CreateTransitGatewayMulticastDomainInput{
		TransitGatewayId: aws.String(d.Get("transit_gateway_id").(string)),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}),
			ec2.ResourceTypeTransitGatewayMulticastDomain),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Multicast Domain: %s", input)
	output, err := conn.CreateTransitGatewayMulticastDomain(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Multicast Domain: %s", err)
	}

	id := aws.StringValue(output.TransitGatewayMulticastDomain.TransitGatewayMulticastDomainId)
	d.SetId(id)

	if err := waitForEc2TransitGatewayMulticastDomainCreation(conn, id); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Multicast Domain (%s) availability: %s", id, err)
	}

	return resourceAwsEc2TransitGatewayMulticastDomainRead(d, meta)
}

func resourceAwsEc2TransitGatewayMulticastDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	id := d.Id()
	multicastDomain, err := ec2DescribeTransitGatewayMulticastDomain(conn, id)

	if isAWSErr(err, "InvalidTransitGatewayMulticastDomainId.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Multicast Domain: %s", err)
	}

	if multicastDomain == nil {
		log.Printf("[WARN] EC2 Transit Gateway Multicast Domain (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if aws.StringValue(multicastDomain.State) == ec2.TransitGatewayStateDeleting || aws.StringValue(multicastDomain.State) == ec2.TransitGatewayStateDeleted {
		log.Printf("[WARN] EC2 Transit Gateway (%s) in deleted state (%s), removing from state", d.Id(),
			aws.StringValue(multicastDomain.State))
		d.SetId("")
		return nil
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(multicastDomain.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error settings tags: %s", err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		id := d.Id()
		if err := keyvaluetags.Ec2UpdateTags(conn, id, o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Multicast Domain (%s) tags: %s", id, err)
		}
	}

	return nil
}

func resourceAwsEc2TransitGatewayMulticastDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	id := d.Id()
	input := &ec2.DeleteTransitGatewayMulticastDomainInput{
		TransitGatewayMulticastDomainId: aws.String(id),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Multicast Domain (%s): %s", id, input)
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteTransitGatewayMulticastDomain(input)

		// TODO: error handling?

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteTransitGatewayMulticastDomain(input)
	}

	if isAWSErr(err, "InvalidTransitGatewayMulticastDomainId.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Multicast Domain: %s", err)
	}

	if err := waitForEc2TransitGatewayMulticastDomainDeletion(conn, d.Get("transit_gateway_id").(string), id); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Multicast Domain (%s) deletion: %s", id, err)
	}

	return nil
}
