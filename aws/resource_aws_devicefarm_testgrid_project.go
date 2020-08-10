package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsDevicefarmTestgridProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDevicefarmTestgridProjectCreate,
		Read:   resourceAwsDevicefarmTestgridProjectRead,
		Update: resourceAwsDevicefarmTestgridProjectUpdate,
		Delete: resourceAwsDevicefarmTestgridProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsDevicefarmTestgridProjectCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).devicefarmconn
	region := meta.(*AWSClient).region

	//	We need to ensure that DeviceFarm is only being run against us-west-2
	//	As this is the only place that AWS currently supports it
	if region != "us-west-2" {
		return fmt.Errorf("DeviceFarm can only be used with us-west-2. You are trying to use it on %s", region)
	}

	input := &devicefarm.CreateTestGridProjectInput{
		Name: aws.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Creating DeviceFarm TestGrid Project: %s", d.Get("name").(string))
	out, err := conn.CreateTestGridProject(input)
	if err != nil {
		return fmt.Errorf("Error creating DeviceFarm TestGrid Project: %s", err)
	}

	log.Printf("[DEBUG] Successsfully Created DeviceFarm TestGrid Project: %s", *out.TestGridProject.Arn)
	d.SetId(*out.TestGridProject.Arn)

	return resourceAwsDevicefarmTestgridProjectRead(d, meta)
}

func resourceAwsDevicefarmTestgridProjectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).devicefarmconn

	input := &devicefarm.GetTestGridProjectInput{
		ProjectArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DeviceFarm TestGrid Project: %s", d.Id())
	out, err := conn.GetTestGridProject(input)
	if err != nil {
		if isAWSErr(err, devicefarm.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] DeviceFarm TestGrid Project (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading DeviceFarm TestGrid Project: %s", err)
	}

	d.Set("name", out.TestGridProject.Name)
	d.Set("arn", out.TestGridProject.Arn)

	return nil
}

func resourceAwsDevicefarmTestgridProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).devicefarmconn

	if d.HasChange("name") {
		input := &devicefarm.UpdateTestGridProjectInput{
			ProjectArn: aws.String(d.Id()),
			Name:       aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating DeviceFarm TestGrid Project: %s", d.Id())
		_, err := conn.UpdateTestGridProject(input)
		if err != nil {
			return fmt.Errorf("Error Updating DeviceFarm TestGrid Project: %s", err)
		}

	}

	return resourceAwsDevicefarmTestgridProjectRead(d, meta)
}

func resourceAwsDevicefarmTestgridProjectDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).devicefarmconn

	input := &devicefarm.DeleteTestGridProjectInput{
		ProjectArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm TestGrid Project: %s", d.Id())
	_, err := conn.DeleteTestGridProject(input)
	if err != nil {
		return fmt.Errorf("Error deleting DeviceFarm TestGrid Project: %s", err)
	}

	return nil
}
