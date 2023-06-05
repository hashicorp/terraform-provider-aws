package sagemaker

import (
	"context"
	"log"
	"regexp"

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

// @SDKResource("aws_sagemaker_model_quality_job_definition", name="Model Quality Job Definition")
// @Tags(identifierAttribute="arn")
func ResourceModelQualityJobDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceModelQualityJobDefinitionCreate,
		ReadWithoutTimeout:   resourceModelQualityJobDefinitionRead,
		UpdateWithoutTimeout: resourceModelQualityJobDefinitionUpdate,
		DeleteWithoutTimeout: resourceModelQualityJobDefinitionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validName,
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
			"model_quality_app_specification": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"problem_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.MonitoringProblemType_Values(), false),
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
			"model_quality_baseline_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"baselining_job_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}`), ""),
								validation.StringLenBetween(1, 63),
							),
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
					},
				},
			},
			"model_quality_job_input": {
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
											},
										},
									},
									"end_time_offset": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 15),
											validation.StringMatch(regexp.MustCompile(`^.?P.*`), ""),
										),
									},
									"features_attribute": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
										Computed: true,
									},
									"inference_attribute": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
										Computed: true,
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
										Computed: true,
									},
									"probability_threshold_attribute": {
										Type:     schema.TypeFloat,
										ForceNew: true,
										Optional: true,
										Computed: true,
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
										Computed: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 15),
											validation.StringMatch(regexp.MustCompile(`^.?P.*`), ""),
										),
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
									"endpoint_name": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validName,
									},
									"end_time_offset": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
										Computed: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 15),
											validation.StringMatch(regexp.MustCompile(`^.?P.*`), ""),
										),
									},
									"features_attribute": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
										Computed: true,
									},
									"inference_attribute": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
										Computed: true,
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
										Computed: true,
									},
									"probability_threshold_attribute": {
										Type:     schema.TypeFloat,
										ForceNew: true,
										Optional: true,
										Computed: true,
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
										Computed: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 15),
											validation.StringMatch(regexp.MustCompile(`^.?P.*`), ""),
										),
									},
								},
							},
						},
						"ground_truth_s3_input": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_uri": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
										Computed: true,
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
			"model_quality_job_output_config": {
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
							MaxItems: 1,
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
							Optional: true,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceModelQualityJobDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = id.UniqueId()
	}

	var roleArn string
	if v, ok := d.GetOk("role_arn"); ok {
		roleArn = v.(string)
	}

	createOpts := &sagemaker.CreateModelQualityJobDefinitionInput{
		ModelQualityAppSpecification: expandModelQualityAppSpecification(d.Get("model_quality_app_specification").([]interface{})),
		ModelQualityBaselineConfig:   expandModelQualityBaselineConfig(d.Get("model_quality_baseline_config").([]interface{})),
		ModelQualityJobInput:         expandModelQualityJobInput(d.Get("model_quality_job_input").([]interface{})),
		ModelQualityJobOutputConfig:  expandMonitoringOutputConfig(d.Get("model_quality_job_output_config").([]interface{})),
		JobDefinitionName:            aws.String(name),
		JobResources:                 expandMonitoringResources(d.Get("job_resources").([]interface{})),
		RoleArn:                      aws.String(roleArn),
		Tags:                         GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("network_config"); ok && len(v.([]interface{})) > 0 {
		createOpts.NetworkConfig = expandMonitoringNetworkConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("stopping_condition"); ok && len(v.([]interface{})) > 0 {
		createOpts.StoppingCondition = expandMonitoringStoppingCondition(v.([]interface{}))
	}

	_, err := conn.CreateModelQualityJobDefinitionWithContext(ctx, createOpts)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Model Quality Job Definition (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceModelQualityJobDefinitionRead(ctx, d, meta)...)
}

func resourceModelQualityJobDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	jobDefinition, err := FindModelQualityJobDefinitionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		log.Printf("[WARN] SageMaker Model Quality Job Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Model Quality Job Definition (%s): %s", d.Id(), err)
	}

	d.Set("arn", jobDefinition.JobDefinitionArn)
	d.Set("name", jobDefinition.JobDefinitionName)
	d.Set("role_arn", jobDefinition.RoleArn)

	if err := d.Set("model_quality_app_specification", flattenModelQualityAppSpecification(jobDefinition.ModelQualityAppSpecification)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting model_quality_app_specification for SageMaker Model Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("model_quality_baseline_config", flattenModelQualityBaselineConfig(jobDefinition.ModelQualityBaselineConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting model_quality_baseline_config for SageMaker Model Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("model_quality_job_input", flattenModelQualityJobInput(jobDefinition.ModelQualityJobInput)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting model_quality_job_input for SageMaker Model Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("model_quality_job_output_config", flattenMonitoringOutputConfig(jobDefinition.ModelQualityJobOutputConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting model_quality_job_output_config for SageMaker Model Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("job_resources", flattenMonitoringResources(jobDefinition.JobResources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting job_resources for SageMaker Model Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("network_config", flattenMonitoringNetworkConfig(jobDefinition.NetworkConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_config for SageMaker Model Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("stopping_condition", flattenMonitoringStoppingCondition(jobDefinition.StoppingCondition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stopping_condition for SageMaker Model Quality Job Definition (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceModelQualityJobDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceModelQualityJobDefinitionRead(ctx, d, meta)...)
}

func resourceModelQualityJobDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	log.Printf("[INFO] Deleting SageMaker Model Quality Job Definition: %s", d.Id())
	_, err := conn.DeleteModelQualityJobDefinitionWithContext(ctx, &sagemaker.DeleteModelQualityJobDefinitionInput{
		JobDefinitionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Model Quality Job Definition (%s): %s", d.Id(), err)
	}

	return diags
}

func FindModelQualityJobDefinitionByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeModelQualityJobDefinitionOutput, error) {
	input := &sagemaker.DescribeModelQualityJobDefinitionInput{
		JobDefinitionName: aws.String(name),
	}

	output, err := conn.DescribeModelQualityJobDefinitionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
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

func flattenMonitoringGroundTruthS3Input(config *sagemaker.MonitoringGroundTruthS3Input) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.S3Uri != nil {
		m["s3_uri"] = aws.StringValue(config.S3Uri)
	}

	return []map[string]interface{}{m}
}

func flattenModelQualityAppSpecification(config *sagemaker.ModelQualityAppSpecification) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.ImageUri != nil {
		m["image_uri"] = aws.StringValue(config.ImageUri)
	}

	if config.Environment != nil {
		m["environment"] = aws.StringValueMap(config.Environment)
	}

	if config.PostAnalyticsProcessorSourceUri != nil {
		m["post_analytics_processor_source_uri"] = aws.StringValue(config.PostAnalyticsProcessorSourceUri)
	}

	if config.RecordPreprocessorSourceUri != nil {
		m["record_preprocessor_source_uri"] = aws.StringValue(config.RecordPreprocessorSourceUri)
	}

	if config.ProblemType != nil {
		m["problem_type"] = aws.StringValue(config.ProblemType)
	}

	return []map[string]interface{}{m}
}

func flattenModelQualityBaselineConfig(config *sagemaker.ModelQualityBaselineConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.BaseliningJobName != nil {
		m["baselining_job_name"] = aws.StringValue(config.BaseliningJobName)
	}

	if config.ConstraintsResource != nil {
		m["constraints_resource"] = flattenConstraintsResource(config.ConstraintsResource)
	}

	return []map[string]interface{}{m}
}

func flattenModelQualityJobInput(config *sagemaker.ModelQualityJobInput) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.EndpointInput != nil {
		m["endpoint_input"] = flattenEndpointInput(config.EndpointInput)
	}

	if config.BatchTransformInput != nil {
		m["batch_transform_input"] = flattenBatchTransformInput(config.BatchTransformInput)
	}

	if config.GroundTruthS3Input != nil {
		m["ground_truth_s3_input"] = flattenMonitoringGroundTruthS3Input(config.GroundTruthS3Input)
	}

	return []map[string]interface{}{m}
}

func expandMonitoringGroundTruthS3Input(configured []interface{}) *sagemaker.MonitoringGroundTruthS3Input {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.MonitoringGroundTruthS3Input{}

	if v, ok := m["s3_uri"].(string); ok && v != "" {
		c.S3Uri = aws.String(v)
	}

	return c
}

func expandModelQualityAppSpecification(configured []interface{}) *sagemaker.ModelQualityAppSpecification {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.ModelQualityAppSpecification{}

	if v, ok := m["image_uri"].(string); ok && v != "" {
		c.ImageUri = aws.String(v)
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

	if v, ok := m["problem_type"].(string); ok && v != "" {
		c.ProblemType = aws.String(v)
	}

	return c
}

func expandModelQualityBaselineConfig(configured []interface{}) *sagemaker.ModelQualityBaselineConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.ModelQualityBaselineConfig{}

	if v, ok := m["baselining_job_name"].(string); ok && len(v) > 0 {
		c.BaseliningJobName = aws.String(v)
	}

	if v, ok := m["constraints_resource"].([]interface{}); ok && len(v) > 0 {
		c.ConstraintsResource = expandMonitoringConstraintsResource(v)
	}

	return c
}

func expandModelQualityJobInput(configured []interface{}) *sagemaker.ModelQualityJobInput {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.ModelQualityJobInput{}

	if v, ok := m["endpoint_input"].([]interface{}); ok && len(v) > 0 {
		c.EndpointInput = expandEndpointInput(v)
	}

	if v, ok := m["batch_transform_input"].([]interface{}); ok && len(v) > 0 {
		c.BatchTransformInput = expandBatchTransformInput(v)
	}

	if v, ok := m["ground_truth_s3_input"].([]interface{}); ok && len(v) > 0 {
		c.GroundTruthS3Input = expandMonitoringGroundTruthS3Input(v)
	}

	return c
}
