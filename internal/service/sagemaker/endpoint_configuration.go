// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_endpoint_configuration", name="Endpoint Configuration")
// @Tags(identifierAttribute="arn")
func resourceEndpointConfiguration() *schema.Resource {
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
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.AsyncNotificationTopicTypes](),
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
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.CaptureMode](),
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
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ProductionVariantAcceleratorType](),
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
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ProductionVariantInferenceAmiVersion](),
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
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ProductionVariantInstanceType](),
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
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.RoutingStrategy](),
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
						"managed_instance_scaling": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_instance_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"min_instance_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									names.AttrStatus: {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ManagedInstanceScalingStatus](),
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
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ProductionVariantAcceleratorType](),
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
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ProductionVariantInferenceAmiVersion](),
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
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ProductionVariantInstanceType](),
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
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.RoutingStrategy](),
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
						"managed_instance_scaling": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_instance_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"min_instance_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									names.AttrStatus: {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ManagedInstanceScalingStatus](),
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

		CustomizeDiff: validateDataCaptureConfigCustomDiff,
	}
}

func validateDataCaptureConfigCustomDiff(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	var diags diag.Diagnostics

	configRaw := d.GetRawConfig()
	if !configRaw.IsKnown() || configRaw.IsNull() {
		return nil
	}

	dataCapturesPath := cty.GetAttrPath("data_capture_config")
	dataCaptures := configRaw.GetAttr("data_capture_config")
	if dataCaptures.IsKnown() && !dataCaptures.IsNull() {
		dataCaptureConfigPlanTimeValidate(dataCapturesPath, dataCaptures, &diags)
	}

	return sdkdiag.DiagnosticsError(diags)
}

func dataCaptureConfigPlanTimeValidate(path cty.Path, dataCaptures cty.Value, diags *diag.Diagnostics) {
	it := dataCaptures.ElementIterator()
	for it.Next() {
		_, dataCapture := it.Element()

		if !dataCapture.IsKnown() {
			break
		}
		if dataCapture.IsNull() {
			break
		}

		captureContentHeaderPath := path.GetAttr("capture_content_type_header")
		captureContentHeaders := dataCapture.GetAttr("capture_content_type_header")

		captureContentTypeHeaderPlanTimeValidate(captureContentHeaderPath, captureContentHeaders, diags)
	}
}

func captureContentTypeHeaderPlanTimeValidate(path cty.Path, captureContentHeaders cty.Value, diags *diag.Diagnostics) {
	it := captureContentHeaders.ElementIterator()
	for it.Next() {
		_, captureContentHeader := it.Element()

		if !captureContentHeader.IsKnown() {
			break
		}
		if captureContentHeader.IsNull() {
			break
		}

		csvContentTypes := captureContentHeader.GetAttr("csv_content_types")
		if csvContentTypes.IsKnown() && !csvContentTypes.IsNull() {
			break
		}

		jsonContentTypes := captureContentHeader.GetAttr("json_content_types")
		if jsonContentTypes.IsKnown() && !jsonContentTypes.IsNull() {
			break
		}

		*diags = append(*diags, errs.NewAtLeastOneOfChildrenError(path,
			cty.GetAttrPath("csv_content_types"),
			cty.GetAttrPath("json_content_types"),
		))
	}
}

func resourceEndpointConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))

	createOpts := &sagemaker.CreateEndpointConfigInput{
		EndpointConfigName: aws.String(name),
		ProductionVariants: expandProductionVariants(d.Get("production_variants").([]any)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		createOpts.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("shadow_production_variants"); ok && len(v.([]any)) > 0 {
		createOpts.ShadowProductionVariants = expandProductionVariants(v.([]any))
	}

	if v, ok := d.GetOk("data_capture_config"); ok {
		createOpts.DataCaptureConfig = expandDataCaptureConfig(v.([]any))
	}

	if v, ok := d.GetOk("async_inference_config"); ok {
		createOpts.AsyncInferenceConfig = expandEndpointConfigAsyncInferenceConfig(v.([]any))
	}

	log.Printf("[DEBUG] SageMaker AI Endpoint Configuration create config: %#v", *createOpts)
	_, err := conn.CreateEndpointConfig(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Endpoint Configuration: %s", err)
	}
	d.SetId(name)

	return append(diags, resourceEndpointConfigurationRead(ctx, d, meta)...)
}

func resourceEndpointConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	endpointConfig, err := findEndpointConfigByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Endpoint Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Endpoint Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, endpointConfig.EndpointConfigArn)
	d.Set(names.AttrName, endpointConfig.EndpointConfigName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(endpointConfig.EndpointConfigName)))
	d.Set(names.AttrKMSKeyARN, endpointConfig.KmsKeyId)

	if err := d.Set("production_variants", flattenProductionVariants(endpointConfig.ProductionVariants)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting production_variants for SageMaker AI Endpoint Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("shadow_production_variants", flattenProductionVariants(endpointConfig.ShadowProductionVariants)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting shadow_production_variants for SageMaker AI Endpoint Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_capture_config", flattenDataCaptureConfig(endpointConfig.DataCaptureConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_capture_config for SageMaker AI Endpoint Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("async_inference_config", flattenEndpointConfigAsyncInferenceConfig(endpointConfig.AsyncInferenceConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting async_inference_config for SageMaker AI Endpoint Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceEndpointConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceEndpointConfigurationRead(ctx, d, meta)...)
}

func resourceEndpointConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	deleteOpts := &sagemaker.DeleteEndpointConfigInput{
		EndpointConfigName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker AI Endpoint Configuration: %s", d.Id())

	_, err := conn.DeleteEndpointConfig(ctx, deleteOpts)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint configuration") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Endpoint Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findEndpointConfigByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeEndpointConfigOutput, error) {
	input := &sagemaker.DescribeEndpointConfigInput{
		EndpointConfigName: aws.String(name),
	}

	output, err := conn.DescribeEndpointConfig(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint configuration") {
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

func expandProductionVariants(configured []any) []awstypes.ProductionVariant {
	containers := make([]awstypes.ProductionVariant, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]any)

		l := awstypes.ProductionVariant{
			ModelName: aws.String(data["model_name"].(string)),
		}

		if v, ok := data["initial_instance_count"].(int); ok && v > 0 {
			l.InitialInstanceCount = aws.Int32(int32(v))
		}

		if v, ok := data["container_startup_health_check_timeout_in_seconds"].(int); ok && v > 0 {
			l.ContainerStartupHealthCheckTimeoutInSeconds = aws.Int32(int32(v))
		}

		if v, ok := data["model_data_download_timeout_in_seconds"].(int); ok && v > 0 {
			l.ModelDataDownloadTimeoutInSeconds = aws.Int32(int32(v))
		}

		if v, ok := data["volume_size_in_gb"].(int); ok && v > 0 {
			l.VolumeSizeInGB = aws.Int32(int32(v))
		}

		if v, ok := data[names.AttrInstanceType].(string); ok && v != "" {
			l.InstanceType = awstypes.ProductionVariantInstanceType(v)
		}

		if v, ok := data["variant_name"].(string); ok && v != "" {
			l.VariantName = aws.String(v)
		} else {
			l.VariantName = aws.String(id.UniqueId())
		}

		if v, ok := data["initial_variant_weight"].(float64); ok {
			l.InitialVariantWeight = aws.Float32(float32(v))
		}

		if v, ok := data["accelerator_type"].(string); ok && v != "" {
			l.AcceleratorType = awstypes.ProductionVariantAcceleratorType(v)
		}

		if v, ok := data["routing_config"].([]any); ok && len(v) > 0 {
			l.RoutingConfig = expandRoutingConfig(v)
		}

		if v, ok := data["serverless_config"].([]any); ok && len(v) > 0 {
			l.ServerlessConfig = expandServerlessConfig(v)
		}

		if v, ok := data["core_dump_config"].([]any); ok && len(v) > 0 {
			l.CoreDumpConfig = expandCoreDumpConfig(v)
		}

		if v, ok := data["enable_ssm_access"].(bool); ok {
			l.EnableSSMAccess = aws.Bool(v)
		}

		if v, ok := data["managed_instance_scaling"].([]any); ok && len(v) > 0 {
			l.ManagedInstanceScaling = expandManagedInstanceScaling(v)
		}

		if v, ok := data["inference_ami_version"].(string); ok && v != "" {
			l.InferenceAmiVersion = awstypes.ProductionVariantInferenceAmiVersion(v)
		}

		containers = append(containers, l)
	}

	return containers
}

func flattenProductionVariants(list []awstypes.ProductionVariant) []map[string]any {
	result := make([]map[string]any, 0, len(list))

	for _, i := range list {
		l := map[string]any{
			"accelerator_type":       i.AcceleratorType,
			names.AttrInstanceType:   i.InstanceType,
			"inference_ami_version":  i.InferenceAmiVersion,
			"initial_variant_weight": aws.ToFloat32(i.InitialVariantWeight),
			"model_name":             aws.ToString(i.ModelName),
			"variant_name":           aws.ToString(i.VariantName),
		}

		if i.InitialInstanceCount != nil {
			l["initial_instance_count"] = aws.ToInt32(i.InitialInstanceCount)
		}

		if i.ContainerStartupHealthCheckTimeoutInSeconds != nil {
			l["container_startup_health_check_timeout_in_seconds"] = aws.ToInt32(i.ContainerStartupHealthCheckTimeoutInSeconds)
		}

		if i.ModelDataDownloadTimeoutInSeconds != nil {
			l["model_data_download_timeout_in_seconds"] = aws.ToInt32(i.ModelDataDownloadTimeoutInSeconds)
		}

		if i.VolumeSizeInGB != nil {
			l["volume_size_in_gb"] = aws.ToInt32(i.VolumeSizeInGB)
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
			l["enable_ssm_access"] = aws.ToBool(i.EnableSSMAccess)
		}

		if i.ManagedInstanceScaling != nil {
			l["managed_instance_scaling"] = flattenManagedInstanceScaling(i.ManagedInstanceScaling)
		}

		result = append(result, l)
	}
	return result
}

func expandDataCaptureConfig(configured []any) *awstypes.DataCaptureConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.DataCaptureConfig{
		InitialSamplingPercentage: aws.Int32(int32(m["initial_sampling_percentage"].(int))),
		DestinationS3Uri:          aws.String(m["destination_s3_uri"].(string)),
		CaptureOptions:            expandCaptureOptions(m["capture_options"].([]any)),
	}

	if v, ok := m["enable_capture"]; ok {
		c.EnableCapture = aws.Bool(v.(bool))
	}

	if v, ok := m[names.AttrKMSKeyID].(string); ok && v != "" {
		c.KmsKeyId = aws.String(v)
	}

	if v, ok := m["capture_content_type_header"].([]any); ok && len(v) > 0 && v[0] != nil {
		c.CaptureContentTypeHeader = expandCaptureContentTypeHeader(v[0].(map[string]any))
	}

	return c
}

