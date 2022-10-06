package deploy

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDeploymentConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeploymentConfigCreate,
		Read:   resourceDeploymentConfigRead,
		Delete: resourceDeploymentConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
					codedeploy.ComputePlatformServer,
					codedeploy.ComputePlatformLambda,
					codedeploy.ComputePlatformEcs,
				}, false),
				Default: codedeploy.ComputePlatformServer,
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
								codedeploy.MinimumHealthyHostsTypeHostCount,
								codedeploy.MinimumHealthyHostsTypeFleetPercent,
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
								codedeploy.TrafficRoutingTypeAllAtOnce,
								codedeploy.TrafficRoutingTypeTimeBasedCanary,
								codedeploy.TrafficRoutingTypeTimeBasedLinear,
							}, false),
							Default: codedeploy.TrafficRoutingTypeAllAtOnce,
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

func resourceDeploymentConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeployConn

	input := &codedeploy.CreateDeploymentConfigInput{
		DeploymentConfigName: aws.String(d.Get("deployment_config_name").(string)),
		ComputePlatform:      aws.String(d.Get("compute_platform").(string)),
		MinimumHealthyHosts:  expandMinimumHealthHostsConfig(d),
		TrafficRoutingConfig: expandTrafficRoutingConfig(d),
	}

	_, err := conn.CreateDeploymentConfig(input)
	if err != nil {
		return err
	}

	d.SetId(d.Get("deployment_config_name").(string))

	return resourceDeploymentConfigRead(d, meta)
}

func resourceDeploymentConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeployConn

	input := &codedeploy.GetDeploymentConfigInput{
		DeploymentConfigName: aws.String(d.Id()),
	}

	resp, err := conn.GetDeploymentConfig(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codedeploy.ErrCodeDeploymentConfigDoesNotExistException) {
		log.Printf("[WARN] CodeDeploy Deployment Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("finding CodeDeploy Deployment Config (%s): %w", d.Id(), err)
	}

	if resp.DeploymentConfigInfo == nil {
		return fmt.Errorf("Cannot find DeploymentConfig %q", d.Id())
	}

	if err := d.Set("minimum_healthy_hosts", flattenMinimumHealthHostsConfig(resp.DeploymentConfigInfo.MinimumHealthyHosts)); err != nil {
		return err
	}

	if err := d.Set("traffic_routing_config", flattenTrafficRoutingConfig(resp.DeploymentConfigInfo.TrafficRoutingConfig)); err != nil {
		return err
	}

	d.Set("deployment_config_id", resp.DeploymentConfigInfo.DeploymentConfigId)
	d.Set("deployment_config_name", resp.DeploymentConfigInfo.DeploymentConfigName)
	d.Set("compute_platform", resp.DeploymentConfigInfo.ComputePlatform)

	return nil
}

func resourceDeploymentConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeployConn

	input := &codedeploy.DeleteDeploymentConfigInput{
		DeploymentConfigName: aws.String(d.Id()),
	}

	_, err := conn.DeleteDeploymentConfig(input)
	return err
}

func expandMinimumHealthHostsConfig(d *schema.ResourceData) *codedeploy.MinimumHealthyHosts {
	hosts, ok := d.GetOk("minimum_healthy_hosts")
	if !ok {
		return nil
	}
	host := hosts.([]interface{})[0].(map[string]interface{})

	minimumHealthyHost := codedeploy.MinimumHealthyHosts{
		Type:  aws.String(host["type"].(string)),
		Value: aws.Int64(int64(host["value"].(int))),
	}

	return &minimumHealthyHost
}

func expandTrafficRoutingConfig(d *schema.ResourceData) *codedeploy.TrafficRoutingConfig {
	block, ok := d.GetOk("traffic_routing_config")
	if !ok {
		return nil
	}
	config := block.([]interface{})[0].(map[string]interface{})
	trafficRoutingConfig := codedeploy.TrafficRoutingConfig{}

	if trafficType, ok := config["type"]; ok {
		trafficRoutingConfig.Type = aws.String(trafficType.(string))
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

func expandTrafficTimeBasedCanaryConfig(config map[string]interface{}) *codedeploy.TimeBasedCanary {
	canary := codedeploy.TimeBasedCanary{}
	if interval, ok := config["interval"]; ok {
		canary.CanaryInterval = aws.Int64(int64(interval.(int)))
	}
	if percentage, ok := config["percentage"]; ok {
		canary.CanaryPercentage = aws.Int64(int64(percentage.(int)))
	}
	return &canary
}

func expandTrafficTimeBasedLinearConfig(config map[string]interface{}) *codedeploy.TimeBasedLinear {
	linear := codedeploy.TimeBasedLinear{}
	if interval, ok := config["interval"]; ok {
		linear.LinearInterval = aws.Int64(int64(interval.(int)))
	}
	if percentage, ok := config["percentage"]; ok {
		linear.LinearPercentage = aws.Int64(int64(percentage.(int)))
	}
	return &linear
}

func flattenMinimumHealthHostsConfig(hosts *codedeploy.MinimumHealthyHosts) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if hosts == nil {
		return result
	}

	item := make(map[string]interface{})

	item["type"] = aws.StringValue(hosts.Type)
	item["value"] = aws.Int64Value(hosts.Value)

	return append(result, item)
}

func flattenTrafficRoutingConfig(config *codedeploy.TrafficRoutingConfig) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if config == nil {
		return result
	}

	item := make(map[string]interface{})

	item["type"] = aws.StringValue(config.Type)
	item["time_based_canary"] = flattenTrafficRoutingCanaryConfig(config.TimeBasedCanary)
	item["time_based_linear"] = flattenTrafficRoutingLinearConfig(config.TimeBasedLinear)

	return append(result, item)
}

func flattenTrafficRoutingCanaryConfig(canary *codedeploy.TimeBasedCanary) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if canary == nil {
		return result
	}

	item := make(map[string]interface{})
	item["interval"] = aws.Int64Value(canary.CanaryInterval)
	item["percentage"] = aws.Int64Value(canary.CanaryPercentage)

	return append(result, item)
}

func flattenTrafficRoutingLinearConfig(linear *codedeploy.TimeBasedLinear) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	if linear == nil {
		return result
	}

	item := make(map[string]interface{})
	item["interval"] = aws.Int64Value(linear.LinearInterval)
	item["percentage"] = aws.Int64Value(linear.LinearPercentage)

	return append(result, item)
}
