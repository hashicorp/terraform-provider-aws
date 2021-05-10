package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAppconfigApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppconfigApplicationCreate,
		Read:   resourceAwsAppconfigApplicationRead,
		Update: resourceAwsAppconfigApplicationUpdate,
		Delete: resourceAwsAppconfigApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1024),
				),
			},
			"tags": tagsSchema(),
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppconfigApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn
	applicationName := d.Get("name").(string)
	applicationDescription := d.Get("description").(string)

	input := &appconfig.CreateApplicationInput{
		Name:        aws.String(applicationName),
		Description: aws.String(applicationDescription),
		Tags:        keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppconfigTags(),
	}

	app, err := conn.CreateApplication(input)
	if err != nil {
		return fmt.Errorf("Error creating AppConfig application: %s", err)
	}

	d.SetId(aws.StringValue(app.Id))

	return resourceAwsAppconfigApplicationRead(d, meta)
}

func resourceAwsAppconfigApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &appconfig.GetApplicationInput{
		ApplicationId: aws.String(d.Id()),
	}

	output, err := conn.GetApplication(input)

	if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Appconfig Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppConfig Application (%s): %s", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting AppConfig Application (%s): empty response", d.Id())
	}

	appARN := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("application/%s", aws.StringValue(output.Id)),
		Service:   "appconfig",
	}.String()

	d.Set("arn", appARN)
	d.Set("name", output.Name)
	d.Set("description", output.Description)

	tags, err := keyvaluetags.AppconfigListTags(conn, appARN)
	if err != nil {
		return fmt.Errorf("error getting tags for AppConfig Application (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAppconfigApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.AppconfigUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppConfig (%s) tags: %s", d.Id(), err)
		}
	}

	appDesc := d.Get("description").(string)
	appName := d.Get("name").(string)

	updateInput := &appconfig.UpdateApplicationInput{
		ApplicationId: aws.String(d.Id()),
		Description:   aws.String(appDesc),
		Name:          aws.String(appName),
	}

	if d.HasChange("description") {
		_, n := d.GetChange("description")
		updateInput.Description = aws.String(n.(string))
	}

	if d.HasChange("name") {
		_, n := d.GetChange("name")
		updateInput.Name = aws.String(n.(string))
	}

	_, err := conn.UpdateApplication(updateInput)
	if err != nil {
		return fmt.Errorf("error updating AppConfig Application(%s): %s", d.Id(), err)
	}

	return resourceAwsAppconfigApplicationRead(d, meta)
}

func resourceAwsAppconfigApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	input := &appconfig.DeleteApplicationInput{
		ApplicationId: aws.String(d.Id()),
	}

	_, err := conn.DeleteApplication(input)

	if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Appconfig Application (%s): %s", d.Id(), err)
	}

	return nil
}
