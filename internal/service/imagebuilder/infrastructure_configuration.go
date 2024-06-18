// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			names.AttrARN: {
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
			names.AttrDescription: {
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
									names.AttrS3BucketName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									names.AttrS3KeyPrefix: {
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResourceTags: tftags.TagsSchema(),
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			names.AttrSNSTopicARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrSubnetID: {
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
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.CreateInfrastructureConfigurationInput{
		ClientToken:                aws.String(id.UniqueId()),
		Tags:                       getTagsIn(ctx),
		TerminateInstanceOnFailure: aws.Bool(d.Get("terminate_instance_on_failure").(bool)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
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

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Logging = expandLogging(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrResourceTags); ok && len(v.(map[string]interface{})) > 0 {
		input.ResourceTags = Tags(tftags.New(ctx, v.(map[string]interface{})))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrSNSTopicARN); ok {
		input.SnsTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSubnetID); ok {
		input.SubnetId = aws.String(v.(string))
	}

	var output *imagebuilder.CreateInfrastructureConfigurationOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		output, err = conn.CreateInfrastructureConfigurationWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, imagebuilder.ErrCodeInvalidParameterValueException, "instance profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateInfrastructureConfigurationWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Infrastructure Configuration: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Infrastructure Configuration: empty response")
	}

	d.SetId(aws.StringValue(output.InfrastructureConfigurationArn))

	return append(diags, resourceInfrastructureConfigurationRead(ctx, d, meta)...)
}

func resourceInfrastructureConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.GetInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.GetInfrastructureConfigurationWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
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

	d.Set(names.AttrARN, infrastructureConfiguration.Arn)
	d.Set("date_created", infrastructureConfiguration.DateCreated)
	d.Set("date_updated", infrastructureConfiguration.DateUpdated)
	d.Set(names.AttrDescription, infrastructureConfiguration.Description)

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
	d.Set(names.AttrName, infrastructureConfiguration.Name)
	d.Set(names.AttrResourceTags, KeyValueTags(ctx, infrastructureConfiguration.ResourceTags).Map())
	d.Set(names.AttrSecurityGroupIDs, aws.StringValueSlice(infrastructureConfiguration.SecurityGroupIds))
	d.Set(names.AttrSNSTopicARN, infrastructureConfiguration.SnsTopicArn)
	d.Set(names.AttrSubnetID, infrastructureConfiguration.SubnetId)

	setTagsOut(ctx, infrastructureConfiguration.Tags)

	d.Set("terminate_instance_on_failure", infrastructureConfiguration.TerminateInstanceOnFailure)

	return diags
}

func resourceInfrastructureConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	if d.HasChanges(
		names.AttrDescription,
		"instance_metadata_options",
		"instance_profile_name",
		"instance_types",
		"key_pair",
		"logging",
		names.AttrResourceTags,
		names.AttrSecurityGroupIDs,
		names.AttrSNSTopicARN,
		names.AttrSubnetID,
		"terminate_instance_on_failure",
	) {
		input := &imagebuilder.UpdateInfrastructureConfigurationInput{
			InfrastructureConfigurationArn: aws.String(d.Id()),
			TerminateInstanceOnFailure:     aws.Bool(d.Get("terminate_instance_on_failure").(bool)),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
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

		if v, ok := d.GetOk(names.AttrResourceTags); ok && len(v.(map[string]interface{})) > 0 {
			input.ResourceTags = Tags(tftags.New(ctx, v.(map[string]interface{})))
		}

		if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
			input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk(names.AttrSNSTopicARN); ok {
			input.SnsTopicArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrSubnetID); ok {
			input.SubnetId = aws.String(v.(string))
		}

		err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
			_, err := conn.UpdateInfrastructureConfigurationWithContext(ctx, input)

			if tfawserr.ErrMessageContains(err, imagebuilder.ErrCodeInvalidParameterValueException, "instance profile does not exist") {
				return retry.RetryableError(err)
			}

			if err != nil {
				return retry.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateInfrastructureConfigurationWithContext(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Image Builder Infrastructure Configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceInfrastructureConfigurationRead(ctx, d, meta)...)
}

func resourceInfrastructureConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.DeleteInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteInfrastructureConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Infrastructure Configuration (%s): %s", d.Id(), err)
	}

	return diags
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

	if v, ok := tfMap[names.AttrS3BucketName].(string); ok && v != "" {
		apiObject.S3BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrS3KeyPrefix].(string); ok && v != "" {
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
		tfMap[names.AttrS3BucketName] = aws.StringValue(v)
	}

	if v := apiObject.S3KeyPrefix; v != nil {
		tfMap[names.AttrS3KeyPrefix] = aws.StringValue(v)
	}

	return tfMap
}
