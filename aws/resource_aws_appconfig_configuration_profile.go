package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func resourceAwsAppconfigConfigurationProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppconfigConfigurationProfileCreate,
		Read:   resourceAwsAppconfigConfigurationProfileRead,
		Update: resourceAwsAppconfigConfigurationProfileUpdate,
		Delete: resourceAwsAppconfigConfigurationProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-z0-9]{4,7}`), ""),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_profile_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"location_uri": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"retrieval_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"validator": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							ValidateFunc: validation.Any(
								validation.StringIsJSON,
								validateArn,
							),
							DiffSuppressFunc: suppressEquivalentJsonDiffs,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appconfig.ValidatorType_Values(), false),
						},
					},
				},
			},
		},
		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsAppconfigConfigurationProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	appId := d.Get("application_id").(string)
	name := d.Get("name").(string)

	input := &appconfig.CreateConfigurationProfileInput{
		ApplicationId: aws.String(appId),
		LocationUri:   aws.String(d.Get("location_uri").(string)),
		Name:          aws.String(name),
		Tags:          tags.IgnoreAws().AppconfigTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retrieval_role_arn"); ok {
		input.RetrievalRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("validator"); ok && v.(*schema.Set).Len() > 0 {
		input.Validators = expandAppconfigValidators(v.(*schema.Set).List())
	}

	profile, err := conn.CreateConfigurationProfile(input)

	if err != nil {
		return fmt.Errorf("error creating AppConfig Configuration Profile (%s) for Application (%s): %w", name, appId, err)
	}

	if profile == nil {
		return fmt.Errorf("error creating AppConfig Configuration Profile (%s) for Application (%s): empty response", name, appId)
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(profile.Id), aws.StringValue(profile.ApplicationId)))

	return resourceAwsAppconfigConfigurationProfileRead(d, meta)
}

func resourceAwsAppconfigConfigurationProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	confProfID, appID, err := resourceAwsAppconfigConfigurationProfileParseID(d.Id())

	if err != nil {
		return err
	}

	input := &appconfig.GetConfigurationProfileInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
	}

	output, err := conn.GetConfigurationProfile(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] AppConfig Configuration Profile (%s) for Application (%s) not found, removing from state", confProfID, appID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppConfig Configuration Profile (%s) for Application (%s): %w", confProfID, appID, err)
	}

	if output == nil {
		return fmt.Errorf("error getting AppConfig Configuration Profile (%s) for Application (%s): empty response", confProfID, appID)
	}

	d.Set("application_id", output.ApplicationId)
	d.Set("configuration_profile_id", output.Id)
	d.Set("description", output.Description)
	d.Set("location_uri", output.LocationUri)
	d.Set("name", output.Name)

	d.Set("retrieval_role_arn", output.RetrievalRoleArn)

	if err := d.Set("validator", flattenAwsAppconfigValidators(output.Validators)); err != nil {
		return fmt.Errorf("error setting validator: %w", err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/configurationprofile/%s", appID, confProfID),
		Service:   "appconfig",
	}.String()

	d.Set("arn", arn)

	tags, err := keyvaluetags.AppconfigListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for AppConfig Configuration Profile (%s): %w", d.Id(), err)
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

func resourceAwsAppconfigConfigurationProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn

	if d.HasChangesExcept("tags", "tags_all") {
		confProfID, appID, err := resourceAwsAppconfigConfigurationProfileParseID(d.Id())

		if err != nil {
			return err
		}

		updateInput := &appconfig.UpdateConfigurationProfileInput{
			ApplicationId:          aws.String(appID),
			ConfigurationProfileId: aws.String(confProfID),
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

		if d.HasChange("validator") {
			updateInput.Validators = expandAppconfigValidators(d.Get("validator").(*schema.Set).List())
		}

		_, err = conn.UpdateConfigurationProfile(updateInput)

		if err != nil {
			return fmt.Errorf("error updating AppConfig Configuration Profile (%s) for Application (%s): %w", confProfID, appID, err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.AppconfigUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppConfig Configuration Profile (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceAwsAppconfigConfigurationProfileRead(d, meta)
}

func resourceAwsAppconfigConfigurationProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppConfigConn

	confProfID, appID, err := resourceAwsAppconfigConfigurationProfileParseID(d.Id())

	if err != nil {
		return err
	}

	input := &appconfig.DeleteConfigurationProfileInput{
		ApplicationId:          aws.String(appID),
		ConfigurationProfileId: aws.String(confProfID),
	}

	_, err = conn.DeleteConfigurationProfile(input)

	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting AppConfig Configuration Profile (%s) for Application (%s): %w", confProfID, appID, err)
	}

	return nil
}

func resourceAwsAppconfigConfigurationProfileParseID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected ConfigurationProfileID:ApplicationID", id)
	}

	return parts[0], parts[1], nil
}

func expandAppconfigValidator(tfMap map[string]interface{}) *appconfig.Validator {
	if tfMap == nil {
		return nil
	}

	validator := &appconfig.Validator{}

	// AppConfig API supports empty content
	if v, ok := tfMap["content"].(string); ok {
		validator.Content = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		validator.Type = aws.String(v)
	}

	return validator
}

func expandAppconfigValidators(tfList []interface{}) []*appconfig.Validator {
	// AppConfig API requires a 0 length slice instead of a nil value
	// when updating from N validators to 0/nil validators
	validators := make([]*appconfig.Validator, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		validator := expandAppconfigValidator(tfMap)

		if validator == nil {
			continue
		}

		validators = append(validators, validator)
	}

	return validators
}

func flattenAwsAppconfigValidator(validator *appconfig.Validator) map[string]interface{} {
	if validator == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := validator.Content; v != nil {
		tfMap["content"] = aws.StringValue(v)
	}

	if v := validator.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenAwsAppconfigValidators(validators []*appconfig.Validator) []interface{} {
	if len(validators) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, validator := range validators {
		if validator == nil {
			continue
		}

		tfList = append(tfList, flattenAwsAppconfigValidator(validator))
	}

	return tfList
}
