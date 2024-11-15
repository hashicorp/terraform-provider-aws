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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

// @SDKResource("aws_sagemaker_data_quality_job_definition", name="Data Quality Job Definition")
// @Tags(identifierAttribute="arn")
func resourceDataQualityJobDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataQualityJobDefinitionCreate,
		ReadWithoutTimeout:   resourceDataQualityJobDefinitionRead,
		UpdateWithoutTimeout: resourceDataQualityJobDefinitionUpdate,
		DeleteWithoutTimeout: resourceDataQualityJobDefinitionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
						names.AttrEnvironment: {
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
								validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
								validation.StringLenBetween(1, 512),
							),
						},
						"record_preprocessor_source_uri": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
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
											validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
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
											validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
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
															names.AttrHeader: {
																Type:     schema.TypeBool,
																Optional: true,
																ForceNew: true,
															},
														},
													},
												},
												names.AttrJSON: {
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
									"local_path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/opt/ml/processing/input",
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1024),
											validation.StringMatch(regexache.MustCompile(`^\/opt\/ml\/processing\/.*`), "Must start with `/opt/ml/processing`."),
										),
									},
									"s3_data_distribution_type": {
										Type:             schema.TypeString,
										ForceNew:         true,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ProcessingS3DataDistributionType](),
									},
									"s3_input_mode": {
										Type:             schema.TypeString,
										ForceNew:         true,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ProcessingS3InputMode](),
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
									"local_path": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "/opt/ml/processing/input",
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1024),
											validation.StringMatch(regexache.MustCompile(`^\/opt\/ml\/processing\/.*`), "Must start with `/opt/ml/processing`."),
										),
									},
									"s3_data_distribution_type": {
										Type:             schema.TypeString,
										ForceNew:         true,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ProcessingS3DataDistributionType](),
									},
									"s3_input_mode": {
										Type:             schema.TypeString,
										ForceNew:         true,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ProcessingS3InputMode](),
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
						names.AttrKMSKeyID: {
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
														validation.StringMatch(regexache.MustCompile(`^\/opt\/ml\/processing\/.*`), "Must start with `/opt/ml/processing`."),
													),
												},
												"s3_upload_mode": {
													Type:             schema.TypeString,
													ForceNew:         true,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[awstypes.ProcessingS3UploadMode](),
												},
												"s3_uri": {
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
									names.AttrInstanceCount: {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									names.AttrInstanceType: {
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ProcessingInstanceType](),
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
			names.AttrName: {
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
						names.AttrVPCConfig: {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrSecurityGroupIDs: {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 5,
										Required: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrSubnets: {
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
			names.AttrRoleARN: {
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

func resourceDataQualityJobDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	var name string
	if v, ok := d.GetOk(names.AttrName); ok {
		name = v.(string)
	} else {
		name = id.UniqueId()
	}

	var roleArn string
	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		roleArn = v.(string)
	}

	createOpts := &sagemaker.CreateDataQualityJobDefinitionInput{
		DataQualityAppSpecification: expandDataQualityAppSpecification(d.Get("data_quality_app_specification").([]interface{})),
		DataQualityJobInput:         expandDataQualityJobInput(d.Get("data_quality_job_input").([]interface{})),
		DataQualityJobOutputConfig:  expandMonitoringOutputConfig(d.Get("data_quality_job_output_config").([]interface{})),
		JobDefinitionName:           aws.String(name),
		JobResources:                expandMonitoringResources(d.Get("job_resources").([]interface{})),
		RoleArn:                     aws.String(roleArn),
		Tags:                        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("data_quality_baseline_config"); ok && len(v.([]interface{})) > 0 {
		createOpts.DataQualityBaselineConfig = expandDataQualityBaselineConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("network_config"); ok && len(v.([]interface{})) > 0 {
		createOpts.NetworkConfig = expandMonitoringNetworkConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("stopping_condition"); ok && len(v.([]interface{})) > 0 {
		createOpts.StoppingCondition = expandMonitoringStoppingCondition(v.([]interface{}))
	}

	_, err := conn.CreateDataQualityJobDefinition(ctx, createOpts)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Data Quality Job Definition (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceDataQualityJobDefinitionRead(ctx, d, meta)...)
}

func resourceDataQualityJobDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	jobDefinition, err := findDataQualityJobDefinitionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Data Quality Job Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, jobDefinition.JobDefinitionArn)
	d.Set(names.AttrName, jobDefinition.JobDefinitionName)
	d.Set(names.AttrRoleARN, jobDefinition.RoleArn)

	if err := d.Set("data_quality_app_specification", flattenDataQualityAppSpecification(jobDefinition.DataQualityAppSpecification)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_app_specification for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_quality_baseline_config", flattenDataQualityBaselineConfig(jobDefinition.DataQualityBaselineConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_baseline_config for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_quality_job_input", flattenDataQualityJobInput(jobDefinition.DataQualityJobInput)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_job_input for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_quality_job_output_config", flattenMonitoringOutputConfig(jobDefinition.DataQualityJobOutputConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_job_output_config for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("job_resources", flattenMonitoringResources(jobDefinition.JobResources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting job_resources for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("network_config", flattenMonitoringNetworkConfig(jobDefinition.NetworkConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_config for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	if err := d.Set("stopping_condition", flattenMonitoringStoppingCondition(jobDefinition.StoppingCondition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stopping_condition for SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}
	return diags
}

func resourceDataQualityJobDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceDataQualityJobDefinitionRead(ctx, d, meta)...)
}

func resourceDataQualityJobDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[INFO] Deleting SageMaker Data Quality Job Definition: %s", d.Id())
	_, err := conn.DeleteDataQualityJobDefinition(ctx, &sagemaker.DeleteDataQualityJobDefinitionInput{
		JobDefinitionName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	return diags
}

func findDataQualityJobDefinitionByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeDataQualityJobDefinitionOutput, error) {
	input := &sagemaker.DescribeDataQualityJobDefinitionInput{
		JobDefinitionName: aws.String(name),
	}

	output, err := conn.DescribeDataQualityJobDefinition(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
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

func flattenDataQualityAppSpecification(config *awstypes.DataQualityAppSpecification) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.ImageUri != nil {
		m["image_uri"] = aws.ToString(config.ImageUri)
	}

	if config.Environment != nil {
		m[names.AttrEnvironment] = aws.StringMap(config.Environment)
	}

	if config.PostAnalyticsProcessorSourceUri != nil {
		m["post_analytics_processor_source_uri"] = aws.ToString(config.PostAnalyticsProcessorSourceUri)
	}

	if config.RecordPreprocessorSourceUri != nil {
		m["record_preprocessor_source_uri"] = aws.ToString(config.RecordPreprocessorSourceUri)
	}

	return []map[string]interface{}{m}
}

func flattenDataQualityBaselineConfig(config *awstypes.DataQualityBaselineConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.ConstraintsResource != nil {
		m["constraints_resource"] = flattenConstraintsResource(config.ConstraintsResource)
	}

	if config.StatisticsResource != nil {
		m["statistics_resource"] = flattenMonitoringStatisticsResource(config.StatisticsResource)
	}

	return []map[string]interface{}{m}
}

func flattenConstraintsResource(config *awstypes.MonitoringConstraintsResource) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.S3Uri != nil {
		m["s3_uri"] = aws.ToString(config.S3Uri)
	}

	return []map[string]interface{}{m}
}

func flattenMonitoringStatisticsResource(config *awstypes.MonitoringStatisticsResource) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.S3Uri != nil {
		m["s3_uri"] = aws.ToString(config.S3Uri)
	}

	return []map[string]interface{}{m}
}

func flattenDataQualityJobInput(config *awstypes.DataQualityJobInput) []map[string]interface{} {
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

	return []map[string]interface{}{m}
}

func flattenBatchTransformInput(config *awstypes.BatchTransformInput) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.LocalPath != nil {
		m["local_path"] = aws.ToString(config.LocalPath)
	}

	if config.DataCapturedDestinationS3Uri != nil {
		m["data_captured_destination_s3_uri"] = aws.ToString(config.DataCapturedDestinationS3Uri)
	}

	if config.DatasetFormat != nil {
		m["dataset_format"] = flattenMonitoringDatasetFormat(config.DatasetFormat)
	}

	m["s3_data_distribution_type"] = config.S3DataDistributionType

	m["s3_input_mode"] = config.S3InputMode

	return []map[string]interface{}{m}
}

func flattenMonitoringDatasetFormat(config *awstypes.MonitoringDatasetFormat) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.Csv != nil {
		m["csv"] = flattenMonitoringCSVDatasetFormat(config.Csv)
	}

	if config.Json != nil {
		m[names.AttrJSON] = flattenMonitoringJSONDatasetFormat(config.Json)
	}

	return []map[string]interface{}{m}
}

func flattenMonitoringCSVDatasetFormat(config *awstypes.MonitoringCsvDatasetFormat) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.Header != nil {
		m[names.AttrHeader] = aws.ToBool(config.Header)
	}

	return []map[string]interface{}{m}
}

func flattenMonitoringJSONDatasetFormat(config *awstypes.MonitoringJsonDatasetFormat) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.Line != nil {
		m["line"] = aws.ToBool(config.Line)
	}

	return []map[string]interface{}{m}
}

