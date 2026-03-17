// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
							ValidateFunc: validHTTPSOrS3URI,
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
												"s3_data_type": {
													Type:             schema.TypeString,
													Required:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.S3ModelDataType](),
												},
												"s3_uri": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validHTTPSOrS3URI,
												},
											},
										},
									},
								},
							},
						},
						"additional_model_data_source": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"channel_name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"s3_data_source": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
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
												"s3_data_type": {
													Type:             schema.TypeString,
													Required:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.S3ModelDataType](),
												},
												"s3_uri": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validHTTPSOrS3URI,
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
							ValidateFunc: validHTTPSOrS3URI,
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
												"s3_data_type": {
													Type:             schema.TypeString,
													Required:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.S3ModelDataType](),
												},
												"s3_uri": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validHTTPSOrS3URI,
												},
											},
										},
									},
								},
							},
						},
						"additional_model_data_source": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"channel_name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"s3_data_source": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
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
												"s3_data_type": {
													Type:             schema.TypeString,
													Required:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.S3ModelDataType](),
												},
												"s3_uri": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validHTTPSOrS3URI,
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
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 5,
							ForceNew: true,
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
		name = sdkid.UniqueId()
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
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func(ctx context.Context) (any, error) {
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

	if !d.IsNewResource() && retry.NotFound(err) {
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

	err := tfresource.Retry(ctx, 5*time.Minute, func(ctx context.Context) *tfresource.RetryError {
		_, err := conn.DeleteModel(ctx, deleteOpts)

		if err != nil {
			if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find model") {
				return nil
			}

			if errs.IsA[*awstypes.ResourceNotFound](err) {
				return tfresource.RetryableError(err)
			}

			return tfresource.NonRetryableError(err)
		}

		return nil
	})

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
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func expandVPCConfigRequest(tfList []any) *awstypes.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	return &awstypes.VpcConfig{
		SecurityGroupIds: flex.ExpandStringValueSet(m[names.AttrSecurityGroupIDs].(*schema.Set)),
		Subnets:          flex.ExpandStringValueSet(m[names.AttrSubnets].(*schema.Set)),
	}
}

func expandContainer(tfMap map[string]any) *awstypes.ContainerDefinition {
	apiObject := awstypes.ContainerDefinition{}

	if v, ok := tfMap["image"]; ok && v.(string) != "" {
		apiObject.Image = aws.String(v.(string))
	}
	if v, ok := tfMap[names.AttrMode]; ok && v.(string) != "" {
		apiObject.Mode = awstypes.ContainerMode(v.(string))
	}
	if v, ok := tfMap["container_hostname"]; ok && v.(string) != "" {
		apiObject.ContainerHostname = aws.String(v.(string))
	}
	if v, ok := tfMap["model_data_url"]; ok && v.(string) != "" {
		apiObject.ModelDataUrl = aws.String(v.(string))
	}
	if v, ok := tfMap["model_package_name"]; ok && v.(string) != "" {
		apiObject.ModelPackageName = aws.String(v.(string))
	}
	if v, ok := tfMap["model_data_source"]; ok {
		apiObject.ModelDataSource = expandModelDataSource(v.([]any))
	}
	if v, ok := tfMap["additional_model_data_source"]; ok {
		apiObject.AdditionalModelDataSources = expandAdditionalModelDataSources(v.([]any))
	}
	if v, ok := tfMap[names.AttrEnvironment].(map[string]any); ok && len(v) > 0 {
		apiObject.Environment = flex.ExpandStringValueMap(v)
	}
	if v, ok := tfMap["image_config"]; ok {
		apiObject.ImageConfig = expandModelImageConfig(v.([]any))
	}
	if v, ok := tfMap["inference_specification_name"]; ok && v.(string) != "" {
		apiObject.InferenceSpecificationName = aws.String(v.(string))
	}
	if v, ok := tfMap["multi_model_config"].([]any); ok && len(v) > 0 {
		apiObject.MultiModelConfig = expandMultiModelConfig(v)
	}

	return &apiObject
}

func expandModelDataSource(tfList []any) *awstypes.ModelDataSource {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := awstypes.ModelDataSource{}
	if v, ok := tfMap["s3_data_source"]; ok {
		apiObject.S3DataSource = expandS3ModelDataSource(v.([]any))
	}

	return &apiObject
}

