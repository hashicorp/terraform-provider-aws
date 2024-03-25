// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_infrastructure_configuration", name="Infrastructure Configuration")
// @Tags(identifierAttribute="id")
func ResourceInfrastructureConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInfrastructureConfigurationCreate,
		ReadWithoutTimeout:   resourceInfrastructureConfigurationRead,
		UpdateWithoutTimeout: resourceInfrastructureConfigurationUpdate,
		DeleteWithoutTimeout: resourceInfrastructureConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"terminate_instance_on_failure": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceInfrastructureConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.CreateInfrastructureConfigurationInput{
		ClientToken:                aws.String(id.UniqueId()),
		Tags:                       getTagsIn(ctx),
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
		input.InstanceTypes = flex.ExpandStringValueSet(v.(*schema.Set))
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
		input.ResourceTags = Tags(tftags.New(ctx, v.(map[string]interface{})))
	}

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("sns_topic_arn"); ok {
		input.SnsTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	var output *imagebuilder.CreateInfrastructureConfigurationOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		output, err = conn.CreateInfrastructureConfiguration(ctx, input)

		if errs.Contains(err, "instance profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateInfrastructureConfiguration(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Infrastructure Configuration: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Infrastructure Configuration: empty response")
	}

	d.SetId(aws.ToString(output.InfrastructureConfigurationArn))

	return append(diags, resourceInfrastructureConfigurationRead(ctx, d, meta)...)
}

func resourceInfrastructureConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.GetInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.GetInfrastructureConfiguration(ctx, input)

	if !d.IsNewResource() && errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
		log.Printf("[WARN] Image Builder Infrastructure Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Infrastructure Configuration (%s): %s", d.Id(), err)
	}

	if output == nil || output.InfrastructureConfiguration == nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Infrastructure Configuration (%s): empty response", d.Id())
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
	d.Set("instance_types", infrastructureConfiguration.InstanceTypes)
	d.Set("key_pair", infrastructureConfiguration.KeyPair)
	if infrastructureConfiguration.Logging != nil {
		d.Set("logging", []interface{}{flattenLogging(infrastructureConfiguration.Logging)})
	} else {
		d.Set("logging", nil)
	}
	d.Set("name", infrastructureConfiguration.Name)
	d.Set("resource_tags", KeyValueTags(ctx, infrastructureConfiguration.ResourceTags).Map())
	d.Set("security_group_ids", infrastructureConfiguration.SecurityGroupIds)
	d.Set("sns_topic_arn", infrastructureConfiguration.SnsTopicArn)
	d.Set("subnet_id", infrastructureConfiguration.SubnetId)

	setTagsOut(ctx, infrastructureConfiguration.Tags)

	d.Set("terminate_instance_on_failure", infrastructureConfiguration.TerminateInstanceOnFailure)

	return diags
}

func resourceInfrastructureConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

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
			input.InstanceTypes = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("key_pair"); ok {
			input.KeyPair = aws.String(v.(string))
		}

		if v, ok := d.GetOk("logging"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Logging = expandLogging(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("resource_tags"); ok && len(v.(map[string]interface{})) > 0 {
			input.ResourceTags = Tags(tftags.New(ctx, v.(map[string]interface{})))
		}

		if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("sns_topic_arn"); ok {
			input.SnsTopicArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("subnet_id"); ok {
			input.SubnetId = aws.String(v.(string))
		}

		err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
			_, err := conn.UpdateInfrastructureConfiguration(ctx, input)

			if errs.Contains(err, "instance profile does not exist") {
				return retry.RetryableError(err)
			}

			if err != nil {
				return retry.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateInfrastructureConfiguration(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Image Builder Infrastructure Configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceInfrastructureConfigurationRead(ctx, d, meta)...)
}

func resourceInfrastructureConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.DeleteInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteInfrastructureConfiguration(ctx, input)

	if errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Infrastructure Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func expandInstanceMetadataOptions(tfMap map[string]interface{}) *awstypes.InstanceMetadataOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InstanceMetadataOptions{}

	if v, ok := tfMap["http_put_response_hop_limit"].(int); ok && v != 0 {
		apiObject.HttpPutResponseHopLimit = aws.Int32(int32(v))
	}

	if v, ok := tfMap["http_tokens"].(string); ok && v != "" {
		apiObject.HttpTokens = aws.String(v)
	}

	return apiObject
}

func expandLogging(tfMap map[string]interface{}) *awstypes.Logging {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Logging{}

	if v, ok := tfMap["s3_logs"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3Logs = expandS3Logs(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandS3Logs(tfMap map[string]interface{}) *awstypes.S3Logs {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.S3Logs{}

	if v, ok := tfMap["s3_bucket_name"].(string); ok && v != "" {
		apiObject.S3BucketName = aws.String(v)
	}

	if v, ok := tfMap["s3_key_prefix"].(string); ok && v != "" {
		apiObject.S3KeyPrefix = aws.String(v)
	}

	return apiObject
}

func flattenInstanceMetadataOptions(apiObject *awstypes.InstanceMetadataOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HttpPutResponseHopLimit; v != nil {
		tfMap["http_put_response_hop_limit"] = aws.ToInt32(v)
	}

	if v := apiObject.HttpTokens; v != nil {
		tfMap["http_tokens"] = aws.ToString(v)
	}

	return tfMap
}

func flattenLogging(apiObject *awstypes.Logging) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3Logs; v != nil {
		tfMap["s3_logs"] = []interface{}{flattenS3Logs(v)}
	}

	return tfMap
}

func flattenS3Logs(apiObject *awstypes.S3Logs) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3BucketName; v != nil {
		tfMap["s3_bucket_name"] = aws.ToString(v)
	}

	if v := apiObject.S3KeyPrefix; v != nil {
		tfMap["s3_key_prefix"] = aws.ToString(v)
	}

	return tfMap
}
