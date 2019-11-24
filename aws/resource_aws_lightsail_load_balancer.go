package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsLightsailLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLightsailLoadBalancerCreate,
		Read:   resourceAwsLightsailLoadBalancerRead,
		Update: resourceAwsLightsailLoadBalancerUpdate,
		Delete: resourceAwsLightsailLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+[^._\-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			"health_check_path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"instance_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ports": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func resourceAwsLightsailLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	req := lightsail.CreateLoadBalancerInput{
		HealthCheckPath:  aws.String(d.Get("health_check_path").(string)),
		InstancePort:     aws.Int64(int64(d.Get("instance_port").(int))),
		LoadBalancerName: aws.String(d.Get("name").(string)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		req.Tags = keyvaluetags.New(v).IgnoreAws().LightsailTags()
	}

	resp, err := conn.CreateLoadBalancer(&req)
	if err != nil {
		return err
	}

	if len(resp.Operations) == 0 {
		return fmt.Errorf("No operations found for CreateInstance request")
	}

	op := resp.Operations[0]
	d.SetId(d.Get("name").(string))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Started"},
		Target:     []string{"Completed", "Succeeded"},
		Refresh:    resourceAwsLightsailLoadBalancerOperationRefreshFunc(op.Id, meta),
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		// We don't return an error here because the Create call succeeded
		log.Printf("[ERR] Error waiting for load balancer (%s) to become ready: %s", d.Id(), err)
	}

	return resourceAwsLightsailLoadBalancerRead(d, meta)
}

func resourceAwsLightsailLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	resp, err := conn.GetLoadBalancer(&lightsail.GetLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				log.Printf("[WARN] Lightsail load balancer (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return err
		}
		return err
	}

	d.Set("arn", resp.LoadBalancer.Arn)
	d.Set("created_at", resp.LoadBalancer.CreatedAt.Format(time.RFC3339))
	d.Set("health_check_path", resp.LoadBalancer.HealthCheckPath)
	d.Set("instance_port", resp.LoadBalancer.InstancePort)
	d.Set("name", resp.LoadBalancer.Name)
	d.Set("protocol", resp.LoadBalancer.Protocol)
	d.Set("public_ports", resp.LoadBalancer.PublicPorts)

	if err := d.Set("tags", keyvaluetags.LightsailKeyValueTags(resp.LoadBalancer.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsLightsailLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	resp, err := conn.DeleteLoadBalancer(&lightsail.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
	})

	op := resp.Operations[0]

	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Started"},
		Target:     []string{"Completed", "Succeeded"},
		Refresh:    resourceAwsLightsailLoadBalancerOperationRefreshFunc(op.Id, meta),
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for load balancer (%s) to become destroyed: %s",
			d.Id(), err)
	}

	return err
}

func resourceAwsLightsailLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	if d.HasChange("health_check_path") {
		_, err := conn.UpdateLoadBalancerAttribute(&lightsail.UpdateLoadBalancerAttributeInput{
			AttributeName:    aws.String("HealthCheckPath"),
			AttributeValue:   aws.String(d.Get("health_check_path").(string)),
			LoadBalancerName: aws.String(d.Get("name").(string)),
		})
		d.SetPartial("health_check_path")
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.LightsailUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Instance (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsLightsailLoadBalancerRead(d, meta)
}

// method to check the status of an Operation, which is returned from
// Create/Delete methods.
// Status's are an aws.OperationStatus enum:
// - NotStarted
// - Started
// - Failed
// - Completed
// - Succeeded (not documented?)
func resourceAwsLightsailLoadBalancerOperationRefreshFunc(
	oid *string, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn := meta.(*AWSClient).lightsailconn
		log.Printf("[DEBUG] Checking if Lightsail Operation (%s) is Completed", *oid)
		o, err := conn.GetOperation(&lightsail.GetOperationInput{
			OperationId: oid,
		})
		if err != nil {
			return o, "FAILED", err
		}

		if o.Operation == nil {
			return nil, "Failed", fmt.Errorf("Error retrieving Operation info for operation (%s)", *oid)
		}

		log.Printf("[DEBUG] Lightsail Operation (%s) is currently %q", *oid, *o.Operation.Status)
		return o, *o.Operation.Status, nil
	}
}
