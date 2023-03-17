package sagemaker

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_sagemaker_data_quality_job_definition")
func ResourceDataQualityJobDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataQualityJobDefinitionCreate,
		ReadWithoutTimeout:   resourceDataQualityJobDefinitionRead,
		UpdateWithoutTimeout: resourceDataQualityJobDefinitionUpdate,
		DeleteWithoutTimeout: resourceDataQualityJobDefinitionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_quality_app_specification": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_arguments": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"container_entrypoint": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"environment": {
							Type:         schema.TypeMap,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validEnvironment,
							Elem:         &schema.Schema{Type: schema.TypeString},
						},
						"image_uri": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validImage,
						},
						"post_analytics_processor_source_uri": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
								validation.StringLenBetween(1, 512),
							),
						},
						"record_preprocessor_source_uri": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
								validation.StringLenBetween(1, 512),
							),
						},
					},
				},
			},
			"data_quality_baseline_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"baselining_job_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validName,
						},
						"constraints_resource": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringMatch(regexp.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
											validation.StringLenBetween(1, 512),
										),
									},
								},
							},
						},
						"statistics_resource": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_uri": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringMatch(regexp.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
											validation.StringLenBetween(1, 512),
										),
									},
								},
							},
						},
					},
				},
			},
			"data_quality_job_input": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"batch_transform_input": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_captured_destination_s3_uri": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringMatch(regexp.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
											validation.StringLenBetween(1, 512),
										),
									},
									"dataset_format": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"csv": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"header": {
																Type:     schema.TypeBool,
																Optional: true,
																ForceNew: true,
															},
														},
													},
												},
												"json": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"line": {
																Type:     schema.TypeBool,
																Optional: true,
																ForceNew: true,
															},
														},
													},
												},
												"parquet": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{},
													},
												},
											},
										},
									},
									"end_time_offset": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"features_attribute": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"inference_attribute": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"local_path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/opt/ml/processing/input",
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1024),
											validation.StringMatch(regexp.MustCompile(`^\/opt\/ml\/processing\/.*`), "Must start with `/opt/ml/processing`."),
										),
									},
									"probability_attribute": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"probability_threshold_attribute": {
										Type:         schema.TypeFloat,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.FloatAtLeast(0),
									},
									"s3_data_distribution_type": {
										Type:         schema.TypeString,
										ForceNew:     true,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.ProcessingS3DataDistributionType_Values(), false),
									},
									"s3_input_mode": {
										Type:         schema.TypeString,
										ForceNew:     true,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.ProcessingS3InputMode_Values(), false),
									},
									"start_time_offset": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
								},
							},
						},
						"endpoint_input": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"end_time_offset": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"endpoint_name": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validName,
									},
									"features_attribute": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"inference_attribute": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"local_path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/opt/ml/processing/input",
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1024),
											validation.StringMatch(regexp.MustCompile(`^\/opt\/ml\/processing\/.*`), "Must start with `/opt/ml/processing`."),
										),
									},
									"probability_attribute": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"probability_threshold_attribute": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.FloatAtLeast(0),
									},
									"s3_data_distribution_type": {
										Type:         schema.TypeString,
										ForceNew:     true,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.ProcessingS3DataDistributionType_Values(), false),
									},
									"s3_input_mode": {
										Type:         schema.TypeString,
										ForceNew:     true,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.ProcessingS3InputMode_Values(), false),
									},
									"start_time_offset": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"data_quality_job_output_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"monitoring_outputs": {
							Type:     schema.TypeList,
							MinItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_output": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"local_path": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "/opt/ml/processing/output",
													ForceNew: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 1024),
														validation.StringMatch(regexp.MustCompile(`^\/opt\/ml\/processing\/.*`), "Must start with `/opt/ml/processing`."),
													),
												},
												"s3_upload_mode": {
													Type:         schema.TypeString,
													ForceNew:     true,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.ProcessingS3UploadMode_Values(), false),
												},
												"s3_uri": {
													Type:     schema.TypeString,
													Required: true,
													ForceNew: true,
													ValidateFunc: validation.All(
														validation.StringMatch(regexp.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
														validation.StringLenBetween(1, 512),
													),
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
			"job_resources": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"instance_count": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"instance_type": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.ProcessingInstanceType_Values(), false),
									},
									"volume_kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
									"volume_size_in_gb": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 512),
									},
								},
							},
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
			"network_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_inter_container_traffic_encryption": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"enable_network_isolation": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"vpc_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"security_group_ids": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 5,
										Required: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"subnets": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 16,
										Required: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"stopping_condition": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_runtime_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 3600),
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDataQualityJobDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	var roleArn string
	if v, ok := d.GetOk("role_arn"); ok {
		roleArn = v.(string)
	}

	createOpts := &sagemaker.CreateDataQualityJobDefinitionInput{
		JobDefinitionName:           aws.String(name),
		DataQualityAppSpecification: expandDataQualityAppSpecification(d.Get("data_quality_app_specification").([]interface{})),
		DataQualityJobInput:         expandDataQualityJobInput(d.Get("data_quality_job_input").([]interface{})),
		DataQualityJobOutputConfig:  expandDataQualityJobOutputConfig(d.Get("data_quality_job_output_config").([]interface{})),
		JobResources:                expandJobResources(d.Get("job_resources").([]interface{})),
		RoleArn:                     aws.String(roleArn),
	}

	if v, ok := d.GetOk("data_quality_baseline_config"); ok && len(v.([]interface{})) > 0 {
		createOpts.DataQualityBaselineConfig = expandDataQualityBaselineConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("network_config"); ok && len(v.([]interface{})) > 0 {
		createOpts.NetworkConfig = expandNetworkConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("stopping_condition"); ok && len(v.([]interface{})) > 0 {
		createOpts.StoppingCondition = expandStoppingCondition(v.([]interface{}))
	}

	if len(tags) > 0 {
		createOpts.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] SageMaker Data Quality Job Definition create config: %#v", *createOpts)
	_, err := conn.CreateDataQualityJobDefinitionWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Data Quality Job Definition: %s", err)
	}
	d.SetId(name)

	return append(diags, resourceDataQualityJobDefinitionRead(ctx, d, meta)...)
}

func resourceDataQualityJobDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	jobDefinition, err := FindDataQualityJobDefinitionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		log.Printf("[WARN] SageMaker Data Quality Job Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	d.Set("arn", jobDefinition.JobDefinitionArn)
	d.Set("name", jobDefinition.JobDefinitionName)
	d.Set("role_arn", jobDefinition.RoleArn)

	if err := d.Set("data_quality_app_specification", flattenDataQualityAppSpecification(jobDefinition.DataQualityAppSpecification)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_app_specification for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_quality_baseline_config", flattenDataQualityBaselineConfig(jobDefinition.DataQualityBaselineConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_baseline_config for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_quality_job_input", flattenDataQualityJobInput(jobDefinition.DataQualityJobInput)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_job_input for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_quality_job_output_config", flattenDataQualityJobOutputConfig(jobDefinition.DataQualityJobOutputConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_job_output_config for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("job_resources", flattenJobResources(jobDefinition.JobResources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting job_resources for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("network_config", flattenNetworkConfig(jobDefinition.NetworkConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_config for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("stopping_condition", flattenStoppingCondition(jobDefinition.StoppingCondition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stopping_condition for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, aws.StringValue(jobDefinition.JobDefinitionArn))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func flattenDataQualityAppSpecification(appSpecification *sagemaker.DataQualityAppSpecification) []map[string]interface{} {
	if appSpecification == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if appSpecification.ImageUri != nil {
		spec["image_uri"] = aws.StringValue(appSpecification.ImageUri)
	}

	if appSpecification.ContainerArguments != nil {
		spec["container_arguments"] = flex.FlattenStringSet(appSpecification.ContainerArguments)
	}

	if appSpecification.ContainerEntrypoint != nil {
		spec["container_entrypoint"] = flex.FlattenStringSet(appSpecification.ContainerEntrypoint)
	}

	if appSpecification.Environment != nil {
		spec["environment"] = aws.StringValueMap(appSpecification.Environment)
	}

	if appSpecification.PostAnalyticsProcessorSourceUri != nil {
		spec["post_analytics_processor_source_uri"] = aws.StringValue(appSpecification.PostAnalyticsProcessorSourceUri)
	}

	if appSpecification.RecordPreprocessorSourceUri != nil {
		spec["record_preprocessor_source_uri"] = aws.StringValue(appSpecification.RecordPreprocessorSourceUri)
	}

	return []map[string]interface{}{spec}
}

func flattenDataQualityBaselineConfig(baselineConfig *sagemaker.DataQualityBaselineConfig) []map[string]interface{} {
	if baselineConfig == nil {
		return []map[string]interface{}{}
	}

	fConfig := map[string]interface{}{}

	if baselineConfig.BaseliningJobName != nil {
		fConfig["baselining_job_name"] = aws.StringValue(baselineConfig.BaseliningJobName)
	}

	if baselineConfig.ConstraintsResource != nil {
		fConfig["constraints_resource"] = flattenConstraintsResource(baselineConfig.ConstraintsResource)
	}

	if baselineConfig.StatisticsResource != nil {
		fConfig["statistics_resource"] = flattenStatisticsResource(baselineConfig.StatisticsResource)
	}

	return []map[string]interface{}{fConfig}
}

func flattenConstraintsResource(constraintsResource *sagemaker.MonitoringConstraintsResource) []map[string]interface{} {
	if constraintsResource == nil {
		return []map[string]interface{}{}
	}

	fResource := map[string]interface{}{}

	if constraintsResource.S3Uri != nil {
		fResource["s3_uri"] = aws.StringValue(constraintsResource.S3Uri)
	}

	return []map[string]interface{}{fResource}
}

func flattenStatisticsResource(statisticsResource *sagemaker.MonitoringStatisticsResource) []map[string]interface{} {
	if statisticsResource == nil {
		return []map[string]interface{}{}
	}

	fResource := map[string]interface{}{}

	if statisticsResource.S3Uri != nil {
		fResource["s3_uri"] = aws.StringValue(statisticsResource.S3Uri)
	}

	return []map[string]interface{}{fResource}
}

func flattenDataQualityJobInput(jobInput *sagemaker.DataQualityJobInput) []map[string]interface{} {
	if jobInput == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if jobInput.EndpointInput != nil {
		spec["endpoint_input"] = flattenEndpointInput(jobInput.EndpointInput)
	}

	if jobInput.BatchTransformInput != nil {
		spec["batch_transform_input"] = flattenBatchTransformInput(jobInput.BatchTransformInput)
	}

	return []map[string]interface{}{spec}
}

func flattenBatchTransformInput(transformInput *sagemaker.BatchTransformInput_) []map[string]interface{} {
	if transformInput == nil {
		return []map[string]interface{}{}
	}

	fInput := map[string]interface{}{}

	if transformInput.LocalPath != nil {
		fInput["local_path"] = aws.StringValue(transformInput.LocalPath)
	}

	if transformInput.DataCapturedDestinationS3Uri != nil {
		fInput["data_captured_destination_s3_uri"] = aws.StringValue(transformInput.DataCapturedDestinationS3Uri)
	}

	if transformInput.DatasetFormat != nil {
		fInput["dataset_format"] = flattenDatasetFormat(transformInput.DatasetFormat)
	}

	if transformInput.EndTimeOffset != nil {
		fInput["end_time_offset"] = aws.StringValue(transformInput.EndTimeOffset)
	}

	if transformInput.FeaturesAttribute != nil {
		fInput["features_attribute"] = aws.StringValue(transformInput.FeaturesAttribute)
	}

	if transformInput.InferenceAttribute != nil {
		fInput["inference_attribute"] = aws.StringValue(transformInput.InferenceAttribute)
	}

	if transformInput.ProbabilityAttribute != nil {
		fInput["probability_attribute"] = aws.StringValue(transformInput.ProbabilityAttribute)
	}

	if transformInput.ProbabilityThresholdAttribute != nil {
		fInput["probability_threshold_attribute"] = aws.Float64Value(transformInput.ProbabilityThresholdAttribute)
	}

	if transformInput.S3DataDistributionType != nil {
		fInput["s3_data_distribution_type"] = aws.StringValue(transformInput.S3DataDistributionType)
	}

	if transformInput.S3InputMode != nil {
		fInput["s3_input_mode"] = aws.StringValue(transformInput.S3InputMode)
	}

	if transformInput.StartTimeOffset != nil {
		fInput["start_time_offset"] = aws.StringValue(transformInput.StartTimeOffset)
	}

	return []map[string]interface{}{fInput}
}

func flattenDatasetFormat(datasetFormat *sagemaker.MonitoringDatasetFormat) []map[string]interface{} {
	if datasetFormat == nil {
		return []map[string]interface{}{}
	}

	fFormat := map[string]interface{}{}

	if datasetFormat.Csv != nil {
		fFormat["csv"] = flattenCsv(datasetFormat.Csv)
	}

	if datasetFormat.Json != nil {
		fFormat["json"] = flattenJson(datasetFormat.Json)
	}

	if datasetFormat.Parquet != nil {
		fFormat["parquet"] = []map[string]interface{}{}
	}

	return []map[string]interface{}{fFormat}
}

func flattenCsv(csv *sagemaker.MonitoringCsvDatasetFormat) []map[string]interface{} {
	if csv == nil {
		return []map[string]interface{}{}
	}

	fCsv := map[string]interface{}{}

	if csv.Header != nil {
		fCsv["header"] = aws.BoolValue(csv.Header)
	}

	return []map[string]interface{}{fCsv}
}

func flattenJson(json *sagemaker.MonitoringJsonDatasetFormat) []map[string]interface{} {
	if json == nil {
		return []map[string]interface{}{}
	}

	fJson := map[string]interface{}{}

	if json.Line != nil {
		fJson["line"] = aws.BoolValue(json.Line)
	}

	return []map[string]interface{}{fJson}
}

func flattenEndpointInput(endpointInput *sagemaker.EndpointInput) []map[string]interface{} {
	if endpointInput == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if endpointInput.EndpointName != nil {
		spec["endpoint_name"] = aws.StringValue(endpointInput.EndpointName)
	}

	if endpointInput.LocalPath != nil {
		spec["local_path"] = aws.StringValue(endpointInput.LocalPath)
	}

	if endpointInput.EndTimeOffset != nil {
		spec["end_time_offset"] = aws.StringValue(endpointInput.EndTimeOffset)
	}

	if endpointInput.FeaturesAttribute != nil {
		spec["features_attribute"] = aws.StringValue(endpointInput.FeaturesAttribute)
	}

	if endpointInput.InferenceAttribute != nil {
		spec["inference_attribute"] = aws.StringValue(endpointInput.InferenceAttribute)
	}

	if endpointInput.ProbabilityAttribute != nil {
		spec["probability_attribute"] = aws.StringValue(endpointInput.ProbabilityAttribute)
	}

	if endpointInput.ProbabilityThresholdAttribute != nil {
		spec["probability_threshold_attribute"] = aws.Float64Value(endpointInput.ProbabilityThresholdAttribute)
	}

	if endpointInput.S3DataDistributionType != nil {
		spec["s3_data_distribution_type"] = aws.StringValue(endpointInput.S3DataDistributionType)
	}

	if endpointInput.S3InputMode != nil {
		spec["s3_input_mode"] = aws.StringValue(endpointInput.S3InputMode)
	}

	if endpointInput.StartTimeOffset != nil {
		spec["start_time_offset"] = aws.StringValue(endpointInput.StartTimeOffset)
	}

	return []map[string]interface{}{spec}
}

func flattenDataQualityJobOutputConfig(outputConfig *sagemaker.MonitoringOutputConfig) []map[string]interface{} {
	if outputConfig == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if outputConfig.KmsKeyId != nil {
		spec["kms_key_id"] = aws.StringValue(outputConfig.KmsKeyId)
	}

	if outputConfig.MonitoringOutputs != nil {
		spec["monitoring_outputs"] = flattenMonitoringOutputs(outputConfig.MonitoringOutputs)
	}

	return []map[string]interface{}{spec}
}

func flattenMonitoringOutputs(list []*sagemaker.MonitoringOutput) []map[string]interface{} {
	containers := make([]map[string]interface{}, 0, len(list))

	for _, lRaw := range list {
		monitoringOutput := make(map[string]interface{})
		monitoringOutput["s3_output"] = flattenS3Output(lRaw.S3Output)
		containers = append(containers, monitoringOutput)
	}

	return containers
}

func flattenS3Output(s3Output *sagemaker.MonitoringS3Output) []map[string]interface{} {
	if s3Output == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if s3Output.LocalPath != nil {
		spec["local_path"] = aws.StringValue(s3Output.LocalPath)
	}

	if s3Output.S3UploadMode != nil {
		spec["s3_upload_mode"] = aws.StringValue(s3Output.S3UploadMode)
	}

	if s3Output.S3Uri != nil {
		spec["s3_uri"] = aws.StringValue(s3Output.S3Uri)
	}

	return []map[string]interface{}{spec}
}

func flattenJobResources(jobResources *sagemaker.MonitoringResources) []map[string]interface{} {
	if jobResources == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if jobResources.ClusterConfig != nil {
		spec["cluster_config"] = flattenClusterConfig(jobResources.ClusterConfig)
	}

	return []map[string]interface{}{spec}
}

func flattenClusterConfig(clusterConfig *sagemaker.MonitoringClusterConfig) []map[string]interface{} {
	if clusterConfig == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if clusterConfig.InstanceCount != nil {
		spec["instance_count"] = aws.Int64Value(clusterConfig.InstanceCount)
	}

	if clusterConfig.InstanceType != nil {
		spec["instance_type"] = aws.StringValue(clusterConfig.InstanceType)
	}

	if clusterConfig.VolumeKmsKeyId != nil {
		spec["volume_kms_key_id"] = aws.StringValue(clusterConfig.VolumeKmsKeyId)
	}

	if clusterConfig.VolumeSizeInGB != nil {
		spec["volume_size_in_gb"] = aws.Int64Value(clusterConfig.VolumeSizeInGB)
	}

	return []map[string]interface{}{spec}
}

func flattenNetworkConfig(networkConfig *sagemaker.MonitoringNetworkConfig) []map[string]interface{} {
	if networkConfig == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if networkConfig.EnableInterContainerTrafficEncryption != nil {
		spec["enable_inter_container_traffic_encryption"] = aws.BoolValue(networkConfig.EnableInterContainerTrafficEncryption)
	}

	if networkConfig.EnableNetworkIsolation != nil {
		spec["enable_network_isolation"] = aws.BoolValue(networkConfig.EnableNetworkIsolation)
	}

	if networkConfig.VpcConfig != nil {
		spec["vpc_config"] = flattenVpcConfig(networkConfig.VpcConfig)
	}

	return []map[string]interface{}{spec}
}

func flattenVpcConfig(vpcConfig *sagemaker.VpcConfig) []map[string]interface{} {
	if vpcConfig == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if vpcConfig.SecurityGroupIds != nil {
		spec["security_group_ids"] = flex.FlattenStringSet(vpcConfig.SecurityGroupIds)
	}

	if vpcConfig.Subnets != nil {
		spec["subnets"] = flex.FlattenStringSet(vpcConfig.Subnets)
	}

	return []map[string]interface{}{spec}
}

func flattenStoppingCondition(stoppingCondition *sagemaker.MonitoringStoppingCondition) []map[string]interface{} {
	if stoppingCondition == nil {
		return []map[string]interface{}{}
	}

	spec := map[string]interface{}{}

	if stoppingCondition.MaxRuntimeInSeconds != nil {
		spec["max_runtime_in_seconds"] = aws.Int64Value(stoppingCondition.MaxRuntimeInSeconds)
	}

	return []map[string]interface{}{spec}
}

func resourceDataQualityJobDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Data Quality Job Definition (%s) tags: %s", d.Id(), err)
		}
	}
	return append(diags, resourceDataQualityJobDefinitionRead(ctx, d, meta)...)
}

func resourceDataQualityJobDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	deleteOpts := &sagemaker.DeleteDataQualityJobDefinitionInput{
		JobDefinitionName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Data Quality Job Definition : %s", d.Id())

	_, err := conn.DeleteDataQualityJobDefinitionWithContext(ctx, deleteOpts)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDataQualityAppSpecification(configured []interface{}) *sagemaker.DataQualityAppSpecification {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.DataQualityAppSpecification{}

	if v, ok := m["image_uri"].(string); ok && v != "" {
		c.ImageUri = aws.String(v)
	}

	if v, ok := m["container_arguments"].(*schema.Set); ok && v.Len() > 0 {
		c.ContainerArguments = flex.ExpandStringSet(v)
	}

	if v, ok := m["container_entrypoint"].(*schema.Set); ok && v.Len() > 0 {
		c.ContainerEntrypoint = flex.ExpandStringSet(v)
	}

	if v, ok := m["environment"].(map[string]interface{}); ok && len(v) > 0 {
		c.Environment = flex.ExpandStringMap(v)
	}

	if v, ok := m["post_analytics_processor_source_uri"].(string); ok && v != "" {
		c.PostAnalyticsProcessorSourceUri = aws.String(v)
	}

	if v, ok := m["record_preprocessor_source_uri"].(string); ok && v != "" {
		c.RecordPreprocessorSourceUri = aws.String(v)
	}

	return c
}

func expandDataQualityBaselineConfig(configured []interface{}) *sagemaker.DataQualityBaselineConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.DataQualityBaselineConfig{}

	if v, ok := m["baselining_job_name"].(string); ok && v != "" {
		c.BaseliningJobName = aws.String(v)
	}

	if v, ok := m["constraints_resource"].([]interface{}); ok && len(v) > 0 {
		c.ConstraintsResource = expandConstraintsResource(v)
	}

	if v, ok := m["statistics_resource"].([]interface{}); ok && len(v) > 0 {
		c.StatisticsResource = expandStatisticsResource(v)
	}

	return c
}

