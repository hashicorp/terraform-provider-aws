// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_endpoint_configuration", name="Endpoint Configuration")
// @Tags(identifierAttribute="arn")
func ResourceEndpointConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointConfigurationCreate,
		ReadWithoutTimeout:   resourceEndpointConfigurationRead,
		UpdateWithoutTimeout: resourceEndpointConfigurationUpdate,
		DeleteWithoutTimeout: resourceEndpointConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"async_inference_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_concurrent_invocations_per_instance": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 1000),
									},
								},
							},
						},
						"output_config": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKMSKeyID: {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
									"notification_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_topic": {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: verify.ValidARN,
												},
												"include_inference_response_in": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(sagemaker.AsyncNotificationTopicTypes_Values(), false),
													},
												},
												"success_topic": {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									"s3_failure_path": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
											validation.StringLenBetween(1, 512),
										),
									},
									"s3_output_path": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
											validation.StringLenBetween(1, 512),
										),
									},
								},
							},
						},
					},
				},
			},
			"data_capture_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capture_content_type_header": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"csv_content_types": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 10,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.All(
												validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z])*\/[0-9A-Za-z](-*[0-9A-Za-z.])*`), ""),
												validation.StringLenBetween(1, 256),
											),
										},
										Optional: true,
										ForceNew: true,
									},
									"json_content_types": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 10,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.All(
												validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z])*\/[0-9A-Za-z](-*[0-9A-Za-z.])*`), ""),
												validation.StringLenBetween(1, 256),
											),
										},
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"capture_options": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 2,
							MinItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capture_mode": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.CaptureMode_Values(), false),
									},
								},
							},
						},
						"destination_s3_uri": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
								validation.StringLenBetween(1, 512),
							),
						},
						"enable_capture": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"initial_sampling_percentage": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 100),
						},
						names.AttrKMSKeyID: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrKMSKeyARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validPrefix,
			},
			"production_variants": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accelerator_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ProductionVariantAcceleratorType_Values(), false),
						},
						"container_startup_health_check_timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(60, 3600),
						},
						"core_dump_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_s3_uri": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
											validation.StringLenBetween(1, 512),
										),
									},
									names.AttrKMSKeyID: {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"enable_ssm_access": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"inference_ami_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ProductionVariantInferenceAmiVersion_Values(), false),
						},
						"initial_instance_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"initial_variant_weight": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.FloatAtLeast(0),
							Default:      1,
						},
						names.AttrInstanceType: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ProductionVariantInstanceType_Values(), false),
						},
						"model_data_download_timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(60, 3600),
						},
						"model_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"routing_config": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"routing_strategy": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.RoutingStrategy_Values(), false),
									},
								},
							},
						},
						"serverless_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_concurrency": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 200),
									},
									"memory_size_in_mb": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntInSlice([]int{1024, 2048, 3072, 4096, 5120, 6144}),
									},
									"provisioned_concurrency": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 200),
									},
								},
							},
						},
						"variant_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"volume_size_in_gb": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 512),
						},
					},
				},
			},
			"shadow_production_variants": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accelerator_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ProductionVariantAcceleratorType_Values(), false),
						},
						"container_startup_health_check_timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(60, 3600),
						},
						"core_dump_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_s3_uri": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
											validation.StringLenBetween(1, 512),
										),
									},
									names.AttrKMSKeyID: {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"enable_ssm_access": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"inference_ami_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ProductionVariantInferenceAmiVersion_Values(), false),
						},
						"initial_instance_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"initial_variant_weight": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.FloatAtLeast(0),
							Default:      1,
						},
						names.AttrInstanceType: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ProductionVariantInstanceType_Values(), false),
						},
						"model_data_download_timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(60, 3600),
						},
						"model_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"routing_config": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"routing_strategy": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.RoutingStrategy_Values(), false),
									},
								},
							},
						},
						"serverless_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_concurrency": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 200),
									},
									"memory_size_in_mb": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntInSlice([]int{1024, 2048, 3072, 4096, 5120, 6144}),
									},
									"provisioned_concurrency": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 200),
									},
								},
							},
						},
						"variant_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"volume_size_in_gb": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 512),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEndpointConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))

	createOpts := &sagemaker.CreateEndpointConfigInput{
		EndpointConfigName: aws.String(name),
		ProductionVariants: expandProductionVariants(d.Get("production_variants").([]interface{})),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		createOpts.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("shadow_production_variants"); ok && len(v.([]interface{})) > 0 {
		createOpts.ShadowProductionVariants = expandProductionVariants(v.([]interface{}))
	}

	if v, ok := d.GetOk("data_capture_config"); ok {
		createOpts.DataCaptureConfig = expandDataCaptureConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("async_inference_config"); ok {
		createOpts.AsyncInferenceConfig = expandEndpointConfigAsyncInferenceConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] SageMaker Endpoint Configuration create config: %#v", *createOpts)
	_, err := conn.CreateEndpointConfigWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Endpoint Configuration: %s", err)
	}
	d.SetId(name)

	return append(diags, resourceEndpointConfigurationRead(ctx, d, meta)...)
}

func resourceEndpointConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	endpointConfig, err := FindEndpointConfigByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Endpoint Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Endpoint Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, endpointConfig.EndpointConfigArn)
	d.Set(names.AttrName, endpointConfig.EndpointConfigName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.StringValue(endpointConfig.EndpointConfigName)))
	d.Set(names.AttrKMSKeyARN, endpointConfig.KmsKeyId)

	if err := d.Set("production_variants", flattenProductionVariants(endpointConfig.ProductionVariants)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting production_variants for SageMaker Endpoint Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("shadow_production_variants", flattenProductionVariants(endpointConfig.ShadowProductionVariants)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting shadow_production_variants for SageMaker Endpoint Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_capture_config", flattenDataCaptureConfig(endpointConfig.DataCaptureConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_capture_config for SageMaker Endpoint Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("async_inference_config", flattenEndpointConfigAsyncInferenceConfig(endpointConfig.AsyncInferenceConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting async_inference_config for SageMaker Endpoint Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceEndpointConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceEndpointConfigurationRead(ctx, d, meta)...)
}

func resourceEndpointConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	deleteOpts := &sagemaker.DeleteEndpointConfigInput{
		EndpointConfigName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Endpoint Configuration: %s", d.Id())

	_, err := conn.DeleteEndpointConfigWithContext(ctx, deleteOpts)

	if tfawserr.ErrMessageContains(err, "ValidationException", "Could not find endpoint configuration") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Endpoint Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func expandProductionVariants(configured []interface{}) []*sagemaker.ProductionVariant {
	containers := make([]*sagemaker.ProductionVariant, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &sagemaker.ProductionVariant{
			ModelName: aws.String(data["model_name"].(string)),
		}

		if v, ok := data["initial_instance_count"].(int); ok && v > 0 {
			l.InitialInstanceCount = aws.Int64(int64(v))
		}

		if v, ok := data["container_startup_health_check_timeout_in_seconds"].(int); ok && v > 0 {
			l.ContainerStartupHealthCheckTimeoutInSeconds = aws.Int64(int64(v))
		}

		if v, ok := data["model_data_download_timeout_in_seconds"].(int); ok && v > 0 {
			l.ModelDataDownloadTimeoutInSeconds = aws.Int64(int64(v))
		}

		if v, ok := data["volume_size_in_gb"].(int); ok && v > 0 {
			l.VolumeSizeInGB = aws.Int64(int64(v))
		}

		if v, ok := data[names.AttrInstanceType].(string); ok && v != "" {
			l.InstanceType = aws.String(v)
		}

		if v, ok := data["variant_name"].(string); ok && v != "" {
			l.VariantName = aws.String(v)
		} else {
			l.VariantName = aws.String(id.UniqueId())
		}

		if v, ok := data["initial_variant_weight"]; ok {
			l.InitialVariantWeight = aws.Float64(v.(float64))
		}

		if v, ok := data["accelerator_type"].(string); ok && v != "" {
			l.AcceleratorType = aws.String(v)
		}

		if v, ok := data["routing_config"].([]interface{}); ok && len(v) > 0 {
			l.RoutingConfig = expandRoutingConfig(v)
		}

		if v, ok := data["serverless_config"].([]interface{}); ok && len(v) > 0 {
			l.ServerlessConfig = expandServerlessConfig(v)
		}

		if v, ok := data["core_dump_config"].([]interface{}); ok && len(v) > 0 {
			l.CoreDumpConfig = expandCoreDumpConfig(v)
		}

		if v, ok := data["enable_ssm_access"].(bool); ok {
			l.EnableSSMAccess = aws.Bool(v)
		}

		if v, ok := data["inference_ami_version"].(string); ok && v != "" {
			l.InferenceAmiVersion = aws.String(v)
		}

		containers = append(containers, l)
	}

	return containers
}

func flattenProductionVariants(list []*sagemaker.ProductionVariant) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))

	for _, i := range list {
		l := map[string]interface{}{
			"accelerator_type":       aws.StringValue(i.AcceleratorType),
			"initial_variant_weight": aws.Float64Value(i.InitialVariantWeight),
			"model_name":             aws.StringValue(i.ModelName),
			"variant_name":           aws.StringValue(i.VariantName),
		}

		if i.InitialInstanceCount != nil {
			l["initial_instance_count"] = aws.Int64Value(i.InitialInstanceCount)
		}

		if i.ContainerStartupHealthCheckTimeoutInSeconds != nil {
			l["container_startup_health_check_timeout_in_seconds"] = aws.Int64Value(i.ContainerStartupHealthCheckTimeoutInSeconds)
		}

		if i.ModelDataDownloadTimeoutInSeconds != nil {
			l["model_data_download_timeout_in_seconds"] = aws.Int64Value(i.ModelDataDownloadTimeoutInSeconds)
		}

		if i.VolumeSizeInGB != nil {
			l["volume_size_in_gb"] = aws.Int64Value(i.VolumeSizeInGB)
		}

		if i.InstanceType != nil {
			l[names.AttrInstanceType] = aws.StringValue(i.InstanceType)
		}

		if i.RoutingConfig != nil {
			l["routing_config"] = flattenRoutingConfig(i.RoutingConfig)
		}

		if i.ServerlessConfig != nil {
			l["serverless_config"] = flattenServerlessConfig(i.ServerlessConfig)
		}

		if i.CoreDumpConfig != nil {
			l["core_dump_config"] = flattenCoreDumpConfig(i.CoreDumpConfig)
		}

		if i.EnableSSMAccess != nil {
			l["enable_ssm_access"] = aws.BoolValue(i.EnableSSMAccess)
		}

		if i.InferenceAmiVersion != nil {
			l["inference_ami_version"] = aws.StringValue(i.InferenceAmiVersion)
		}

		result = append(result, l)
	}
	return result
}

