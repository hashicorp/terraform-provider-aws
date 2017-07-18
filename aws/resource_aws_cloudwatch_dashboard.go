package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCloudWatchDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchDashboardCreate,
		Read:   resourceAwsCloudWatchDashboardRead,
		Update: resourceAwsCloudWatchDashboardUpdate,
		Delete: resourceAwsCloudWatchDashboardDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		// Note that we specify both the `dashboard_body` and
		// the `dashboard_name` as being required, even though
		// according to the REST API documentation both are
		// optional: http://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_PutDashboard.html#API_PutDashboard_RequestParameters
		Schema: map[string]*schema.Schema{
			"dashboard_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dashboard_body": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateJsonString,
				StateFunc: func(v interface{}) string {
					json, _ := normalizeJsonString(v)
					return json
				},
			},
			"dashboard_name": {
				Type:     schema.TypeString,
				Required: true,
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
	dashboardName := d.Get("dashboard_name").(string)
	log.Printf("[DEBUG] Reading CloudWatch Dashboard: %s", dashboardName)
	conn := meta.(*AWSClient).cloudwatchconn

	params := cloudwatch.GetDashboardInput{
		DashboardName: aws.String(d.Id()),
	}

	resp, err := conn.GetDashboard(&params)
	if err != nil {
		if isResourceNotFoundErr(err) {
			log.Printf("[WARN] CloudWatch Dashboard %q not found, removing", dashboardName)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading dashboard failed: %s", err)
	}

	d.Set("dashboard_name", resp.DashboardName)
	d.Set("dashboard_body", resp.DashboardBody)
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
	log.Println("[INFO] Deleting CloudWatch Dashboard")
	conn := meta.(*AWSClient).cloudwatchconn
	params := cloudwatch.DeleteDashboardsInput{
		DashboardNames: []*string{aws.String(d.Id())},
	}

	if _, err := conn.DeleteDashboards(&params); err != nil {
		if isResourceNotFoundErr(err) {
			log.Printf("[WARN] CloudWatch Dashboard %s is already gone", d.Id())
			return nil
		}
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

func isResourceNotFoundErr(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "ResourceNotFoundException"
}
