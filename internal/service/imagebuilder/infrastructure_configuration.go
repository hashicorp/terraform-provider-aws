// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
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
func resourceInfrastructureConfiguration() *schema.Resource {
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
	}
}

func resourceInfrastructureConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.CreateInfrastructureConfigurationInput{
		ClientToken:                aws.String(id.UniqueId()),
		Tags:                       getTagsIn(ctx),
		TerminateInstanceOnFailure: aws.Bool(d.Get("terminate_instance_on_failure").(bool)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_metadata_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.InstanceMetadataOptions = expandInstanceMetadataOptions(v.([]any)[0].(map[string]any))
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

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Logging = expandLogging(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrResourceTags); ok && len(v.(map[string]any)) > 0 {
		input.ResourceTags = svcTags(tftags.New(ctx, v.(map[string]any)))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrSNSTopicARN); ok {
		input.SnsTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSubnetID); ok {
		input.SubnetId = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (any, error) {
			return conn.CreateInfrastructureConfiguration(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.Contains(err, "instance profile does not exist") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Infrastructure Configuration: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*imagebuilder.CreateInfrastructureConfigurationOutput).InfrastructureConfigurationArn))

	return append(diags, resourceInfrastructureConfigurationRead(ctx, d, meta)...)
}

func resourceInfrastructureConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	infrastructureConfiguration, err := findInfrastructureConfigurationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Image Builder Infrastructure Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Infrastructure Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, infrastructureConfiguration.Arn)
	d.Set("date_created", infrastructureConfiguration.DateCreated)
	d.Set("date_updated", infrastructureConfiguration.DateUpdated)
	d.Set(names.AttrDescription, infrastructureConfiguration.Description)
	if infrastructureConfiguration.InstanceMetadataOptions != nil {
		if err := d.Set("instance_metadata_options", []any{flattenInstanceMetadataOptions(infrastructureConfiguration.InstanceMetadataOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting instance_metadata_options: %s", err)
		}
	} else {
		d.Set("instance_metadata_options", nil)
	}
	d.Set("instance_profile_name", infrastructureConfiguration.InstanceProfileName)
	d.Set("instance_types", infrastructureConfiguration.InstanceTypes)
	d.Set("key_pair", infrastructureConfiguration.KeyPair)
	if infrastructureConfiguration.Logging != nil {
		if err := d.Set("logging", []any{flattenLogging(infrastructureConfiguration.Logging)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging: %s", err)
		}
	} else {
		d.Set("logging", nil)
	}
	d.Set(names.AttrName, infrastructureConfiguration.Name)
	d.Set(names.AttrResourceTags, keyValueTags(ctx, infrastructureConfiguration.ResourceTags).Map())
	d.Set(names.AttrSecurityGroupIDs, infrastructureConfiguration.SecurityGroupIds)
	d.Set(names.AttrSNSTopicARN, infrastructureConfiguration.SnsTopicArn)
	d.Set(names.AttrSubnetID, infrastructureConfiguration.SubnetId)
	d.Set("terminate_instance_on_failure", infrastructureConfiguration.TerminateInstanceOnFailure)

	setTagsOut(ctx, infrastructureConfiguration.Tags)

	return diags
}

func resourceInfrastructureConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

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

		if v, ok := d.GetOk("instance_metadata_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.InstanceMetadataOptions = expandInstanceMetadataOptions(v.([]any)[0].(map[string]any))
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

		if v, ok := d.GetOk("logging"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.Logging = expandLogging(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk(names.AttrResourceTags); ok && len(v.(map[string]any)) > 0 {
			input.ResourceTags = svcTags(tftags.New(ctx, v.(map[string]any)))
		}

		if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
			input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk(names.AttrSNSTopicARN); ok {
			input.SnsTopicArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrSubnetID); ok {
			input.SubnetId = aws.String(v.(string))
		}

		_, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (any, error) {
				return conn.UpdateInfrastructureConfiguration(ctx, input)
			},
			func(err error) (bool, error) {
				if errs.Contains(err, "instance profile does not exist") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Image Builder Infrastructure Configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceInfrastructureConfigurationRead(ctx, d, meta)...)
}

func resourceInfrastructureConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	log.Printf("[DEBUG] Deleting Image Infrastructure Configuration: %s", d.Id())
	_, err := conn.DeleteInfrastructureConfiguration(ctx, &imagebuilder.DeleteInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Infrastructure Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findInfrastructureConfigurationByARN(ctx context.Context, conn *imagebuilder.Client, arn string) (*awstypes.InfrastructureConfiguration, error) {
	input := &imagebuilder.GetInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(arn),
	}

	output, err := conn.GetInfrastructureConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.InfrastructureConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.InfrastructureConfiguration, nil
}

func expandInstanceMetadataOptions(tfMap map[string]any) *awstypes.InstanceMetadataOptions {
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

func expandLogging(tfMap map[string]any) *awstypes.Logging {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Logging{}

	if v, ok := tfMap["s3_logs"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3Logs = expandS3Logs(v[0].(map[string]any))
	}

	return apiObject
}

func expandS3Logs(tfMap map[string]any) *awstypes.S3Logs {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.S3Logs{}

	if v, ok := tfMap[names.AttrS3BucketName].(string); ok && v != "" {
		apiObject.S3BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrS3KeyPrefix].(string); ok && v != "" {
		apiObject.S3KeyPrefix = aws.String(v)
	}

	return apiObject
}

func flattenInstanceMetadataOptions(apiObject *awstypes.InstanceMetadataOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.HttpPutResponseHopLimit; v != nil {
		tfMap["http_put_response_hop_limit"] = aws.ToInt32(v)
	}

	if v := apiObject.HttpTokens; v != nil {
		tfMap["http_tokens"] = aws.ToString(v)
	}

	return tfMap
}

func flattenLogging(apiObject *awstypes.Logging) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.S3Logs; v != nil {
		tfMap["s3_logs"] = []any{flattenS3Logs(v)}
	}

	return tfMap
}

func flattenS3Logs(apiObject *awstypes.S3Logs) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.S3BucketName; v != nil {
		tfMap[names.AttrS3BucketName] = aws.ToString(v)
	}

	if v := apiObject.S3KeyPrefix; v != nil {
		tfMap[names.AttrS3KeyPrefix] = aws.ToString(v)
	}

	return tfMap
}
