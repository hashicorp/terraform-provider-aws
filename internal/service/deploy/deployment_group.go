// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codedeploy_deployment_group", name="Deployment Group")
// @Tags(identifierAttribute="arn")
func ResourceDeploymentGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeploymentGroupCreate,
		ReadWithoutTimeout:   resourceDeploymentGroupRead,
		UpdateWithoutTimeout: resourceDeploymentGroupUpdate,
		DeleteWithoutTimeout: resourceDeploymentGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")

				if len(idParts) != 2 {
					return []*schema.ResourceData{}, fmt.Errorf("expected ID in format ApplicationName:DeploymentGroupName, received: %s", d.Id())
				}

				applicationName := idParts[0]
				deploymentGroupName := idParts[1]
				conn := meta.(*conns.AWSClient).DeployClient(ctx)

				group, err := FindDeploymentGroupByTwoPartKey(ctx, conn, applicationName, deploymentGroupName)

				if err != nil {
					return []*schema.ResourceData{}, err
				}

				d.SetId(aws.ToString(group.DeploymentGroupId))
				d.Set("app_name", applicationName)
				d.Set("deployment_group_name", deploymentGroupName)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"alarm_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alarms": {
							Type:     schema.TypeSet,
							MaxItems: 10,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"ignore_poll_alarm_failure": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"app_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_rollback_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"events": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"autoscaling_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"blue_green_deployment_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"deployment_ready_option": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action_on_timeout": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(flattenReadyActionValues(types.DeploymentReadyAction("").Values()), false),
									},
									"wait_time_in_minutes": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"green_fleet_provisioning_option": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(flattenGreenFleetProvisioningActionValues(types.GreenFleetProvisioningAction("").Values()), false),
									},
								},
							},
						},
						"terminate_blue_instances_on_deployment_success": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(flattenInstanceActionValues(types.InstanceAction("").Values()), false),
									},
									"termination_wait_time_in_minutes": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtMost(2880),
									},
								},
							},
						},
					},
				},
			},
			"compute_platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_config_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "CodeDeployDefault.OneAtATime",
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"deployment_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"deployment_style": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"deployment_option": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      string(types.DeploymentOptionWithoutTrafficControl),
							ValidateFunc: validation.StringInSlice(flattenOptionValues(types.DeploymentOption("").Values()), false),
						},
						"deployment_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      string(types.DeploymentTypeInPlace),
							ValidateFunc: validation.StringInSlice(flattenTypeValues(types.DeploymentType("").Values()), false),
						},
					},
				},
			},
			"ec2_tag_filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validTagFilters,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceTagFilterHash,
			},
			"ec2_tag_set": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ec2_tag_filter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validTagFilters,
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							Set: resourceTagFilterHash,
						},
					},
				},
				Set: resourceTagSetHash,
			},
			"ecs_service": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"service_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},
			"load_balancer_info": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"elb_info": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      LoadBalancerInfoHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"target_group_info": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      LoadBalancerInfoHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"target_group_pair_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"prod_traffic_route": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"listener_arns": {
													Type:     schema.TypeSet,
													Required: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidARN,
													},
												},
											},
										},
									},
									"target_group": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 2,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.NoZeroValues,
												},
											},
										},
									},
									"test_traffic_route": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"listener_arns": {
													Type:     schema.TypeSet,
													Required: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidARN,
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
			},
			"on_premises_instance_tag_filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validTagFilters,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceTagFilterHash,
			},
			"outdated_instances_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      string(types.OutdatedInstancesStrategyUpdate),
				ValidateFunc: validation.StringInSlice(flattenOutdatedInstancesStrategyValues(types.OutdatedInstancesStrategy("").Values()), false),
			},
			"service_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"trigger_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"trigger_events": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(flattenTriggerEventTypeValues(types.TriggerEventType("").Values()), false),
							},
						},
						"trigger_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"trigger_target_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
				Set: resourceTriggerHashConfig,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDeploymentGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	applicationName := d.Get("app_name").(string)
	deploymentGroupName := d.Get("deployment_group_name").(string)
	serviceRoleArn := d.Get("service_role_arn").(string)
	input := codedeploy.CreateDeploymentGroupInput{
		ApplicationName:     aws.String(applicationName),
		DeploymentGroupName: aws.String(deploymentGroupName),
		ServiceRoleArn:      aws.String(serviceRoleArn),
		Tags:                getTagsIn(ctx),
	}

	if attr, ok := d.GetOk("deployment_style"); ok {
		input.DeploymentStyle = ExpandDeploymentStyle(attr.([]interface{}))
	}

	if attr, ok := d.GetOk("deployment_config_name"); ok {
		input.DeploymentConfigName = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("autoscaling_groups"); ok {
		input.AutoScalingGroups = flex.ExpandStringValueSet(attr.(*schema.Set))
	}

	if attr, ok := d.GetOk("on_premises_instance_tag_filter"); ok {
		onPremFilters := buildOnPremTagFilters(attr.(*schema.Set).List())
		input.OnPremisesInstanceTagFilters = onPremFilters
	}

	if attr, ok := d.GetOk("ec2_tag_set"); ok {
		input.Ec2TagSet = buildEC2TagSet(attr.(*schema.Set).List())
	}

	if attr, ok := d.GetOk("ec2_tag_filter"); ok {
		input.Ec2TagFilters = buildEC2TagFilters(attr.(*schema.Set).List())
	}

	if attr, ok := d.GetOk("ecs_service"); ok {
		input.EcsServices = expandECSServices(attr.([]interface{}))
	}

	if attr, ok := d.GetOk("trigger_configuration"); ok {
		triggerConfigs := BuildTriggerConfigs(attr.(*schema.Set).List())
		input.TriggerConfigurations = triggerConfigs
	}

	if attr, ok := d.GetOk("auto_rollback_configuration"); ok {
		input.AutoRollbackConfiguration = BuildAutoRollbackConfig(attr.([]interface{}))
	}

	if attr, ok := d.GetOk("alarm_configuration"); ok {
		input.AlarmConfiguration = BuildAlarmConfig(attr.([]interface{}))
	}

	if attr, ok := d.GetOk("load_balancer_info"); ok {
		input.LoadBalancerInfo = ExpandLoadBalancerInfo(attr.([]interface{}))
	}

	if attr, ok := d.GetOk("blue_green_deployment_config"); ok {
		input.BlueGreenDeploymentConfiguration = ExpandBlueGreenDeploymentConfig(attr.([]interface{}))
	}

	if attr, ok := d.GetOk("outdated_instances_strategy"); ok {
		input.OutdatedInstancesStrategy = types.OutdatedInstancesStrategy(attr.(string))
	}

	var resp *codedeploy.CreateDeploymentGroupOutput
	var err error
	err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		resp, err = conn.CreateDeploymentGroup(ctx, &input)

		if errs.IsA[*types.InvalidRoleException](err) {
			return retry.RetryableError(err)
		}

		if errs.IsAErrorMessageContains[*types.InvalidTriggerConfigException](err, "Topic ARN") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.CreateDeploymentGroup(ctx, &input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeDeploy Deployment Group: %s", err)
	}

	d.SetId(aws.ToString(resp.DeploymentGroupId))

	return append(diags, resourceDeploymentGroupRead(ctx, d, meta)...)
}

func resourceDeploymentGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	group, err := FindDeploymentGroupByTwoPartKey(ctx, conn, d.Get("app_name").(string), d.Get("deployment_group_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeDeploy Deployment Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeDeploy Deployment Group (%s): %s", d.Id(), err)
	}

	appName := aws.ToString(group.ApplicationName)
	groupName := aws.ToString(group.DeploymentGroupName)
	groupArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "codedeploy",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("deploymentgroup:%s/%s", appName, groupName),
	}.String()

	d.Set("arn", groupArn)
	d.Set("app_name", appName)
	d.Set("deployment_config_name", group.DeploymentConfigName)
	d.Set("deployment_group_name", group.DeploymentGroupName)
	d.Set("deployment_group_id", group.DeploymentGroupId)
	d.Set("compute_platform", group.ComputePlatform)
	d.Set("service_role_arn", group.ServiceRoleArn)
	d.Set("outdated_instances_strategy", group.OutdatedInstancesStrategy)

	autoScalingGroups := make([]string, len(group.AutoScalingGroups))
	for i, autoScalingGroup := range group.AutoScalingGroups {
		autoScalingGroups[i] = aws.ToString(autoScalingGroup.Name)
	}
	if err := d.Set("autoscaling_groups", autoScalingGroups); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting autoscaling_groups: %s", err)
	}

	if err := d.Set("deployment_style", FlattenDeploymentStyle(group.DeploymentStyle)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting deployment_style: %s", err)
	}

	if err := d.Set("ec2_tag_set", ec2TagSetToMap(group.Ec2TagSet)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ec2_tag_set: %s", err)
	}

	if err := d.Set("ec2_tag_filter", ec2TagFiltersToMap(group.Ec2TagFilters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ec2_tag_filter: %s", err)
	}

	if err := d.Set("ecs_service", flattenECSServices(group.EcsServices)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ecs_service: %s", err)
	}

	if err := d.Set("on_premises_instance_tag_filter", onPremisesTagFiltersToMap(group.OnPremisesInstanceTagFilters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting on_premises_instance_tag_filter: %s", err)
	}

	if err := d.Set("trigger_configuration", TriggerConfigsToMap(group.TriggerConfigurations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting trigger_configuration: %s", err)
	}

	if err := d.Set("auto_rollback_configuration", AutoRollbackConfigToMap(group.AutoRollbackConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting auto_rollback_configuration: %s", err)
	}

	if err := d.Set("alarm_configuration", AlarmConfigToMap(group.AlarmConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting alarm_configuration: %s", err)
	}

	if err := d.Set("load_balancer_info", FlattenLoadBalancerInfo(group.LoadBalancerInfo)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting load_balancer_info: %s", err)
	}

	if err := d.Set("blue_green_deployment_config", FlattenBlueGreenDeploymentConfig(group.BlueGreenDeploymentConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting blue_green_deployment_config: %s", err)
	}

	return diags
}

func resourceDeploymentGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		// required fields
		applicationName := d.Get("app_name").(string)
		serviceRoleArn := d.Get("service_role_arn").(string)

		input := codedeploy.UpdateDeploymentGroupInput{
			ApplicationName: aws.String(applicationName),
			ServiceRoleArn:  aws.String(serviceRoleArn),
		}

		if d.HasChange("deployment_group_name") {
			o, n := d.GetChange("deployment_group_name")
			input.CurrentDeploymentGroupName = aws.String(o.(string))
			input.NewDeploymentGroupName = aws.String(n.(string))
		} else {
			input.CurrentDeploymentGroupName = aws.String(d.Get("deployment_group_name").(string))
		}

		if d.HasChange("deployment_style") {
			_, n := d.GetChange("deployment_style")
			input.DeploymentStyle = ExpandDeploymentStyle(n.([]interface{}))
		}

		if d.HasChange("deployment_config_name") {
			_, n := d.GetChange("deployment_config_name")
			input.DeploymentConfigName = aws.String(n.(string))
		}

		// include (original or new) autoscaling groups when blue_green_deployment_config changes except for ECS
		if _, isEcs := d.GetOk("ecs_service"); d.HasChange("autoscaling_groups") || (d.HasChange("blue_green_deployment_config") && !isEcs) {
			_, n := d.GetChange("autoscaling_groups")
			input.AutoScalingGroups = flex.ExpandStringValueSet(n.(*schema.Set))
		}

		// TagFilters aren't like tags. They don't append. They simply replace.
		if d.HasChange("on_premises_instance_tag_filter") {
			_, n := d.GetChange("on_premises_instance_tag_filter")
			onPremFilters := buildOnPremTagFilters(n.(*schema.Set).List())
			input.OnPremisesInstanceTagFilters = onPremFilters
		}

		if d.HasChange("ec2_tag_set") {
			_, n := d.GetChange("ec2_tag_set")
			ec2TagSet := buildEC2TagSet(n.(*schema.Set).List())
			input.Ec2TagSet = ec2TagSet
		}

		if d.HasChange("ec2_tag_filter") {
			_, n := d.GetChange("ec2_tag_filter")
			ec2Filters := buildEC2TagFilters(n.(*schema.Set).List())
			input.Ec2TagFilters = ec2Filters
		}

		if d.HasChange("ecs_service") {
			input.EcsServices = expandECSServices(d.Get("ecs_service").([]interface{}))
		}

		if d.HasChange("trigger_configuration") {
			_, n := d.GetChange("trigger_configuration")
			triggerConfigs := BuildTriggerConfigs(n.(*schema.Set).List())
			input.TriggerConfigurations = triggerConfigs
		}

		if d.HasChange("auto_rollback_configuration") {
			_, n := d.GetChange("auto_rollback_configuration")
			input.AutoRollbackConfiguration = BuildAutoRollbackConfig(n.([]interface{}))
		}

		if d.HasChange("alarm_configuration") {
			_, n := d.GetChange("alarm_configuration")
			input.AlarmConfiguration = BuildAlarmConfig(n.([]interface{}))
		}

		if d.HasChange("load_balancer_info") {
			_, n := d.GetChange("load_balancer_info")
			input.LoadBalancerInfo = ExpandLoadBalancerInfo(n.([]interface{}))
		}

		if d.HasChange("blue_green_deployment_config") {
			_, n := d.GetChange("blue_green_deployment_config")
			input.BlueGreenDeploymentConfiguration = ExpandBlueGreenDeploymentConfig(n.([]interface{}))
		}

		if d.HasChange("outdated_instances_strategy") {
			o, n := d.GetChange("outdated_instances_strategy")
			if n.(string) == "" && o.(string) == string(types.OutdatedInstancesStrategyIgnore) { // if the user is trying to remove the strategy, set it to update (the default)
				input.OutdatedInstancesStrategy = types.OutdatedInstancesStrategyUpdate
			} else if n.(string) != "" { //
				input.OutdatedInstancesStrategy = types.OutdatedInstancesStrategy(n.(string))
			}
		}

		log.Printf("[DEBUG] Updating CodeDeploy DeploymentGroup %s", d.Id())

		var err error
		err = retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
			_, err = conn.UpdateDeploymentGroup(ctx, &input)

			if errs.IsA[*types.InvalidRoleException](err) {
				return retry.RetryableError(err)
			}

			if errs.IsAErrorMessageContains[*types.InvalidTriggerConfigException](err, "Topic ARN") {
				return retry.RetryableError(err)
			}

			if err != nil {
				return retry.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateDeploymentGroup(ctx, &input)
		}
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeDeploy deployment group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDeploymentGroupRead(ctx, d, meta)...)
}

func resourceDeploymentGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	log.Printf("[DEBUG] Deleting CodeDeploy Deployment Group: %s", d.Id())
	_, err := conn.DeleteDeploymentGroup(ctx, &codedeploy.DeleteDeploymentGroupInput{
		ApplicationName:     aws.String(d.Get("app_name").(string)),
		DeploymentGroupName: aws.String(d.Get("deployment_group_name").(string)),
	})

	if err != nil {
		if errs.IsA[*types.DeploymentGroupDoesNotExistException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting CodeDeploy Deployment Group (%s): %s", d.Id(), err)
	}

	return diags
}

func FindDeploymentGroupByTwoPartKey(ctx context.Context, conn *codedeploy.Client, applicationName, deploymentGroupName string) (*types.DeploymentGroupInfo, error) {
	input := &codedeploy.GetDeploymentGroupInput{
		ApplicationName:     aws.String(applicationName),
		DeploymentGroupName: aws.String(deploymentGroupName),
	}

	output, err := conn.GetDeploymentGroup(ctx, input)

	if errs.IsA[*types.ApplicationDoesNotExistException](err) || errs.IsA[*types.DeploymentGroupDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DeploymentGroupInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DeploymentGroupInfo, nil
}

// buildOnPremTagFilters converts raw schema lists into a list of
// types.TagFilters.
func buildOnPremTagFilters(configured []interface{}) []types.TagFilter {
	filters := make([]types.TagFilter, 0)
	for _, raw := range configured {
		var filter types.TagFilter
		m := raw.(map[string]interface{})

		if v, ok := m["key"]; ok {
			filter.Key = aws.String(v.(string))
		}
		if v, ok := m["type"]; ok {
			filter.Type = types.TagFilterType(v.(string))
		}
		if v, ok := m["value"]; ok {
			filter.Value = aws.String(v.(string))
		}

		filters = append(filters, filter)
	}

	return filters
}

// buildEC2TagFilters converts raw schema lists into a list of
// types.EC2TagFilters.
func buildEC2TagFilters(configured []interface{}) []types.EC2TagFilter {
	filters := make([]types.EC2TagFilter, 0)
	for _, raw := range configured {
		var filter types.EC2TagFilter
		m := raw.(map[string]interface{})

		filter.Key = aws.String(m["key"].(string))
		filter.Type = types.EC2TagFilterType(m["type"].(string))
		filter.Value = aws.String(m["value"].(string))

		filters = append(filters, filter)
	}

	return filters
}

// buildEC2TagSet converts raw schema lists into a types.EC2TagSet.
func buildEC2TagSet(configured []interface{}) *types.EC2TagSet {
	filterSets := make([][]types.EC2TagFilter, 0)
	for _, raw := range configured {
		m := raw.(map[string]interface{})
		rawFilters := m["ec2_tag_filter"].(*schema.Set)
		filters := buildEC2TagFilters(rawFilters.List())
		filterSets = append(filterSets, filters)
	}
	return &types.EC2TagSet{Ec2TagSetList: filterSets}
}

// BuildTriggerConfigs converts a raw schema list into a list of
// types.TriggerConfig.
func BuildTriggerConfigs(configured []interface{}) []types.TriggerConfig {
	configs := make([]types.TriggerConfig, 0, len(configured))
	for _, raw := range configured {
		var config types.TriggerConfig
		m := raw.(map[string]interface{})

		config.TriggerEvents = expandTriggerEventTypes(flex.ExpandStringValueSet(m["trigger_events"].(*schema.Set)))
		config.TriggerName = aws.String(m["trigger_name"].(string))
		config.TriggerTargetArn = aws.String(m["trigger_target_arn"].(string))

		configs = append(configs, config)
	}
	return configs
}

// BuildAutoRollbackConfig converts a raw schema list containing a map[string]interface{}
// into a single types.AutoRollbackConfiguration
func BuildAutoRollbackConfig(configured []interface{}) *types.AutoRollbackConfiguration {
	result := &types.AutoRollbackConfiguration{}

	if len(configured) == 1 {
		config := configured[0].(map[string]interface{})
		result.Enabled = config["enabled"].(bool)
		result.Events = expandAutoRollbackEvents(flex.ExpandStringValueSet(config["events"].(*schema.Set)))
	} else { // delete the configuration
		result.Enabled = false
		result.Events = make([]types.AutoRollbackEvent, 0)
	}

	return result
}

// BuildAlarmConfig converts a raw schema list containing a map[string]interface{}
// into a single types.AlarmConfiguration
func BuildAlarmConfig(configured []interface{}) *types.AlarmConfiguration {
	result := &types.AlarmConfiguration{}

	if len(configured) == 1 {
		config := configured[0].(map[string]interface{})
		names := flex.ExpandStringSet(config["alarms"].(*schema.Set))
		alarms := make([]types.Alarm, 0, len(names))

		for _, name := range names {
			alarm := types.Alarm{
				Name: name,
			}
			alarms = append(alarms, alarm)
		}

		result.Alarms = alarms
		result.Enabled = config["enabled"].(bool)
		result.IgnorePollAlarmFailure = config["ignore_poll_alarm_failure"].(bool)
	} else { // delete the configuration
		result.Alarms = make([]types.Alarm, 0)
		result.Enabled = false
		result.IgnorePollAlarmFailure = false
	}

	return result
}

func expandECSServices(l []interface{}) []types.ECSService {
	ecsServices := make([]types.ECSService, 0)

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}

		m := mRaw.(map[string]interface{})

		ecsService := types.ECSService{
			ClusterName: aws.String(m["cluster_name"].(string)),
			ServiceName: aws.String(m["service_name"].(string)),
		}

		ecsServices = append(ecsServices, ecsService)
	}

	return ecsServices
}

func expandELBInfo(l []interface{}) []types.ELBInfo {
	elbInfos := []types.ELBInfo{}

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}

		m := mRaw.(map[string]interface{})

		elbInfo := types.ELBInfo{
			Name: aws.String(m["name"].(string)),
		}

		elbInfos = append(elbInfos, elbInfo)
	}

	return elbInfos
}