func expandConstraintsResource(configured []interface{}) *sagemaker.MonitoringConstraintsResource {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringConstraintsResource{}

	if v, ok := m["s3_uri"].(string); ok && v != "" {
		c.S3Uri = aws.String(v)
	}

	return c
}

func expandStatisticsResource(configured []interface{}) *sagemaker.MonitoringStatisticsResource {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringStatisticsResource{}

	if v, ok := m["s3_uri"].(string); ok && v != "" {
		c.S3Uri = aws.String(v)
	}

	return c
}

func expandDataQualityJobInput(configured []interface{}) *sagemaker.DataQualityJobInput {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.DataQualityJobInput{}

	if v, ok := m["endpoint_input"].([]interface{}); ok && len(v) > 0 {
		c.EndpointInput = expandEndpointInput(v)
	}

	if v, ok := m["batch_transform_input"].([]interface{}); ok && len(v) > 0 {
		c.BatchTransformInput = expandBatchTransformInput(v)
	}

	return c
}

func expandEndpointInput(configured []interface{}) *sagemaker.EndpointInput {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.EndpointInput{}

	if v, ok := m["endpoint_name"].(string); ok && v != "" {
		c.EndpointName = aws.String(v)
	}

	if v, ok := m["end_time_offset"].(string); ok && v != "" {
		c.EndTimeOffset = aws.String(v)
	}

	if v, ok := m["features_attribute"].(string); ok && v != "" {
		c.FeaturesAttribute = aws.String(v)
	}

	if v, ok := m["inference_attribute"].(string); ok && v != "" {
		c.InferenceAttribute = aws.String(v)
	}

	if v, ok := m["local_path"].(string); ok && v != "" {
		c.LocalPath = aws.String(v)
	}

	if v, ok := m["probability_attribute"].(string); ok && v != "" {
		c.ProbabilityAttribute = aws.String(v)
	}

	if v, ok := m["probability_threshold_attribute"].(float64); ok && v > 0 {
		c.ProbabilityThresholdAttribute = aws.Float64(v)
	}

	if v, ok := m["s3_data_distribution_type"].(string); ok && v != "" {
		c.S3DataDistributionType = aws.String(v)
	}

	if v, ok := m["s3_input_mode"].(string); ok && v != "" {
		c.S3InputMode = aws.String(v)
	}

	if v, ok := m["start_time_offset"].(string); ok && v != "" {
		c.StartTimeOffset = aws.String(v)
	}

	return c
}