func flattenDataCaptureConfig(dataCaptureConfig *awstypes.DataCaptureConfig) []map[string]any {
	if dataCaptureConfig == nil {
		return []map[string]any{}
	}

	cfg := map[string]any{
		"initial_sampling_percentage": aws.ToInt32(dataCaptureConfig.InitialSamplingPercentage),
		"destination_s3_uri":          aws.ToString(dataCaptureConfig.DestinationS3Uri),
		"capture_options":             flattenCaptureOptions(dataCaptureConfig.CaptureOptions),
	}

	if dataCaptureConfig.EnableCapture != nil {
		cfg["enable_capture"] = aws.ToBool(dataCaptureConfig.EnableCapture)
	}

	if dataCaptureConfig.KmsKeyId != nil {
		cfg[names.AttrKMSKeyID] = aws.ToString(dataCaptureConfig.KmsKeyId)
	}

	if dataCaptureConfig.CaptureContentTypeHeader != nil {
		cfg["capture_content_type_header"] = flattenCaptureContentTypeHeader(dataCaptureConfig.CaptureContentTypeHeader)
	}

	return []map[string]any{cfg}
}

func expandCaptureOptions(configured []any) []awstypes.CaptureOption {
	containers := make([]awstypes.CaptureOption, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]any)

		l := awstypes.CaptureOption{
			CaptureMode: awstypes.CaptureMode(data["capture_mode"].(string)),
		}
		containers = append(containers, l)
	}

	return containers
}

func flattenCaptureOptions(list []awstypes.CaptureOption) []map[string]any {
	containers := make([]map[string]any, 0, len(list))

	for _, lRaw := range list {
		captureOption := make(map[string]any)
		captureOption["capture_mode"] = lRaw.CaptureMode
		containers = append(containers, captureOption)
	}

	return containers
}

func expandCaptureContentTypeHeader(m map[string]any) *awstypes.CaptureContentTypeHeader {
	c := &awstypes.CaptureContentTypeHeader{}

	if v, ok := m["csv_content_types"].(*schema.Set); ok && v.Len() > 0 {
		c.CsvContentTypes = flex.ExpandStringValueSet(v)
	}

	if v, ok := m["json_content_types"].(*schema.Set); ok && v.Len() > 0 {
		c.JsonContentTypes = flex.ExpandStringValueSet(v)
	}

	return c
}

