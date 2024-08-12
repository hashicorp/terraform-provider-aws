// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	var name string
	if v, ok := d.GetOk(names.AttrName); ok {
		name = v.(string)
	} else {
		name = id.UniqueId()
	}

	createOpts := &sagemaker.CreateEndpointInput{
		EndpointName:       aws.String(name),
		EndpointConfigName: aws.String(d.Get("endpoint_config_name").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("deployment_config"); ok && (len(v.([]interface{})) > 0) {
		createOpts.DeploymentConfig = expandEndpointDeploymentConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] SageMaker Endpoint create config: %#v", *createOpts)
	_, err := conn.CreateEndpoint(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Endpoint: %s", err)
	}

	d.SetId(name)

	if err := waitEndpointInService(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Endpoint (%s) to be in service: %s", name, err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	endpoint, err := findEndpointByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Endpoint (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, endpoint.EndpointName)
	d.Set("endpoint_config_name", endpoint.EndpointConfigName)
	d.Set(names.AttrARN, endpoint.EndpointArn)

	if err := d.Set("deployment_config", flattenEndpointDeploymentConfig(endpoint.LastDeploymentConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting deployment_config for SageMaker Endpoint (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChanges("endpoint_config_name", "deployment_config") {
		modifyOpts := &sagemaker.UpdateEndpointInput{
			EndpointName:       aws.String(d.Id()),
			EndpointConfigName: aws.String(d.Get("endpoint_config_name").(string)),
		}

		if v, ok := d.GetOk("deployment_config"); ok && (len(v.([]interface{})) > 0) {
			modifyOpts.DeploymentConfig = expandEndpointDeploymentConfig(v.([]interface{}))
		}

		log.Printf("[INFO] Modifying endpoint_config_name attribute for %s: %#v", d.Id(), modifyOpts)
		if _, err := conn.UpdateEndpoint(ctx, modifyOpts); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Endpoint (%s): %s", d.Id(), err)
		}

		err := waitEndpointInService(ctx, conn, d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Endpoint (%s) to be in service: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	deleteEndpointOpts := &sagemaker.DeleteEndpointInput{
		EndpointName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Endpoint: %s", d.Id())

	_, err := conn.DeleteEndpoint(ctx, deleteEndpointOpts)

	if tfawserr.ErrCodeEquals(err, ErrCodeValidationException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitEndpointDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Endpoint (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func findEndpointByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeEndpointOutput, error) {
	input := &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.DescribeEndpoint(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint") {
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

	if output.EndpointStatus == awstypes.EndpointStatusDeleting {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandEndpointDeploymentConfig(configured []interface{}) *awstypes.DeploymentConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.DeploymentConfig{
		BlueGreenUpdatePolicy: expandEndpointDeploymentConfigBlueGreenUpdatePolicy(m["blue_green_update_policy"].([]interface{})),
	}

	if v, ok := m["auto_rollback_configuration"].([]interface{}); ok && len(v) > 0 {
		c.AutoRollbackConfiguration = expandEndpointDeploymentConfigAutoRollbackConfig(v)
	}

	if v, ok := m["rolling_update_policy"].([]interface{}); ok && len(v) > 0 {
		c.RollingUpdatePolicy = expandEndpointDeploymentConfigRollingUpdatePolicy(v)
	}

	return c
}

func flattenEndpointDeploymentConfig(configured *awstypes.DeploymentConfig) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"blue_green_update_policy": flattenEndpointDeploymentConfigBlueGreenUpdatePolicy(configured.BlueGreenUpdatePolicy),
	}

	if configured.AutoRollbackConfiguration != nil {
		cfg["auto_rollback_configuration"] = flattenEndpointDeploymentConfigAutoRollbackConfig(configured.AutoRollbackConfiguration)
	}

	if configured.RollingUpdatePolicy != nil {
		cfg["rolling_update_policy"] = flattenEndpointDeploymentConfigRollingUpdatePolicy(configured.RollingUpdatePolicy)
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentConfigBlueGreenUpdatePolicy(configured []interface{}) *awstypes.BlueGreenUpdatePolicy {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.BlueGreenUpdatePolicy{
		TerminationWaitInSeconds:    aws.Int32(int32(m["termination_wait_in_seconds"].(int))),
		TrafficRoutingConfiguration: expandEndpointDeploymentConfigTrafficRoutingConfiguration(m["traffic_routing_configuration"].([]interface{})),
	}

	if v, ok := m["maximum_execution_timeout_in_seconds"].(int); ok && v > 0 {
		c.MaximumExecutionTimeoutInSeconds = aws.Int32(int32(v))
	}

	return c
}

func flattenEndpointDeploymentConfigBlueGreenUpdatePolicy(configured *awstypes.BlueGreenUpdatePolicy) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"termination_wait_in_seconds":   aws.ToInt32(configured.TerminationWaitInSeconds),
		"traffic_routing_configuration": flattenEndpointDeploymentConfigTrafficRoutingConfiguration(configured.TrafficRoutingConfiguration),
	}

	if configured.MaximumExecutionTimeoutInSeconds != nil {
		cfg["maximum_execution_timeout_in_seconds"] = aws.ToInt32(configured.MaximumExecutionTimeoutInSeconds)
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentConfigTrafficRoutingConfiguration(configured []interface{}) *awstypes.TrafficRoutingConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.TrafficRoutingConfig{
		Type:                  awstypes.TrafficRoutingConfigType(m[names.AttrType].(string)),
		WaitIntervalInSeconds: aws.Int32(int32(m["wait_interval_in_seconds"].(int))),
	}

	if v, ok := m["canary_size"].([]interface{}); ok && len(v) > 0 {
		c.CanarySize = expandEndpointDeploymentCapacitySize(v)
	}

	if v, ok := m["linear_step_size"].([]interface{}); ok && len(v) > 0 {
		c.LinearStepSize = expandEndpointDeploymentCapacitySize(v)
	}

	return c
}

func flattenEndpointDeploymentConfigTrafficRoutingConfiguration(configured *awstypes.TrafficRoutingConfig) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		names.AttrType:             configured.Type,
		"wait_interval_in_seconds": aws.ToInt32(configured.WaitIntervalInSeconds),
	}

	if configured.CanarySize != nil {
		cfg["canary_size"] = flattenEndpointDeploymentCapacitySize(configured.CanarySize)
	}

	if configured.LinearStepSize != nil {
		cfg["linear_step_size"] = flattenEndpointDeploymentCapacitySize(configured.LinearStepSize)
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentCapacitySize(configured []interface{}) *awstypes.CapacitySize {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.CapacitySize{
		Type:  awstypes.CapacitySizeType(m[names.AttrType].(string)),
		Value: aws.Int32(int32(m[names.AttrValue].(int))),
	}

	return c
}

func flattenEndpointDeploymentCapacitySize(configured *awstypes.CapacitySize) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		names.AttrType:  configured.Type,
		names.AttrValue: aws.ToInt32(configured.Value),
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentConfigAutoRollbackConfig(configured []interface{}) *awstypes.AutoRollbackConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.AutoRollbackConfig{
		Alarms: expandEndpointDeploymentConfigAutoRollbackConfigAlarms(m["alarms"].(*schema.Set).List()),
	}

	return c
}

func flattenEndpointDeploymentConfigAutoRollbackConfig(configured *awstypes.AutoRollbackConfig) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"alarms": flattenEndpointDeploymentConfigAutoRollbackConfigAlarms(configured.Alarms),
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentConfigRollingUpdatePolicy(configured []interface{}) *awstypes.RollingUpdatePolicy {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &awstypes.RollingUpdatePolicy{
		WaitIntervalInSeconds: aws.Int32(int32(m["wait_interval_in_seconds"].(int))),
	}

	if v, ok := m["maximum_execution_timeout_in_seconds"].(int); ok && v > 0 {
		c.MaximumExecutionTimeoutInSeconds = aws.Int32(int32(v))
	}

	if v, ok := m["maximum_batch_size"].([]interface{}); ok && len(v) > 0 {
		c.MaximumBatchSize = expandEndpointDeploymentCapacitySize(v)
	}

	if v, ok := m["rollback_maximum_batch_size"].([]interface{}); ok && len(v) > 0 {
		c.RollbackMaximumBatchSize = expandEndpointDeploymentCapacitySize(v)
	}

	return c
}

func flattenEndpointDeploymentConfigRollingUpdatePolicy(configured *awstypes.RollingUpdatePolicy) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"maximum_execution_timeout_in_seconds": aws.ToInt32(configured.MaximumExecutionTimeoutInSeconds),
		"wait_interval_in_seconds":             aws.ToInt32(configured.WaitIntervalInSeconds),
		"maximum_batch_size":                   flattenEndpointDeploymentCapacitySize(configured.MaximumBatchSize),
		"rollback_maximum_batch_size":          flattenEndpointDeploymentCapacitySize(configured.RollbackMaximumBatchSize),
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentConfigAutoRollbackConfigAlarms(configured []interface{}) []awstypes.Alarm {
	if len(configured) == 0 {
		return nil
	}

	alarms := make([]awstypes.Alarm, 0, len(configured))

	for _, alarmRaw := range configured {
		m := alarmRaw.(map[string]interface{})

		alarm := awstypes.Alarm{
			AlarmName: aws.String(m["alarm_name"].(string)),
		}

		alarms = append(alarms, alarm)
	}

	return alarms
}

func flattenEndpointDeploymentConfigAutoRollbackConfigAlarms(configured []awstypes.Alarm) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(configured))

	for _, i := range configured {
		l := map[string]interface{}{
			"alarm_name": aws.ToString(i.AlarmName),
		}

		result = append(result, l)
	}
	return result
}