func expandBatchTransformInput(configured []interface{}) *sagemaker.BatchTransformInput_ {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.BatchTransformInput_{}

	if v, ok := m["data_captured_destination_s3_uri"].(string); ok && v != "" {
		c.DataCapturedDestinationS3Uri = aws.String(v)
	}

	if v, ok := m["dataset_format"].([]interface{}); ok && len(v) > 0 {
		c.DatasetFormat = expandDatasetFormat(v)
	}

	if v, ok := m["end_time_offset"].(string); ok && v != "" {
		c.EndTimeOffset = aws.String(v)
	}

	if v, ok := m["features_attribute"].(string); ok && v != "" {
		c.FeaturesAttribute = aws.String(v)
	}

	if v, ok := m["inference_attribute"].(string); ok && v != "" {
		c.InferenceAttribute = aws.String(v)
	}

	if v, ok := m["local_path"].(string); ok && v != "" {
		c.LocalPath = aws.String(v)
	}

	if v, ok := m["probability_attribute"].(string); ok && v != "" {
		c.ProbabilityAttribute = aws.String(v)
	}

	if v, ok := m["probability_threshold_attribute"].(float64); ok && v > 0 {
		c.ProbabilityThresholdAttribute = aws.Float64(v)
	}

	if v, ok := m["s3_data_distribution_type"].(string); ok && v != "" {
		c.S3DataDistributionType = aws.String(v)
	}

	if v, ok := m["s3_input_mode"].(string); ok && v != "" {
		c.S3InputMode = aws.String(v)
	}

	if v, ok := m["start_time_offset"].(string); ok && v != "" {
		c.StartTimeOffset = aws.String(v)
	}

	return c
}

