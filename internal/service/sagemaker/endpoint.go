// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_endpoint", name="Endpoint")
// @Tags(identifierAttribute="arn")
func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointCreate,
		ReadWithoutTimeout:   resourceEndpointRead,
		UpdateWithoutTimeout: resourceEndpointUpdate,
		DeleteWithoutTimeout: resourceEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_rollback_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"alarms": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 1,
										MaxItems: 10,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"alarm_name": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"blue_green_update_policy": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"deployment_config.0.blue_green_update_policy",
								"deployment_config.0.rolling_update_policy",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"maximum_execution_timeout_in_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(600, 14400),
									},
									"termination_wait_in_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 3600),
									},
									"traffic_routing_configuration": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"canary_size": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrType: {
																Type:             schema.TypeString,
																Required:         true,
																ValidateDiagFunc: enum.Validate[awstypes.CapacitySizeType](),
															},
															names.AttrValue: {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												"linear_step_size": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrType: {
																Type:             schema.TypeString,
																Required:         true,
																ValidateDiagFunc: enum.Validate[awstypes.CapacitySizeType](),
															},
															names.AttrValue: {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												names.AttrType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.TrafficRoutingConfigType](),
												},
												"wait_interval_in_seconds": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(0, 3600),
												},
											},
										},
									},
								},
							},
						},
						"rolling_update_policy": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"deployment_config.0.blue_green_update_policy",
								"deployment_config.0.rolling_update_policy",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"maximum_batch_size": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.CapacitySizeType](),
												},
												names.AttrValue: {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
											},
										},
									},
									"maximum_execution_timeout_in_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(600, 14400),
									},
									"rollback_maximum_batch_size": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.CapacitySizeType](),
												},
												names.AttrValue: {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
											},
										},
									},
									"wait_interval_in_seconds": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(0, 3600),
									},
								},
							},
						},
					},
				},
			},
			"endpoint_config_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validName,
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

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	var name string
	if v, ok := d.GetOk(names.AttrName); ok {
		name = v.(string)
	} else {
		name = sdkid.UniqueId()
	}
	input := sagemaker.CreateEndpointInput{
		EndpointName:       aws.String(name),
		EndpointConfigName: aws.String(d.Get("endpoint_config_name").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("deployment_config"); ok && (len(v.([]any)) > 0) {
		input.DeploymentConfig = expandDeploymentConfig(v.([]any))
	}

	err := tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		_, err := conn.CreateEndpoint(ctx, &input)

		if err != nil {
			return tfresource.NonRetryableError(fmt.Errorf("creating SageMaker AI Endpoint (%s): %w", name, err))
		}

		_, err = waitEndpointInService(ctx, conn, name)

		// unexpected state 'Failed', wanted target 'InService'. last error: The execution role ARN "..." is invalid. Please ensure that the role exists and that its trust relationship policy allows the action "sts:AssumeRole" for the service principal "sagemaker.amazonaws.com"
		if errs.Contains(err, `Please ensure that the role exists and that its trust relationship policy allows the action "sts:AssumeRole" for the service principal "sagemaker.amazonaws.com"`) {
			d := resourceEndpoint().Data(nil)
			d.SetId(name)
			if diags := resourceEndpointDelete(ctx, d, meta); diags.HasError() {
				return tfresource.NonRetryableError(sdkdiag.DiagnosticsError(diags))
			}

			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(fmt.Errorf("waiting for SageMaker AI Endpoint (%s) create: %w", name, err))
		}

		return nil
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(name)

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	endpoint, err := findEndpointByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Endpoint (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, endpoint.EndpointArn)
	if err := d.Set("deployment_config", flattenDeploymentConfig(endpoint.LastDeploymentConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting deployment_config: %s", err)
	}
	d.Set("endpoint_config_name", endpoint.EndpointConfigName)
	d.Set(names.AttrName, endpoint.EndpointName)

	return diags
}

func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChanges("endpoint_config_name", "deployment_config") {
		_, n := d.GetChange("endpoint_config_name")
		input := sagemaker.UpdateEndpointInput{
			EndpointName:       aws.String(d.Id()),
			EndpointConfigName: aws.String(n.(string)),
		}

		if v, ok := d.GetOk("deployment_config"); ok && (len(v.([]any)) > 0) {
			input.DeploymentConfig = expandDeploymentConfig(v.([]any))
		}

		_, err := conn.UpdateEndpoint(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Endpoint (%s): %s", d.Id(), err)
		}

		if _, err := waitEndpointInService(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Endpoint (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[INFO] Deleting SageMaker AI Endpoint: %s", d.Id())
	input := sagemaker.DeleteEndpointInput{
		EndpointName: aws.String(d.Id()),
	}
	_, err := conn.DeleteEndpoint(ctx, &input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findEndpointByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeEndpointOutput, error) {
	input := sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(name),
	}

	output, err := findEndpoint(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := output.EndpointStatus; status == awstypes.EndpointStatusDeleting {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func findEndpoint(ctx context.Context, conn *sagemaker.Client, input *sagemaker.DescribeEndpointInput) (*sagemaker.DescribeEndpointOutput, error) {
	output, err := conn.DescribeEndpoint(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint") {
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

func statusEndpoint(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findEndpointByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.EndpointStatus), nil
	}
}

func waitEndpointInService(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeEndpointOutput, error) { //nolint:unparam
	const (
		timeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EndpointStatusCreating, awstypes.EndpointStatusUpdating, awstypes.EndpointStatusSystemUpdating),
		Target:  enum.Slice(awstypes.EndpointStatusInService),
		Refresh: statusEndpoint(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeEndpointOutput); ok {
		if failureReason := output.FailureReason; failureReason != nil {
			retry.SetLastError(err, errors.New(aws.ToString(failureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitEndpointDeleted(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeEndpointOutput, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EndpointStatusDeleting),
		Target:  []string{},
		Refresh: statusEndpoint(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeEndpointOutput); ok {
		if failureReason := output.FailureReason; failureReason != nil {
			retry.SetLastError(err, errors.New(aws.ToString(failureReason)))
		}

		return output, err
	}

	return nil, err
}

func expandDeploymentConfig(tfList []any) *awstypes.DeploymentConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.DeploymentConfig{
		BlueGreenUpdatePolicy: expandBlueGreenUpdatePolicy(tfMap["blue_green_update_policy"].([]any)),
	}

	if v, ok := tfMap["auto_rollback_configuration"].([]any); ok && len(v) > 0 {
		apiObject.AutoRollbackConfiguration = expandAutoRollbackConfig(v)
	}

	if v, ok := tfMap["rolling_update_policy"].([]any); ok && len(v) > 0 {
		apiObject.RollingUpdatePolicy = expandRollingUpdatePolicy(v)
	}

	return apiObject
}

func flattenDeploymentConfig(apiObject *awstypes.DeploymentConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"blue_green_update_policy": flattenBlueGreenUpdatePolicy(apiObject.BlueGreenUpdatePolicy),
	}

	if apiObject.AutoRollbackConfiguration != nil {
		tfMap["auto_rollback_configuration"] = flattenAutoRollbackConfig(apiObject.AutoRollbackConfiguration)
	}

	if apiObject.RollingUpdatePolicy != nil {
		tfMap["rolling_update_policy"] = flattenRollingUpdatePolicy(apiObject.RollingUpdatePolicy)
	}

	return []any{tfMap}
}

func expandBlueGreenUpdatePolicy(tfList []any) *awstypes.BlueGreenUpdatePolicy {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.BlueGreenUpdatePolicy{
		TerminationWaitInSeconds:    aws.Int32(int32(tfMap["termination_wait_in_seconds"].(int))),
		TrafficRoutingConfiguration: expandTrafficRoutingConfig(tfMap["traffic_routing_configuration"].([]any)),
	}

	if v, ok := tfMap["maximum_execution_timeout_in_seconds"].(int); ok && v > 0 {
		apiObject.MaximumExecutionTimeoutInSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenBlueGreenUpdatePolicy(apiObject *awstypes.BlueGreenUpdatePolicy) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"termination_wait_in_seconds":   aws.ToInt32(apiObject.TerminationWaitInSeconds),
		"traffic_routing_configuration": flattenTrafficRoutingConfig(apiObject.TrafficRoutingConfiguration),
	}

	if apiObject.MaximumExecutionTimeoutInSeconds != nil {
		tfMap["maximum_execution_timeout_in_seconds"] = aws.ToInt32(apiObject.MaximumExecutionTimeoutInSeconds)
	}

	return []any{tfMap}
}

func expandTrafficRoutingConfig(tfList []any) *awstypes.TrafficRoutingConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.TrafficRoutingConfig{
		Type:                  awstypes.TrafficRoutingConfigType(tfMap[names.AttrType].(string)),
		WaitIntervalInSeconds: aws.Int32(int32(tfMap["wait_interval_in_seconds"].(int))),
	}

	if v, ok := tfMap["canary_size"].([]any); ok && len(v) > 0 {
		apiObject.CanarySize = expandCapacitySize(v)
	}

	if v, ok := tfMap["linear_step_size"].([]any); ok && len(v) > 0 {
		apiObject.LinearStepSize = expandCapacitySize(v)
	}

	return apiObject
}

func flattenTrafficRoutingConfig(apiObject *awstypes.TrafficRoutingConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrType:             apiObject.Type,
		"wait_interval_in_seconds": aws.ToInt32(apiObject.WaitIntervalInSeconds),
	}

	if apiObject.CanarySize != nil {
		tfMap["canary_size"] = flattenCapacitySize(apiObject.CanarySize)
	}

	if apiObject.LinearStepSize != nil {
		tfMap["linear_step_size"] = flattenCapacitySize(apiObject.LinearStepSize)
	}

	return []any{tfMap}
}

func expandCapacitySize(tfList []any) *awstypes.CapacitySize {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.CapacitySize{
		Type:  awstypes.CapacitySizeType(tfMap[names.AttrType].(string)),
		Value: aws.Int32(int32(tfMap[names.AttrValue].(int))),
	}

	return apiObject
}

func flattenCapacitySize(apiObject *awstypes.CapacitySize) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrType:  apiObject.Type,
		names.AttrValue: aws.ToInt32(apiObject.Value),
	}

	return []any{tfMap}
}

func expandAutoRollbackConfig(tfList []any) *awstypes.AutoRollbackConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.AutoRollbackConfig{
		Alarms: expandAlarms(tfMap["alarms"].(*schema.Set).List()),
	}

	return apiObject
}

func flattenAutoRollbackConfig(apiObject *awstypes.AutoRollbackConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"alarms": flattenAlarms(apiObject.Alarms),
	}

	return []any{tfMap}
}

func expandRollingUpdatePolicy(tfList []any) *awstypes.RollingUpdatePolicy {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.RollingUpdatePolicy{
		WaitIntervalInSeconds: aws.Int32(int32(tfMap["wait_interval_in_seconds"].(int))),
	}

	if v, ok := tfMap["maximum_batch_size"].([]any); ok && len(v) > 0 {
		apiObject.MaximumBatchSize = expandCapacitySize(v)
	}

	if v, ok := tfMap["maximum_execution_timeout_in_seconds"].(int); ok && v > 0 {
		apiObject.MaximumExecutionTimeoutInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["rollback_maximum_batch_size"].([]any); ok && len(v) > 0 {
		apiObject.RollbackMaximumBatchSize = expandCapacitySize(v)
	}

	return apiObject
}

func flattenRollingUpdatePolicy(apiObject *awstypes.RollingUpdatePolicy) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"maximum_batch_size":                   flattenCapacitySize(apiObject.MaximumBatchSize),
		"maximum_execution_timeout_in_seconds": aws.ToInt32(apiObject.MaximumExecutionTimeoutInSeconds),
		"rollback_maximum_batch_size":          flattenCapacitySize(apiObject.RollbackMaximumBatchSize),
		"wait_interval_in_seconds":             aws.ToInt32(apiObject.WaitIntervalInSeconds),
	}

	return []any{tfMap}
}

func expandAlarms(tfList []any) []awstypes.Alarm {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]awstypes.Alarm, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := awstypes.Alarm{
			AlarmName: aws.String(tfMap["alarm_name"].(string)),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenAlarms(apiObjects []awstypes.Alarm) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"alarm_name": aws.ToString(apiObject.AlarmName),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