func flattenEndpointInput(config *awstypes.EndpointInput) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.EndpointName != nil {
		m["endpoint_name"] = aws.ToString(config.EndpointName)
	}

	if config.LocalPath != nil {
		m["local_path"] = aws.ToString(config.LocalPath)
	}

	m["s3_data_distribution_type"] = config.S3DataDistributionType

	m["s3_input_mode"] = config.S3InputMode

	return []map[string]interface{}{m}
}

func flattenMonitoringOutputConfig(config *awstypes.MonitoringOutputConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.KmsKeyId != nil {
		m[names.AttrKMSKeyID] = aws.ToString(config.KmsKeyId)
	}

	if config.MonitoringOutputs != nil {
		m["monitoring_outputs"] = flattenMonitoringOutputs(config.MonitoringOutputs)
	}

	return []map[string]interface{}{m}
}

func flattenMonitoringOutputs(list []awstypes.MonitoringOutput) []map[string]interface{} {
	outputs := make([]map[string]interface{}, 0, len(list))

	for _, lRaw := range list {
		m := make(map[string]interface{})
		m["s3_output"] = flattenMonitoringS3Output(lRaw.S3Output)
		outputs = append(outputs, m)
	}

	return outputs
}

func flattenMonitoringS3Output(config *awstypes.MonitoringS3Output) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.LocalPath != nil {
		m["local_path"] = aws.ToString(config.LocalPath)
	}

	m["s3_upload_mode"] = config.S3UploadMode

	if config.S3Uri != nil {
		m["s3_uri"] = aws.ToString(config.S3Uri)
	}

	return []map[string]interface{}{m}
}

