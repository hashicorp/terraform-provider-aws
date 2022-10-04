package events

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAPIDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceAPIDestinationCreate,
		Read:   resourceAPIDestinationRead,
		Update: resourceAPIDestinationUpdate,
		Delete: resourceAPIDestinationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+`), ""),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"invocation_endpoint": {
				Type:     schema.TypeString,
				Required: true,
			},
			"invocation_rate_limit_per_second": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Default:      300,
			},
			"http_method": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(eventbridge.ApiDestinationHttpMethod_Values(), true),
			},
			"connection_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAPIDestinationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	input := &eventbridge.CreateApiDestinationInput{}

	if name, ok := d.GetOk("name"); ok {
		input.Name = aws.String(name.(string))
	}
	if description, ok := d.GetOk("description"); ok {
		input.Description = aws.String(description.(string))
	}
	if invocationEndpoint, ok := d.GetOk("invocation_endpoint"); ok {
		input.InvocationEndpoint = aws.String(invocationEndpoint.(string))
	}
	if invocationRateLimitPerSecond, ok := d.GetOk("invocation_rate_limit_per_second"); ok {
		input.InvocationRateLimitPerSecond = aws.Int64(int64(invocationRateLimitPerSecond.(int)))
	}
	if httpMethod, ok := d.GetOk("http_method"); ok {
		input.HttpMethod = aws.String(httpMethod.(string))
	}
	if connectionArn, ok := d.GetOk("connection_arn"); ok {
		input.ConnectionArn = aws.String(connectionArn.(string))
	}

	_, err := conn.CreateApiDestination(input)
	if err != nil {
		return fmt.Errorf("Creating EventBridge API Destination (%s) failed: %w", aws.StringValue(input.Name), err)
	}

	d.SetId(aws.StringValue(input.Name))

	log.Printf("[INFO] EventBridge API Destination (%s) created", d.Id())

	return resourceAPIDestinationRead(d, meta)
}

func resourceAPIDestinationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	input := &eventbridge.DescribeApiDestinationInput{
		Name: aws.String(d.Id()),
	}

	output, err := conn.DescribeApiDestination(input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] EventBridge API Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading EventBridge API Destination (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.ApiDestinationArn)
	d.Set("name", output.Name)
	d.Set("description", output.Description)
	d.Set("invocation_endpoint", output.InvocationEndpoint)
	d.Set("invocation_rate_limit_per_second", output.InvocationRateLimitPerSecond)
	d.Set("http_method", output.HttpMethod)
	d.Set("connection_arn", output.ConnectionArn)

	return nil
}

func resourceAPIDestinationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	input := &eventbridge.UpdateApiDestinationInput{}

	if name, ok := d.GetOk("name"); ok {
		input.Name = aws.String(name.(string))
	}
	if description, ok := d.GetOk("description"); ok {
		input.Description = aws.String(description.(string))
	}
	if invocationEndpoint, ok := d.GetOk("invocation_endpoint"); ok {
		input.InvocationEndpoint = aws.String(invocationEndpoint.(string))
	}
	if invocationRateLimitPerSecond, ok := d.GetOk("invocation_rate_limit_per_second"); ok {
		input.InvocationRateLimitPerSecond = aws.Int64(int64(invocationRateLimitPerSecond.(int)))
	}
	if httpMethod, ok := d.GetOk("http_method"); ok {
		input.HttpMethod = aws.String(httpMethod.(string))
	}
	if connectionArn, ok := d.GetOk("connection_arn"); ok {
		input.ConnectionArn = aws.String(connectionArn.(string))
	}

	log.Printf("[DEBUG] Updating EventBridge API Destination: %s", input)
	_, err := conn.UpdateApiDestination(input)
	if err != nil {
		return fmt.Errorf("error updating EventBridge API Destination (%s): %w", d.Id(), err)
	}
	return resourceAPIDestinationRead(d, meta)
}

func resourceAPIDestinationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	log.Printf("[INFO] Deleting EventBridge API Destination (%s)", d.Id())
	input := &eventbridge.DeleteApiDestinationInput{
		Name: aws.String(d.Id()),
	}

	_, err := conn.DeleteApiDestination(input)

	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] EventBridge API Destination (%s) not found", d.Id())
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting EventBridge API Destination (%s): %w", d.Id(), err)
	}
	log.Printf("[INFO] EventBridge API Destination (%s) deleted", d.Id())

	return nil
}