func expandTargetGroupInfo(l []interface{}) []types.TargetGroupInfo {
	targetGroupInfos := []types.TargetGroupInfo{}

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}

		m := mRaw.(map[string]interface{})

		targetGroupInfo := types.TargetGroupInfo{
			Name: aws.String(m["name"].(string)),
		}

		targetGroupInfos = append(targetGroupInfos, targetGroupInfo)
	}

	return targetGroupInfos
}

func expandTargetGroupPairInfo(l []interface{}) []types.TargetGroupPairInfo {
	targetGroupPairInfos := []types.TargetGroupPairInfo{}

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}

		m := mRaw.(map[string]interface{})

		targetGroupPairInfo := types.TargetGroupPairInfo{
			ProdTrafficRoute: expandTrafficRoute(m["prod_traffic_route"].([]interface{})),
			TargetGroups:     expandTargetGroupInfo(m["target_group"].([]interface{})),
			TestTrafficRoute: expandTrafficRoute(m["test_traffic_route"].([]interface{})),
		}

		targetGroupPairInfos = append(targetGroupPairInfos, targetGroupPairInfo)
	}

	return targetGroupPairInfos
}

func expandTrafficRoute(l []interface{}) *types.TrafficRoute {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	trafficRoute := &types.TrafficRoute{
		ListenerArns: flex.ExpandStringValueSet(m["listener_arns"].(*schema.Set)),
	}

	return trafficRoute
}

