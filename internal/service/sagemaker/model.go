// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
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

// @SDKResource("aws_sagemaker_model", name="Model")
// @Tags(identifierAttribute="arn")
func ResourceModel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceModelCreate,
		ReadWithoutTimeout:   resourceModelRead,
		UpdateWithoutTimeout: resourceModelUpdate,
		DeleteWithoutTimeout: resourceModelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
						"environment": {
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
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.RepositoryAccessMode_Values(), false),
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
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      sagemaker.ContainerModeSingleModel,
							ValidateFunc: validation.StringInSlice(sagemaker.ContainerMode_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.S3ModelDataType_Values(), false),
												},
												"compression_type": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.ModelCompressionType_Values(), false),
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
			"enable_network_isolation": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"execution_role_arn": {
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
						"mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.InferenceExecutionMode_Values(), false),
						},
					},
				},
			},
			"name": {
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
						"environment": {
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
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.RepositoryAccessMode_Values(), false),
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
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      sagemaker.ContainerModeSingleModel,
							ValidateFunc: validation.StringInSlice(sagemaker.ContainerMode_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.S3ModelDataType_Values(), false),
												},
												"compression_type": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.ModelCompressionType_Values(), false),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 16,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceModelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = id.UniqueId()
	}

	createOpts := &sagemaker.CreateModelInput{
		ModelName: aws.String(name),
		Tags:      getTagsIn(ctx),
	}

	if v, ok := d.GetOk("primary_container"); ok {
		createOpts.PrimaryContainer = expandContainer(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("container"); ok {
		createOpts.Containers = expandContainers(v.([]interface{}))
	}

	if v, ok := d.GetOk("execution_role_arn"); ok {
		createOpts.ExecutionRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		createOpts.VpcConfig = expandVPCConfigRequest(v.([]interface{}))
	}

	if v, ok := d.GetOk("enable_network_isolation"); ok {
		createOpts.EnableNetworkIsolation = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("inference_execution_config"); ok {
		createOpts.InferenceExecutionConfig = expandModelInferenceExecutionConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] SageMaker model create config: %#v", *createOpts)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.CreateModelWithContext(ctx, createOpts)
	}, "ValidationException")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker model: %s", err)
	}
	d.SetId(name)

	return append(diags, resourceModelRead(ctx, d, meta)...)
}

func expandVPCConfigRequest(l []interface{}) *sagemaker.VpcConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &sagemaker.VpcConfig{
		SecurityGroupIds: flex.ExpandStringSet(m["security_group_ids"].(*schema.Set)),
		Subnets:          flex.ExpandStringSet(m["subnets"].(*schema.Set)),
	}
}

func resourceModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	request := &sagemaker.DescribeModelInput{
		ModelName: aws.String(d.Id()),
	}

	model, err := conn.DescribeModelWithContext(ctx, request)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, "ValidationException") {
			log.Printf("[INFO] unable to find the sagemaker model resource and therefore it is removed from the state: %s", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker model %s: %s", d.Id(), err)
	}

	arn := aws.StringValue(model.ModelArn)
	d.Set("arn", arn)
	d.Set("name", model.ModelName)
	d.Set("execution_role_arn", model.ExecutionRoleArn)
	d.Set("enable_network_isolation", model.EnableNetworkIsolation)

	if err := d.Set("primary_container", flattenContainer(model.PrimaryContainer)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting primary_container: %s", err)
	}

	if err := d.Set("container", flattenContainers(model.Containers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting container: %s", err)
	}

	if err := d.Set("vpc_config", flattenVPCConfigResponse(model.VpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
	}

	if err := d.Set("inference_execution_config", flattenModelInferenceExecutionConfig(model.InferenceExecutionConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting inference_execution_config: %s", err)
	}

	return diags
}

func flattenVPCConfigResponse(vpcConfig *sagemaker.VpcConfig) []map[string]interface{} {
	if vpcConfig == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"security_group_ids": flex.FlattenStringSet(vpcConfig.SecurityGroupIds),
		"subnets":            flex.FlattenStringSet(vpcConfig.Subnets),
	}

	return []map[string]interface{}{m}
}

func resourceModelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceModelRead(ctx, d, meta)...)
}

func resourceModelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	deleteOpts := &sagemaker.DeleteModelInput{
		ModelName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker model: %s", d.Id())

	err := retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteModelWithContext(ctx, deleteOpts)
		if err == nil {
			return nil
		}

		if tfawserr.ErrCodeEquals(err, "ResourceNotFound") {
			return retry.RetryableError(err)
		}
		return retry.NonRetryableError(err)
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteModelWithContext(ctx, deleteOpts)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting sagemaker model: %s", err)
	}
	return diags
}

func expandContainer(m map[string]interface{}) *sagemaker.ContainerDefinition {
	container := sagemaker.ContainerDefinition{}

	if v, ok := m["image"]; ok && v.(string) != "" {
		container.Image = aws.String(v.(string))
	}

	if v, ok := m["mode"]; ok && v.(string) != "" {
		container.Mode = aws.String(v.(string))
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
		container.ModelDataSource = expandModelDataSource(v.([]interface{}))
	}
	if v, ok := m["environment"].(map[string]interface{}); ok && len(v) > 0 {
		container.Environment = flex.ExpandStringMap(v)
	}

	if v, ok := m["image_config"]; ok {
		container.ImageConfig = expandModelImageConfig(v.([]interface{}))
	}

	return &container
}

