package elb

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAttachmentCreate,
		Read:   resourceAttachmentRead,
		Delete: resourceAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"elb": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"instance": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn
	elbName := d.Get("elb").(string)

	instance := d.Get("instance").(string)

	registerInstancesOpts := elb.RegisterInstancesWithLoadBalancerInput{
		LoadBalancerName: aws.String(elbName),
		Instances:        []*elb.Instance{{InstanceId: aws.String(instance)}},
	}

	log.Printf("[INFO] registering instance %s with ELB %s", instance, elbName)

	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.RegisterInstancesWithLoadBalancer(&registerInstancesOpts)

		if tfawserr.ErrCodeEquals(err, "InvalidTarget") {
			return resource.RetryableError(fmt.Errorf("Error attaching instance to ELB, retrying: %s", err))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.RegisterInstancesWithLoadBalancer(&registerInstancesOpts)
	}
	if err != nil {
		return fmt.Errorf("Failure registering instances with ELB: %s", err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", elbName)))

	return nil
}

func resourceAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn
	elbName := d.Get("elb").(string)

	// only add the instance that was previously defined for this resource
	expected := d.Get("instance").(string)

	// Retrieve the ELB properties to get a list of attachments
	describeElbOpts := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(elbName)},
	}

	resp, err := conn.DescribeLoadBalancers(describeElbOpts)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, elb.ErrCodeAccessPointNotFoundException) {
			log.Printf("[WARN] ELB Classic LB (%s) not found, removing from state", elbName)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error retrieving ELB Classic LB (%s): %w", elbName, err)
	}
	if !d.IsNewResource() && len(resp.LoadBalancerDescriptions) != 1 {
		log.Printf("[WARN] ELB Classic LB (%s) not found, removing from state", elbName)
		d.SetId("")
		return nil
	}

	// only set the instance Id that this resource manages
	found := false
	for _, i := range resp.LoadBalancerDescriptions[0].Instances {
		if expected == aws.StringValue(i.InstanceId) {
			d.Set("instance", expected)
			found = true
		}
	}

	if !d.IsNewResource() && !found {
		log.Printf("[WARN] instance %s not found in elb attachments", expected)
		d.SetId("")
	}

	return nil
}

func resourceAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBConn
	elbName := d.Get("elb").(string)

	instance := d.Get("instance").(string)

	log.Printf("[INFO] Deleting Attachment %s from: %s", instance, elbName)

	deRegisterInstancesOpts := elb.DeregisterInstancesFromLoadBalancerInput{
		LoadBalancerName: aws.String(elbName),
		Instances:        []*elb.Instance{{InstanceId: aws.String(instance)}},
	}

	_, err := conn.DeregisterInstancesFromLoadBalancer(&deRegisterInstancesOpts)
	if err != nil {
		return fmt.Errorf("Failure deregistering instances from ELB: %s", err)
	}

	return nil
}
