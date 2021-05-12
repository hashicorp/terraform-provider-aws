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

func resourceAwsAppconfigConfigurationProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppconfigConfigurationProfileCreate,
		Read:   resourceAwsAppconfigConfigurationProfileRead,
		Update: resourceAwsAppconfigConfigurationProfileUpdate,
		Delete: resourceAwsAppconfigConfigurationProfileDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppconfigConfigurationProfileImport,
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
			"location_uri": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 2048),
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1024),
				),
			},
			"retrieval_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(20, 2048),
				),
			},
			"validators": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(0, 32768),
							),
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"JSON_SCHEMA", "LAMBDA",
							}, false),
						},
					},
				},
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppconfigConfigurationProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	input := &appconfig.CreateConfigurationProfileInput{
		Name:             aws.String(d.Get("name").(string)),
		Description:      aws.String(d.Get("description").(string)),
		LocationUri:      aws.String(d.Get("location_uri").(string)),
		RetrievalRoleArn: aws.String(d.Get("retrieval_role_arn").(string)),
		ApplicationId:    aws.String(d.Get("application_id").(string)),
		Validators:       expandAppconfigValidators(d.Get("validators").([]interface{})),
		Tags:             keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppconfigTags(),
	}

	profile, err := conn.CreateConfigurationProfile(input)
	if err != nil {
		return fmt.Errorf("Error creating AppConfig ConfigurationProfile: %s", err)
	}

	d.SetId(aws.StringValue(profile.Id))

	return resourceAwsAppconfigConfigurationProfileRead(d, meta)
}

func expandAppconfigValidators(list []interface{}) []*appconfig.Validator {
	validators := make([]*appconfig.Validator, len(list))
	for i, validatorInterface := range list {
		m := validatorInterface.(map[string]interface{})
		validators[i] = &appconfig.Validator{
			Content: aws.String(m["content"].(string)),
			Type:    aws.String(m["type"].(string)),
		}
	}
	return validators
}

func flattenAwsAppconfigValidators(validators []*appconfig.Validator) []interface{} {
	list := make([]interface{}, len(validators))
	for i, validator := range validators {
		list[i] = map[string]interface{}{
			"content": aws.StringValue(validator.Content),
			"type":    aws.StringValue(validator.Type),
		}
	}
	return list
}

func resourceAwsAppconfigConfigurationProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	appID := d.Get("application_id").(string)

	input := &appconfig.GetConfigurationProfileInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(d.Id()),
	}

	output, err := conn.GetConfigurationProfile(input)

	if !d.IsNewResource() && isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Appconfig ConfigurationProfile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppConfig ConfigurationProfile (%s): %s", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting AppConfig ConfigurationProfile (%s): empty response", d.Id())
	}

	d.Set("name", output.Name)
	d.Set("description", output.Description)
	d.Set("application_id", output.ApplicationId)
	d.Set("location_uri", output.LocationUri)
	d.Set("retrieval_role_arn", output.RetrievalRoleArn)
	d.Set("validators", flattenAwsAppconfigValidators(output.Validators))

	profileARN := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("application/%s/configurationprofile/%s", appID, d.Id()),
		Service:   "appconfig",
	}.String()
	d.Set("arn", profileARN)

	tags, err := keyvaluetags.AppconfigListTags(conn, profileARN)
	if err != nil {
		return fmt.Errorf("error getting tags for AppConfig ConfigurationProfile (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAppconfigConfigurationProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	updateInput := &appconfig.UpdateConfigurationProfileInput{
		ConfigurationProfileId: aws.String(d.Id()),
		ApplicationId:          aws.String(d.Get("application_id").(string)),
	}

	if d.HasChange("description") {
		updateInput.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("name") {
		updateInput.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("retrieval_role_arn") {
		updateInput.RetrievalRoleArn = aws.String(d.Get("retrieval_role_arn").(string))
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.AppconfigUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppConfig (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("validators") {
		updateInput.Validators = expandAppconfigValidators(d.Get("validators").([]interface{}))
	}

	_, err := conn.UpdateConfigurationProfile(updateInput)
	if err != nil {
		return fmt.Errorf("error updating AppConfig ConfigurationProfile (%s): %s", d.Id(), err)
	}

	return resourceAwsAppconfigConfigurationProfileRead(d, meta)
}

func resourceAwsAppconfigConfigurationProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appconfigconn

	input := &appconfig.DeleteConfigurationProfileInput{
		ConfigurationProfileId: aws.String(d.Id()),
		ApplicationId:          aws.String(d.Get("application_id").(string)),
	}

	_, err := conn.DeleteConfigurationProfile(input)

	if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Appconfig ConfigurationProfile (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsAppconfigConfigurationProfileImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'application-id/configurationprofile-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("application_id", parts[0])

	return []*schema.ResourceData{d}, nil
}
