package aws

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"reflect"
	"regexp"
)

func resourceAwsCloudWatchDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchDashboardCreate,
		Read:   resourceAwsCloudWatchDashboardRead,
		Update: resourceAwsCloudWatchDashboardUpdate,
		Delete: resourceAwsCloudWatchDashboardDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsCloudWatchDashboardName,
			},
			"body": &schema.Schema{
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         false,
				ValidateFunc:     validateJsonString,
				DiffSuppressFunc: suppressSameJson,
			},
			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCloudWatchDashboardCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsCloudWatchDashboardUpdate(d, meta)
}

func resourceAwsCloudWatchDashboardUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchconn

	dashboardName := d.Get("name").(string)

	putDashboardInput := cloudwatch.PutDashboardInput{
		DashboardName: aws.String(dashboardName),
		DashboardBody: aws.String(d.Get("body").(string)),
	}

	log.Printf("[DEBUG] Creating Cloudwatch Dashboard: %s", dashboardName)

	_, err := conn.PutDashboard(&putDashboardInput)
	if err != nil {
		return fmt.Errorf("Failed to create dashboard: %s '%s'", dashboardName, err)
	}

	d.SetId(dashboardName)

	log.Printf("[INFO] Cloudwatch dashboard created: %s", dashboardName)

	return resourceAwsCloudWatchDashboardRead(d, meta)
}

func resourceAwsCloudWatchDashboardRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchconn

	getDashboardInput := cloudwatch.GetDashboardInput{
		DashboardName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading dashboard %s", getDashboardInput.DashboardName)

	dashboardOutput, err := conn.GetDashboard(&getDashboardInput)
	if err != nil {
		if err.(awserr.Error).Code() == cloudwatch.ErrCodeDashboardNotFoundError {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieveing dashboard '%s': %s", *getDashboardInput.DashboardName, err)
	}

	d.Set("arn", dashboardOutput.DashboardArn)
	d.Set("body", dashboardOutput.DashboardBody)
	d.Set("arn", dashboardOutput.DashboardArn)

	log.Printf("[INFO] Retrieved dashboard %s", getDashboardInput.DashboardName)

	return nil
}

func resourceAwsCloudWatchDashboardDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchconn

	dashboardName := d.Get("name").(string)
	deleteDashboardInput := cloudwatch.DeleteDashboardsInput{
		DashboardNames: []*string{&dashboardName},
	}

	log.Printf("[DEBUG] Deleting dashboard %s", dashboardName)

	_, err := conn.DeleteDashboards(&deleteDashboardInput)
	if err != nil {
		if err.(awserr.Error).Code() == cloudwatch.ErrCodeDashboardNotFoundError {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error deleting dashboard %s: %s", dashboardName, err)
	}

	log.Printf("[INFO] Deleted dashboard: %s", dashboardName)

	return nil
}

func validateAwsCloudWatchDashboardName(v interface{}, k string) (ws []string, errors []error) {
	name := v.(string)

	if len(name) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 255 chars: %q", k, name,
		))
	}

	pattern := `^[A-Za-z0-9\.\-_]+$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		errors = append(errors, fmt.Errorf(
			"%q has invalid chars (%q): %q", k, pattern, name,
		))
	}

	return
}

func suppressSameJson(k, old, new string, d *schema.ResourceData) bool {
	var obj1 interface{}
	var obj2 interface{}

	err := json.Unmarshal([]byte(old), &obj1)
	if err != nil {
		return false
	}

	err = json.Unmarshal([]byte(new), &obj2)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(obj1, obj2)

}