func flattenMonitoringResources(config *awstypes.MonitoringResources) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.ClusterConfig != nil {
		m["cluster_config"] = flattenMonitoringClusterConfig(config.ClusterConfig)
	}

	return []map[string]interface{}{m}
}

func flattenMonitoringClusterConfig(config *awstypes.MonitoringClusterConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.InstanceCount != nil {
		m[names.AttrInstanceCount] = aws.ToInt32(config.InstanceCount)
	}

	m[names.AttrInstanceType] = config.InstanceType

	if config.VolumeKmsKeyId != nil {
		m["volume_kms_key_id"] = aws.ToString(config.VolumeKmsKeyId)
	}

	if config.VolumeSizeInGB != nil {
		m["volume_size_in_gb"] = aws.ToInt32(config.VolumeSizeInGB)
	}

	return []map[string]interface{}{m}
}

func flattenMonitoringNetworkConfig(config *awstypes.MonitoringNetworkConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.EnableInterContainerTrafficEncryption != nil {
		m["enable_inter_container_traffic_encryption"] = aws.ToBool(config.EnableInterContainerTrafficEncryption)
	}

	if config.EnableNetworkIsolation != nil {
		m["enable_network_isolation"] = aws.ToBool(config.EnableNetworkIsolation)
	}

	if config.VpcConfig != nil {
		m[names.AttrVPCConfig] = flattenVPCConfig(config.VpcConfig)
	}

	return []map[string]interface{}{m}
}

func flattenVPCConfig(config *awstypes.VpcConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.SecurityGroupIds != nil {
		m[names.AttrSecurityGroupIDs] = flex.FlattenStringValueSet(config.SecurityGroupIds)
	}

	if config.Subnets != nil {
		m[names.AttrSubnets] = flex.FlattenStringValueSet(config.Subnets)
	}

	return []map[string]interface{}{m}
}

func flattenMonitoringStoppingCondition(config *awstypes.MonitoringStoppingCondition) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.MaxRuntimeInSeconds != nil {
		m["max_runtime_in_seconds"] = aws.ToInt32(config.MaxRuntimeInSeconds)
	}

	return []map[string]interface{}{m}
}