// ExpandDeploymentStyle converts a raw schema list containing a map[string]interface{}
// into a single types.DeploymentStyle object
func ExpandDeploymentStyle(list []interface{}) *types.DeploymentStyle {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	style := list[0].(map[string]interface{})
	result := &types.DeploymentStyle{}

	if v, ok := style["deployment_option"]; ok {
		result.DeploymentOption = types.DeploymentOption(v.(string))
	}
	if v, ok := style["deployment_type"]; ok {
		result.DeploymentType = types.DeploymentType(v.(string))
	}

	return result
}

// ExpandLoadBalancerInfo converts a raw schema list containing a map[string]interface{}
// into a single types.LoadBalancerInfo object. Returns an empty object if list is nil.
func ExpandLoadBalancerInfo(list []interface{}) *types.LoadBalancerInfo {
	loadBalancerInfo := &types.LoadBalancerInfo{}
	if len(list) == 0 || list[0] == nil {
		return loadBalancerInfo
	}

	lbInfo := list[0].(map[string]interface{})

	if attr, ok := lbInfo["elb_info"]; ok && attr.(*schema.Set).Len() > 0 {
		loadBalancerInfo.ElbInfoList = expandELBInfo(attr.(*schema.Set).List())
	}

	if attr, ok := lbInfo["target_group_info"]; ok && attr.(*schema.Set).Len() > 0 {
		loadBalancerInfo.TargetGroupInfoList = expandTargetGroupInfo(attr.(*schema.Set).List())
	}

	if attr, ok := lbInfo["target_group_pair_info"]; ok && len(attr.([]interface{})) > 0 {
		loadBalancerInfo.TargetGroupPairInfoList = expandTargetGroupPairInfo(attr.([]interface{}))
	}

	return loadBalancerInfo
}

