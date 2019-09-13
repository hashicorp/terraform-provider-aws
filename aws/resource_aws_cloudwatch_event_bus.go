package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
)

func resourceAwsCloudWatchEventBus() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchEventBusCreate,
		Read:   resourceAwsCloudWatchEventBusRead,
		Delete: resourceAwsCloudWatchEventBusDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCloudWatchEventBusName,
			},
			"event_source_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateCloudWatchEventSourceName,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v.(string))
					return json
				},
			},
		},
	}
}

func resourceAwsCloudWatchEventBusCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	eventBusName := d.Get("name").(string)
	params := &cloudwatchevents.CreateEventBusInput{
		Name: aws.String(eventBusName),
	}

	if v, ok := d.GetOk("event_source_name"); ok {
		if eventBusName != v.(string) {
			return fmt.Errorf("Event bus name %v must match event source name %v", eventBusName, v)
		}
		params.EventSourceName = aws.String(v.(string))
	} else if strings.HasPrefix(eventBusName, "aws.") {
		return fmt.Errorf("EventBus name starting with 'aws.' is not valid: %v", eventBusName)
	}

	log.Printf("[DEBUG] Creating CloudWatch Event Bus: %v", params)

	_, err := conn.CreateEventBus(params)
	if err != nil {
		return fmt.Errorf("Creating CloudWatch Event Bus %v failed: %v", eventBusName, err)
	}

	d.SetId(eventBusName)

	log.Printf("[INFO] CloudWatch Event Bus %v created", d.Id())

	return resourceAwsCloudWatchEventBusRead(d, meta)
}

func resourceAwsCloudWatchEventBusRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	log.Printf("[DEBUG] Reading CloudWatch Event Bus: %v", d.Id())

	input := &cloudwatchevents.DescribeEventBusInput{
		Name: aws.String(d.Id()),
	}

	output, err := conn.DescribeEventBus(input)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Found CloudWatch Event bus: %#v", *output)

	d.Set("arn", output.Arn)
	d.Set("name", output.Name)
	d.Set("policy", output.Policy)

	return nil
}

func resourceAwsCloudWatchEventBusDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	log.Printf("[INFO] Deleting CloudWatch Event Bus: %v", d.Id())
	_, err := conn.DeleteEventBus(&cloudwatchevents.DeleteEventBusInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting CloudWatch Event Bus %v: %v", d.Id(), err)
	}
	log.Printf("[INFO] CloudWatch Event Bus %v deleted", d.Id())

	return nil
}