func expandModelDataSource(l []interface{}) *sagemaker.ModelDataSource {
	if len(l) == 0 {
		return nil
	}

	modelDataSource := sagemaker.ModelDataSource{}

	m := l[0].(map[string]interface{})

	if v, ok := m["s3_data_source"]; ok {
		modelDataSource.S3DataSource = expandS3ModelDataSource(v.([]interface{}))
	}

	return &modelDataSource
}

func expandS3ModelDataSource(l []interface{}) *sagemaker.S3ModelDataSource {
	if len(l) == 0 {
		return nil
	}

	s3ModelDataSource := sagemaker.S3ModelDataSource{}

	m := l[0].(map[string]interface{})

	if v, ok := m["s3_uri"]; ok && v.(string) != "" {
		s3ModelDataSource.S3Uri = aws.String(v.(string))
	}
	if v, ok := m["s3_data_type"]; ok && v.(string) != "" {
		s3ModelDataSource.S3DataType = aws.String(v.(string))
	}
	if v, ok := m["compression_type"]; ok && v.(string) != "" {
		s3ModelDataSource.CompressionType = aws.String(v.(string))
	}

	return &s3ModelDataSource
}

func expandModelImageConfig(l []interface{}) *sagemaker.ImageConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	imageConfig := &sagemaker.ImageConfig{
		RepositoryAccessMode: aws.String(m["repository_access_mode"].(string)),
	}

	if v, ok := m["repository_auth_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		imageConfig.RepositoryAuthConfig = expandRepositoryAuthConfig(v[0].(map[string]interface{}))
	}

	return imageConfig
}

func expandRepositoryAuthConfig(tfMap map[string]interface{}) *sagemaker.RepositoryAuthConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &sagemaker.RepositoryAuthConfig{}

	if v, ok := tfMap["repository_credentials_provider_arn"].(string); ok && v != "" {
		apiObject.RepositoryCredentialsProviderArn = aws.String(v)
	}

	return apiObject
}

func expandContainers(a []interface{}) []*sagemaker.ContainerDefinition {
	containers := make([]*sagemaker.ContainerDefinition, 0, len(a))

	for _, m := range a {
		containers = append(containers, expandContainer(m.(map[string]interface{})))
	}

	return containers
}

func flattenContainer(container *sagemaker.ContainerDefinition) []interface{} {
	if container == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})

	if container.Image != nil {
		cfg["image"] = aws.StringValue(container.Image)
	}

	if container.Mode != nil {
		cfg["mode"] = aws.StringValue(container.Mode)
	}

	if container.ContainerHostname != nil {
		cfg["container_hostname"] = aws.StringValue(container.ContainerHostname)
	}
	if container.ModelDataUrl != nil {
		cfg["model_data_url"] = aws.StringValue(container.ModelDataUrl)
	}
	if container.ModelDataSource != nil {
		cfg["model_data_source"] = flattenModelDataSource(container.ModelDataSource)
	}
	if container.ModelPackageName != nil {
		cfg["model_package_name"] = aws.StringValue(container.ModelPackageName)
	}
	if container.Environment != nil {
		cfg["environment"] = aws.StringValueMap(container.Environment)
	}

	if container.ImageConfig != nil {
		cfg["image_config"] = flattenImageConfig(container.ImageConfig)
	}

	return []interface{}{cfg}
}

func flattenModelDataSource(modelDataSource *sagemaker.ModelDataSource) []interface{} {
	if modelDataSource == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})

	if modelDataSource.S3DataSource != nil {
		cfg["s3_data_source"] = flattenS3ModelDataSource(modelDataSource.S3DataSource)
	}

	return []interface{}{cfg}
}

func flattenS3ModelDataSource(s3ModelDataSource *sagemaker.S3ModelDataSource) []interface{} {
	if s3ModelDataSource == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})

	if s3ModelDataSource.S3Uri != nil {
		cfg["s3_uri"] = aws.StringValue(s3ModelDataSource.S3Uri)
	}
	if s3ModelDataSource.S3DataType != nil {
		cfg["s3_data_type"] = aws.StringValue(s3ModelDataSource.S3DataType)
	}
	if s3ModelDataSource.CompressionType != nil {
		cfg["compression_type"] = aws.StringValue(s3ModelDataSource.CompressionType)
	}

	return []interface{}{cfg}
}

func flattenImageConfig(imageConfig *sagemaker.ImageConfig) []interface{} {
	if imageConfig == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})

	cfg["repository_access_mode"] = aws.StringValue(imageConfig.RepositoryAccessMode)

	if tfMap := flattenRepositoryAuthConfig(imageConfig.RepositoryAuthConfig); len(tfMap) > 0 {
		cfg["repository_auth_config"] = []interface{}{tfMap}
	}

	return []interface{}{cfg}
}

func flattenRepositoryAuthConfig(apiObject *sagemaker.RepositoryAuthConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.RepositoryCredentialsProviderArn; v != nil {
		tfMap["repository_credentials_provider_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenContainers(containers []*sagemaker.ContainerDefinition) []interface{} {
	fContainers := make([]interface{}, 0, len(containers))
	for _, container := range containers {
		fContainers = append(fContainers, flattenContainer(container)[0].(map[string]interface{}))
	}
	return fContainers
}

func expandModelInferenceExecutionConfig(l []interface{}) *sagemaker.InferenceExecutionConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.InferenceExecutionConfig{
		Mode: aws.String(m["mode"].(string)),
	}

	return config
}

func flattenModelInferenceExecutionConfig(config *sagemaker.InferenceExecutionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})

	cfg["mode"] = aws.StringValue(config.Mode)

	return []interface{}{cfg}
}