func expandAdditionalModelDataSources(tfList []any) []awstypes.AdditionalModelDataSource {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]awstypes.AdditionalModelDataSource, 0, len(tfList))
	for _, m := range tfList {
		tfMap, ok := m.(map[string]any)
		if !ok {
			continue
		}
		apiObjects = append(apiObjects, expandAdditionalModelDataSource(tfMap))
	}

	return apiObjects
}

func expandAdditionalModelDataSource(tfMap map[string]any) awstypes.AdditionalModelDataSource {
	apiObject := awstypes.AdditionalModelDataSource{}

	if v, ok := tfMap["channel_name"]; ok && v.(string) != "" {
		apiObject.ChannelName = aws.String(v.(string))
	}
	if v, ok := tfMap["s3_data_source"]; ok {
		apiObject.S3DataSource = expandS3ModelDataSource(v.([]any))
	}

	return apiObject
}

func expandS3ModelDataSource(tfList []any) *awstypes.S3ModelDataSource {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	apiObject := awstypes.S3ModelDataSource{}
	if v, ok := m["s3_uri"]; ok && v.(string) != "" {
		apiObject.S3Uri = aws.String(v.(string))
	}
	if v, ok := m["s3_data_type"]; ok && v.(string) != "" {
		apiObject.S3DataType = awstypes.S3ModelDataType(v.(string))
	}
	if v, ok := m["compression_type"]; ok && v.(string) != "" {
		apiObject.CompressionType = awstypes.ModelCompressionType(v.(string))
	}
	if v, ok := m["model_access_config"].([]any); ok && len(v) > 0 {
		apiObject.ModelAccessConfig = expandModelAccessConfig(v)
	}

	return &apiObject
}

func expandModelImageConfig(tfList []any) *awstypes.ImageConfig {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	apiObject := &awstypes.ImageConfig{
		RepositoryAccessMode: awstypes.RepositoryAccessMode(m["repository_access_mode"].(string)),
	}

	if v, ok := m["repository_auth_config"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.RepositoryAuthConfig = expandRepositoryAuthConfig(v[0].(map[string]any))
	}

	return apiObject
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
	apiObjects := make([]awstypes.ContainerDefinition, 0, len(a))

	for _, m := range a {
		apiObjects = append(apiObjects, *expandContainer(m.(map[string]any)))
	}

	return apiObjects
}

func expandModelAccessConfig(tfList []any) *awstypes.ModelAccessConfig {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	apiObject := &awstypes.ModelAccessConfig{}
	if v, ok := m["accept_eula"].(bool); ok {
		apiObject.AcceptEula = aws.Bool(v)
	}

	return apiObject
}

func expandMultiModelConfig(tfList []any) *awstypes.MultiModelConfig {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	apiObject := &awstypes.MultiModelConfig{}
	if v, ok := m["model_cache_setting"].(string); ok && v != "" {
		apiObject.ModelCacheSetting = awstypes.ModelCacheSetting(v)
	}

	return apiObject
}

func expandModelInferenceExecutionConfig(tfList []any) *awstypes.InferenceExecutionConfig {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)
	apiObject := &awstypes.InferenceExecutionConfig{
		Mode: awstypes.InferenceExecutionMode(m[names.AttrMode].(string)),
	}

	return apiObject
}

func flattenVPCConfigResponse(apiObject *awstypes.VpcConfig) []map[string]any {
	if apiObject == nil {
		return []map[string]any{}
	}

	tfMap := map[string]any{
		names.AttrSecurityGroupIDs: flex.FlattenStringValueSet(apiObject.SecurityGroupIds),
		names.AttrSubnets:          flex.FlattenStringValueSet(apiObject.Subnets),
	}

	return []map[string]any{tfMap}
}