// ExpandBlueGreenDeploymentConfig converts a raw schema list containing a map[string]interface{}
// into a single types.BlueGreenDeploymentConfiguration object
func ExpandBlueGreenDeploymentConfig(list []interface{}) *types.BlueGreenDeploymentConfiguration {
	if len(list) == 0 || list[0] == nil {
		return nil
	}

	config := list[0].(map[string]interface{})
	blueGreenDeploymentConfig := &types.BlueGreenDeploymentConfiguration{}

	if attr, ok := config["deployment_ready_option"]; ok {
		a := attr.([]interface{})

		if len(a) > 0 && a[0] != nil {
			m := a[0].(map[string]interface{})

			deploymentReadyOption := &types.DeploymentReadyOption{}
			if v, ok := m["action_on_timeout"]; ok {
				deploymentReadyOption.ActionOnTimeout = types.DeploymentReadyAction(v.(string))
			}
			if v, ok := m["wait_time_in_minutes"]; ok {
				deploymentReadyOption.WaitTimeInMinutes = int32(v.(int))
			}
			blueGreenDeploymentConfig.DeploymentReadyOption = deploymentReadyOption
		}
	}

	if attr, ok := config["green_fleet_provisioning_option"]; ok {
		a := attr.([]interface{})

		if len(a) > 0 && a[0] != nil {
			m := a[0].(map[string]interface{})

			greenFleetProvisioningOption := &types.GreenFleetProvisioningOption{}
			if v, ok := m["action"]; ok {
				greenFleetProvisioningOption.Action = types.GreenFleetProvisioningAction(v.(string))
			}
			blueGreenDeploymentConfig.GreenFleetProvisioningOption = greenFleetProvisioningOption
		}
	}

	if attr, ok := config["terminate_blue_instances_on_deployment_success"]; ok {
		a := attr.([]interface{})

		if len(a) > 0 && a[0] != nil {
			m := a[0].(map[string]interface{})

			blueInstanceTerminationOption := &types.BlueInstanceTerminationOption{}
			if v, ok := m["action"]; ok {
				blueInstanceTerminationOption.Action = types.InstanceAction(v.(string))
			}
			if v, ok := m["termination_wait_time_in_minutes"]; ok {
				blueInstanceTerminationOption.TerminationWaitTimeInMinutes = int32(v.(int))
			}
			blueGreenDeploymentConfig.TerminateBlueInstancesOnDeploymentSuccess = blueInstanceTerminationOption
		}
	}

	return blueGreenDeploymentConfig
}

