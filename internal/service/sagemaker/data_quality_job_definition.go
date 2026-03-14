// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
	}
}

func resourceDataQualityJobDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	var name string
	if v, ok := d.GetOk(names.AttrName); ok {
		name = v.(string)
	} else {
		name = sdkid.UniqueId()
	}

	var roleArn string
	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		roleArn = v.(string)
	}

	input := sagemaker.CreateDataQualityJobDefinitionInput{
		DataQualityAppSpecification: expandDataQualityAppSpecification(d.Get("data_quality_app_specification").([]any)),
		DataQualityJobInput:         expandDataQualityJobInput(d.Get("data_quality_job_input").([]any)),
		DataQualityJobOutputConfig:  expandMonitoringOutputConfig(d.Get("data_quality_job_output_config").([]any)),
		JobDefinitionName:           aws.String(name),
		JobResources:                expandMonitoringResources(d.Get("job_resources").([]any)),
		RoleArn:                     aws.String(roleArn),
		Tags:                        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("data_quality_baseline_config"); ok && len(v.([]any)) > 0 {
		input.DataQualityBaselineConfig = expandDataQualityBaselineConfig(v.([]any))
	}

	if v, ok := d.GetOk("network_config"); ok && len(v.([]any)) > 0 {
		input.NetworkConfig = expandMonitoringNetworkConfig(v.([]any))
	}

	if v, ok := d.GetOk("stopping_condition"); ok && len(v.([]any)) > 0 {
		input.StoppingCondition = expandMonitoringStoppingCondition(v.([]any))
	}

	_, err := conn.CreateDataQualityJobDefinition(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Data Quality Job Definition (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceDataQualityJobDefinitionRead(ctx, d, meta)...)
}

func resourceDataQualityJobDefinitionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	jobDefinition, err := findDataQualityJobDefinitionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Data Quality Job Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, jobDefinition.JobDefinitionArn)
	if err := d.Set("data_quality_app_specification", flattenDataQualityAppSpecification(jobDefinition.DataQualityAppSpecification)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_app_specification: %s", err)
	}
	if err := d.Set("data_quality_baseline_config", flattenDataQualityBaselineConfig(jobDefinition.DataQualityBaselineConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_baseline_config: %s", err)
	}
	if err := d.Set("data_quality_job_input", flattenDataQualityJobInput(jobDefinition.DataQualityJobInput)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_job_input: %s", err)
	}
	if err := d.Set("data_quality_job_output_config", flattenMonitoringOutputConfig(jobDefinition.DataQualityJobOutputConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_quality_job_output_config: %s", err)
	}
	if err := d.Set("job_resources", flattenMonitoringResources(jobDefinition.JobResources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting job_resources: %s", err)
	}
	d.Set(names.AttrName, jobDefinition.JobDefinitionName)
	if err := d.Set("network_config", flattenMonitoringNetworkConfig(jobDefinition.NetworkConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_config: %s", err)
	}
	d.Set(names.AttrRoleARN, jobDefinition.RoleArn)
	if err := d.Set("stopping_condition", flattenMonitoringStoppingCondition(jobDefinition.StoppingCondition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stopping_condition: %s", err)
	}

	return diags
}

func resourceDataQualityJobDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceDataQualityJobDefinitionRead(ctx, d, meta)...)
}

func resourceDataQualityJobDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[INFO] Deleting SageMaker AI Data Quality Job Definition: %s", d.Id())
	input := sagemaker.DeleteDataQualityJobDefinitionInput{
		JobDefinitionName: aws.String(d.Id()),
	}
	_, err := conn.DeleteDataQualityJobDefinition(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Data Quality Job Definition (%s): %s", d.Id(), err)
	}

	return diags
}

func findDataQualityJobDefinitionByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeDataQualityJobDefinitionOutput, error) {
	input := sagemaker.DescribeDataQualityJobDefinitionInput{
		JobDefinitionName: aws.String(name),
	}

	return findDataQualityJobDefinition(ctx, conn, &input)
}