func expandDataCaptureConfig(configured []interface{}) *sagemaker.DataCaptureConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.DataCaptureConfig{
		InitialSamplingPercentage: aws.Int64(int64(m["initial_sampling_percentage"].(int))),
		DestinationS3Uri:          aws.String(m["destination_s3_uri"].(string)),
		CaptureOptions:            expandCaptureOptions(m["capture_options"].([]interface{})),
	}

	if v, ok := m["enable_capture"]; ok {
		c.EnableCapture = aws.Bool(v.(bool))
	}

	if v, ok := m[names.AttrKMSKeyID].(string); ok && v != "" {
		c.KmsKeyId = aws.String(v)
	}

	if v, ok := m["capture_content_type_header"].([]interface{}); ok && (len(v) > 0) {
		c.CaptureContentTypeHeader = expandCaptureContentTypeHeader(v[0].(map[string]interface{}))
	}

	return c
}

func flattenDataCaptureConfig(dataCaptureConfig *sagemaker.DataCaptureConfig) []map[string]interface{} {
	if dataCaptureConfig == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"initial_sampling_percentage": aws.Int64Value(dataCaptureConfig.InitialSamplingPercentage),
		"destination_s3_uri":          aws.StringValue(dataCaptureConfig.DestinationS3Uri),
		"capture_options":             flattenCaptureOptions(dataCaptureConfig.CaptureOptions),
	}

	if dataCaptureConfig.EnableCapture != nil {
		cfg["enable_capture"] = aws.BoolValue(dataCaptureConfig.EnableCapture)
	}

	if dataCaptureConfig.KmsKeyId != nil {
		cfg[names.AttrKMSKeyID] = aws.StringValue(dataCaptureConfig.KmsKeyId)
	}

	if dataCaptureConfig.CaptureContentTypeHeader != nil {
		cfg["capture_content_type_header"] = flattenCaptureContentTypeHeader(dataCaptureConfig.CaptureContentTypeHeader)
	}

	return []map[string]interface{}{cfg}
}

