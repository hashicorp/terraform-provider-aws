package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAppconfigEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppconfigEnvironmentCreate,
		Read:   resourceAwsAppconfigEnvironmentRead,
		Update: resourceAwsAppconfigEnvironmentUpdate,
		Delete: resourceAwsAppconfigEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppconfigEnvironmentImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
				),
			},
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(4, 7),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1024),
				),
			},
			// TODO monitors
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppconfigEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	input := &appconfig.CreateEnvironmentInput{
		Name:          aws.String(d.Get("name").(string)),
		Description:   aws.String(d.Get("description").(string)),
		ApplicationId: aws.String(d.Get("application_id").(string)),
		Tags:          keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppconfigTags(),
		// TODO monitors
	}

	environment, err := conn.CreateEnvironment(input)
	if err != nil {
		return fmt.Errorf("Error creating AppConfig Environment: %s", err)
	}

	d.SetId(aws.StringValue(environment.Id))

	return resourceAwsAppconfigEnvironmentRead(d, meta)
}

func resourceAwsAppconfigEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	appID := d.Get("application_id").(string)

	input := &appconfig.GetEnvironmentInput{
		ApplicationId: aws.String(appID),
		EnvironmentId: aws.String(d.Id()),
	}

	output, err := conn.GetEnvironment(input)

	if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Appconfig Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppConfig Environment (%s): %s", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting AppConfig Environment (%s): empty response", d.Id())
	}

	d.Set("name", output.Name)
	d.Set("description", output.Description)
	d.Set("application_id", output.ApplicationId)
	// TODO monitors

	environmentARN := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("application/%s/environment/%s", appID, d.Id()),
		Service:   "appconfig",
	}.String()
	d.Set("arn", environmentARN)

	tags, err := keyvaluetags.AppconfigListTags(conn, environmentARN)
	if err != nil {
		return fmt.Errorf("error getting tags for AppConfig Environment (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAppconfigEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	updateInput := &appconfig.UpdateEnvironmentInput{
		EnvironmentId: aws.String(d.Id()),
		ApplicationId: aws.String(d.Get("application_id").(string)),
	}

	if d.HasChange("description") {
		_, n := d.GetChange("description")
		updateInput.Description = aws.String(n.(string))
	}

	if d.HasChange("name") {
		_, n := d.GetChange("name")
		updateInput.Name = aws.String(n.(string))
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.AppconfigUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppConfig (%s) tags: %s", d.Id(), err)
		}
	}
	// TODO monitors

	_, err := conn.UpdateEnvironment(updateInput)
	if err != nil {
		return fmt.Errorf("error updating AppConfig Environment(%s): %s", d.Id(), err)
	}

	return resourceAwsAppconfigEnvironmentRead(d, meta)
}

func resourceAwsAppconfigEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	input := &appconfig.DeleteEnvironmentInput{
		EnvironmentId: aws.String(d.Id()),
		ApplicationId: aws.String(d.Get("application_id").(string)),
	}

	_, err := conn.DeleteEnvironment(input)

	if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Appconfig Environment (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsAppconfigEnvironmentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'application-id/environment-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("application_id", parts[0])

	return []*schema.ResourceData{d}, nil
}
