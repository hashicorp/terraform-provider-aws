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
			"zonal_config": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"first_zone_monitor_duration_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"minimum_healthy_hosts_per_zone": {
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
										ValidateDiagFunc: enum.Validate[types.MinimumHealthyHostsPerZoneType](),
									},
									names.AttrValue: {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"monitor_duration_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceDeploymentConfigCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	name := d.Get("deployment_config_name").(string)
	input := &codedeploy.CreateDeploymentConfigInput{
		ComputePlatform:      types.ComputePlatform(d.Get("compute_platform").(string)),
		DeploymentConfigName: aws.String(name),
		MinimumHealthyHosts:  expandMinimumHealthyHosts(d),
		TrafficRoutingConfig: expandTrafficRoutingConfig(d),
		ZonalConfig:          expandZonalConfig(d),
	}

	_, err := conn.CreateDeploymentConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeDeploy Deployment Config (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceDeploymentConfigRead(ctx, d, meta)...)
}

func resourceDeploymentConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "codedeploy",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
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
	if err := d.Set("zonal_config", flattenZonalConfig(deploymentConfig.ZonalConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting zonal_config: %s", err)
	}

	return diags
}

func resourceDeploymentConfigDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	log.Printf("[INFO] Deleting CodeDeploy Deployment Config: %s", d.Id())
	input := codedeploy.DeleteDeploymentConfigInput{
		DeploymentConfigName: aws.String(d.Id()),
	}
	_, err := conn.DeleteDeploymentConfig(ctx, &input)

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
	v, ok := d.GetOk("minimum_healthy_hosts")
	if !ok {
		return nil
	}

	tfMap := v.([]any)[0].(map[string]any)

	apiObject := &types.MinimumHealthyHosts{
		Type:  types.MinimumHealthyHostsType(tfMap[names.AttrType].(string)),
		Value: int32(tfMap[names.AttrValue].(int)),
	}

	return apiObject
}

func expandTrafficRoutingConfig(d *schema.ResourceData) *types.TrafficRoutingConfig {
	v, ok := d.GetOk("traffic_routing_config")
	if !ok {
		return nil
	}

	tfMap := v.([]any)[0].(map[string]any)
	apiObject := &types.TrafficRoutingConfig{}

	if v, ok := tfMap["time_based_canary"]; ok && len(v.([]any)) > 0 {
		apiObject.TimeBasedCanary = expandTimeBasedCanary(v.([]any)[0].(map[string]any))
	}
	if v, ok := tfMap["time_based_linear"]; ok && len(v.([]any)) > 0 {
		apiObject.TimeBasedLinear = expandTimeBasedLinear(v.([]any)[0].(map[string]any))
	}
	if v, ok := tfMap[names.AttrType]; ok {
		apiObject.Type = types.TrafficRoutingType(v.(string))
	}

	return apiObject
}

func expandTimeBasedCanary(tfMap map[string]any) *types.TimeBasedCanary {
	apiObject := &types.TimeBasedCanary{}

	if v, ok := tfMap[names.AttrInterval]; ok {
		apiObject.CanaryInterval = int32(v.(int))
	}
	if v, ok := tfMap["percentage"]; ok {
		apiObject.CanaryPercentage = int32(v.(int))
	}

	return apiObject
}

func expandTimeBasedLinear(tfMap map[string]any) *types.TimeBasedLinear {
	apiObject := &types.TimeBasedLinear{}

	if v, ok := tfMap[names.AttrInterval]; ok {
		apiObject.LinearInterval = int32(v.(int))
	}
	if v, ok := tfMap["percentage"]; ok {
		apiObject.LinearPercentage = int32(v.(int))
	}

	return apiObject
}

func expandZonalConfig(d *schema.ResourceData) *types.ZonalConfig {
	v, ok := d.GetOk("zonal_config")
	if !ok {
		return nil
	}

	tfMap := v.([]any)[0].(map[string]any)
	apiObject := &types.ZonalConfig{}

	if v, ok := tfMap["first_zone_monitor_duration_in_seconds"].(int); ok {
		apiObject.FirstZoneMonitorDurationInSeconds = aws.Int64(int64(v))
	}
	if v, ok := tfMap["minimum_healthy_hosts_per_zone"]; ok && len(v.([]any)) > 0 {
		apiObject.MinimumHealthyHostsPerZone = expandMinimumHealthyHostsPerZone(v.([]any)[0].(map[string]any))
	}
	if v, ok := tfMap["monitor_duration_in_seconds"].(int); ok {
		apiObject.MonitorDurationInSeconds = aws.Int64(int64(v))
	}

	return apiObject
}

func expandMinimumHealthyHostsPerZone(tfMap map[string]any) *types.MinimumHealthyHostsPerZone {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.MinimumHealthyHostsPerZone{
		Type:  types.MinimumHealthyHostsPerZoneType(tfMap[names.AttrType].(string)),
		Value: int32(tfMap[names.AttrValue].(int)),
	}

	return apiObject
}

func flattenMinimumHealthHosts(apiObject *types.MinimumHealthyHosts) []any {
	tfList := make([]any, 0)

	if apiObject == nil {
		return tfList
	}

	tfMap := make(map[string]any)
	tfMap[names.AttrType] = apiObject.Type
	tfMap[names.AttrValue] = apiObject.Value

	return append(tfList, tfMap)
}

func flattenTrafficRoutingConfig(apiObject *types.TrafficRoutingConfig) []any {
	tfList := make([]any, 0)

	if apiObject == nil {
		return tfList
	}

	tfMap := make(map[string]any)
	tfMap["time_based_canary"] = flattenTimeBasedCanary(apiObject.TimeBasedCanary)
	tfMap["time_based_linear"] = flattenTimeBasedLinear(apiObject.TimeBasedLinear)
	tfMap[names.AttrType] = apiObject.Type

	return append(tfList, tfMap)
}

func flattenTimeBasedCanary(apiObject *types.TimeBasedCanary) []any {
	tfList := make([]any, 0)

	if apiObject == nil {
		return tfList
	}

	tfMap := make(map[string]any)
	tfMap[names.AttrInterval] = apiObject.CanaryInterval
	tfMap["percentage"] = apiObject.CanaryPercentage

	return append(tfList, tfMap)
}

func flattenTimeBasedLinear(apiObject *types.TimeBasedLinear) []any {
	tfList := make([]any, 0)

	if apiObject == nil {
		return tfList
	}

	tfMap := make(map[string]any)
	tfMap[names.AttrInterval] = apiObject.LinearInterval
	tfMap["percentage"] = apiObject.LinearPercentage

	return append(tfList, tfMap)
}

func flattenZonalConfig(apiObject *types.ZonalConfig) []any {
	tfList := make([]any, 0)

	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)
	tfMap["first_zone_monitor_duration_in_seconds"] = aws.ToInt64(apiObject.FirstZoneMonitorDurationInSeconds)
	tfMap["minimum_healthy_hosts_per_zone"] = flattenMinimumHealthHostsPerZone(apiObject.MinimumHealthyHostsPerZone)
	tfMap["monitor_duration_in_seconds"] = aws.ToInt64(apiObject.MonitorDurationInSeconds)

	return append(tfList, tfMap)
}

func flattenMinimumHealthHostsPerZone(apiObject *types.MinimumHealthyHostsPerZone) []any {
	tfList := make([]any, 0)

	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)
	tfMap[names.AttrType] = apiObject.Type
	tfMap[names.AttrValue] = apiObject.Value

	return append(tfList, tfMap)
}