func findDataQualityJobDefinition(ctx context.Context, conn *sagemaker.Client, input *sagemaker.DescribeDataQualityJobDefinitionInput) (*sagemaker.DescribeDataQualityJobDefinitionOutput, error) {
	output, err := conn.DescribeDataQualityJobDefinition(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
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

func flattenDataQualityAppSpecification(apiObject *awstypes.DataQualityAppSpecification) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.Environment != nil {
		tfMap[names.AttrEnvironment] = apiObject.Environment
	}

	if apiObject.ImageUri != nil {
		tfMap["image_uri"] = aws.ToString(apiObject.ImageUri)
	}

	if apiObject.PostAnalyticsProcessorSourceUri != nil {
		tfMap["post_analytics_processor_source_uri"] = aws.ToString(apiObject.PostAnalyticsProcessorSourceUri)
	}

	if apiObject.RecordPreprocessorSourceUri != nil {
		tfMap["record_preprocessor_source_uri"] = aws.ToString(apiObject.RecordPreprocessorSourceUri)
	}

	return []any{tfMap}
}

func flattenDataQualityBaselineConfig(apiObject *awstypes.DataQualityBaselineConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.ConstraintsResource != nil {
		tfMap["constraints_resource"] = flattenMonitoringConstraintsResource(apiObject.ConstraintsResource)
	}

	if apiObject.StatisticsResource != nil {
		tfMap["statistics_resource"] = flattenMonitoringStatisticsResource(apiObject.StatisticsResource)
	}

	return []any{tfMap}
}

func flattenMonitoringConstraintsResource(apiObject *awstypes.MonitoringConstraintsResource) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.S3Uri != nil {
		tfMap["s3_uri"] = aws.ToString(apiObject.S3Uri)
	}

	return []any{tfMap}
}

func flattenMonitoringStatisticsResource(apiObject *awstypes.MonitoringStatisticsResource) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.S3Uri != nil {
		tfMap["s3_uri"] = aws.ToString(apiObject.S3Uri)
	}

	return []any{tfMap}
}

func flattenDataQualityJobInput(apiObject *awstypes.DataQualityJobInput) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.BatchTransformInput != nil {
		tfMap["batch_transform_input"] = flattenBatchTransformInput(apiObject.BatchTransformInput)
	}

	if apiObject.EndpointInput != nil {
		tfMap["endpoint_input"] = flattenEndpointInput(apiObject.EndpointInput)
	}

	return []any{tfMap}
}

func flattenBatchTransformInput(apiObject *awstypes.BatchTransformInput) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.DataCapturedDestinationS3Uri != nil {
		tfMap["data_captured_destination_s3_uri"] = aws.ToString(apiObject.DataCapturedDestinationS3Uri)
	}

	if apiObject.DatasetFormat != nil {
		tfMap["dataset_format"] = flattenMonitoringDatasetFormat(apiObject.DatasetFormat)
	}

	if apiObject.LocalPath != nil {
		tfMap["local_path"] = aws.ToString(apiObject.LocalPath)
	}

	tfMap["s3_data_distribution_type"] = apiObject.S3DataDistributionType
	tfMap["s3_input_mode"] = apiObject.S3InputMode

	return []any{tfMap}
}

func flattenMonitoringDatasetFormat(apiObject *awstypes.MonitoringDatasetFormat) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.Csv != nil {
		tfMap["csv"] = flattenMonitoringCSVDatasetFormat(apiObject.Csv)
	}

	if apiObject.Json != nil {
		tfMap[names.AttrJSON] = flattenMonitoringJSONDatasetFormat(apiObject.Json)
	}

	return []any{tfMap}
}

func flattenMonitoringCSVDatasetFormat(apiObject *awstypes.MonitoringCsvDatasetFormat) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.Header != nil {
		tfMap[names.AttrHeader] = aws.ToBool(apiObject.Header)
	}

	return []any{tfMap}
}

func flattenMonitoringJSONDatasetFormat(apiObject *awstypes.MonitoringJsonDatasetFormat) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.Line != nil {
		tfMap["line"] = aws.ToBool(apiObject.Line)
	}

	return []any{tfMap}
}

