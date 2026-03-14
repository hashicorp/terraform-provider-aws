// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

import (
	"context"
	"errors"
	"log"
	"time"

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

// @SDKResource("aws_sagemaker_monitoring_schedule", name="Monitoring Schedule")
// @Tags(identifierAttribute="arn")
func resourceMonitoringSchedule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMonitoringScheduleCreate,
		ReadWithoutTimeout:   resourceMonitoringScheduleRead,
		UpdateWithoutTimeout: resourceMonitoringScheduleUpdate,
		DeleteWithoutTimeout: resourceMonitoringScheduleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"monitoring_schedule_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"monitoring_job_definition": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"baseline": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"baselining_job_name": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"constraints_resource": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"s3_uri": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validHTTPSOrS3URI,
															},
														},
													},
												},
												"statistics_resource": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"s3_uri": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validHTTPSOrS3URI,
															},
														},
													},
												},
											},
										},
									},
									names.AttrEnvironment: {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"monitoring_app_specification": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"container_arguments": {
													Type:     schema.TypeList,
													MaxItems: 50,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"container_entrypoint": {
													Type:     schema.TypeList,
													MaxItems: 100,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"image_uri": {
													Type:     schema.TypeString,
													Required: true,
												},
												"post_analytics_processor_source_uri": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validHTTPSOrS3URI,
												},
												"record_preprocessor_source_uri": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validHTTPSOrS3URI,
												},
											},
										},
									},
									"monitoring_inputs": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"batch_transform_input": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"data_captured_destination_s3_uri": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validHTTPSOrS3URI,
															},
															"dataset_format": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Required: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"csv": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Optional: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					names.AttrHeader: {
																						Type:     schema.TypeBool,
																						Optional: true,
																					},
																				},
																			},
																		},
																		names.AttrJSON: {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Optional: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"line": {
																						Type:     schema.TypeBool,
																						Optional: true,
																					},
																				},
																			},
																		},
																	},
																},
															},
															"end_time_offset": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"exclude_features_attribute": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"features_attribute": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"inference_attribute": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"local_path": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 256),
															},
															"probability_attribute": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(1, 256),
															},
															"probability_threshold_attribute": {
																Type:         schema.TypeFloat,
																Optional:     true,
																ValidateFunc: validation.FloatBetween(0, 1),
															},
															"s3_data_distribution_type": {
																Type:             schema.TypeString,
																Optional:         true,
																Computed:         true,
																ValidateDiagFunc: enum.Validate[awstypes.ProcessingS3DataDistributionType](),
															},
															"s3_input_mode": {
																Type:             schema.TypeString,
																Optional:         true,
																Computed:         true,
																ValidateDiagFunc: enum.Validate[awstypes.ProcessingS3InputMode](),
															},
															"start_time_offset": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
												"endpoint_input": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"end_time_offset": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"endpoint_name": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 63),
															},
															"exclude_features_attribute": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"features_attribute": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"inference_attribute": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"local_path": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 256),
															},
															"probability_attribute": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"probability_threshold_attribute": {
																Type:     schema.TypeFloat,
																Optional: true,
															},
															"s3_data_distribution_type": {
																Type:             schema.TypeString,
																Optional:         true,
																Computed:         true,
																ValidateDiagFunc: enum.Validate[awstypes.ProcessingS3DataDistributionType](),
															},
															"s3_input_mode": {
																Type:             schema.TypeString,
																Optional:         true,
																Computed:         true,
																ValidateDiagFunc: enum.Validate[awstypes.ProcessingS3InputMode](),
															},
															"start_time_offset": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"monitoring_output_config": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrKMSKeyID: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"monitoring_outputs": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"s3_output": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Required: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"local_path": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																		"s3_upload_mode": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.ProcessingS3UploadMode](),
																		},
																		"s3_uri": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validHTTPSOrS3URI,
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
									"monitoring_resources": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cluster_config": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrInstanceCount: {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(1, 100),
															},
															names.AttrInstanceType: {
																Type:     schema.TypeString,
																Required: true,
															},
															"volume_kms_key_id": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: verify.ValidARN,
															},
															"volume_size_in_gb": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(1, 16384),
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
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enable_inter_container_traffic_encryption": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"enable_network_isolation": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												names.AttrVPCConfig: {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrSecurityGroupIDs: {
																Type:     schema.TypeSet,
																Required: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															names.AttrSubnets: {
																Type:     schema.TypeSet,
																Required: true,
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
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"max_runtime_in_seconds": {
													Type:         schema.TypeInt,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.IntBetween(1, 86400),
												},
											},
										},
									},
								},
							},
						},
						"monitoring_job_definition_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validName,
						},
						"monitoring_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.MonitoringType](),
						},
						"schedule_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrScheduleExpression: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceMonitoringScheduleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	var name string
	if v, ok := d.GetOk(names.AttrName); ok {
		name = v.(string)
	} else {
		name = sdkid.UniqueId()
	}
	input := sagemaker.CreateMonitoringScheduleInput{
		MonitoringScheduleConfig: expandMonitoringScheduleConfig(d.Get("monitoring_schedule_config").([]any)),
		MonitoringScheduleName:   aws.String(name),
		Tags:                     getTagsIn(ctx),
	}

	_, err := conn.CreateMonitoringSchedule(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Monitoring Schedule (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitMonitoringScheduleScheduled(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Monitoring Schedule (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceMonitoringScheduleRead(ctx, d, meta)...)
}

func resourceMonitoringScheduleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	monitoringSchedule, err := findMonitoringScheduleByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Monitoring Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Monitoring Schedule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, monitoringSchedule.MonitoringScheduleArn)
	if err := d.Set("monitoring_schedule_config", flattenMonitoringScheduleConfig(monitoringSchedule.MonitoringScheduleConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting monitoring_schedule_config: %s", err)
	}
	d.Set(names.AttrName, monitoringSchedule.MonitoringScheduleName)

	return diags
}

func resourceMonitoringScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChanges("monitoring_schedule_config") {
		input := sagemaker.UpdateMonitoringScheduleInput{
			MonitoringScheduleName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("monitoring_schedule_config"); ok && (len(v.([]any)) > 0) {
			input.MonitoringScheduleConfig = expandMonitoringScheduleConfig(v.([]any))
		}

		_, err := conn.UpdateMonitoringSchedule(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Monitoring Schedule (%s): %s", d.Id(), err)
		}

		if _, err := waitMonitoringScheduleScheduled(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Monitoring Schedule (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMonitoringScheduleRead(ctx, d, meta)...)
}

func resourceMonitoringScheduleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[INFO] Deleting SageMaker AI Monitoring Schedule: %s", d.Id())
	input := sagemaker.DeleteMonitoringScheduleInput{
		MonitoringScheduleName: aws.String(d.Id()),
	}
	_, err := conn.DeleteMonitoringSchedule(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Monitoring Schedule (%s): %s", d.Id(), err)
	}

	if _, err := waitMonitoringScheduleDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Monitoring Schedule (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findMonitoringScheduleByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeMonitoringScheduleOutput, error) {
	input := sagemaker.DescribeMonitoringScheduleInput{
		MonitoringScheduleName: aws.String(name),
	}

	return findMonitoringSchedule(ctx, conn, &input)
}

func findMonitoringSchedule(ctx context.Context, conn *sagemaker.Client, input *sagemaker.DescribeMonitoringScheduleInput) (*sagemaker.DescribeMonitoringScheduleOutput, error) {
	output, err := conn.DescribeMonitoringSchedule(ctx, input)

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

func statusMonitoringSchedule(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findMonitoringScheduleByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.MonitoringScheduleStatus), nil
	}
}

func waitMonitoringScheduleScheduled(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeMonitoringScheduleOutput, error) { //nolint:unparam
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScheduleStatusPending),
		Target:  enum.Slice(awstypes.ScheduleStatusScheduled),
		Refresh: statusMonitoringSchedule(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeMonitoringScheduleOutput); ok {
		if status, reason := output.MonitoringScheduleStatus, aws.ToString(output.FailureReason); status == awstypes.ScheduleStatusFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitMonitoringScheduleDeleted(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeMonitoringScheduleOutput, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScheduleStatusScheduled, awstypes.ScheduleStatusPending, awstypes.ScheduleStatusStopped),
		Target:  []string{},
		Refresh: statusMonitoringSchedule(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeMonitoringScheduleOutput); ok {
		if status, reason := output.MonitoringScheduleStatus, aws.ToString(output.FailureReason); status == awstypes.ScheduleStatusFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func expandMonitoringScheduleConfig(tfList []any) *awstypes.MonitoringScheduleConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringScheduleConfig{}

	if v, ok := tfMap["monitoring_job_definition"].([]any); ok && len(v) > 0 {
		apiObject.MonitoringJobDefinition = expandMonitoringJobDefinition(v)
	}

	if v, ok := tfMap["monitoring_job_definition_name"].(string); ok && v != "" {
		apiObject.MonitoringJobDefinitionName = aws.String(v)
	}

	if v, ok := tfMap["monitoring_type"].(string); ok && v != "" {
		apiObject.MonitoringType = awstypes.MonitoringType(v)
	}

	if v, ok := tfMap["schedule_config"].([]any); ok && len(v) > 0 {
		apiObject.ScheduleConfig = expandScheduleConfig(v)
	}

	return apiObject
}

func expandMonitoringJobDefinition(tfList []any) *awstypes.MonitoringJobDefinition {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringJobDefinition{}

	if v, ok := tfMap["baseline"].([]any); ok && len(v) > 0 {
		apiObject.BaselineConfig = expandMonitoringBaselineConfig(v)
	}

	if v, ok := tfMap[names.AttrEnvironment].(map[string]any); ok && len(v) > 0 {
		apiObject.Environment = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["monitoring_app_specification"].([]any); ok && len(v) > 0 {
		apiObject.MonitoringAppSpecification = expandMonitoringAppSpecification(v)
	}

	if v, ok := tfMap["monitoring_inputs"].([]any); ok && len(v) > 0 {
		apiObject.MonitoringInputs = expandMonitoringInputs(v)
	}

	if v, ok := tfMap["monitoring_output_config"].([]any); ok && len(v) > 0 {
		apiObject.MonitoringOutputConfig = expandMonitoringOutputConfig(v)
	}

	if v, ok := tfMap["monitoring_resources"].([]any); ok && len(v) > 0 {
		apiObject.MonitoringResources = expandMonitoringResources(v)
	}

	if v, ok := tfMap["network_config"].([]any); ok && len(v) > 0 {
		apiObject.NetworkConfig = expandNetworkConfig(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["stopping_condition"].([]any); ok && len(v) > 0 {
		apiObject.StoppingCondition = expandMonitoringStoppingCondition(v)
	}

	return apiObject
}

func expandMonitoringBaselineConfig(tfList []any) *awstypes.MonitoringBaselineConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringBaselineConfig{}

	if v, ok := tfMap["baselining_job_name"].(string); ok && v != "" {
		apiObject.BaseliningJobName = aws.String(v)
	}

	if v, ok := tfMap["constraints_resource"].([]any); ok && len(v) > 0 {
		apiObject.ConstraintsResource = expandMonitoringConstraintsResource(v)
	}

	if v, ok := tfMap["statistics_resource"].([]any); ok && len(v) > 0 {
		apiObject.StatisticsResource = expandMonitoringStatisticsResource(v)
	}

	return apiObject
}

func expandMonitoringAppSpecification(tfList []any) *awstypes.MonitoringAppSpecification {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.MonitoringAppSpecification{}

	if v, ok := tfMap["container_arguments"].([]any); ok && len(v) > 0 {
		apiObject.ContainerArguments = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["container_entrypoint"].([]any); ok && len(v) > 0 {
		apiObject.ContainerEntrypoint = flex.ExpandStringValueList(v)
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

func expandMonitoringInputs(tfList []any) []awstypes.MonitoringInput {
	apiObjects := make([]awstypes.MonitoringInput, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.MonitoringInput{}

		if v, ok := tfMap["batch_transform_input"].([]any); ok && len(v) > 0 {
			apiObject.BatchTransformInput = expandBatchTransformInput(v)
		}

		if v, ok := tfMap["endpoint_input"].([]any); ok && len(v) > 0 {
			apiObject.EndpointInput = expandEndpointInput(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandNetworkConfig(tfList []any) *awstypes.NetworkConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.NetworkConfig{}

	if v, ok := tfMap["enable_inter_container_traffic_encryption"].(bool); ok {
		apiObject.EnableInterContainerTrafficEncryption = aws.Bool(v)
	}

	if v, ok := tfMap["enable_network_isolation"].(bool); ok {
		apiObject.EnableNetworkIsolation = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrVPCConfig].([]any); ok && len(v) > 0 {
		apiObject.VpcConfig = expandVPCConfig(v)
	}

	return apiObject
}

func expandScheduleConfig(tfList []any) *awstypes.ScheduleConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ScheduleConfig{}

	if v, ok := tfMap[names.AttrScheduleExpression].(string); ok && v != "" {
		apiObject.ScheduleExpression = aws.String(v)
	}

	return apiObject
}

func flattenMonitoringScheduleConfig(apiObject *awstypes.MonitoringScheduleConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.MonitoringJobDefinition != nil {
		tfMap["monitoring_job_definition"] = flattenMonitoringJobDefinition(apiObject.MonitoringJobDefinition)
	}

	if apiObject.MonitoringJobDefinitionName != nil {
		tfMap["monitoring_job_definition_name"] = aws.ToString(apiObject.MonitoringJobDefinitionName)
	}

	tfMap["monitoring_type"] = apiObject.MonitoringType

	if apiObject.ScheduleConfig != nil {
		tfMap["schedule_config"] = flattenScheduleConfig(apiObject.ScheduleConfig)
	}

	return []any{tfMap}
}

func flattenMonitoringJobDefinition(apiObject *awstypes.MonitoringJobDefinition) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.BaselineConfig != nil {
		tfMap["baseline"] = flattenMonitoringBaselineConfig(apiObject.BaselineConfig)
	}

	if apiObject.Environment != nil {
		tfMap[names.AttrEnvironment] = apiObject.Environment
	}

	if apiObject.MonitoringAppSpecification != nil {
		tfMap["monitoring_app_specification"] = flattenMonitoringAppSpecification(apiObject.MonitoringAppSpecification)
	}

	if apiObject.MonitoringInputs != nil {
		tfMap["monitoring_inputs"] = flattenMonitoringInputs(apiObject.MonitoringInputs)
	}

	if apiObject.MonitoringOutputConfig != nil {
		tfMap["monitoring_output_config"] = flattenMonitoringOutputConfig(apiObject.MonitoringOutputConfig)
	}

	if apiObject.MonitoringResources != nil {
		tfMap["monitoring_resources"] = flattenMonitoringResources(apiObject.MonitoringResources)
	}

	if apiObject.NetworkConfig != nil {
		tfMap["network_config"] = flattenNetworkConfig(apiObject.NetworkConfig)
	}

	if apiObject.RoleArn != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(apiObject.RoleArn)
	}

	if apiObject.StoppingCondition != nil {
		tfMap["stopping_condition"] = flattenMonitoringStoppingCondition(apiObject.StoppingCondition)
	}

	return []any{tfMap}
}

func flattenMonitoringBaselineConfig(apiObject *awstypes.MonitoringBaselineConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.BaseliningJobName != nil {
		tfMap["baselining_job_name"] = aws.ToString(apiObject.BaseliningJobName)
	}

	if apiObject.ConstraintsResource != nil {
		tfMap["constraints_resource"] = flattenMonitoringConstraintsResource(apiObject.ConstraintsResource)
	}

	if apiObject.StatisticsResource != nil {
		tfMap["statistics_resource"] = flattenMonitoringStatisticsResource(apiObject.StatisticsResource)
	}

	return []any{tfMap}
}

func flattenMonitoringAppSpecification(apiObject *awstypes.MonitoringAppSpecification) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"container_arguments":  apiObject.ContainerArguments,
		"container_entrypoint": apiObject.ContainerEntrypoint,
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

func flattenMonitoringInputs(apiObjects []awstypes.MonitoringInput) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		if apiObject.BatchTransformInput != nil {
			tfMap["batch_transform_input"] = flattenBatchTransformInput(apiObject.BatchTransformInput)
		}

		if apiObject.EndpointInput != nil {
			tfMap["endpoint_input"] = flattenEndpointInput(apiObject.EndpointInput)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenNetworkConfig(apiObject *awstypes.NetworkConfig) []any {
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

func flattenScheduleConfig(apiObject *awstypes.ScheduleConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.ScheduleExpression != nil {
		tfMap[names.AttrScheduleExpression] = aws.ToString(apiObject.ScheduleExpression)
	}

	return []any{tfMap}
}
