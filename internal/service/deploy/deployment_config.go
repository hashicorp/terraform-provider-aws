// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codedeploy_deployment_config", name="Deployment Config")
func resourceDeploymentConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeploymentConfigCreate,
		ReadWithoutTimeout:   resourceDeploymentConfigRead,
		DeleteWithoutTimeout: resourceDeploymentConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_platform": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.ComputePlatformServer,
				ValidateDiagFunc: enum.Validate[types.ComputePlatform](),
			},
			"deployment_config_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_config_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"minimum_healthy_hosts": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.MinimumHealthyHostsType](),
						},
						names.AttrValue: {
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
						"time_based_canary": {
							Type:          schema.TypeList,
							Optional:      true,
							ForceNew:      true,
							MaxItems:      1,
							ConflictsWith: []string{"traffic_routing_config.0.time_based_linear"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrInterval: {
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
							MaxItems:      1,
							ConflictsWith: []string{"traffic_routing_config.0.time_based_canary"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrInterval: {
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
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          types.TrafficRoutingTypeAllAtOnce,
							ValidateDiagFunc: enum.Validate[types.TrafficRoutingType](),
						},
					},
				},
			},
		},
	}
}

func resourceDeploymentConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	name := d.Get("deployment_config_name").(string)
	input := &codedeploy.CreateDeploymentConfigInput{
		ComputePlatform:      types.ComputePlatform(d.Get("compute_platform").(string)),
		DeploymentConfigName: aws.String(name),
		MinimumHealthyHosts:  expandMinimumHealthyHosts(d),
		TrafficRoutingConfig: expandTrafficRoutingConfig(d),
	}

	_, err := conn.CreateDeploymentConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeDeploy Deployment Config (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceDeploymentConfigRead(ctx, d, meta)...)
}

func resourceDeploymentConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	deploymentConfig, err := findDeploymentConfigByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeDeploy Deployment Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeDeploy Deployment Config (%s): %s", d.Id(), err)
	}

	deploymentConfigName := aws.ToString(deploymentConfig.DeploymentConfigName)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "codedeploy",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "deploymentconfig:" + deploymentConfigName,
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("compute_platform", deploymentConfig.ComputePlatform)
	d.Set("deployment_config_id", deploymentConfig.DeploymentConfigId)
	d.Set("deployment_config_name", deploymentConfigName)
	if err := d.Set("minimum_healthy_hosts", flattenMinimumHealthHosts(deploymentConfig.MinimumHealthyHosts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting minimum_healthy_hosts: %s", err)
	}
	if err := d.Set("traffic_routing_config", flattenTrafficRoutingConfig(deploymentConfig.TrafficRoutingConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting traffic_routing_config: %s", err)
	}

	return diags
}

func resourceDeploymentConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	log.Printf("[INFO] Deleting CodeDeploy Deployment Config: %s", d.Id())
	_, err := conn.DeleteDeploymentConfig(ctx, &codedeploy.DeleteDeploymentConfigInput{
		DeploymentConfigName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeDeploy Deployment Config (%s): %s", d.Id(), err)
	}

	return diags
}

func findDeploymentConfigByName(ctx context.Context, conn *codedeploy.Client, name string) (*types.DeploymentConfigInfo, error) {
	input := &codedeploy.GetDeploymentConfigInput{
		DeploymentConfigName: aws.String(name),
	}

	output, err := conn.GetDeploymentConfig(ctx, input)

	if errs.IsA[*types.DeploymentConfigDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DeploymentConfigInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DeploymentConfigInfo, nil
}

func expandMinimumHealthyHosts(d *schema.ResourceData) *types.MinimumHealthyHosts {
	hosts, ok := d.GetOk("minimum_healthy_hosts")
	if !ok {
		return nil
	}
	host := hosts.([]interface{})[0].(map[string]interface{})

	minimumHealthyHost := types.MinimumHealthyHosts{
		Type:  types.MinimumHealthyHostsType(host[names.AttrType].(string)),
		Value: int32(host[names.AttrValue].(int)),
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

	if trafficType, ok := config[names.AttrType]; ok {
		trafficRoutingConfig.Type = types.TrafficRoutingType(trafficType.(string))
	}
	if canary, ok := config["time_based_canary"]; ok && len(canary.([]interface{})) > 0 {
		canaryConfig := canary.([]interface{})[0].(map[string]interface{})
		trafficRoutingConfig.TimeBasedCanary = expandTimeBasedCanary(canaryConfig)
	}
	if linear, ok := config["time_based_linear"]; ok && len(linear.([]interface{})) > 0 {
		linearConfig := linear.([]interface{})[0].(map[string]interface{})
		trafficRoutingConfig.TimeBasedLinear = expandTimeBasedLinear(linearConfig)
	}

	return &trafficRoutingConfig
}

func expandTimeBasedCanary(config map[string]interface{}) *types.TimeBasedCanary {
	canary := types.TimeBasedCanary{}
	if interval, ok := config[names.AttrInterval]; ok {
		canary.CanaryInterval = int32(interval.(int))
	}
	if percentage, ok := config["percentage"]; ok {
		canary.CanaryPercentage = int32(percentage.(int))
	}
	return &canary
}

func expandTimeBasedLinear(config map[string]interface{}) *types.TimeBasedLinear {
	linear := types.TimeBasedLinear{}
	if interval, ok := config[names.AttrInterval]; ok {
		linear.LinearInterval = int32(interval.(int))
	}
	if percentage, ok := config["percentage"]; ok {
		linear.LinearPercentage = int32(percentage.(int))
	}
	return &linear
}

func flattenMinimumHealthHosts(hosts *types.MinimumHealthyHosts) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if hosts == nil {
		return result
	}

	item := make(map[string]interface{})

	item[names.AttrType] = string(hosts.Type)
	item[names.AttrValue] = hosts.Value

	return append(result, item)
}

func flattenTrafficRoutingConfig(config *types.TrafficRoutingConfig) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if config == nil {
		return result
	}

	item := make(map[string]interface{})

	item[names.AttrType] = string(config.Type)
	item["time_based_canary"] = flattenTimeBasedCanary(config.TimeBasedCanary)
	item["time_based_linear"] = flattenTimeBasedLinear(config.TimeBasedLinear)

	return append(result, item)
}

func flattenTimeBasedCanary(canary *types.TimeBasedCanary) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if canary == nil {
		return result
	}

	item := make(map[string]interface{})
	item[names.AttrInterval] = canary.CanaryInterval
	item["percentage"] = canary.CanaryPercentage

	return append(result, item)
}

func flattenTimeBasedLinear(linear *types.TimeBasedLinear) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if linear == nil {
		return result
	}

	item := make(map[string]interface{})
	item[names.AttrInterval] = linear.LinearInterval
	item["percentage"] = linear.LinearPercentage

	return append(result, item)
}