// ec2TagFiltersToMap converts lists of tag filters into a []map[string]interface{}.
func ec2TagFiltersToMap(list []types.EC2TagFilter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, tf := range list {
		l := make(map[string]interface{})
		if v := tf.Key; aws.ToString(v) != "" {
			l["key"] = aws.ToString(v)
		}
		if v := tf.Value; aws.ToString(v) != "" {
			l["value"] = aws.ToString(v)
		}
		if v := tf.Type; types.EC2TagFilterType(v) != "" {
			l["type"] = string(v)
		}
		result = append(result, l)
	}
	return result
}

// onPremisesTagFiltersToMap converts lists of on-prem tag filters into a []map[string]string.
func onPremisesTagFiltersToMap(list []types.TagFilter) []map[string]string {
	result := make([]map[string]string, 0, len(list))
	for _, tf := range list {
		l := make(map[string]string)
		if v := tf.Key; aws.ToString(v) != "" {
			l["key"] = aws.ToString(v)
		}
		if v := tf.Value; aws.ToString(v) != "" {
			l["value"] = aws.ToString(v)
		}
		if v := tf.Type; string(v) != "" {
			l["type"] = string(v)
		}
		result = append(result, l)
	}
	return result
}

// ec2TagSetToMap converts lists of tag filters into a [][]map[string]string.
func ec2TagSetToMap(tagSet *types.EC2TagSet) []map[string]interface{} {
	var result []map[string]interface{}
	if tagSet == nil {
		result = make([]map[string]interface{}, 0)
	} else {
		result = make([]map[string]interface{}, 0, len(tagSet.Ec2TagSetList))
		for _, filterSet := range tagSet.Ec2TagSetList {
			filters := ec2TagFiltersToMap(filterSet)
			filtersAsIntfSlice := make([]interface{}, 0, len(filters))
			for _, item := range filters {
				filtersAsIntfSlice = append(filtersAsIntfSlice, item)
			}
			tagFilters := map[string]interface{}{
				"ec2_tag_filter": schema.NewSet(resourceTagFilterHash, filtersAsIntfSlice),
			}
			result = append(result, tagFilters)
		}
	}
	return result
}