func flattenCaptureContentTypeHeader(contentTypeHeader *awstypes.CaptureContentTypeHeader) []map[string]any {
	if contentTypeHeader == nil {
		return []map[string]any{}
	}

	l := make(map[string]any)

	if contentTypeHeader.CsvContentTypes != nil {
		l["csv_content_types"] = flex.FlattenStringValueSet(contentTypeHeader.CsvContentTypes)
	}

	if contentTypeHeader.JsonContentTypes != nil {
		l["json_content_types"] = flex.FlattenStringValueSet(contentTypeHeader.JsonContentTypes)
	}

	return []map[string]any{l}
}

func expandEndpointConfigAsyncInferenceConfig(configured []any) *awstypes.AsyncInferenceConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.AsyncInferenceConfig{}

	if v, ok := m["client_config"].([]any); ok && len(v) > 0 {
		c.ClientConfig = expandEndpointConfigClientConfig(v)
	}

	if v, ok := m["output_config"].([]any); ok && len(v) > 0 {
		c.OutputConfig = expandEndpointConfigOutputConfig(v)
	}

	return c
}

func expandEndpointConfigClientConfig(configured []any) *awstypes.AsyncInferenceClientConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.AsyncInferenceClientConfig{}

	if v, ok := m["max_concurrent_invocations_per_instance"]; ok {
		c.MaxConcurrentInvocationsPerInstance = aws.Int32(int32(v.(int)))
	}

	return c
}

func expandEndpointConfigOutputConfig(configured []any) *awstypes.AsyncInferenceOutputConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.AsyncInferenceOutputConfig{
		S3OutputPath: aws.String(m["s3_output_path"].(string)),
	}

	if v, ok := m[names.AttrKMSKeyID].(string); ok && v != "" {
		c.KmsKeyId = aws.String(v)
	}

	if v, ok := m["s3_failure_path"].(string); ok && v != "" {
		c.S3FailurePath = aws.String(v)
	}

	if v, ok := m["notification_config"].([]any); ok && len(v) > 0 {
		c.NotificationConfig = expandEndpointConfigNotificationConfig(v)
	}

	return c
}

func expandEndpointConfigNotificationConfig(configured []any) *awstypes.AsyncInferenceNotificationConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.AsyncInferenceNotificationConfig{}

	if v, ok := m["error_topic"].(string); ok && v != "" {
		c.ErrorTopic = aws.String(v)
	}

	if v, ok := m["success_topic"].(string); ok && v != "" {
		c.SuccessTopic = aws.String(v)
	}

	if v, ok := m["include_inference_response_in"].(*schema.Set); ok && v.Len() > 0 {
		c.IncludeInferenceResponseIn = flex.ExpandStringyValueSet[awstypes.AsyncNotificationTopicTypes](v)
	}

	return c
}

func expandRoutingConfig(configured []any) *awstypes.ProductionVariantRoutingConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.ProductionVariantRoutingConfig{}

	if v, ok := m["routing_strategy"].(string); ok && v != "" {
		c.RoutingStrategy = awstypes.RoutingStrategy(v)
	}

	return c
}

func expandServerlessConfig(configured []any) *awstypes.ProductionVariantServerlessConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.ProductionVariantServerlessConfig{}

	if v, ok := m["max_concurrency"].(int); ok {
		c.MaxConcurrency = aws.Int32(int32(v))
	}

	if v, ok := m["memory_size_in_mb"].(int); ok {
		c.MemorySizeInMB = aws.Int32(int32(v))
	}

	if v, ok := m["provisioned_concurrency"].(int); ok && v > 0 {
		c.ProvisionedConcurrency = aws.Int32(int32(v))
	}

	return c
}

func expandCoreDumpConfig(configured []any) *awstypes.ProductionVariantCoreDumpConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.ProductionVariantCoreDumpConfig{}

	if v, ok := m["destination_s3_uri"].(string); ok {
		c.DestinationS3Uri = aws.String(v)
	}

	if v, ok := m[names.AttrKMSKeyID].(string); ok {
		c.KmsKeyId = aws.String(v)
	}

	return c
}