func expandCaptureOptions(configured []interface{}) []*sagemaker.CaptureOption {
	containers := make([]*sagemaker.CaptureOption, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &sagemaker.CaptureOption{
			CaptureMode: aws.String(data["capture_mode"].(string)),
		}
		containers = append(containers, l)
	}

	return containers
}

func flattenCaptureOptions(list []*sagemaker.CaptureOption) []map[string]interface{} {
	containers := make([]map[string]interface{}, 0, len(list))

	for _, lRaw := range list {
		captureOption := make(map[string]interface{})
		captureOption["capture_mode"] = aws.StringValue(lRaw.CaptureMode)
		containers = append(containers, captureOption)
	}

	return containers
}

func expandCaptureContentTypeHeader(m map[string]interface{}) *sagemaker.CaptureContentTypeHeader {
	c := &sagemaker.CaptureContentTypeHeader{}

	if v, ok := m["csv_content_types"].(*schema.Set); ok && v.Len() > 0 {
		c.CsvContentTypes = flex.ExpandStringSet(v)
	}

	if v, ok := m["json_content_types"].(*schema.Set); ok && v.Len() > 0 {
		c.JsonContentTypes = flex.ExpandStringSet(v)
	}

	return c
}

func flattenCaptureContentTypeHeader(contentTypeHeader *sagemaker.CaptureContentTypeHeader) []map[string]interface{} {
	if contentTypeHeader == nil {
		return []map[string]interface{}{}
	}

	l := make(map[string]interface{})

	if contentTypeHeader.CsvContentTypes != nil {
		l["csv_content_types"] = flex.FlattenStringSet(contentTypeHeader.CsvContentTypes)
	}

	if contentTypeHeader.JsonContentTypes != nil {
		l["json_content_types"] = flex.FlattenStringSet(contentTypeHeader.JsonContentTypes)
	}

	return []map[string]interface{}{l}
}

func expandEndpointConfigAsyncInferenceConfig(configured []interface{}) *sagemaker.AsyncInferenceConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.AsyncInferenceConfig{}

	if v, ok := m["client_config"].([]interface{}); ok && len(v) > 0 {
		c.ClientConfig = expandEndpointConfigClientConfig(v)
	}

	if v, ok := m["output_config"].([]interface{}); ok && len(v) > 0 {
		c.OutputConfig = expandEndpointConfigOutputConfig(v)
	}

	return c
}

func expandEndpointConfigClientConfig(configured []interface{}) *sagemaker.AsyncInferenceClientConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.AsyncInferenceClientConfig{}

	if v, ok := m["max_concurrent_invocations_per_instance"]; ok {
		c.MaxConcurrentInvocationsPerInstance = aws.Int64(int64(v.(int)))
	}

	return c
}

func expandEndpointConfigOutputConfig(configured []interface{}) *sagemaker.AsyncInferenceOutputConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.AsyncInferenceOutputConfig{
		S3OutputPath: aws.String(m["s3_output_path"].(string)),
	}

	if v, ok := m[names.AttrKMSKeyID].(string); ok && v != "" {
		c.KmsKeyId = aws.String(v)
	}

	if v, ok := m["s3_failure_path"].(string); ok && v != "" {
		c.S3FailurePath = aws.String(v)
	}

	if v, ok := m["notification_config"].([]interface{}); ok && len(v) > 0 {
		c.NotificationConfig = expandEndpointConfigNotificationConfig(v)
	}

	return c
}

func expandEndpointConfigNotificationConfig(configured []interface{}) *sagemaker.AsyncInferenceNotificationConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.AsyncInferenceNotificationConfig{}

	if v, ok := m["error_topic"].(string); ok && v != "" {
		c.ErrorTopic = aws.String(v)
	}

	if v, ok := m["success_topic"].(string); ok && v != "" {
		c.SuccessTopic = aws.String(v)
	}

	if v, ok := m["include_inference_response_in"].(*schema.Set); ok && v.Len() > 0 {
		c.IncludeInferenceResponseIn = flex.ExpandStringSet(v)
	}

	return c
}

func expandRoutingConfig(configured []interface{}) *sagemaker.ProductionVariantRoutingConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.ProductionVariantRoutingConfig{}

	if v, ok := m["routing_strategy"].(string); ok && v != "" {
		c.RoutingStrategy = aws.String(v)
	}

	return c
}

