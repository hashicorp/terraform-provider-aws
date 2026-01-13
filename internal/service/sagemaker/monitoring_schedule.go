// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
								},
							},
						},
						"monitoring_job_definition_name": {
							Type:         schema.TypeString,
							Required:     true,
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
		name = id.UniqueId()
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
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
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

func statusMonitoringSchedule(ctx context.Context, conn *sagemaker.Client, name string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
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
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScheduleStatusPending),
		Target:  enum.Slice(awstypes.ScheduleStatusScheduled),
		Refresh: statusMonitoringSchedule(ctx, conn, name),
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
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScheduleStatusScheduled, awstypes.ScheduleStatusPending, awstypes.ScheduleStatusStopped),
		Target:  []string{},
		Refresh: statusMonitoringSchedule(ctx, conn, name),
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

	if v, ok := tfMap["monitoring_app_specification"].([]any); ok && len(v) > 0 {
		apiObject.MonitoringAppSpecification = expandMonitoringAppSpecification(v)
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

	if apiObject.MonitoringAppSpecification != nil {
		tfMap["monitoring_app_specification"] = flattenMonitoringAppSpecification(apiObject.MonitoringAppSpecification)
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
