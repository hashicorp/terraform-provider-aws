package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"

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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
	if isAWSErr(err, cloudwatchevents.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] CloudWatch Event Bus (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading CloudWatch Event Bus: %s", err)
	}

	log.Printf("[DEBUG] Found CloudWatch Event bus: %#v", *output)

	d.Set("arn", output.Arn)
	d.Set("name", output.Name)

	return nil
}

func resourceAwsCloudWatchEventBusDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	log.Printf("[INFO] Deleting CloudWatch Event Bus: %v", d.Id())
	_, err := conn.DeleteEventBus(&cloudwatchevents.DeleteEventBusInput{
		Name: aws.String(d.Id()),
	})
	if isAWSErr(err, cloudwatchevents.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] CloudWatch Event Bus (%s) not found", d.Id())
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting CloudWatch Event Bus %v: %v", d.Id(), err)
	}
	log.Printf("[INFO] CloudWatch Event Bus %v deleted", d.Id())

	return nil
}