func expandServerlessConfig(configured []interface{}) *sagemaker.ProductionVariantServerlessConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.ProductionVariantServerlessConfig{}

	if v, ok := m["max_concurrency"].(int); ok {
		c.MaxConcurrency = aws.Int64(int64(v))
	}

	if v, ok := m["memory_size_in_mb"].(int); ok {
		c.MemorySizeInMB = aws.Int64(int64(v))
	}

	if v, ok := m["provisioned_concurrency"].(int); ok && v > 0 {
		c.ProvisionedConcurrency = aws.Int64(int64(v))
	}

	return c
}

func expandCoreDumpConfig(configured []interface{}) *sagemaker.ProductionVariantCoreDumpConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.ProductionVariantCoreDumpConfig{}

	if v, ok := m["destination_s3_uri"].(string); ok {
		c.DestinationS3Uri = aws.String(v)
	}

	if v, ok := m[names.AttrKMSKeyID].(string); ok {
		c.KmsKeyId = aws.String(v)
	}

	return c
}

func flattenEndpointConfigAsyncInferenceConfig(config *sagemaker.AsyncInferenceConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{}

	if config.ClientConfig != nil {
		cfg["client_config"] = flattenEndpointConfigClientConfig(config.ClientConfig)
	}

	if config.OutputConfig != nil {
		cfg["output_config"] = flattenEndpointConfigOutputConfig(config.OutputConfig)
	}

	return []map[string]interface{}{cfg}
}

func flattenEndpointConfigClientConfig(config *sagemaker.AsyncInferenceClientConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{}

	if config.MaxConcurrentInvocationsPerInstance != nil {
		cfg["max_concurrent_invocations_per_instance"] = aws.Int64Value(config.MaxConcurrentInvocationsPerInstance)
	}

	return []map[string]interface{}{cfg}
}

func flattenEndpointConfigOutputConfig(config *sagemaker.AsyncInferenceOutputConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"s3_output_path": aws.StringValue(config.S3OutputPath),
	}

	if config.KmsKeyId != nil {
		cfg[names.AttrKMSKeyID] = aws.StringValue(config.KmsKeyId)
	}

	if config.NotificationConfig != nil {
		cfg["notification_config"] = flattenEndpointConfigNotificationConfig(config.NotificationConfig)
	}

	if config.S3FailurePath != nil {
		cfg["s3_failure_path"] = aws.StringValue(config.S3FailurePath)
	}

	return []map[string]interface{}{cfg}
}

func flattenEndpointConfigNotificationConfig(config *sagemaker.AsyncInferenceNotificationConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{}

	if config.ErrorTopic != nil {
		cfg["error_topic"] = aws.StringValue(config.ErrorTopic)
	}

	if config.SuccessTopic != nil {
		cfg["success_topic"] = aws.StringValue(config.SuccessTopic)
	}

	if config.IncludeInferenceResponseIn != nil {
		cfg["include_inference_response_in"] = flex.FlattenStringSet(config.IncludeInferenceResponseIn)
	}

	return []map[string]interface{}{cfg}
}

func flattenRoutingConfig(config *sagemaker.ProductionVariantRoutingConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{}

	if config.RoutingStrategy != nil {
		cfg["routing_strategy"] = aws.StringValue(config.RoutingStrategy)
	}

	return []map[string]interface{}{cfg}
}

func flattenServerlessConfig(config *sagemaker.ProductionVariantServerlessConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{}

	if config.MaxConcurrency != nil {
		cfg["max_concurrency"] = aws.Int64Value(config.MaxConcurrency)
	}

	if config.MemorySizeInMB != nil {
		cfg["memory_size_in_mb"] = aws.Int64Value(config.MemorySizeInMB)
	}

	if config.ProvisionedConcurrency != nil {
		cfg["provisioned_concurrency"] = aws.Int64Value(config.ProvisionedConcurrency)
	}

	return []map[string]interface{}{cfg}
}

func flattenCoreDumpConfig(config *sagemaker.ProductionVariantCoreDumpConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{}

	if config.DestinationS3Uri != nil {
		cfg["destination_s3_uri"] = aws.StringValue(config.DestinationS3Uri)
	}

	if config.KmsKeyId != nil {
		cfg[names.AttrKMSKeyID] = aws.StringValue(config.KmsKeyId)
	}

	return []map[string]interface{}{cfg}
}
