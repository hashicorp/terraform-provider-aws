package imagebuilder

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceInfrastructureConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceInfrastructureConfigurationCreate,
		Read:   resourceInfrastructureConfigurationRead,
		Update: resourceInfrastructureConfigurationUpdate,
		Delete: resourceInfrastructureConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"instance_metadata_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_put_response_hop_limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 64),
						},
						"http_tokens": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"required", "optional"}, false),
						},
					},
				},
			},
			"instance_profile_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"instance_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"key_pair": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"logging": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_logs": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_bucket_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"s3_key_prefix": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
										Default:      "/",
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_tags": tftags.TagsSchema(),
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"terminate_instance_on_failure": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceInfrastructureConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ImageBuilderConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &imagebuilder.CreateInfrastructureConfigurationInput{
		ClientToken:                aws.String(resource.UniqueId()),
		TerminateInstanceOnFailure: aws.Bool(d.Get("terminate_instance_on_failure").(bool)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_metadata_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.InstanceMetadataOptions = expandInstanceMetadataOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("instance_profile_name"); ok {
		input.InstanceProfileName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_types"); ok && v.(*schema.Set).Len() > 0 {
		input.InstanceTypes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("key_pair"); ok {
		input.KeyPair = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Logging = expandLogging(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("resource_tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.ResourceTags = Tags(tftags.New(v.(map[string]interface{})))
	}

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("sns_topic_arn"); ok {
		input.SnsTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	var output *imagebuilder.CreateInfrastructureConfigurationOutput
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreateInfrastructureConfiguration(input)

		if tfawserr.ErrMessageContains(err, imagebuilder.ErrCodeInvalidParameterValueException, "instance profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateInfrastructureConfiguration(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Image Builder Infrastructure Configuration: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Image Builder Infrastructure Configuration: empty response")
	}

	d.SetId(aws.StringValue(output.InfrastructureConfigurationArn))

	return resourceInfrastructureConfigurationRead(d, meta)
}

func resourceInfrastructureConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ImageBuilderConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.GetInfrastructureConfiguration(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Image Builder Infrastructure Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): %w", d.Id(), err)
	}

	if output == nil || output.InfrastructureConfiguration == nil {
		return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): empty response", d.Id())
	}

	infrastructureConfiguration := output.InfrastructureConfiguration

	d.Set("arn", infrastructureConfiguration.Arn)
	d.Set("date_created", infrastructureConfiguration.DateCreated)
	d.Set("date_updated", infrastructureConfiguration.DateUpdated)
	d.Set("description", infrastructureConfiguration.Description)

	if infrastructureConfiguration.InstanceMetadataOptions != nil {
		d.Set("instance_metadata_options", []interface{}{
			flattenInstanceMetadataOptions(infrastructureConfiguration.InstanceMetadataOptions),
		})
	} else {
		d.Set("instance_metadata_options", nil)
	}

	d.Set("instance_profile_name", infrastructureConfiguration.InstanceProfileName)
	d.Set("instance_types", aws.StringValueSlice(infrastructureConfiguration.InstanceTypes))
	d.Set("key_pair", infrastructureConfiguration.KeyPair)
	if infrastructureConfiguration.Logging != nil {
		d.Set("logging", []interface{}{flattenLogging(infrastructureConfiguration.Logging)})
	} else {
		d.Set("logging", nil)
	}
	d.Set("name", infrastructureConfiguration.Name)
	d.Set("resource_tags", KeyValueTags(infrastructureConfiguration.ResourceTags).Map())
	d.Set("security_group_ids", aws.StringValueSlice(infrastructureConfiguration.SecurityGroupIds))
	d.Set("sns_topic_arn", infrastructureConfiguration.SnsTopicArn)
	d.Set("subnet_id", infrastructureConfiguration.SubnetId)
	tags := KeyValueTags(infrastructureConfiguration.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}
	d.Set("terminate_instance_on_failure", infrastructureConfiguration.TerminateInstanceOnFailure)

	return nil
}

func resourceInfrastructureConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ImageBuilderConn

	if d.HasChanges(
		"description",
		"instance_metadata_options",
		"instance_profile_name",
		"instance_types",
		"key_pair",
		"logging",
		"resource_tags",
		"security_group_ids",
		"sns_topic_arn",
		"subnet_id",
		"terminate_instance_on_failure",
	) {
		input := &imagebuilder.UpdateInfrastructureConfigurationInput{
			InfrastructureConfigurationArn: aws.String(d.Id()),
			TerminateInstanceOnFailure:     aws.Bool(d.Get("terminate_instance_on_failure").(bool)),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("instance_metadata_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.InstanceMetadataOptions = expandInstanceMetadataOptions(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("instance_profile_name"); ok {
			input.InstanceProfileName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("instance_types"); ok && v.(*schema.Set).Len() > 0 {
			input.InstanceTypes = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("key_pair"); ok {
			input.KeyPair = aws.String(v.(string))
		}

		if v, ok := d.GetOk("logging"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Logging = expandLogging(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("resource_tags"); ok && len(v.(map[string]interface{})) > 0 {
			input.ResourceTags = Tags(tftags.New(v.(map[string]interface{})))
		}

		if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("sns_topic_arn"); ok {
			input.SnsTopicArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("subnet_id"); ok {
			input.SubnetId = aws.String(v.(string))
		}

		err := resource.Retry(propagationTimeout, func() *resource.RetryError {
			_, err := conn.UpdateInfrastructureConfiguration(input)

			if tfawserr.ErrMessageContains(err, imagebuilder.ErrCodeInvalidParameterValueException, "instance profile does not exist") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateInfrastructureConfiguration(input)
		}

		if err != nil {
			return fmt.Errorf("error updating Image Builder Infrastructure Configuration (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for Image Builder Infrastructure Configuration (%s): %w", d.Id(), err)
		}
	}

	return resourceInfrastructureConfigurationRead(d, meta)
}

func resourceInfrastructureConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ImageBuilderConn

	input := &imagebuilder.DeleteInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteInfrastructureConfiguration(input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Image Builder Infrastructure Configuration (%s): %w", d.Id(), err)
	}

	return nil
}

func expandInstanceMetadataOptions(tfMap map[string]interface{}) *imagebuilder.InstanceMetadataOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.InstanceMetadataOptions{}

	if v, ok := tfMap["http_put_response_hop_limit"].(int); ok && v != 0 {
		apiObject.HttpPutResponseHopLimit = aws.Int64(int64(v))
	}

	if v, ok := tfMap["http_tokens"].(string); ok && v != "" {
		apiObject.HttpTokens = aws.String(v)
	}

	return apiObject
}

func expandLogging(tfMap map[string]interface{}) *imagebuilder.Logging {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.Logging{}

	if v, ok := tfMap["s3_logs"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3Logs = expandS3Logs(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandS3Logs(tfMap map[string]interface{}) *imagebuilder.S3Logs {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.S3Logs{}

	if v, ok := tfMap["s3_bucket_name"].(string); ok && v != "" {
		apiObject.S3BucketName = aws.String(v)
	}

	if v, ok := tfMap["s3_key_prefix"].(string); ok && v != "" {
		apiObject.S3KeyPrefix = aws.String(v)
	}

	return apiObject
}

func flattenInstanceMetadataOptions(apiObject *imagebuilder.InstanceMetadataOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HttpPutResponseHopLimit; v != nil {
		tfMap["http_put_response_hop_limit"] = aws.Int64Value(v)
	}

	if v := apiObject.HttpTokens; v != nil {
		tfMap["http_tokens"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLogging(apiObject *imagebuilder.Logging) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3Logs; v != nil {
		tfMap["s3_logs"] = []interface{}{flattenS3Logs(v)}
	}

	return tfMap
}

func flattenS3Logs(apiObject *imagebuilder.S3Logs) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3BucketName; v != nil {
		tfMap["s3_bucket_name"] = aws.StringValue(v)
	}

	if v := apiObject.S3KeyPrefix; v != nil {
		tfMap["s3_key_prefix"] = aws.StringValue(v)
	}

	return tfMap
}
