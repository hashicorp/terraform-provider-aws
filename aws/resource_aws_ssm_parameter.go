package aws

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	// Maximum amount of time to wait for asynchronous validation on SSM Parameter creation.
	ssmParameterCreationValidationTimeout = 2 * time.Minute
)

func resourceAwsSsmParameter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmParameterPut,
		Read:   resourceAwsSsmParameterRead,
		Update: resourceAwsSsmParameterPut,
		Delete: resourceAwsSsmParameterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"tier": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ssm.ParameterTierStandard,
				ValidateFunc: validation.StringInSlice(ssm.ParameterTier_Values(), false),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(ssm.ParameterType_Values(), false),
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"data_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"aws:ec2:image",
					"text",
				}, false),
			},
			"overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"allowed_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"policies": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.Any(
					validation.StringIsJSON,
					validation.StringLenBetween(1, 4096),
				),
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"tags": tagsSchema(),
		},

		CustomizeDiff: customdiff.All(
			// Prevent the following error during tier update from Advanced to Standard:
			// ValidationException: This parameter uses the advanced-parameter tier. You can't downgrade a parameter from the advanced-parameter tier to the standard-parameter tier. If necessary, you can delete the advanced parameter and recreate it as a standard parameter.
			customdiff.ForceNewIfChange("tier", func(_ context.Context, old, new, meta interface{}) bool {
				return old.(string) == ssm.ParameterTierAdvanced && new.(string) == ssm.ParameterTierStandard
			}),
		),
	}
}

func resourceAwsSsmParameterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading SSM Parameter: %s", d.Id())

	input := &ssm.GetParameterInput{
		Name:           aws.String(d.Id()),
		WithDecryption: aws.Bool(true),
	}

	var resp *ssm.GetParameterOutput
	err := resource.Retry(ssmParameterCreationValidationTimeout, func() *resource.RetryError {
		var err error
		resp, err = conn.GetParameter(input)

		if isAWSErr(err, ssm.ErrCodeParameterNotFound, "") && d.IsNewResource() && d.Get("data_type").(string) == "aws:ec2:image" {
			return resource.RetryableError(fmt.Errorf("error reading SSM Parameter (%s) after creation: this can indicate that the provided parameter value could not be validated by SSM", d.Id()))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		resp, err = conn.GetParameter(input)
	}

	if isAWSErr(err, ssm.ErrCodeParameterNotFound, "") && !d.IsNewResource() {
		log.Printf("[WARN] SSM Parameter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SSM Parameter (%s): %w", d.Id(), err)
	}

	param := resp.Parameter
	name := aws.StringValue(param.Name)
	d.Set("name", name)
	d.Set("type", param.Type)
	d.Set("value", param.Value)
	d.Set("version", param.Version)

	describeParamsInput := &ssm.DescribeParametersInput{
		ParameterFilters: []*ssm.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Option: aws.String("Equals"),
				Values: []*string{aws.String(name)},
			},
		},
	}
	describeResp, err := conn.DescribeParameters(describeParamsInput)
	if err != nil {
		return fmt.Errorf("error describing SSM parameter: %w", err)
	}

	if describeResp == nil || len(describeResp.Parameters) == 0 || describeResp.Parameters[0] == nil {
		log.Printf("[WARN] SSM Parameter %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	detail := describeResp.Parameters[0]
	d.Set("key_id", detail.KeyId)
	d.Set("description", detail.Description)
	d.Set("tier", ssm.ParameterTierStandard)
	if detail.Tier != nil {
		d.Set("tier", detail.Tier)
	}
	d.Set("allowed_pattern", detail.AllowedPattern)
	d.Set("data_type", detail.DataType)
	d.Set("policies", flattenAwsSsmParameterPolicies(detail.Policies))

	tags, err := keyvaluetags.SsmListTags(conn, name, ssm.ResourceTypeForTaggingParameter)

	if err != nil {
		return fmt.Errorf("error listing tags for SSM Parameter (%s): %w", name, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("arn", param.ARN)

	return nil
}

func resourceAwsSsmParameterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Deleting SSM Parameter: %s", d.Id())

	_, err := conn.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return fmt.Errorf("error deleting SSM Parameter (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsSsmParameterPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	ssmParamName := d.Get("name").(string)
	log.Printf("[INFO] Creating SSM Parameter: %s", ssmParamName)

	paramInput := &ssm.PutParameterInput{
		Name:           aws.String(ssmParamName),
		Type:           aws.String(d.Get("type").(string)),
		Tier:           aws.String(d.Get("tier").(string)),
		Value:          aws.String(d.Get("value").(string)),
		Overwrite:      aws.Bool(shouldUpdateSsmParameter(d)),
		AllowedPattern: aws.String(d.Get("allowed_pattern").(string)),
	}

	if v, ok := d.GetOk("data_type"); ok {
		paramInput.DataType = aws.String(v.(string))
	}

	if d.HasChange("description") {
		_, n := d.GetChange("description")
		paramInput.Description = aws.String(n.(string))
	}

	if v, ok := d.GetOk("policies"); ok {
		paramInput.Policies = aws.String(v.(string))
	}

	if keyID, ok := d.GetOk("key_id"); ok && d.Get("type").(string) == ssm.ParameterTypeSecureString {
		paramInput.SetKeyId(keyID.(string))
	}

	log.Printf("[DEBUG] Waiting for SSM Parameter %v to be updated", ssmParamName)
	_, err := conn.PutParameter(paramInput)

	if isAWSErr(err, "ValidationException", "Tier is not supported") {
		paramInput.Tier = nil
		_, err = conn.PutParameter(paramInput)
	}

	if err != nil {
		return fmt.Errorf("error creating SSM parameter: %w", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SsmUpdateTags(conn, ssmParamName, ssm.ResourceTypeForTaggingParameter, o, n); err != nil {
			return fmt.Errorf("error updating SSM Parameter (%s) tags: %w", ssmParamName, err)
		}
	}

	d.SetId(ssmParamName)

	return resourceAwsSsmParameterRead(d, meta)
}

func shouldUpdateSsmParameter(d *schema.ResourceData) bool {
	// If the user has specified a preference, return their preference
	if value, ok := d.GetOkExists("overwrite"); ok {
		return value.(bool)
	}

	// Since the user has not specified a preference, obey lifecycle rules
	// if it is not a new resource, otherwise overwrite should be set to false.
	return !d.IsNewResource()
}

func flattenAwsSsmParameterPolicies(policies []*ssm.ParameterInlinePolicy) string {
	if len(policies) == 0 {
		return ""
	}

	strPolicies := make([]string, 0)

	for _, policy := range policies {
		strPolicies = append(strPolicies, aws.StringValue(policy.PolicyText))
	}

	return fmt.Sprintf("[%s]", strings.Join(strPolicies, ","))
}