func expandManagedInstanceScaling(configured []any) *awstypes.ProductionVariantManagedInstanceScaling {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]any)

	c := &awstypes.ProductionVariantManagedInstanceScaling{}

	if v, ok := m[names.AttrStatus].(string); ok {
		c.Status = awstypes.ManagedInstanceScalingStatus(v)
	}

	if v, ok := m["min_instance_count"].(int); ok {
		c.MinInstanceCount = aws.Int32(int32(v))
	}

	if v, ok := m["max_instance_count"].(int); ok && v > 0 {
		c.MaxInstanceCount = aws.Int32(int32(v))
	}

	return c
}

func flattenEndpointConfigAsyncInferenceConfig(config *awstypes.AsyncInferenceConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	cfg := map[string]any{}

	if config.ClientConfig != nil {
		cfg["client_config"] = flattenEndpointConfigClientConfig(config.ClientConfig)
	}

	if config.OutputConfig != nil {
		cfg["output_config"] = flattenEndpointConfigOutputConfig(config.OutputConfig)
	}

	return []map[string]any{cfg}
}

func flattenEndpointConfigClientConfig(config *awstypes.AsyncInferenceClientConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	cfg := map[string]any{}

	if config.MaxConcurrentInvocationsPerInstance != nil {
		cfg["max_concurrent_invocations_per_instance"] = aws.ToInt32(config.MaxConcurrentInvocationsPerInstance)
	}

	return []map[string]any{cfg}
}

func flattenEndpointConfigOutputConfig(config *awstypes.AsyncInferenceOutputConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	cfg := map[string]any{
		"s3_output_path": aws.ToString(config.S3OutputPath),
	}

	if config.KmsKeyId != nil {
		cfg[names.AttrKMSKeyID] = aws.ToString(config.KmsKeyId)
	}

	if config.NotificationConfig != nil {
		cfg["notification_config"] = flattenEndpointConfigNotificationConfig(config.NotificationConfig)
	}

	if config.S3FailurePath != nil {
		cfg["s3_failure_path"] = aws.ToString(config.S3FailurePath)
	}

	return []map[string]any{cfg}
}

func flattenEndpointConfigNotificationConfig(config *awstypes.AsyncInferenceNotificationConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	cfg := map[string]any{}

	if config.ErrorTopic != nil {
		cfg["error_topic"] = aws.ToString(config.ErrorTopic)
	}

	if config.SuccessTopic != nil {
		cfg["success_topic"] = aws.ToString(config.SuccessTopic)
	}

	if config.IncludeInferenceResponseIn != nil {
		cfg["include_inference_response_in"] = flex.FlattenStringyValueSet[awstypes.AsyncNotificationTopicTypes](config.IncludeInferenceResponseIn)
	}

	return []map[string]any{cfg}
}

func flattenRoutingConfig(config *awstypes.ProductionVariantRoutingConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	cfg := map[string]any{
		"routing_strategy": config.RoutingStrategy,
	}

	return []map[string]any{cfg}
}

func flattenServerlessConfig(config *awstypes.ProductionVariantServerlessConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	cfg := map[string]any{}

	if config.MaxConcurrency != nil {
		cfg["max_concurrency"] = aws.ToInt32(config.MaxConcurrency)
	}

	if config.MemorySizeInMB != nil {
		cfg["memory_size_in_mb"] = aws.ToInt32(config.MemorySizeInMB)
	}

	if config.ProvisionedConcurrency != nil {
		cfg["provisioned_concurrency"] = aws.ToInt32(config.ProvisionedConcurrency)
	}

	return []map[string]any{cfg}
}

func flattenCoreDumpConfig(config *awstypes.ProductionVariantCoreDumpConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	cfg := map[string]any{}

	if config.DestinationS3Uri != nil {
		cfg["destination_s3_uri"] = aws.ToString(config.DestinationS3Uri)
	}

	if config.KmsKeyId != nil {
		cfg[names.AttrKMSKeyID] = aws.ToString(config.KmsKeyId)
	}

	return []map[string]any{cfg}
}

func flattenManagedInstanceScaling(config *awstypes.ProductionVariantManagedInstanceScaling) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	cfg := map[string]any{}

	if config.Status != "" {
		cfg[names.AttrStatus] = config.Status
	}

	if config.MinInstanceCount != nil {
		cfg["min_instance_count"] = aws.ToInt32(config.MinInstanceCount)
	}

	if config.MaxInstanceCount != nil {
		cfg["max_instance_count"] = aws.ToInt32(config.MaxInstanceCount)
	}

	return []map[string]any{cfg}
}
