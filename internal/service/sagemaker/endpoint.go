// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_endpoint", name="Endpoint")
// @Tags(identifierAttribute="arn")
func ResourceEndpoint() *schema.Resource {
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
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(sagemaker.CapacitySizeType_Values(), false),
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
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(sagemaker.CapacitySizeType_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.TrafficRoutingConfigType_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.CapacitySizeType_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.CapacitySizeType_Values(), false),
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
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

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
	_, err := conn.CreateEndpointWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Endpoint: %s", err)
	}

	d.SetId(name)

	describeInput := &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(name),
	}

	if err := conn.WaitUntilEndpointInServiceWithContext(ctx, describeInput); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Endpoint (%s) to be in service: %s", name, err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	endpoint, err := FindEndpointByName(ctx, conn, d.Id())

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
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	if d.HasChanges("endpoint_config_name", "deployment_config") {
		modifyOpts := &sagemaker.UpdateEndpointInput{
			EndpointName:       aws.String(d.Id()),
			EndpointConfigName: aws.String(d.Get("endpoint_config_name").(string)),
		}

		if v, ok := d.GetOk("deployment_config"); ok && (len(v.([]interface{})) > 0) {
			modifyOpts.DeploymentConfig = expandEndpointDeploymentConfig(v.([]interface{}))
		}

		log.Printf("[INFO] Modifying endpoint_config_name attribute for %s: %#v", d.Id(), modifyOpts)
		if _, err := conn.UpdateEndpointWithContext(ctx, modifyOpts); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Endpoint (%s): %s", d.Id(), err)
		}

		describeInput := &sagemaker.DescribeEndpointInput{
			EndpointName: aws.String(d.Id()),
		}

		err := conn.WaitUntilEndpointInServiceWithContext(ctx, describeInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Endpoint (%s) to be in service: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	deleteEndpointOpts := &sagemaker.DeleteEndpointInput{
		EndpointName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Endpoint: %s", d.Id())

	_, err := conn.DeleteEndpointWithContext(ctx, deleteEndpointOpts)

	if tfawserr.ErrCodeEquals(err, "ValidationException") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Endpoint (%s): %s", d.Id(), err)
	}

	describeInput := &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(d.Id()),
	}

	if err := conn.WaitUntilEndpointDeletedWithContext(ctx, describeInput); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Endpoint (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func expandEndpointDeploymentConfig(configured []interface{}) *sagemaker.DeploymentConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.DeploymentConfig{
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

func flattenEndpointDeploymentConfig(configured *sagemaker.DeploymentConfig) []map[string]interface{} {
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

func expandEndpointDeploymentConfigBlueGreenUpdatePolicy(configured []interface{}) *sagemaker.BlueGreenUpdatePolicy {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.BlueGreenUpdatePolicy{
		TerminationWaitInSeconds:    aws.Int64(int64(m["termination_wait_in_seconds"].(int))),
		TrafficRoutingConfiguration: expandEndpointDeploymentConfigTrafficRoutingConfiguration(m["traffic_routing_configuration"].([]interface{})),
	}

	if v, ok := m["maximum_execution_timeout_in_seconds"].(int); ok && v > 0 {
		c.MaximumExecutionTimeoutInSeconds = aws.Int64(int64(v))
	}

	return c
}

func flattenEndpointDeploymentConfigBlueGreenUpdatePolicy(configured *sagemaker.BlueGreenUpdatePolicy) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"termination_wait_in_seconds":   aws.Int64Value(configured.TerminationWaitInSeconds),
		"traffic_routing_configuration": flattenEndpointDeploymentConfigTrafficRoutingConfiguration(configured.TrafficRoutingConfiguration),
	}

	if configured.MaximumExecutionTimeoutInSeconds != nil {
		cfg["maximum_execution_timeout_in_seconds"] = aws.Int64Value(configured.MaximumExecutionTimeoutInSeconds)
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentConfigTrafficRoutingConfiguration(configured []interface{}) *sagemaker.TrafficRoutingConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.TrafficRoutingConfig{
		Type:                  aws.String(m[names.AttrType].(string)),
		WaitIntervalInSeconds: aws.Int64(int64(m["wait_interval_in_seconds"].(int))),
	}

	if v, ok := m["canary_size"].([]interface{}); ok && len(v) > 0 {
		c.CanarySize = expandEndpointDeploymentCapacitySize(v)
	}

	if v, ok := m["linear_step_size"].([]interface{}); ok && len(v) > 0 {
		c.LinearStepSize = expandEndpointDeploymentCapacitySize(v)
	}

	return c
}

func flattenEndpointDeploymentConfigTrafficRoutingConfiguration(configured *sagemaker.TrafficRoutingConfig) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		names.AttrType:             aws.StringValue(configured.Type),
		"wait_interval_in_seconds": aws.Int64Value(configured.WaitIntervalInSeconds),
	}

	if configured.CanarySize != nil {
		cfg["canary_size"] = flattenEndpointDeploymentCapacitySize(configured.CanarySize)
	}

	if configured.LinearStepSize != nil {
		cfg["linear_step_size"] = flattenEndpointDeploymentCapacitySize(configured.LinearStepSize)
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentCapacitySize(configured []interface{}) *sagemaker.CapacitySize {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.CapacitySize{
		Type:  aws.String(m[names.AttrType].(string)),
		Value: aws.Int64(int64(m[names.AttrValue].(int))),
	}

	return c
}

func flattenEndpointDeploymentCapacitySize(configured *sagemaker.CapacitySize) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		names.AttrType:  aws.StringValue(configured.Type),
		names.AttrValue: aws.Int64Value(configured.Value),
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentConfigAutoRollbackConfig(configured []interface{}) *sagemaker.AutoRollbackConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.AutoRollbackConfig{
		Alarms: expandEndpointDeploymentConfigAutoRollbackConfigAlarms(m["alarms"].(*schema.Set).List()),
	}

	return c
}

func flattenEndpointDeploymentConfigAutoRollbackConfig(configured *sagemaker.AutoRollbackConfig) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"alarms": flattenEndpointDeploymentConfigAutoRollbackConfigAlarms(configured.Alarms),
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentConfigRollingUpdatePolicy(configured []interface{}) *sagemaker.RollingUpdatePolicy {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.RollingUpdatePolicy{
		WaitIntervalInSeconds: aws.Int64(int64(m["wait_interval_in_seconds"].(int))),
	}

	if v, ok := m["maximum_execution_timeout_in_seconds"].(int); ok && v > 0 {
		c.MaximumExecutionTimeoutInSeconds = aws.Int64(int64(v))
	}

	if v, ok := m["maximum_batch_size"].([]interface{}); ok && len(v) > 0 {
		c.MaximumBatchSize = expandEndpointDeploymentCapacitySize(v)
	}

	if v, ok := m["rollback_maximum_batch_size"].([]interface{}); ok && len(v) > 0 {
		c.RollbackMaximumBatchSize = expandEndpointDeploymentCapacitySize(v)
	}

	return c
}

func flattenEndpointDeploymentConfigRollingUpdatePolicy(configured *sagemaker.RollingUpdatePolicy) []map[string]interface{} {
	if configured == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"maximum_execution_timeout_in_seconds": aws.Int64Value(configured.MaximumExecutionTimeoutInSeconds),
		"wait_interval_in_seconds":             aws.Int64Value(configured.WaitIntervalInSeconds),
		"maximum_batch_size":                   flattenEndpointDeploymentCapacitySize(configured.MaximumBatchSize),
		"rollback_maximum_batch_size":          flattenEndpointDeploymentCapacitySize(configured.RollbackMaximumBatchSize),
	}

	return []map[string]interface{}{cfg}
}

func expandEndpointDeploymentConfigAutoRollbackConfigAlarms(configured []interface{}) []*sagemaker.Alarm {
	if len(configured) == 0 {
		return nil
	}

	alarms := make([]*sagemaker.Alarm, 0, len(configured))

	for _, alarmRaw := range configured {
		m := alarmRaw.(map[string]interface{})

		alarm := &sagemaker.Alarm{
			AlarmName: aws.String(m["alarm_name"].(string)),
		}

		alarms = append(alarms, alarm)
	}

	return alarms
}

func flattenEndpointDeploymentConfigAutoRollbackConfigAlarms(configured []*sagemaker.Alarm) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(configured))

	for _, i := range configured {
		l := map[string]interface{}{
			"alarm_name": aws.StringValue(i.AlarmName),
		}

		result = append(result, l)
	}
	return result
}