// TriggerConfigsToMap converts a list of []types.TriggerConfig into a []map[string]interface{}
func TriggerConfigsToMap(list []types.TriggerConfig) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, tc := range list {
		item := make(map[string]interface{})
		item["trigger_events"] = flattenTriggerEventTypeValues(tc.TriggerEvents)
		item["trigger_name"] = aws.ToString(tc.TriggerName)
		item["trigger_target_arn"] = aws.ToString(tc.TriggerTargetArn)
		result = append(result, item)
	}
	return result
}

// AutoRollbackConfigToMap converts a types.AutoRollbackConfiguration
// into a []map[string]interface{} list containing a single item
func AutoRollbackConfigToMap(config *types.AutoRollbackConfiguration) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	// only create configurations that are enabled or temporarily disabled (retaining events)
	// otherwise empty configurations will be created
	if config != nil && (config.Enabled || len(config.Events) > 0) {
		item := make(map[string]interface{})
		item["enabled"] = config.Enabled
		item["events"] = flattenAutoRollbackEvents(config.Events)
		result = append(result, item)
	}

	return result
}

// AlarmConfigToMap converts a types.AlarmConfiguration
// into a []map[string]interface{} list containing a single item
func AlarmConfigToMap(config *types.AlarmConfiguration) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	// only create configurations that are enabled or temporarily disabled (retaining alarms)
	// otherwise empty configurations will be created
	if config != nil && (config.Enabled || len(config.Alarms) > 0) {
		names := make([]*string, 0, len(config.Alarms))
		for _, alarm := range config.Alarms {
			names = append(names, alarm.Name)
		}

		item := make(map[string]interface{})
		item["alarms"] = flex.FlattenStringSet(names)
		item["enabled"] = config.Enabled
		item["ignore_poll_alarm_failure"] = config.IgnorePollAlarmFailure

		result = append(result, item)
	}

	return result
}

func flattenECSServices(ecsServices []types.ECSService) []interface{} {
	l := make([]interface{}, 0)

	for _, ecsService := range ecsServices {
		m := map[string]interface{}{
			"cluster_name": aws.ToString(ecsService.ClusterName),
			"service_name": aws.ToString(ecsService.ServiceName),
		}

		l = append(l, m)
	}

	return l
}

func flattenELBInfo(elbInfos []types.ELBInfo) []interface{} {
	l := make([]interface{}, 0)

	for _, elbInfo := range elbInfos {
		m := map[string]interface{}{
			"name": aws.ToString(elbInfo.Name),
		}

		l = append(l, m)
	}

	return l
}

func flattenTargetGroupInfo(targetGroupInfos []types.TargetGroupInfo) []interface{} {
	l := make([]interface{}, 0)

	for _, targetGroupInfo := range targetGroupInfos {
		m := map[string]interface{}{
			"name": aws.ToString(targetGroupInfo.Name),
		}

		l = append(l, m)
	}

	return l
}

func flattenTargetGroupPairInfo(targetGroupPairInfos []types.TargetGroupPairInfo) []interface{} {
	l := make([]interface{}, 0)

	for _, targetGroupPairInfo := range targetGroupPairInfos {
		m := map[string]interface{}{
			"prod_traffic_route": flattenTrafficRoute(targetGroupPairInfo.ProdTrafficRoute),
			"target_group":       flattenTargetGroupInfo(targetGroupPairInfo.TargetGroups),
			"test_traffic_route": flattenTrafficRoute(targetGroupPairInfo.TestTrafficRoute),
		}

		l = append(l, m)
	}

	return l
}

func flattenTrafficRoute(trafficRoute *types.TrafficRoute) []interface{} {
	if trafficRoute == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"listener_arns": flex.FlattenStringValueSet(trafficRoute.ListenerArns),
	}

	return []interface{}{m}
}

// FlattenDeploymentStyle converts a types.DeploymentStyle object
// into a []map[string]interface{} list containing a single item
func FlattenDeploymentStyle(style *types.DeploymentStyle) []map[string]interface{} {
	if style == nil {
		return nil
	}

	item := make(map[string]interface{})
	if v := string(style.DeploymentOption); v != "" {
		item["deployment_option"] = v
	}
	if v := string(style.DeploymentType); v != "" {
		item["deployment_type"] = v
	}

	result := make([]map[string]interface{}, 0, 1)
	result = append(result, item)
	return result
}

