// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_model", name="Model")
// @Tags(identifierAttribute="arn")
func resourceModel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceModelCreate,
		ReadWithoutTimeout:   resourceModelRead,
		UpdateWithoutTimeout: resourceModelUpdate,
		DeleteWithoutTimeout: resourceModelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_hostname": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validName,
						},
						names.AttrEnvironment: {
							Type:         schema.TypeMap,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validEnvironment,
							Elem:         &schema.Schema{Type: schema.TypeString},
						},
						"image": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validImage,
						},
						"image_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"repository_access_mode": {
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.RepositoryAccessMode](),
									},
									"repository_auth_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"repository_credentials_provider_arn": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
								},
							},
						},
						names.AttrMode: {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          awstypes.ContainerModeSingleModel,
							ValidateDiagFunc: enum.Validate[awstypes.ContainerMode](),
						},
						"model_data_url": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validModelDataURL,
						},
						"model_package_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"model_data_source": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_data_source": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"s3_uri": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validModelDataURL,
												},
												"s3_data_type": {
													Type:             schema.TypeString,
													Required:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.S3ModelDataType](),
												},
												"compression_type": {
													Type:             schema.TypeString,
													Required:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.ModelCompressionType](),
												},
												"model_access_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"accept_eula": {
																Type:     schema.TypeBool,
																Required: true,
																ForceNew: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"inference_specification_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validName,
						},
						"multi_model_config": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"model_cache_setting": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ModelCacheSetting](),
									},
								},
							},
						},
					},
				},
			},
			"enable_network_isolation": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			names.AttrExecutionRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"inference_execution_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMode: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.InferenceExecutionMode](),
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},
			"primary_container": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_hostname": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validName,
						},
						names.AttrEnvironment: {
							Type:         schema.TypeMap,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validEnvironment,
							Elem:         &schema.Schema{Type: schema.TypeString},
						},
						"image": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validImage,
						},
						"image_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"repository_access_mode": {
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.RepositoryAccessMode](),
									},
									"repository_auth_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"repository_credentials_provider_arn": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
								},
							},
						},
						names.AttrMode: {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          awstypes.ContainerModeSingleModel,
							ValidateDiagFunc: enum.Validate[awstypes.ContainerMode](),
						},
						"model_data_url": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validModelDataURL,
						},
						"model_package_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"model_data_source": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_data_source": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"s3_uri": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validModelDataURL,
												},
												"s3_data_type": {
													Type:             schema.TypeString,
													Required:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.S3ModelDataType](),
												},
												"compression_type": {
													Type:             schema.TypeString,
													Required:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.ModelCompressionType](),
												},
												"model_access_config": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"accept_eula": {
																Type:     schema.TypeBool,
																Required: true,
																ForceNew: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"inference_specification_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validName,
						},
						"multi_model_config": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"model_cache_setting": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ModelCacheSetting](),
									},
								},
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCConfig: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 16,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceModelCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	var name string
	if v, ok := d.GetOk(names.AttrName); ok {
		name = v.(string)
	} else {
		name = id.UniqueId()
	}

	createOpts := &sagemaker.CreateModelInput{
		ModelName: aws.String(name),
		Tags:      getTagsIn(ctx),
	}

	if v, ok := d.GetOk("primary_container"); ok {
		createOpts.PrimaryContainer = expandContainer(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("container"); ok {
		createOpts.Containers = expandContainers(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrExecutionRoleARN); ok {
		createOpts.ExecutionRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVPCConfig); ok {
		createOpts.VpcConfig = expandVPCConfigRequest(v.([]any))
	}

	if v, ok := d.GetOk("enable_network_isolation"); ok {
		createOpts.EnableNetworkIsolation = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("inference_execution_config"); ok {
		createOpts.InferenceExecutionConfig = expandModelInferenceExecutionConfig(v.([]any))
	}

	log.Printf("[DEBUG] SageMaker AI model create config: %#v", *createOpts)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (any, error) {
		return conn.CreateModel(ctx, createOpts)
	}, ErrCodeValidationException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI model: %s", err)
	}
	d.SetId(name)

	return append(diags, resourceModelRead(ctx, d, meta)...)
}

func resourceModelRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	output, err := findModelByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[INFO] unable to find the sagemaker model resource and therefore it is removed from the state: %s", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI model %s: %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ModelArn)
	d.Set(names.AttrName, output.ModelName)
	d.Set(names.AttrExecutionRoleARN, output.ExecutionRoleArn)
	d.Set("enable_network_isolation", output.EnableNetworkIsolation)

	if err := d.Set("primary_container", flattenContainer(output.PrimaryContainer)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting primary_container: %s", err)
	}

	if err := d.Set("container", flattenContainers(output.Containers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting container: %s", err)
	}

	if err := d.Set(names.AttrVPCConfig, flattenVPCConfigResponse(output.VpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
	}

	if err := d.Set("inference_execution_config", flattenModelInferenceExecutionConfig(output.InferenceExecutionConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting inference_execution_config: %s", err)
	}

	return diags
}

func resourceModelUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceModelRead(ctx, d, meta)...)
}

func resourceModelDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	deleteOpts := &sagemaker.DeleteModelInput{
		ModelName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker AI model: %s", d.Id())

	err := retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteModel(ctx, deleteOpts)

		if err != nil {
			if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find model") {
				return nil
			}

			if errs.IsA[*awstypes.ResourceNotFound](err) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteModel(ctx, deleteOpts)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting sagemaker model: %s", err)
	}
	return diags
}

func findModelByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeModelOutput, error) {
	input := &sagemaker.DescribeModelInput{
		ModelName: aws.String(name),
	}

	output, err := conn.DescribeModel(ctx, input)

	if tfawserr.ErrCodeContains(err, ErrCodeValidationException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandVPCConfigRequest(l []any) *awstypes.VpcConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]any)

	return &awstypes.VpcConfig{
		SecurityGroupIds: flex.ExpandStringValueSet(m[names.AttrSecurityGroupIDs].(*schema.Set)),
		Subnets:          flex.ExpandStringValueSet(m[names.AttrSubnets].(*schema.Set)),
	}
}