func expandDataQualityAppSpecification(configured []interface{}) *awstypes.DataQualityAppSpecification {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.DataQualityAppSpecification{}

	if v, ok := m["image_uri"].(string); ok && v != "" {
		c.ImageUri = aws.String(v)
	}

	if v, ok := m[names.AttrEnvironment].(map[string]interface{}); ok && len(v) > 0 {
		c.Environment = flex.ExpandStringValueMap(v)
	}

	if v, ok := m["post_analytics_processor_source_uri"].(string); ok && v != "" {
		c.PostAnalyticsProcessorSourceUri = aws.String(v)
	}

	if v, ok := m["record_preprocessor_source_uri"].(string); ok && v != "" {
		c.RecordPreprocessorSourceUri = aws.String(v)
	}

	return c
}

func expandDataQualityBaselineConfig(configured []interface{}) *awstypes.DataQualityBaselineConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.DataQualityBaselineConfig{}

	if v, ok := m["constraints_resource"].([]interface{}); ok && len(v) > 0 {
		c.ConstraintsResource = expandMonitoringConstraintsResource(v)
	}

	if v, ok := m["statistics_resource"].([]interface{}); ok && len(v) > 0 {
		c.StatisticsResource = expandMonitoringStatisticsResource(v)
	}

	return c
}

func expandMonitoringConstraintsResource(configured []interface{}) *awstypes.MonitoringConstraintsResource {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.MonitoringConstraintsResource{}

	if v, ok := m["s3_uri"].(string); ok && v != "" {
		c.S3Uri = aws.String(v)
	}

	return c
}

func expandMonitoringStatisticsResource(configured []interface{}) *awstypes.MonitoringStatisticsResource {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.MonitoringStatisticsResource{}

	if v, ok := m["s3_uri"].(string); ok && v != "" {
		c.S3Uri = aws.String(v)
	}

	return c
}

func expandDataQualityJobInput(configured []interface{}) *awstypes.DataQualityJobInput {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.DataQualityJobInput{}

	if v, ok := m["endpoint_input"].([]interface{}); ok && len(v) > 0 {
		c.EndpointInput = expandEndpointInput(v)
	}

	if v, ok := m["batch_transform_input"].([]interface{}); ok && len(v) > 0 {
		c.BatchTransformInput = expandBatchTransformInput(v)
	}

	return c
}

func expandEndpointInput(configured []interface{}) *awstypes.EndpointInput {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.EndpointInput{}

	if v, ok := m["endpoint_name"].(string); ok && v != "" {
		c.EndpointName = aws.String(v)
	}

	if v, ok := m["local_path"].(string); ok && v != "" {
		c.LocalPath = aws.String(v)
	}

	if v, ok := m["s3_data_distribution_type"].(string); ok && v != "" {
		c.S3DataDistributionType = awstypes.ProcessingS3DataDistributionType(v)
	}

	if v, ok := m["s3_input_mode"].(string); ok && v != "" {
		c.S3InputMode = awstypes.ProcessingS3InputMode(v)
	}

	return c
}

func expandBatchTransformInput(configured []interface{}) *awstypes.BatchTransformInput {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.BatchTransformInput{}

	if v, ok := m["data_captured_destination_s3_uri"].(string); ok && v != "" {
		c.DataCapturedDestinationS3Uri = aws.String(v)
	}

	if v, ok := m["dataset_format"].([]interface{}); ok && len(v) > 0 {
		c.DatasetFormat = expandMonitoringDatasetFormat(v)
	}

	if v, ok := m["local_path"].(string); ok && v != "" {
		c.LocalPath = aws.String(v)
	}

	if v, ok := m["s3_data_distribution_type"].(string); ok && v != "" {
		c.S3DataDistributionType = awstypes.ProcessingS3DataDistributionType(v)
	}

	if v, ok := m["s3_input_mode"].(string); ok && v != "" {
		c.S3InputMode = awstypes.ProcessingS3InputMode(v)
	}

	return c
}

func expandMonitoringDatasetFormat(configured []interface{}) *awstypes.MonitoringDatasetFormat {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.MonitoringDatasetFormat{}

	if v, ok := m["csv"].([]interface{}); ok && len(v) > 0 {
		c.Csv = expandMonitoringCSVDatasetFormat(v)
	}

	if v, ok := m[names.AttrJSON].([]interface{}); ok && len(v) > 0 {
		c.Json = expandMonitoringJSONDatasetFormat(v)
	}

	return c
}

