package ssm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for asynchronous validation on SSM Parameter creation.
	parameterCreationValidationTimeout = 2 * time.Minute
)

func ResourceParameter() *schema.Resource {
	return &schema.Resource{
		Create: resourceParameterCreate,
		Read:   resourceParameterRead,
		Update: resourceParameterUpdate,
		Delete: resourceParameterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"tier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ssm.ParameterTier_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old != "" {
						return new == ssm.ParameterTierIntelligentTiering
					}
					return false
				},
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
				Optional: true,
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
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			// Prevent the following error during tier update from Advanced to Standard:
			// ValidationException: This parameter uses the advanced-parameter tier. You can't downgrade a parameter from the advanced-parameter tier to the standard-parameter tier. If necessary, you can delete the advanced parameter and recreate it as a standard parameter.
			customdiff.ForceNewIfChange("tier", func(_ context.Context, old, new, meta interface{}) bool {
				return old.(string) == ssm.ParameterTierAdvanced && new.(string) == ssm.ParameterTierStandard
			}),
			customdiff.ComputedIf("version", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("value")
			}),
			verify.SetTagsDiff,
		),
	}
}

func resourceParameterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	paramInput := &ssm.PutParameterInput{
		Name:           aws.String(name),
		Type:           aws.String(d.Get("type").(string)),
		Value:          aws.String(d.Get("value").(string)),
		AllowedPattern: aws.String(d.Get("allowed_pattern").(string)),
	}

	if v, ok := d.GetOk("tier"); ok {
		paramInput.Tier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_type"); ok {
		paramInput.DataType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		paramInput.Description = aws.String(v.(string))
	}

	if keyID, ok := d.GetOk("key_id"); ok && d.Get("type").(string) == ssm.ParameterTypeSecureString {
		paramInput.SetKeyId(keyID.(string))
	}

	if len(tags) > 0 {
		paramInput.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.PutParameter(paramInput)

	if tfawserr.ErrMessageContains(err, "ValidationException", "Tier is not supported") {
		log.Printf("[WARN] Creating SSM Parameter (%s): tier %q not supported, using default", name, d.Get("tier").(string))
		paramInput.Tier = nil
		_, err = conn.PutParameter(paramInput)
	}

	if err != nil {
		return fmt.Errorf("error creating SSM Parameter (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceParameterRead(d, meta)
}

func resourceParameterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ssm.GetParameterInput{
		Name:           aws.String(d.Id()),
		WithDecryption: aws.Bool(true),
	}

	var resp *ssm.GetParameterOutput
	err := resource.Retry(parameterCreationValidationTimeout, func() *resource.RetryError {
		var err error
		resp, err = conn.GetParameter(input)

		if tfawserr.ErrCodeEquals(err, ssm.ErrCodeParameterNotFound) && d.IsNewResource() && d.Get("data_type").(string) == "aws:ec2:image" {
			return resource.RetryableError(fmt.Errorf("error reading SSM Parameter (%s) after creation: this can indicate that the provided parameter value could not be validated by SSM", d.Id()))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.GetParameter(input)
	}

	if tfawserr.ErrCodeEquals(err, ssm.ErrCodeParameterNotFound) && !d.IsNewResource() {
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
		return fmt.Errorf("error describing SSM parameter (%s): %w", d.Id(), err)
	}

	if !d.IsNewResource() && (describeResp == nil || len(describeResp.Parameters) == 0 || describeResp.Parameters[0] == nil) {
		log.Printf("[WARN] SSM Parameter %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	detail := describeResp.Parameters[0]
	d.Set("key_id", detail.KeyId)
	d.Set("description", detail.Description)
	d.Set("tier", detail.Tier)
	d.Set("allowed_pattern", detail.AllowedPattern)
	d.Set("data_type", detail.DataType)

	tags, err := ListTags(conn, name, ssm.ResourceTypeForTaggingParameter)

	if err != nil {
		return fmt.Errorf("error listing tags for SSM Parameter (%s): %w", name, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("arn", param.ARN)

	return nil
}

func resourceParameterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	if d.HasChangesExcept("tags", "tags_all") {
		paramInput := &ssm.PutParameterInput{
			Name:           aws.String(d.Get("name").(string)),
			Type:           aws.String(d.Get("type").(string)),
			Tier:           aws.String(d.Get("tier").(string)),
			Value:          aws.String(d.Get("value").(string)),
			Overwrite:      aws.Bool(ShouldUpdateParameter(d)),
			AllowedPattern: aws.String(d.Get("allowed_pattern").(string)),
		}

		// Retrieve the value set in the config directly to counteract the DiffSuppressFunc above
		tier := d.GetRawConfig().GetAttr("tier")
		if tier.IsKnown() && !tier.IsNull() {
			paramInput.Tier = aws.String(tier.AsString())
		}

		if d.HasChange("data_type") {
			paramInput.DataType = aws.String(d.Get("data_type").(string))
		}

		if d.HasChange("description") {
			paramInput.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("key_id") && d.Get("type").(string) == ssm.ParameterTypeSecureString {
			paramInput.SetKeyId(d.Get("key_id").(string))
		}

		_, err := conn.PutParameter(paramInput)

		if tfawserr.ErrMessageContains(err, "ValidationException", "Tier is not supported") {
			log.Printf("[WARN] Updating SSM Parameter (%s): tier %q not supported, using default", d.Get("name").(string), d.Get("tier").(string))
			paramInput.Tier = nil
			_, err = conn.PutParameter(paramInput)
		}

		if err != nil {
			return fmt.Errorf("error updating SSM Parameter (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), ssm.ResourceTypeForTaggingParameter, o, n); err != nil {
			return fmt.Errorf("error updating SSM Parameter (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceParameterRead(d, meta)
}

func resourceParameterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	_, err := conn.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(d.Get("name").(string)),
	})

	if tfawserr.ErrCodeEquals(err, ssm.ErrCodeParameterNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SSM Parameter (%s): %s", d.Id(), err)
	}

	return nil
}

func ShouldUpdateParameter(d *schema.ResourceData) bool {
	// If the user has specified a preference, return their preference
	if value, ok := d.GetOkExists("overwrite"); ok {
		return value.(bool)
	}

	// Since the user has not specified a preference, obey lifecycle rules
	// if it is not a new resource, otherwise overwrite should be set to false.
	return !d.IsNewResource()
}