func flattenVPCConfigResponse(vpcConfig *awstypes.VpcConfig) []map[string]any {
	if vpcConfig == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrSecurityGroupIDs: flex.FlattenStringValueSet(vpcConfig.SecurityGroupIds),
		names.AttrSubnets:          flex.FlattenStringValueSet(vpcConfig.Subnets),
	}

	return []map[string]any{m}
}

func expandContainer(m map[string]any) *awstypes.ContainerDefinition {
	container := awstypes.ContainerDefinition{}

	if v, ok := m["image"]; ok && v.(string) != "" {
		container.Image = aws.String(v.(string))
	}

	if v, ok := m[names.AttrMode]; ok && v.(string) != "" {
		container.Mode = awstypes.ContainerMode(v.(string))
	}

	if v, ok := m["container_hostname"]; ok && v.(string) != "" {
		container.ContainerHostname = aws.String(v.(string))
	}
	if v, ok := m["model_data_url"]; ok && v.(string) != "" {
		container.ModelDataUrl = aws.String(v.(string))
	}
	if v, ok := m["model_package_name"]; ok && v.(string) != "" {
		container.ModelPackageName = aws.String(v.(string))
	}
	if v, ok := m["model_data_source"]; ok {
		container.ModelDataSource = expandModelDataSource(v.([]any))
	}
	if v, ok := m[names.AttrEnvironment].(map[string]any); ok && len(v) > 0 {
		container.Environment = flex.ExpandStringValueMap(v)
	}

	if v, ok := m["image_config"]; ok {
		container.ImageConfig = expandModelImageConfig(v.([]any))
	}

	if v, ok := m["inference_specification_name"]; ok && v.(string) != "" {
		container.InferenceSpecificationName = aws.String(v.(string))
	}

	if v, ok := m["multi_model_config"].([]any); ok && len(v) > 0 {
		container.MultiModelConfig = expandMultiModelConfig(v)
	}

	return &container
}