func expandMonitoringJSONDatasetFormat(configured []interface{}) *awstypes.MonitoringJsonDatasetFormat {
	if len(configured) == 0 {
		return nil
	}

	c := &awstypes.MonitoringJsonDatasetFormat{}

	if configured[0] == nil {
		return c
	}

	m := configured[0].(map[string]interface{})
	if v, ok := m["line"]; ok {
		c.Line = aws.Bool(v.(bool))
	}

	return c
}

func expandMonitoringCSVDatasetFormat(configured []interface{}) *awstypes.MonitoringCsvDatasetFormat {
	if len(configured) == 0 {
		return nil
	}

	c := &awstypes.MonitoringCsvDatasetFormat{}

	if configured[0] == nil {
		return c
	}

	m := configured[0].(map[string]interface{})
	if v, ok := m[names.AttrHeader]; ok {
		c.Header = aws.Bool(v.(bool))
	}

	return c
}

func expandMonitoringOutputConfig(configured []interface{}) *awstypes.MonitoringOutputConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.MonitoringOutputConfig{}

	if v, ok := m[names.AttrKMSKeyID].(string); ok && v != "" {
		c.KmsKeyId = aws.String(v)
	}

	if v, ok := m["monitoring_outputs"].([]interface{}); ok && len(v) > 0 {
		c.MonitoringOutputs = expandMonitoringOutputs(v)
	}

	return c
}

func expandMonitoringOutputs(configured []interface{}) []awstypes.MonitoringOutput {
	containers := make([]awstypes.MonitoringOutput, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := awstypes.MonitoringOutput{
			S3Output: expandMonitoringS3Output(data["s3_output"].([]interface{})),
		}
		containers = append(containers, l)
	}

	return containers
}

func expandMonitoringS3Output(configured []interface{}) *awstypes.MonitoringS3Output {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.MonitoringS3Output{}

	if v, ok := m["local_path"].(string); ok && v != "" {
		c.LocalPath = aws.String(v)
	}

	if v, ok := m["s3_upload_mode"].(string); ok && v != "" {
		c.S3UploadMode = awstypes.ProcessingS3UploadMode(v)
	}

	if v, ok := m["s3_uri"].(string); ok && v != "" {
		c.S3Uri = aws.String(v)
	}

	return c
}

func expandMonitoringResources(configured []interface{}) *awstypes.MonitoringResources {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.MonitoringResources{}

	if v, ok := m["cluster_config"].([]interface{}); ok && len(v) > 0 {
		c.ClusterConfig = expandMonitoringClusterConfig(v)
	}

	return c
}

func expandMonitoringClusterConfig(configured []interface{}) *awstypes.MonitoringClusterConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.MonitoringClusterConfig{}

	if v, ok := m[names.AttrInstanceCount].(int); ok && v > 0 {
		c.InstanceCount = aws.Int32(int32(v))
	}

	if v, ok := m[names.AttrInstanceType].(string); ok && v != "" {
		c.InstanceType = awstypes.ProcessingInstanceType(v)
	}

	if v, ok := m["volume_kms_key_id"].(string); ok && v != "" {
		c.VolumeKmsKeyId = aws.String(v)
	}

	if v, ok := m["volume_size_in_gb"].(int); ok && v > 0 {
		c.VolumeSizeInGB = aws.Int32(int32(v))
	}

	return c
}

func expandMonitoringNetworkConfig(configured []interface{}) *awstypes.MonitoringNetworkConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.MonitoringNetworkConfig{}

	if v, ok := m["enable_inter_container_traffic_encryption"]; ok {
		c.EnableInterContainerTrafficEncryption = aws.Bool(v.(bool))
	}

	if v, ok := m["enable_network_isolation"]; ok {
		c.EnableNetworkIsolation = aws.Bool(v.(bool))
	}

	if v, ok := m[names.AttrVPCConfig].([]interface{}); ok && len(v) > 0 {
		c.VpcConfig = expandVPCConfig(v)
	}

	return c
}

func expandVPCConfig(configured []interface{}) *awstypes.VpcConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.VpcConfig{}

	if v, ok := m[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		c.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := m[names.AttrSubnets].(*schema.Set); ok && v.Len() > 0 {
		c.Subnets = flex.ExpandStringValueSet(v)
	}

	return c
}

func expandMonitoringStoppingCondition(configured []interface{}) *awstypes.MonitoringStoppingCondition {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.MonitoringStoppingCondition{}

	if v, ok := m["max_runtime_in_seconds"].(int); ok && v > 0 {
		c.MaxRuntimeInSeconds = aws.Int32(int32(v))
	}

	return c
}