func flattenContainer(apiObject *awstypes.ContainerDefinition) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrMode: apiObject.Mode,
	}

	if apiObject.Image != nil {
		tfMap["image"] = aws.ToString(apiObject.Image)
	}
	if apiObject.ContainerHostname != nil {
		tfMap["container_hostname"] = aws.ToString(apiObject.ContainerHostname)
	}
	if apiObject.ModelDataUrl != nil {
		tfMap["model_data_url"] = aws.ToString(apiObject.ModelDataUrl)
	}
	if apiObject.ModelDataSource != nil {
		tfMap["model_data_source"] = flattenModelDataSource(apiObject.ModelDataSource)
	}
	if len(apiObject.AdditionalModelDataSources) > 0 {
		tfMap["additional_model_data_source"] = flattenAdditionalModelDataSources(apiObject.AdditionalModelDataSources)
	}
	if apiObject.ModelPackageName != nil {
		tfMap["model_package_name"] = aws.ToString(apiObject.ModelPackageName)
	}
	if apiObject.Environment != nil {
		tfMap[names.AttrEnvironment] = aws.StringMap(apiObject.Environment)
	}
	if apiObject.ImageConfig != nil {
		tfMap["image_config"] = flattenImageConfig(apiObject.ImageConfig)
	}
	if apiObject.InferenceSpecificationName != nil {
		tfMap["inference_specification_name"] = aws.ToString(apiObject.InferenceSpecificationName)
	}
	if apiObject.MultiModelConfig != nil {
		tfMap["multi_model_config"] = flattenMultiModelConfig(apiObject.MultiModelConfig)
	}

	return []any{tfMap}
}

func flattenModelDataSource(apiObject *awstypes.ModelDataSource) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)
	if apiObject.S3DataSource != nil {
		tfMap["s3_data_source"] = flattenS3ModelDataSource(apiObject.S3DataSource)
	}

	return []any{tfMap}
}

func flattenAdditionalModelDataSources(apiObjects []awstypes.AdditionalModelDataSource) []any {
	tfList := make([]any, 0, len(apiObjects))
	for _, obj := range apiObjects {
		tfList = append(tfList, flattenAdditionalModelDataSource(obj))
	}

	return tfList
}

func flattenAdditionalModelDataSource(apiObject awstypes.AdditionalModelDataSource) map[string]any {
	tfMap := make(map[string]any)
	if apiObject.ChannelName != nil {
		tfMap["channel_name"] = aws.ToString(apiObject.ChannelName)
	}
	if apiObject.S3DataSource != nil {
		tfMap["s3_data_source"] = flattenS3ModelDataSource(apiObject.S3DataSource)
	}

	return tfMap
}

func flattenS3ModelDataSource(apiObject *awstypes.S3ModelDataSource) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"s3_data_type":     apiObject.S3DataType,
		"compression_type": apiObject.CompressionType,
	}

	if apiObject.ModelAccessConfig != nil {
		tfMap["model_access_config"] = flattenModelAccessConfig(apiObject.ModelAccessConfig)
	}
	if apiObject.S3Uri != nil {
		tfMap["s3_uri"] = aws.ToString(apiObject.S3Uri)
	}

	return []any{tfMap}
}

func flattenImageConfig(apiObject *awstypes.ImageConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"repository_access_mode": apiObject.RepositoryAccessMode,
	}
	if v := flattenRepositoryAuthConfig(apiObject.RepositoryAuthConfig); len(v) > 0 {
		tfMap["repository_auth_config"] = []any{v}
	}

	return []any{tfMap}
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

func flattenContainers(apiObjects []awstypes.ContainerDefinition) []any {
	tfList := make([]any, 0, len(apiObjects))
	for _, obj := range apiObjects {
		tfList = append(tfList, flattenContainer(&obj)[0].(map[string]any))
	}
	return tfList
}

func flattenModelAccessConfig(apiObject *awstypes.ModelAccessConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"accept_eula": aws.ToBool(apiObject.AcceptEula),
	}

	return []any{tfMap}
}

func flattenMultiModelConfig(apiObject *awstypes.MultiModelConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := make(map[string]any)
	if apiObject.ModelCacheSetting != "" {
		tfMap["model_cache_setting"] = apiObject.ModelCacheSetting
	}

	return []any{tfMap}
}

func flattenModelInferenceExecutionConfig(apiObject *awstypes.InferenceExecutionConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrMode: apiObject.Mode,
	}

	return []any{tfMap}
}