func expandModelDataSource(l []any) *awstypes.ModelDataSource {
	if len(l) == 0 {
		return nil
	}

	modelDataSource := awstypes.ModelDataSource{}

	m := l[0].(map[string]any)

	if v, ok := m["s3_data_source"]; ok {
		modelDataSource.S3DataSource = expandS3ModelDataSource(v.([]any))
	}

	return &modelDataSource
}

func expandS3ModelDataSource(l []any) *awstypes.S3ModelDataSource {
	if len(l) == 0 {
		return nil
	}

	s3ModelDataSource := awstypes.S3ModelDataSource{}

	m := l[0].(map[string]any)

	if v, ok := m["s3_uri"]; ok && v.(string) != "" {
		s3ModelDataSource.S3Uri = aws.String(v.(string))
	}
	if v, ok := m["s3_data_type"]; ok && v.(string) != "" {
		s3ModelDataSource.S3DataType = awstypes.S3ModelDataType(v.(string))
	}
	if v, ok := m["compression_type"]; ok && v.(string) != "" {
		s3ModelDataSource.CompressionType = awstypes.ModelCompressionType(v.(string))
	}

	if v, ok := m["model_access_config"].([]any); ok && len(v) > 0 {
		s3ModelDataSource.ModelAccessConfig = expandModelAccessConfig(v)
	}

	return &s3ModelDataSource
}

func expandModelImageConfig(l []any) *awstypes.ImageConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]any)

	imageConfig := &awstypes.ImageConfig{
		RepositoryAccessMode: awstypes.RepositoryAccessMode(m["repository_access_mode"].(string)),
	}

	if v, ok := m["repository_auth_config"].([]any); ok && len(v) > 0 && v[0] != nil {
		imageConfig.RepositoryAuthConfig = expandRepositoryAuthConfig(v[0].(map[string]any))
	}

	return imageConfig
}

func expandRepositoryAuthConfig(tfMap map[string]any) *awstypes.RepositoryAuthConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RepositoryAuthConfig{}

	if v, ok := tfMap["repository_credentials_provider_arn"].(string); ok && v != "" {
		apiObject.RepositoryCredentialsProviderArn = aws.String(v)
	}

	return apiObject
}

func expandContainers(a []any) []awstypes.ContainerDefinition {
	containers := make([]awstypes.ContainerDefinition, 0, len(a))

	for _, m := range a {
		containers = append(containers, *expandContainer(m.(map[string]any)))
	}

	return containers
}

func expandModelAccessConfig(l []any) *awstypes.ModelAccessConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]any)

	modelAccessConfig := &awstypes.ModelAccessConfig{}

	if v, ok := m["accept_eula"].(bool); ok {
		modelAccessConfig.AcceptEula = aws.Bool(v)
	}

	return modelAccessConfig
}

func expandMultiModelConfig(l []any) *awstypes.MultiModelConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]any)

	multiModelConfig := &awstypes.MultiModelConfig{}

	if v, ok := m["model_cache_setting"].(string); ok && v != "" {
		multiModelConfig.ModelCacheSetting = awstypes.ModelCacheSetting(v)
	}

	return multiModelConfig
}