func flattenEndpointInput(apiObject *awstypes.EndpointInput) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.EndpointName != nil {
		tfMap["endpoint_name"] = aws.ToString(apiObject.EndpointName)
	}

	if apiObject.LocalPath != nil {
		tfMap["local_path"] = aws.ToString(apiObject.LocalPath)
	}

	tfMap["s3_data_distribution_type"] = apiObject.S3DataDistributionType
	tfMap["s3_input_mode"] = apiObject.S3InputMode

	return []any{tfMap}
}

func flattenMonitoringOutputConfig(apiObject *awstypes.MonitoringOutputConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.KmsKeyId != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(apiObject.KmsKeyId)
	}

	if apiObject.MonitoringOutputs != nil {
		tfMap["monitoring_outputs"] = flattenMonitoringOutputs(apiObject.MonitoringOutputs)
	}

	return []any{tfMap}
}

func flattenMonitoringOutputs(apiObjects []awstypes.MonitoringOutput) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)
		tfMap["s3_output"] = flattenMonitoringS3Output(apiObject.S3Output)
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenMonitoringS3Output(apiObject *awstypes.MonitoringS3Output) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.LocalPath != nil {
		tfMap["local_path"] = aws.ToString(apiObject.LocalPath)
	}

	tfMap["s3_upload_mode"] = apiObject.S3UploadMode

	if apiObject.S3Uri != nil {
		tfMap["s3_uri"] = aws.ToString(apiObject.S3Uri)
	}

	return []any{tfMap}
}

func flattenMonitoringResources(apiObjects *awstypes.MonitoringResources) []any {
	if apiObjects == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObjects.ClusterConfig != nil {
		tfMap["cluster_config"] = flattenMonitoringClusterConfig(apiObjects.ClusterConfig)
	}

	return []any{tfMap}
}

func flattenMonitoringClusterConfig(apiObject *awstypes.MonitoringClusterConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.InstanceCount != nil {
		tfMap[names.AttrInstanceCount] = aws.ToInt32(apiObject.InstanceCount)
	}

	tfMap[names.AttrInstanceType] = apiObject.InstanceType

	if apiObject.VolumeKmsKeyId != nil {
		tfMap["volume_kms_key_id"] = aws.ToString(apiObject.VolumeKmsKeyId)
	}

	if apiObject.VolumeSizeInGB != nil {
		tfMap["volume_size_in_gb"] = aws.ToInt32(apiObject.VolumeSizeInGB)
	}

	return []any{tfMap}
}

func flattenMonitoringNetworkConfig(apiObject *awstypes.MonitoringNetworkConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.EnableInterContainerTrafficEncryption != nil {
		tfMap["enable_inter_container_traffic_encryption"] = aws.ToBool(apiObject.EnableInterContainerTrafficEncryption)
	}

	if apiObject.EnableNetworkIsolation != nil {
		tfMap["enable_network_isolation"] = aws.ToBool(apiObject.EnableNetworkIsolation)
	}

	if apiObject.VpcConfig != nil {
		tfMap[names.AttrVPCConfig] = flattenVPCConfig(apiObject.VpcConfig)
	}

	return []any{tfMap}
}

func flattenVPCConfig(apiObject *awstypes.VpcConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrSecurityGroupIDs: apiObject.SecurityGroupIds,
		names.AttrSubnets:          apiObject.Subnets,
	}

	return []any{tfMap}
}

func flattenMonitoringStoppingCondition(apiObject *awstypes.MonitoringStoppingCondition) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.MaxRuntimeInSeconds != nil {
		tfMap["max_runtime_in_seconds"] = aws.ToInt32(apiObject.MaxRuntimeInSeconds)
	}

	return []any{tfMap}
}