func FlattenLoadBalancerInfo(loadBalancerInfo *types.LoadBalancerInfo) []interface{} {
	if loadBalancerInfo == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"elb_info":               schema.NewSet(LoadBalancerInfoHash, flattenELBInfo(loadBalancerInfo.ElbInfoList)),
		"target_group_info":      schema.NewSet(LoadBalancerInfoHash, flattenTargetGroupInfo(loadBalancerInfo.TargetGroupInfoList)),
		"target_group_pair_info": flattenTargetGroupPairInfo(loadBalancerInfo.TargetGroupPairInfoList),
	}

	return []interface{}{m}
}

// FlattenBlueGreenDeploymentConfig converts a types.BlueGreenDeploymentConfiguration object
// into a []map[string]interface{} list containing a single item
func FlattenBlueGreenDeploymentConfig(config *types.BlueGreenDeploymentConfiguration) []map[string]interface{} {
	if config == nil {
		return nil
	}

	m := make(map[string]interface{})

	if config.DeploymentReadyOption != nil {
		a := make([]map[string]interface{}, 0)
		deploymentReadyOption := make(map[string]interface{})

		if v := string(config.DeploymentReadyOption.ActionOnTimeout); v != "" {
			deploymentReadyOption["action_on_timeout"] = v
		}
		if v := config.DeploymentReadyOption.WaitTimeInMinutes; v != 0 {
			deploymentReadyOption["wait_time_in_minutes"] = int32(v)
		}

		m["deployment_ready_option"] = append(a, deploymentReadyOption)
	}

	if config.GreenFleetProvisioningOption != nil {
		b := make([]map[string]interface{}, 0)
		greenFleetProvisioningOption := make(map[string]interface{})

		if v := string(config.GreenFleetProvisioningOption.Action); v != "" {
			greenFleetProvisioningOption["action"] = v
		}

		m["green_fleet_provisioning_option"] = append(b, greenFleetProvisioningOption)
	}

	if config.TerminateBlueInstancesOnDeploymentSuccess != nil {
		c := make([]map[string]interface{}, 0)
		blueInstanceTerminationOption := make(map[string]interface{})

		if v := string(config.TerminateBlueInstancesOnDeploymentSuccess.Action); v != "" {
			blueInstanceTerminationOption["action"] = v
		}
		if v := config.TerminateBlueInstancesOnDeploymentSuccess.TerminationWaitTimeInMinutes; v != 0 {
			blueInstanceTerminationOption["termination_wait_time_in_minutes"] = int32(v)
		}

		m["terminate_blue_instances_on_deployment_success"] = append(c, blueInstanceTerminationOption)
	}

	list := make([]map[string]interface{}, 0)
	list = append(list, m)
	return list
}

func resourceTagFilterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	// Nothing's actually required in tag filters, so we must check the
	// presence of all values before attempting a hash.
	if v, ok := m["key"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["value"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}

func resourceTagSetHash(v interface{}) int {
	tagSetMap := v.(map[string]interface{})
	filterSet := tagSetMap["ec2_tag_filter"]
	filterSetSlice := filterSet.(*schema.Set).List()

	var x uint64 = 1
	for i, filter := range filterSetSlice {
		x = ((x << 7) | (x >> (64 - 7))) ^ uint64(i) ^ uint64(resourceTagFilterHash(filter))
	}
	return int(x)
}

func resourceTriggerHashConfig(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["trigger_name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["trigger_target_arn"].(string)))

	if triggerEvents, ok := m["trigger_events"]; ok {
		names := triggerEvents.(*schema.Set).List()
		strings := make([]string, len(names))
		for i, raw := range names {
			strings[i] = raw.(string)
		}
		sort.Strings(strings)

		for _, s := range strings {
			buf.WriteString(fmt.Sprintf("%s-", s))
		}
	}
	return create.StringHashcode(buf.String())
}

func LoadBalancerInfoHash(v interface{}) int {
	var buf bytes.Buffer

	if v == nil {
		return create.StringHashcode(buf.String())
	}

	m := v.(map[string]interface{})
	if v, ok := m["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}

func flattenReadyActionValues(t []types.DeploymentReadyAction) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func flattenGreenFleetProvisioningActionValues(t []types.GreenFleetProvisioningAction) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func flattenInstanceActionValues(t []types.InstanceAction) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func flattenOptionValues(t []types.DeploymentOption) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func flattenTypeValues(t []types.DeploymentType) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func flattenOutdatedInstancesStrategyValues(t []types.OutdatedInstancesStrategy) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func flattenTriggerEventTypeValues(t []types.TriggerEventType) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}

func expandTriggerEventTypes(t []string) []types.TriggerEventType {
	var out []types.TriggerEventType

	for _, v := range t {
		out = append(out, types.TriggerEventType(v))
	}

	return out
}

func expandAutoRollbackEvents(t []string) []types.AutoRollbackEvent {
	var out []types.AutoRollbackEvent

	for _, v := range t {
		out = append(out, types.AutoRollbackEvent(v))
	}

	return out
}

func flattenAutoRollbackEvents(t []types.AutoRollbackEvent) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}