func flattenContainer(container *awstypes.ContainerDefinition) []any {
	if container == nil {
		return []any{}
	}

	cfg := make(map[string]any)

	if container.Image != nil {
		cfg["image"] = aws.ToString(container.Image)
	}

	cfg[names.AttrMode] = container.Mode

	if container.ContainerHostname != nil {
		cfg["container_hostname"] = aws.ToString(container.ContainerHostname)
	}
	if container.ModelDataUrl != nil {
		cfg["model_data_url"] = aws.ToString(container.ModelDataUrl)
	}
	if container.ModelDataSource != nil {
		cfg["model_data_source"] = flattenModelDataSource(container.ModelDataSource)
	}
	if container.ModelPackageName != nil {
		cfg["model_package_name"] = aws.ToString(container.ModelPackageName)
	}
	if container.Environment != nil {
		cfg[names.AttrEnvironment] = aws.StringMap(container.Environment)
	}

	if container.ImageConfig != nil {
		cfg["image_config"] = flattenImageConfig(container.ImageConfig)
	}

	if container.InferenceSpecificationName != nil {
		cfg["inference_specification_name"] = aws.ToString(container.InferenceSpecificationName)
	}

	if container.MultiModelConfig != nil {
		cfg["multi_model_config"] = flattenMultiModelConfig(container.MultiModelConfig)
	}

	return []any{cfg}
}

func flattenModelDataSource(modelDataSource *awstypes.ModelDataSource) []any {
	if modelDataSource == nil {
		return []any{}
	}

	cfg := make(map[string]any)

	if modelDataSource.S3DataSource != nil {
		cfg["s3_data_source"] = flattenS3ModelDataSource(modelDataSource.S3DataSource)
	}

	return []any{cfg}
}

func flattenS3ModelDataSource(s3ModelDataSource *awstypes.S3ModelDataSource) []any {
	if s3ModelDataSource == nil {
		return []any{}
	}

	cfg := make(map[string]any)

	if s3ModelDataSource.S3Uri != nil {
		cfg["s3_uri"] = aws.ToString(s3ModelDataSource.S3Uri)
	}

	cfg["s3_data_type"] = s3ModelDataSource.S3DataType

	cfg["compression_type"] = s3ModelDataSource.CompressionType

	if s3ModelDataSource.ModelAccessConfig != nil {
		cfg["model_access_config"] = flattenModelAccessConfig(s3ModelDataSource.ModelAccessConfig)
	}

	return []any{cfg}
}

func flattenImageConfig(imageConfig *awstypes.ImageConfig) []any {
	if imageConfig == nil {
		return []any{}
	}

	cfg := make(map[string]any)

	cfg["repository_access_mode"] = imageConfig.RepositoryAccessMode

	if tfMap := flattenRepositoryAuthConfig(imageConfig.RepositoryAuthConfig); len(tfMap) > 0 {
		cfg["repository_auth_config"] = []any{tfMap}
	}

	return []any{cfg}
}

func flattenRepositoryAuthConfig(apiObject *awstypes.RepositoryAuthConfig) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)

	if v := apiObject.RepositoryCredentialsProviderArn; v != nil {
		tfMap["repository_credentials_provider_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenContainers(containers []awstypes.ContainerDefinition) []any {
	fContainers := make([]any, 0, len(containers))
	for _, container := range containers {
		fContainers = append(fContainers, flattenContainer(&container)[0].(map[string]any))
	}
	return fContainers
}

func flattenModelAccessConfig(config *awstypes.ModelAccessConfig) []any {
	if config == nil {
		return []any{}
	}

	cfg := make(map[string]any)

	cfg["accept_eula"] = aws.ToBool(config.AcceptEula)

	return []any{cfg}
}

func flattenMultiModelConfig(config *awstypes.MultiModelConfig) []any {
	if config == nil {
		return []any{}
	}

	cfg := make(map[string]any)

	if config.ModelCacheSetting != "" {
		cfg["model_cache_setting"] = config.ModelCacheSetting
	}

	return []any{cfg}
}

func expandModelInferenceExecutionConfig(l []any) *awstypes.InferenceExecutionConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.InferenceExecutionConfig{
		Mode: awstypes.InferenceExecutionMode(m[names.AttrMode].(string)),
	}

	return config
}

func flattenModelInferenceExecutionConfig(config *awstypes.InferenceExecutionConfig) []any {
	if config == nil {
		return []any{}
	}

	cfg := make(map[string]any)

	cfg[names.AttrMode] = config.Mode

	return []any{cfg}
}
