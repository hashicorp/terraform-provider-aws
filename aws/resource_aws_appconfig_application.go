package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsAppconfigApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	applicationName := d.Get("name").(string)

	input := &appconfig.CreateApplicationInput{
		Name: aws.String(applicationName),
		Tags: tags.IgnoreAws().AppconfigTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	app, err := conn.CreateApplication(input)

	if err != nil {
		return fmt.Errorf("error creating AppConfig Application (%s): %w", applicationName, err)
	}

	if app == nil {
		return fmt.Errorf("error creating AppConfig Application (%s): empty response", applicationName)
	}

	d.SetId(aws.StringValue(app.Id))

	return resourceAwsAppconfigApplicationRead(d, meta)
}

func resourceAwsAppconfigApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &appconfig.GetApplicationInput{
		ApplicationId: aws.String(d.Id()),
	}

	output, err := conn.GetApplication(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Appconfig Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppConfig Application (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting AppConfig Application (%s): empty response", d.Id())
	}

	arn := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("application/%s", aws.StringValue(output.Id)),
		Service:   "appconfig",
	}.String()

	d.Set("arn", arn)
	d.Set("name", output.Name)
	d.Set("description", output.Description)

	tags, err := keyvaluetags.AppconfigListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for AppConfig Application (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsAppconfigApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	if d.HasChangesExcept("tags", "tags_all") {

		updateInput := &appconfig.UpdateApplicationInput{
			ApplicationId: aws.String(d.Id()),
		}

		if d.HasChange("description") {
			updateInput.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			updateInput.Name = aws.String(d.Get("name").(string))
		}

		_, err := conn.UpdateApplication(updateInput)

		if err != nil {
			return fmt.Errorf("error updating AppConfig Application(%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.AppconfigUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppConfig Application (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceAwsAppconfigApplicationRead(d, meta)
}

func resourceAwsAppconfigApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	input := &appconfig.DeleteApplicationInput{
		ApplicationId: aws.String(d.Id()),
	}

	_, err := conn.DeleteApplication(input)

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Appconfig Application (%s): %w", d.Id(), err)
	}

	return nil
}