func expandDataQualityAppSpecification(tfList []any) *awstypes.DataQualityAppSpecification {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.DataQualityAppSpecification{}

	if v, ok := tfMap[names.AttrEnvironment].(map[string]any); ok && len(v) > 0 {
		apiObject.Environment = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["image_uri"].(string); ok && v != "" {
		apiObject.ImageUri = aws.String(v)
	}

	if v, ok := tfMap["post_analytics_processor_source_uri"].(string); ok && v != "" {
		apiObject.PostAnalyticsProcessorSourceUri = aws.String(v)
	}

	if v, ok := tfMap["record_preprocessor_source_uri"].(string); ok && v != "" {
		apiObject.RecordPreprocessorSourceUri = aws.String(v)
	}

	return apiObject
}

func expandDataQualityBaselineConfig(tfList []any) *awstypes.DataQualityBaselineConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.DataQualityBaselineConfig{}

	if v, ok := tfMap["constraints_resource"].([]any); ok && len(v) > 0 {
		apiObject.ConstraintsResource = expandMonitoringConstraintsResource(v)
	}

	if v, ok := tfMap["statistics_resource"].([]any); ok && len(v) > 0 {
		apiObject.StatisticsResource = expandMonitoringStatisticsResource(v)
	}

	return apiObject
}

func expandMonitoringConstraintsResource(tfList []any) *awstypes.MonitoringConstraintsResource {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringConstraintsResource{}

	if v, ok := tfMap["s3_uri"].(string); ok && v != "" {
		apiObject.S3Uri = aws.String(v)
	}

	return apiObject
}

func expandMonitoringStatisticsResource(tfList []any) *awstypes.MonitoringStatisticsResource {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringStatisticsResource{}

	if v, ok := tfMap["s3_uri"].(string); ok && v != "" {
		apiObject.S3Uri = aws.String(v)
	}

	return apiObject
}

func expandDataQualityJobInput(tfList []any) *awstypes.DataQualityJobInput {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.DataQualityJobInput{}

	if v, ok := tfMap["batch_transform_input"].([]any); ok && len(v) > 0 {
		apiObject.BatchTransformInput = expandBatchTransformInput(v)
	}

	if v, ok := tfMap["endpoint_input"].([]any); ok && len(v) > 0 {
		apiObject.EndpointInput = expandEndpointInput(v)
	}

	return apiObject
}

func expandEndpointInput(tfList []any) *awstypes.EndpointInput {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.EndpointInput{}

	if v, ok := tfMap["endpoint_name"].(string); ok && v != "" {
		apiObject.EndpointName = aws.String(v)
	}

	if v, ok := tfMap["local_path"].(string); ok && v != "" {
		apiObject.LocalPath = aws.String(v)
	}

	if v, ok := tfMap["s3_data_distribution_type"].(string); ok && v != "" {
		apiObject.S3DataDistributionType = awstypes.ProcessingS3DataDistributionType(v)
	}

	if v, ok := tfMap["s3_input_mode"].(string); ok && v != "" {
		apiObject.S3InputMode = awstypes.ProcessingS3InputMode(v)
	}

	return apiObject
}

func expandBatchTransformInput(tfList []any) *awstypes.BatchTransformInput {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.BatchTransformInput{}

	if v, ok := tfMap["data_captured_destination_s3_uri"].(string); ok && v != "" {
		apiObject.DataCapturedDestinationS3Uri = aws.String(v)
	}

	if v, ok := tfMap["dataset_format"].([]any); ok && len(v) > 0 {
		apiObject.DatasetFormat = expandMonitoringDatasetFormat(v)
	}

	if v, ok := tfMap["local_path"].(string); ok && v != "" {
		apiObject.LocalPath = aws.String(v)
	}

	if v, ok := tfMap["s3_data_distribution_type"].(string); ok && v != "" {
		apiObject.S3DataDistributionType = awstypes.ProcessingS3DataDistributionType(v)
	}

	if v, ok := tfMap["s3_input_mode"].(string); ok && v != "" {
		apiObject.S3InputMode = awstypes.ProcessingS3InputMode(v)
	}

	return apiObject
}

func expandMonitoringDatasetFormat(tfList []any) *awstypes.MonitoringDatasetFormat {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringDatasetFormat{}

	if v, ok := tfMap["csv"].([]any); ok && len(v) > 0 {
		apiObject.Csv = expandMonitoringCSVDatasetFormat(v)
	}

	if v, ok := tfMap[names.AttrJSON].([]any); ok && len(v) > 0 {
		apiObject.Json = expandMonitoringJSONDatasetFormat(v)
	}

	return apiObject
}