func expandDatasetFormat(configured []interface{}) *sagemaker.MonitoringDatasetFormat {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringDatasetFormat{}

	if v, ok := m["csv"].([]interface{}); ok && len(v) > 0 {
		c.Csv = expandCsv(v)
	}

	if v, ok := m["json"].([]interface{}); ok && len(v) > 0 {
		c.Json = expandJson(v)
	}

	if v, ok := m["parquet"].([]interface{}); ok && len(v) > 0 {
		c.Parquet = &sagemaker.MonitoringParquetDatasetFormat{}
	}

	return c
}

func expandJson(configured []interface{}) *sagemaker.MonitoringJsonDatasetFormat {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringJsonDatasetFormat{}

	if v, ok := m["line"]; ok {
		c.Line = aws.Bool(v.(bool))
	}

	return c
}

func expandCsv(configured []interface{}) *sagemaker.MonitoringCsvDatasetFormat {
	if len(configured) == 0 {
		return nil
	}

	c := &sagemaker.MonitoringCsvDatasetFormat{}

	if configured[0] == nil {
		return c
	}

	m := configured[0].(map[string]interface{})
	if v, ok := m["header"]; ok {
		c.Header = aws.Bool(v.(bool))
	}

	return c
}

func expandDataQualityJobOutputConfig(configured []interface{}) *sagemaker.MonitoringOutputConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringOutputConfig{}

	if v, ok := m["kms_key_id"].(string); ok && v != "" {
		c.KmsKeyId = aws.String(v)
	}

	if v, ok := m["monitoring_outputs"].([]interface{}); ok && len(v) > 0 {
		c.MonitoringOutputs = expandMonitoringOutputs(v)
	}

	return c
}

