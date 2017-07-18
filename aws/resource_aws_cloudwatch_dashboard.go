package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCloudWatchDashboard() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAwsCloudWatchDashboardCreate,
		Read:          resourceAwsCloudWatchDashboardRead,
		Update:        resourceAwsCloudWatchDashboardUpdate,
		Delete:        resourceAwsCloudWatchDashboardDelete,
		SchemaVersion: 1,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"dashboard_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dashboard_body": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateJsonString,
				StateFunc: func(v interface{}) string {
					json, _ := normalizeJsonString(v)
					return json
				},
			},
			"dashboard_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsCloudWatchDashboardCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchconn

	params := getAwsCloudWatchPutDashboardInput(d)

	log.Printf("[DEBUG] Creating CloudWatch Dashboard: %#v", params)
	_, err := conn.PutDashboard(&params)
	if err != nil {
		return fmt.Errorf("Creating dashboard failed: %s", err)
	}
	d.SetId(d.Get("dashboard_name").(string))
	log.Println("[INFO] CloudWatch Dashboard created")

	return resourceAwsCloudWatchDashboardRead(d, meta)
}

func resourceAwsCloudWatchDashboardRead(d *schema.ResourceData, meta interface{}) error {
	a, err := getAwsCloudWatchDashboard(d, meta)
	if err != nil {
		return err
	}
	if a == nil {
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Reading CloudWatch Dashboard: %s", d.Get("dashboard_name"))

	d.Set("dashboard_name", a.DashboardName)
	d.Set("dashboard_body", a.DashboardBody)
	return nil
}

func resourceAwsCloudWatchDashboardUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchconn
	params := getAwsCloudWatchPutDashboardInput(d)

	log.Printf("[DEBUG] Updating CloudWatch Dashboard: %#v", params)
	_, err := conn.PutDashboard(&params)
	if err != nil {
		return fmt.Errorf("Updating dashboard failed: %s", err)
	}
	log.Println("[INFO] CloudWatch Dashboard updated")

	return resourceAwsCloudWatchDashboardRead(d, meta)
}

func resourceAwsCloudWatchDashboardDelete(d *schema.ResourceData, meta interface{}) error {
	p, err := getAwsCloudWatchDashboard(d, meta)
	if err != nil {
		return err
	}
	if p == nil {
		log.Printf("[DEBUG] CloudWatch Dashboard %s is already gone", d.Id())
		return nil
	}

	log.Printf("[INFO] Deleting CloudWatch Dashboard: %s", d.Id())

	conn := meta.(*AWSClient).cloudwatchconn
	params := cloudwatch.DeleteDashboardsInput{
		DashboardNames: []*string{aws.String(d.Id())},
	}

	if _, err := conn.DeleteDashboards(&params); err != nil {
		return fmt.Errorf("Error deleting CloudWatch Dashboard: %s", err)
	}
	log.Println("[INFO] CloudWatch Dashboard deleted")

	d.SetId("")
	return nil
}

func getAwsCloudWatchPutDashboardInput(d *schema.ResourceData) cloudwatch.PutDashboardInput {
	return cloudwatch.PutDashboardInput{
		DashboardBody: aws.String(d.Get("dashboard_body").(string)),
		DashboardName: aws.String(d.Get("dashboard_name").(string)),
	}
}

func getAwsCloudWatchDashboard(d *schema.ResourceData, meta interface{}) (*cloudwatch.GetDashboardOutput, error) {
	conn := meta.(*AWSClient).cloudwatchconn

	params := cloudwatch.GetDashboardInput{
		DashboardName: aws.String(d.Id()),
	}

	resp, err := conn.GetDashboard(&params)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