func expandMonitoringCSVDatasetFormat(tfList []any) *awstypes.MonitoringCsvDatasetFormat {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &awstypes.MonitoringCsvDatasetFormat{}

	if tfList[0] == nil {
		return apiObject
	}

	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap[names.AttrHeader]; ok {
		apiObject.Header = aws.Bool(v.(bool))
	}

	return apiObject
}

func expandMonitoringJSONDatasetFormat(tfList []any) *awstypes.MonitoringJsonDatasetFormat {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &awstypes.MonitoringJsonDatasetFormat{}

	if tfList[0] == nil {
		return apiObject
	}

	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap["line"]; ok {
		apiObject.Line = aws.Bool(v.(bool))
	}

	return apiObject
}

func expandMonitoringOutputConfig(tfList []any) *awstypes.MonitoringOutputConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringOutputConfig{}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["monitoring_outputs"].([]any); ok && len(v) > 0 {
		apiObject.MonitoringOutputs = expandMonitoringOutputs(v)
	}

	return apiObject
}

func expandMonitoringOutputs(tfList []any) []awstypes.MonitoringOutput {
	apiObjects := make([]awstypes.MonitoringOutput, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.MonitoringOutput{
			S3Output: expandMonitoringS3Output(tfMap["s3_output"].([]any)),
		}
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandMonitoringS3Output(tfList []any) *awstypes.MonitoringS3Output {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringS3Output{}

	if v, ok := tfMap["local_path"].(string); ok && v != "" {
		apiObject.LocalPath = aws.String(v)
	}

	if v, ok := tfMap["s3_upload_mode"].(string); ok && v != "" {
		apiObject.S3UploadMode = awstypes.ProcessingS3UploadMode(v)
	}

	if v, ok := tfMap["s3_uri"].(string); ok && v != "" {
		apiObject.S3Uri = aws.String(v)
	}

	return apiObject
}

func expandMonitoringResources(tfList []any) *awstypes.MonitoringResources {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringResources{}

	if v, ok := tfMap["cluster_config"].([]any); ok && len(v) > 0 {
		apiObject.ClusterConfig = expandMonitoringClusterConfig(v)
	}

	return apiObject
}

func expandMonitoringClusterConfig(tfList []any) *awstypes.MonitoringClusterConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringClusterConfig{}

	if v, ok := tfMap[names.AttrInstanceCount].(int); ok && v > 0 {
		apiObject.InstanceCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrInstanceType].(string); ok && v != "" {
		apiObject.InstanceType = awstypes.ProcessingInstanceType(v)
	}

	if v, ok := tfMap["volume_kms_key_id"].(string); ok && v != "" {
		apiObject.VolumeKmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["volume_size_in_gb"].(int); ok && v > 0 {
		apiObject.VolumeSizeInGB = aws.Int32(int32(v))
	}

	return apiObject
}

func expandMonitoringNetworkConfig(tfList []any) *awstypes.MonitoringNetworkConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringNetworkConfig{}

	if v, ok := tfMap["enable_inter_container_traffic_encryption"]; ok {
		apiObject.EnableInterContainerTrafficEncryption = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["enable_network_isolation"]; ok {
		apiObject.EnableNetworkIsolation = aws.Bool(v.(bool))
	}

	if v, ok := tfMap[names.AttrVPCConfig].([]any); ok && len(v) > 0 {
		apiObject.VpcConfig = expandVPCConfig(v)
	}

	return apiObject
}

func expandVPCConfig(tfList []any) *awstypes.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.VpcConfig{}

	if v, ok := tfMap[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnets].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandMonitoringStoppingCondition(tfList []any) *awstypes.MonitoringStoppingCondition {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringStoppingCondition{}

	if v, ok := tfMap["max_runtime_in_seconds"].(int); ok && v > 0 {
		apiObject.MaxRuntimeInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}