func expandMonitoringOutputs(configured []interface{}) []*sagemaker.MonitoringOutput {
	containers := make([]*sagemaker.MonitoringOutput, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &sagemaker.MonitoringOutput{
			S3Output: expandS3Output(data["s3_output"].([]interface{})),
		}
		containers = append(containers, l)
	}

	return containers
}

func expandS3Output(configured []interface{}) *sagemaker.MonitoringS3Output {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringS3Output{}

	if v, ok := m["local_path"].(string); ok && v != "" {
		c.LocalPath = aws.String(v)
	}

	if v, ok := m["s3_upload_mode"].(string); ok && v != "" {
		c.S3UploadMode = aws.String(v)
	}

	if v, ok := m["s3_uri"].(string); ok && v != "" {
		c.S3Uri = aws.String(v)
	}

	return c
}

func expandJobResources(configured []interface{}) *sagemaker.MonitoringResources {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringResources{}

	if v, ok := m["cluster_config"].([]interface{}); ok && len(v) > 0 {
		c.ClusterConfig = expandClusterConfig(v)
	}

	return c
}

func expandClusterConfig(configured []interface{}) *sagemaker.MonitoringClusterConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringClusterConfig{}

	if v, ok := m["instance_count"].(int); ok && v > 0 {
		c.InstanceCount = aws.Int64(int64(v))
	}

	if v, ok := m["instance_type"].(string); ok && v != "" {
		c.InstanceType = aws.String(v)
	}

	if v, ok := m["volume_kms_key_id"].(string); ok && v != "" {
		c.VolumeKmsKeyId = aws.String(v)
	}

	if v, ok := m["volume_size_in_gb"].(int); ok && v > 0 {
		c.VolumeSizeInGB = aws.Int64(int64(v))
	}

	return c
}

