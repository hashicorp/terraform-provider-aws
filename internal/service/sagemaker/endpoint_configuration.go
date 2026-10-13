// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

import (
	"context"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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

// @SDKResource("aws_sagemaker_endpoint_configuration", name="Endpoint Configuration")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("name")
// @Testing(idAttrDuplicates="name")
// @Testing(preIdentityVersion="v6.53.0")
func resourceEndpointConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointConfigurationCreate,
		ReadWithoutTimeout:   resourceEndpointConfigurationRead,
		UpdateWithoutTimeout: resourceEndpointConfigurationUpdate,
		DeleteWithoutTimeout: resourceEndpointConfigurationDelete,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
				names.AttrExecutionRoleARN: {
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: verify.ValidARN,
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
							"capacity_reservation_config": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								ForceNew: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"capacity_reservation_preference": {
											Type:             schema.TypeString,
											Optional:         true,
											ForceNew:         true,
											ValidateDiagFunc: enum.Validate[awstypes.CapacityReservationPreference](),
										},
										"ml_reservation_arn": {
											Type:         schema.TypeString,
											Optional:     true,
											ForceNew:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
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
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									// Suppress diff when model_name is empty (Inference Components)
									// AWS returns nil but schema has default of 1 (there for backwards compatibility)
									if strings.Contains(k, "production_variants") || strings.Contains(k, "shadow_production_variants") {
										parts := strings.Split(k, ".")
										if len(parts) >= 2 {
											prefix := strings.Join(parts[:len(parts)-1], ".")
											modelName := d.Get(prefix + ".model_name").(string)
											return modelName == ""
										}
									}
									return false
								},
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
								Optional: true,
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
							"capacity_reservation_config": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								ForceNew: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"capacity_reservation_preference": {
											Type:             schema.TypeString,
											Optional:         true,
											ForceNew:         true,
											ValidateDiagFunc: enum.Validate[awstypes.CapacityReservationPreference](),
										},
										"ml_reservation_arn": {
											Type:         schema.TypeString,
											Optional:     true,
											ForceNew:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
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
								DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
									// Suppress diff when model_name is empty (Inference Components)
									// AWS returns nil but schema has default of 1 (there for backwards compatibility)
									if strings.Contains(k, "production_variants") || strings.Contains(k, "shadow_production_variants") {
										parts := strings.Split(k, ".")
										if len(parts) >= 2 {
											prefix := strings.Join(parts[:len(parts)-1], ".")
											modelName := d.Get(prefix + ".model_name").(string)
											return modelName == ""
										}
									}
									return false
								},
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
								Optional: true,
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
			}
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

	name := create.Name(ctx, d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := sagemaker.CreateEndpointConfigInput{
		EndpointConfigName: aws.String(name),
		ProductionVariants: expandProductionVariants(d.Get("production_variants").([]any)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("async_inference_config"); ok {
		input.AsyncInferenceConfig = expandEndpointConfigAsyncInferenceConfig(v.([]any))
	}

	if v, ok := d.GetOk("data_capture_config"); ok {
		input.DataCaptureConfig = expandDataCaptureConfig(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrExecutionRoleARN); ok {
		input.ExecutionRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("shadow_production_variants"); ok && len(v.([]any)) > 0 {
		input.ShadowProductionVariants = expandProductionVariants(v.([]any))
	}

	_, err := conn.CreateEndpointConfig(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Endpoint Configuration (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceEndpointConfigurationRead(ctx, d, meta)...)
}

func resourceEndpointConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	endpointConfig, err := findEndpointConfigByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Endpoint Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Endpoint Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, endpointConfig.EndpointConfigArn)
	if err := d.Set("async_inference_config", flattenEndpointConfigAsyncInferenceConfig(endpointConfig.AsyncInferenceConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting async_inference_config: %s", err)
	}
	if err := d.Set("data_capture_config", flattenDataCaptureConfig(endpointConfig.DataCaptureConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_capture_config: %s", err)
	}
	d.Set(names.AttrExecutionRoleARN, endpointConfig.ExecutionRoleArn)
	d.Set(names.AttrKMSKeyARN, endpointConfig.KmsKeyId)
	d.Set(names.AttrName, endpointConfig.EndpointConfigName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(endpointConfig.EndpointConfigName)))
	if err := d.Set("production_variants", flattenProductionVariants(endpointConfig.ProductionVariants)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting production_variants: %s", err)
	}
	if err := d.Set("shadow_production_variants", flattenProductionVariants(endpointConfig.ShadowProductionVariants)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting shadow_production_variants: %s", err)
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

	log.Printf("[INFO] Deleting SageMaker AI Endpoint Configuration: %s", d.Id())
	input := sagemaker.DeleteEndpointConfigInput{
		EndpointConfigName: aws.String(d.Id()),
	}
	_, err := conn.DeleteEndpointConfig(ctx, &input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint configuration") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Endpoint Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findEndpointConfigByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeEndpointConfigOutput, error) {
	input := sagemaker.DescribeEndpointConfigInput{
		EndpointConfigName: aws.String(name),
	}

	return findEndpointConfig(ctx, conn, &input)
}

func findEndpointConfig(ctx context.Context, conn *sagemaker.Client, input *sagemaker.DescribeEndpointConfigInput) (*sagemaker.DescribeEndpointConfigOutput, error) {
	output, err := conn.DescribeEndpointConfig(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint configuration") {
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

func expandProductionVariants(configured []any) []awstypes.ProductionVariant {
	containers := make([]awstypes.ProductionVariant, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]any)

		l := awstypes.ProductionVariant{}

		// Traditional endpoint: set ModelName
		// IC endpoint: omit ModelName
		// Special traditional/IC handling
		if v, ok := data["model_name"].(string); ok && v != "" {
			l.ModelName = aws.String(v)

			// Traditional endpoint: set InitialVariantWeight
			// IC endpoint: must not be set but pre-existing default value of 1
			if v, ok := data["initial_variant_weight"].(float64); ok {
				l.InitialVariantWeight = aws.Float32(float32(v))
			}

			// Traditional endpoint: set EnableSSMAccess
			// IC endpoints: must not be set
			if v, ok := data["enable_ssm_access"].(bool); ok {
				l.EnableSSMAccess = aws.Bool(v)
			}
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
			l.VariantName = aws.String(sdkid.UniqueId())
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

		if v, ok := data["managed_instance_scaling"].([]any); ok && len(v) > 0 {
			l.ManagedInstanceScaling = expandManagedInstanceScaling(v)
		}

		if v, ok := data["inference_ami_version"].(string); ok && v != "" {
			l.InferenceAmiVersion = awstypes.ProductionVariantInferenceAmiVersion(v)
		}

		if v, ok := data["capacity_reservation_config"].([]any); ok && len(v) > 0 {
			l.CapacityReservationConfig = expandCapacityReservationConfig(v)
		}

		containers = append(containers, l)
	}

	return containers
}

func flattenProductionVariants(list []awstypes.ProductionVariant) []map[string]any {
	result := make([]map[string]any, 0, len(list))

	for _, i := range list {
		l := map[string]any{
			"accelerator_type":      i.AcceleratorType,
			names.AttrInstanceType:  i.InstanceType,
			"inference_ami_version": i.InferenceAmiVersion,
			"variant_name":          aws.ToString(i.VariantName),
		}

		// Traditional endpoints have model_name set
		// Inference Component endpoints do not have model_name set
		// Special handling
		if i.ModelName != nil && aws.ToString(i.ModelName) != "" {
			l["model_name"] = aws.ToString(i.ModelName)

			if i.InitialVariantWeight != nil {
				l["initial_variant_weight"] = aws.ToFloat32(i.InitialVariantWeight)
			}

			if i.EnableSSMAccess != nil {
				l["enable_ssm_access"] = aws.ToBool(i.EnableSSMAccess)
			}
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

		if i.ManagedInstanceScaling != nil {
			l["managed_instance_scaling"] = flattenManagedInstanceScaling(i.ManagedInstanceScaling)
		}

		if i.CapacityReservationConfig != nil {
			l["capacity_reservation_config"] = flattenCapacityReservationConfig(i.CapacityReservationConfig)
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
	if configured[0] == nil {
		return &awstypes.AsyncInferenceNotificationConfig{}
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

func expandCapacityReservationConfig(tfList []any) *awstypes.ProductionVariantCapacityReservationConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ProductionVariantCapacityReservationConfig{}

	if v, ok := tfMap["capacity_reservation_preference"].(string); ok {
		apiObject.CapacityReservationPreference = awstypes.CapacityReservationPreference(v)
	}

	if v, ok := tfMap["ml_reservation_arn"].(string); ok {
		apiObject.MlReservationArn = aws.String(v)
	}

	return apiObject
}

func flattenEndpointConfigAsyncInferenceConfig(apiObject *awstypes.AsyncInferenceConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.ClientConfig != nil {
		tfMap["client_config"] = flattenEndpointConfigClientConfig(apiObject.ClientConfig)
	}

	if apiObject.OutputConfig != nil {
		tfMap["output_config"] = flattenEndpointConfigOutputConfig(apiObject.OutputConfig)
	}

	return []any{tfMap}
}

func flattenEndpointConfigClientConfig(apiObject *awstypes.AsyncInferenceClientConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.MaxConcurrentInvocationsPerInstance != nil {
		tfMap["max_concurrent_invocations_per_instance"] = aws.ToInt32(apiObject.MaxConcurrentInvocationsPerInstance)
	}

	return []any{tfMap}
}

func flattenEndpointConfigOutputConfig(apiObject *awstypes.AsyncInferenceOutputConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"s3_output_path": aws.ToString(apiObject.S3OutputPath),
	}

	if apiObject.KmsKeyId != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(apiObject.KmsKeyId)
	}

	if apiObject.NotificationConfig != nil {
		tfMap["notification_config"] = flattenEndpointConfigNotificationConfig(apiObject.NotificationConfig)
	}

	if apiObject.S3FailurePath != nil {
		tfMap["s3_failure_path"] = aws.ToString(apiObject.S3FailurePath)
	}

	return []any{tfMap}
}

func flattenEndpointConfigNotificationConfig(apiObject *awstypes.AsyncInferenceNotificationConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.ErrorTopic != nil {
		tfMap["error_topic"] = aws.ToString(apiObject.ErrorTopic)
	}

	if apiObject.SuccessTopic != nil {
		tfMap["success_topic"] = aws.ToString(apiObject.SuccessTopic)
	}

	if apiObject.IncludeInferenceResponseIn != nil {
		tfMap["include_inference_response_in"] = apiObject.IncludeInferenceResponseIn
	}

	return []any{tfMap}
}

func flattenRoutingConfig(apiObject *awstypes.ProductionVariantRoutingConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"routing_strategy": apiObject.RoutingStrategy,
	}

	return []any{tfMap}
}

func flattenServerlessConfig(apiObject *awstypes.ProductionVariantServerlessConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.MaxConcurrency != nil {
		tfMap["max_concurrency"] = aws.ToInt32(apiObject.MaxConcurrency)
	}

	if apiObject.MemorySizeInMB != nil {
		tfMap["memory_size_in_mb"] = aws.ToInt32(apiObject.MemorySizeInMB)
	}

	if apiObject.ProvisionedConcurrency != nil {
		tfMap["provisioned_concurrency"] = aws.ToInt32(apiObject.ProvisionedConcurrency)
	}

	return []any{tfMap}
}

func flattenCoreDumpConfig(apiObject *awstypes.ProductionVariantCoreDumpConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.DestinationS3Uri != nil {
		tfMap["destination_s3_uri"] = aws.ToString(apiObject.DestinationS3Uri)
	}

	if apiObject.KmsKeyId != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(apiObject.KmsKeyId)
	}

	return []any{tfMap}
}

func flattenManagedInstanceScaling(apiObject *awstypes.ProductionVariantManagedInstanceScaling) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.Status != "" {
		tfMap[names.AttrStatus] = apiObject.Status
	}

	if apiObject.MinInstanceCount != nil {
		tfMap["min_instance_count"] = aws.ToInt32(apiObject.MinInstanceCount)
	}

	if apiObject.MaxInstanceCount != nil {
		tfMap["max_instance_count"] = aws.ToInt32(apiObject.MaxInstanceCount)
	}

	return []any{tfMap}
}

func flattenCapacityReservationConfig(apiObject *awstypes.ProductionVariantCapacityReservationConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.CapacityReservationPreference != "" {
		tfMap["capacity_reservation_preference"] = apiObject.CapacityReservationPreference
	}

	if apiObject.MlReservationArn != nil {
		tfMap["ml_reservation_arn"] = aws.ToString(apiObject.MlReservationArn)
	}

	return []any{tfMap}
}
