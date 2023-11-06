// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_codedeploy_deployment_config")
func ResourceDeploymentConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeploymentConfigCreate,
		ReadWithoutTimeout:   resourceDeploymentConfigRead,
		DeleteWithoutTimeout: resourceDeploymentConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"deployment_config_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"compute_platform": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(types.ComputePlatformServer),
					string(types.ComputePlatformLambda),
					string(types.ComputePlatformEcs),
				}, false),
				Default: types.ComputePlatformServer,
			},

			"minimum_healthy_hosts": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(types.MinimumHealthyHostsTypeHostCount),
								string(types.MinimumHealthyHostsTypeFleetPercent),
							}, false),
						},
						"value": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},

			"traffic_routing_config": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(types.TrafficRoutingTypeAllAtOnce),
								string(types.TrafficRoutingTypeTimeBasedCanary),
								string(types.TrafficRoutingTypeTimeBasedLinear),
							}, false),
							Default: string(types.TrafficRoutingTypeAllAtOnce),
						},

						"time_based_canary": {
							Type:          schema.TypeList,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"traffic_routing_config.0.time_based_linear"},
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"interval": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
									"percentage": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},

						"time_based_linear": {
							Type:          schema.TypeList,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"traffic_routing_config.0.time_based_canary"},
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"interval": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
									"percentage": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},

			"deployment_config_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDeploymentConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	input := &codedeploy.CreateDeploymentConfigInput{
		DeploymentConfigName: aws.String(d.Get("deployment_config_name").(string)),
		ComputePlatform:      types.ComputePlatform(d.Get("compute_platform").(string)),
		MinimumHealthyHosts:  expandMinimumHealthHostsConfig(d),
		TrafficRoutingConfig: expandTrafficRoutingConfig(d),
	}

	_, err := conn.CreateDeploymentConfig(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeDeploy Deployment Config (%s): %s", d.Get("deployment_config_name").(string), err)
	}

	d.SetId(d.Get("deployment_config_name").(string))

	return append(diags, resourceDeploymentConfigRead(ctx, d, meta)...)
}

func resourceDeploymentConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	input := &codedeploy.GetDeploymentConfigInput{
		DeploymentConfigName: aws.String(d.Id()),
	}

	resp, err := conn.GetDeploymentConfig(ctx, input)

	if !d.IsNewResource() && errs.IsA[*types.DeploymentConfigDoesNotExistException](err) {
		log.Printf("[WARN] CodeDeploy Deployment Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeDeploy Deployment Config (%s): %s", d.Id(), err)
	}

	if resp.DeploymentConfigInfo == nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeDeploy Deployment Config (%s): empty result", d.Id())
	}

	if err := d.Set("minimum_healthy_hosts", flattenMinimumHealthHostsConfig(resp.DeploymentConfigInfo.MinimumHealthyHosts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeDeploy Deployment Config (%s): %s", d.Id(), err)
	}

	if err := d.Set("traffic_routing_config", flattenTrafficRoutingConfig(resp.DeploymentConfigInfo.TrafficRoutingConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeDeploy Deployment Config (%s): %s", d.Id(), err)
	}

	d.Set("deployment_config_id", resp.DeploymentConfigInfo.DeploymentConfigId)
	d.Set("deployment_config_name", resp.DeploymentConfigInfo.DeploymentConfigName)
	d.Set("compute_platform", resp.DeploymentConfigInfo.ComputePlatform)

	return diags
}

func resourceDeploymentConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	input := &codedeploy.DeleteDeploymentConfigInput{
		DeploymentConfigName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteDeploymentConfig(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeDeploy Deployment Config (%s): %s", d.Id(), err)
	}
	return diags
}

func expandMinimumHealthHostsConfig(d *schema.ResourceData) *types.MinimumHealthyHosts {
	hosts, ok := d.GetOk("minimum_healthy_hosts")
	if !ok {
		return nil
	}
	host := hosts.([]interface{})[0].(map[string]interface{})

	minimumHealthyHost := types.MinimumHealthyHosts{
		Type:  types.MinimumHealthyHostsType(host["type"].(string)),
		Value: int32(host["value"].(int)),
	}

	return &minimumHealthyHost
}

func expandTrafficRoutingConfig(d *schema.ResourceData) *types.TrafficRoutingConfig {
	block, ok := d.GetOk("traffic_routing_config")
	if !ok {
		return nil
	}
	config := block.([]interface{})[0].(map[string]interface{})
	trafficRoutingConfig := types.TrafficRoutingConfig{}

	if trafficType, ok := config["type"]; ok {
		trafficRoutingConfig.Type = types.TrafficRoutingType(trafficType.(string))
	}
	if canary, ok := config["time_based_canary"]; ok && len(canary.([]interface{})) > 0 {
		canaryConfig := canary.([]interface{})[0].(map[string]interface{})
		trafficRoutingConfig.TimeBasedCanary = expandTrafficTimeBasedCanaryConfig(canaryConfig)
	}
	if linear, ok := config["time_based_linear"]; ok && len(linear.([]interface{})) > 0 {
		linearConfig := linear.([]interface{})[0].(map[string]interface{})
		trafficRoutingConfig.TimeBasedLinear = expandTrafficTimeBasedLinearConfig(linearConfig)
	}

	return &trafficRoutingConfig
}

func expandTrafficTimeBasedCanaryConfig(config map[string]interface{}) *types.TimeBasedCanary {
	canary := types.TimeBasedCanary{}
	if interval, ok := config["interval"]; ok {
		canary.CanaryInterval = int32(interval.(int))
	}
	if percentage, ok := config["percentage"]; ok {
		canary.CanaryPercentage = int32(percentage.(int))
	}
	return &canary
}

func expandTrafficTimeBasedLinearConfig(config map[string]interface{}) *types.TimeBasedLinear {
	linear := types.TimeBasedLinear{}
	if interval, ok := config["interval"]; ok {
		linear.LinearInterval = int32(interval.(int))
	}
	if percentage, ok := config["percentage"]; ok {
		linear.LinearPercentage = int32(percentage.(int))
	}
	return &linear
}

func flattenMinimumHealthHostsConfig(hosts *types.MinimumHealthyHosts) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if hosts == nil {
		return result
	}

	item := make(map[string]interface{})

	item["type"] = string(hosts.Type)
	item["value"] = int32(hosts.Value)

	return append(result, item)
}

func flattenTrafficRoutingConfig(config *types.TrafficRoutingConfig) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if config == nil {
		return result
	}

	item := make(map[string]interface{})

	item["type"] = string(config.Type)
	item["time_based_canary"] = flattenTrafficRoutingCanaryConfig(config.TimeBasedCanary)
	item["time_based_linear"] = flattenTrafficRoutingLinearConfig(config.TimeBasedLinear)

	return append(result, item)
}

func flattenTrafficRoutingCanaryConfig(canary *types.TimeBasedCanary) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if canary == nil {
		return result
	}

	item := make(map[string]interface{})
	item["interval"] = int32(canary.CanaryInterval)
	item["percentage"] = int32(canary.CanaryPercentage)

	return append(result, item)
}

func flattenTrafficRoutingLinearConfig(linear *types.TimeBasedLinear) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if linear == nil {
		return result
	}

	item := make(map[string]interface{})
	item["interval"] = int32(linear.LinearInterval)
	item["percentage"] = int32(linear.LinearPercentage)

	return append(result, item)
}