func expandNetworkConfig(configured []interface{}) *sagemaker.MonitoringNetworkConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringNetworkConfig{}

	if v, ok := m["enable_inter_container_traffic_encryption"]; ok {
		c.EnableInterContainerTrafficEncryption = aws.Bool(v.(bool))
	}

	if v, ok := m["enable_network_isolation"]; ok {
		c.EnableNetworkIsolation = aws.Bool(v.(bool))
	}

	if v, ok := m["vpc_config"].([]interface{}); ok && len(v) > 0 {
		c.VpcConfig = expandVpcConfig(v)
	}

	return c
}

func expandVpcConfig(configured []interface{}) *sagemaker.VpcConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.VpcConfig{}

	if v, ok := m["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		c.SecurityGroupIds = flex.ExpandStringSet(v)
	}

	if v, ok := m["subnets"].(*schema.Set); ok && v.Len() > 0 {
		c.Subnets = flex.ExpandStringSet(v)
	}

	return c
}

func expandStoppingCondition(configured []interface{}) *sagemaker.MonitoringStoppingCondition {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringStoppingCondition{}

	if v, ok := m["max_runtime_in_seconds"].(int); ok && v > 0 {
		c.MaxRuntimeInSeconds = aws.Int64(int64(v))
	}

	return c
}
