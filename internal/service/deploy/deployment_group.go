// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy

import (
	"context"
	"fmt"
	"log"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codedeploy_deployment_group", name="Deployment Group")
// @Tags(identifierAttribute="arn")
func resourceDeploymentGroup() *schema.Resource {
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

				group, err := findDeploymentGroupByTwoPartKey(ctx, conn, applicationName, deploymentGroupName)

				if err != nil {
					return []*schema.ResourceData{}, fmt.Errorf("reading CodeDeploy Deployment Group (%s): %s", d.Id(), err)
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
						names.AttrEnabled: {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_rollback_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
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
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.DeploymentReadyAction](),
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
									names.AttrAction: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.GreenFleetProvisioningAction](),
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
									names.AttrAction: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.InstanceAction](),
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
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.DeploymentOptionWithoutTrafficControl,
							ValidateDiagFunc: enum.Validate[types.DeploymentOption](),
						},
						"deployment_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.DeploymentTypeInPlace,
							ValidateDiagFunc: enum.Validate[types.DeploymentType](),
						},
					},
				},
			},
			"ec2_tag_filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.EC2TagFilterType](),
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
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
									names.AttrKey: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.EC2TagFilterType](),
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"ecs_service": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrClusterName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						names.AttrServiceName: {
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
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"target_group_info": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
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
												names.AttrName: {
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
						names.AttrKey: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.TagFilterType](),
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"outdated_instances_strategy": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.OutdatedInstancesStrategyUpdate,
				ValidateDiagFunc: enum.Validate[types.OutdatedInstancesStrategy](),
			},
			names.AttrServiceRoleARN: {
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
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[types.TriggerEventType](),
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
	serviceRoleArn := d.Get(names.AttrServiceRoleARN).(string)
	input := &codedeploy.CreateDeploymentGroupInput{
		ApplicationName:     aws.String(applicationName),
		DeploymentGroupName: aws.String(deploymentGroupName),
		ServiceRoleArn:      aws.String(serviceRoleArn),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk("alarm_configuration"); ok {
		input.AlarmConfiguration = expandAlarmConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("auto_rollback_configuration"); ok {
		input.AutoRollbackConfiguration = expandAutoRollbackConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("autoscaling_groups"); ok {
		input.AutoScalingGroups = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("blue_green_deployment_config"); ok {
		input.BlueGreenDeploymentConfiguration = expandBlueGreenDeploymentConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("deployment_style"); ok {
		input.DeploymentStyle = expandDeploymentStyle(v.([]interface{}))
	}

	if v, ok := d.GetOk("deployment_config_name"); ok {
		input.DeploymentConfigName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ec2_tag_set"); ok {
		input.Ec2TagSet = expandEC2TagSet(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ec2_tag_filter"); ok {
		input.Ec2TagFilters = expandEC2TagFilters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ecs_service"); ok {
		input.EcsServices = expandECSServices(v.([]interface{}))
	}

	if v, ok := d.GetOk("load_balancer_info"); ok {
		input.LoadBalancerInfo = expandLoadBalancerInfo(v.([]interface{}))
	}

	if v, ok := d.GetOk("on_premises_instance_tag_filter"); ok {
		input.OnPremisesInstanceTagFilters = expandTagFilters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("outdated_instances_strategy"); ok {
		input.OutdatedInstancesStrategy = types.OutdatedInstancesStrategy(v.(string))
	}

	if v, ok := d.GetOk("trigger_configuration"); ok {
		input.TriggerConfigurations = expandTriggerConfigs(v.(*schema.Set).List())
	}

	outputRaw, err := tfresource.RetryWhen(ctx, 5*time.Minute,
		func() (interface{}, error) {
			return conn.CreateDeploymentGroup(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*types.InvalidRoleException](err) {
				return true, err
			}

			if errs.IsAErrorMessageContains[*types.InvalidTriggerConfigException](err, "Topic ARN") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeDeploy Deployment Group (%s): %s", deploymentGroupName, err)
	}

	d.SetId(aws.ToString(outputRaw.(*codedeploy.CreateDeploymentGroupOutput).DeploymentGroupId))

	return append(diags, resourceDeploymentGroupRead(ctx, d, meta)...)
}

func resourceDeploymentGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	group, err := findDeploymentGroupByTwoPartKey(ctx, conn, d.Get("app_name").(string), d.Get("deployment_group_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeDeploy Deployment Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeDeploy Deployment Group (%s): %s", d.Id(), err)
	}

	if err := d.Set("alarm_configuration", flattenAlarmConfiguration(group.AlarmConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting alarm_configuration: %s", err)
	}
	appName := aws.ToString(group.ApplicationName)
	groupName := aws.ToString(group.DeploymentGroupName)
	d.Set("app_name", appName)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "codedeploy",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("deploymentgroup:%s/%s", appName, groupName),
	}.String()
	d.Set(names.AttrARN, arn)
	if err := d.Set("auto_rollback_configuration", flattenAutoRollbackConfiguration(group.AutoRollbackConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting auto_rollback_configuration: %s", err)
	}
	d.Set("autoscaling_groups", tfslices.ApplyToAll(group.AutoScalingGroups, func(v types.AutoScalingGroup) string {
		return aws.ToString(v.Name)
	}))
	if err := d.Set("blue_green_deployment_config", flattenBlueGreenDeploymentConfiguration(group.BlueGreenDeploymentConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting blue_green_deployment_config: %s", err)
	}
	d.Set("compute_platform", group.ComputePlatform)
	d.Set("deployment_config_name", group.DeploymentConfigName)
	d.Set("deployment_group_id", group.DeploymentGroupId)
	d.Set("deployment_group_name", group.DeploymentGroupName)
	if err := d.Set("deployment_style", flattenDeploymentStyle(group.DeploymentStyle)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting deployment_style: %s", err)
	}
	if err := d.Set("ec2_tag_filter", flattenEC2TagFilters(group.Ec2TagFilters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ec2_tag_filter: %s", err)
	}
	if err := d.Set("ec2_tag_set", flattenEC2TagSet(group.Ec2TagSet)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ec2_tag_set: %s", err)
	}
	if err := d.Set("ecs_service", flattenECSServices(group.EcsServices)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ecs_service: %s", err)
	}
	if err := d.Set("load_balancer_info", flattenLoadBalancerInfo(group.LoadBalancerInfo)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting load_balancer_info: %s", err)
	}
	if err := d.Set("on_premises_instance_tag_filter", flattenTagFilters(group.OnPremisesInstanceTagFilters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting on_premises_instance_tag_filter: %s", err)
	}
	d.Set("outdated_instances_strategy", group.OutdatedInstancesStrategy)
	d.Set(names.AttrServiceRoleARN, group.ServiceRoleArn)
	if err := d.Set("trigger_configuration", flattenTriggerConfigs(group.TriggerConfigurations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting trigger_configuration: %s", err)
	}

	return diags
}

func resourceDeploymentGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		// required fields
		applicationName := d.Get("app_name").(string)
		serviceRoleArn := d.Get(names.AttrServiceRoleARN).(string)

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
			input.DeploymentStyle = expandDeploymentStyle(n.([]interface{}))
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
			onPremFilters := expandTagFilters(n.(*schema.Set).List())
			input.OnPremisesInstanceTagFilters = onPremFilters
		}

		if d.HasChange("ec2_tag_set") {
			_, n := d.GetChange("ec2_tag_set")
			ec2TagSet := expandEC2TagSet(n.(*schema.Set).List())
			input.Ec2TagSet = ec2TagSet
		}

		if d.HasChange("ec2_tag_filter") {
			_, n := d.GetChange("ec2_tag_filter")
			ec2Filters := expandEC2TagFilters(n.(*schema.Set).List())
			input.Ec2TagFilters = ec2Filters
		}

		if d.HasChange("ecs_service") {
			input.EcsServices = expandECSServices(d.Get("ecs_service").([]interface{}))
		}

		if d.HasChange("trigger_configuration") {
			_, n := d.GetChange("trigger_configuration")
			triggerConfigs := expandTriggerConfigs(n.(*schema.Set).List())
			input.TriggerConfigurations = triggerConfigs
		}

		if d.HasChange("auto_rollback_configuration") {
			_, n := d.GetChange("auto_rollback_configuration")
			input.AutoRollbackConfiguration = expandAutoRollbackConfiguration(n.([]interface{}))
		}

		if d.HasChange("alarm_configuration") {
			_, n := d.GetChange("alarm_configuration")
			input.AlarmConfiguration = expandAlarmConfiguration(n.([]interface{}))
		}

		if d.HasChange("load_balancer_info") {
			_, n := d.GetChange("load_balancer_info")
			input.LoadBalancerInfo = expandLoadBalancerInfo(n.([]interface{}))
		}

		if d.HasChange("blue_green_deployment_config") {
			_, n := d.GetChange("blue_green_deployment_config")
			input.BlueGreenDeploymentConfiguration = expandBlueGreenDeploymentConfiguration(n.([]interface{}))
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

	if errs.IsA[*types.DeploymentGroupDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeDeploy Deployment Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findDeploymentGroupByTwoPartKey(ctx context.Context, conn *codedeploy.Client, applicationName, deploymentGroupName string) (*types.DeploymentGroupInfo, error) {
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

func expandTagFilters(configured []interface{}) []types.TagFilter {
	filters := make([]types.TagFilter, 0)
	for _, raw := range configured {
		var filter types.TagFilter
		m := raw.(map[string]interface{})

		if v, ok := m[names.AttrKey]; ok {
			filter.Key = aws.String(v.(string))
		}
		if v, ok := m[names.AttrType]; ok {
			filter.Type = types.TagFilterType(v.(string))
		}
		if v, ok := m[names.AttrValue]; ok {
			filter.Value = aws.String(v.(string))
		}

		filters = append(filters, filter)
	}

	return filters
}

func expandEC2TagFilters(configured []interface{}) []types.EC2TagFilter {
	filters := make([]types.EC2TagFilter, 0)
	for _, raw := range configured {
		var filter types.EC2TagFilter
		m := raw.(map[string]interface{})

		filter.Key = aws.String(m[names.AttrKey].(string))
		filter.Type = types.EC2TagFilterType(m[names.AttrType].(string))
		filter.Value = aws.String(m[names.AttrValue].(string))

		filters = append(filters, filter)
	}

	return filters
}

func expandEC2TagSet(configured []interface{}) *types.EC2TagSet {
	filterSets := make([][]types.EC2TagFilter, 0)
	for _, raw := range configured {
		m := raw.(map[string]interface{})
		rawFilters := m["ec2_tag_filter"].(*schema.Set)
		filters := expandEC2TagFilters(rawFilters.List())
		filterSets = append(filterSets, filters)
	}
	return &types.EC2TagSet{Ec2TagSetList: filterSets}
}

func expandTriggerConfigs(configured []interface{}) []types.TriggerConfig {
	configs := make([]types.TriggerConfig, 0, len(configured))
	for _, raw := range configured {
		var config types.TriggerConfig
		m := raw.(map[string]interface{})

		config.TriggerEvents = flex.ExpandStringyValueSet[types.TriggerEventType](m["trigger_events"].(*schema.Set))
		config.TriggerName = aws.String(m["trigger_name"].(string))
		config.TriggerTargetArn = aws.String(m["trigger_target_arn"].(string))

		configs = append(configs, config)
	}
	return configs
}

func expandAutoRollbackConfiguration(configured []interface{}) *types.AutoRollbackConfiguration {
	result := &types.AutoRollbackConfiguration{}

	if len(configured) == 1 {
		config := configured[0].(map[string]interface{})
		result.Enabled = config[names.AttrEnabled].(bool)
		result.Events = flex.ExpandStringyValueSet[types.AutoRollbackEvent](config["events"].(*schema.Set))
	} else { // delete the configuration
		result.Enabled = false
		result.Events = make([]types.AutoRollbackEvent, 0)
	}

	return result
}

func expandAlarmConfiguration(configured []interface{}) *types.AlarmConfiguration {
	result := &types.AlarmConfiguration{}

	if len(configured) == 1 {
		config := configured[0].(map[string]interface{})
		n := flex.ExpandStringSet(config["alarms"].(*schema.Set))
		alarms := make([]types.Alarm, 0, len(n))

		for _, name := range n {
			alarm := types.Alarm{
				Name: name,
			}
			alarms = append(alarms, alarm)
		}

		result.Alarms = alarms
		result.Enabled = config[names.AttrEnabled].(bool)
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
			ClusterName: aws.String(m[names.AttrClusterName].(string)),
			ServiceName: aws.String(m[names.AttrServiceName].(string)),
		}

		ecsServices = append(ecsServices, ecsService)
	}

	return ecsServices
}

func expandELBInfos(l []interface{}) []types.ELBInfo {
	elbInfos := []types.ELBInfo{}

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}

		m := mRaw.(map[string]interface{})

		elbInfo := types.ELBInfo{
			Name: aws.String(m[names.AttrName].(string)),
		}

		elbInfos = append(elbInfos, elbInfo)
	}

	return elbInfos
}

func expandTargetGroupInfos(l []interface{}) []types.TargetGroupInfo {
	targetGroupInfos := []types.TargetGroupInfo{}

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}

		m := mRaw.(map[string]interface{})

		targetGroupInfo := types.TargetGroupInfo{
			Name: aws.String(m[names.AttrName].(string)),
		}

		targetGroupInfos = append(targetGroupInfos, targetGroupInfo)
	}

	return targetGroupInfos
}

func expandTargetGroupPairInfos(l []interface{}) []types.TargetGroupPairInfo {
	targetGroupPairInfos := []types.TargetGroupPairInfo{}

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}

		m := mRaw.(map[string]interface{})

		targetGroupPairInfo := types.TargetGroupPairInfo{
			ProdTrafficRoute: expandTrafficRoute(m["prod_traffic_route"].([]interface{})),
			TargetGroups:     expandTargetGroupInfos(m["target_group"].([]interface{})),
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

func expandDeploymentStyle(list []interface{}) *types.DeploymentStyle {
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

func expandLoadBalancerInfo(list []interface{}) *types.LoadBalancerInfo {
	loadBalancerInfo := &types.LoadBalancerInfo{}
	if len(list) == 0 || list[0] == nil {
		return loadBalancerInfo
	}

	lbInfo := list[0].(map[string]interface{})

	if attr, ok := lbInfo["elb_info"]; ok && attr.(*schema.Set).Len() > 0 {
		loadBalancerInfo.ElbInfoList = expandELBInfos(attr.(*schema.Set).List())
	}

	if attr, ok := lbInfo["target_group_info"]; ok && attr.(*schema.Set).Len() > 0 {
		loadBalancerInfo.TargetGroupInfoList = expandTargetGroupInfos(attr.(*schema.Set).List())
	}

	if attr, ok := lbInfo["target_group_pair_info"]; ok && len(attr.([]interface{})) > 0 {
		loadBalancerInfo.TargetGroupPairInfoList = expandTargetGroupPairInfos(attr.([]interface{}))
	}

	return loadBalancerInfo
}

func expandBlueGreenDeploymentConfiguration(list []interface{}) *types.BlueGreenDeploymentConfiguration {
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
			if v, ok := m[names.AttrAction]; ok {
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
			if v, ok := m[names.AttrAction]; ok {
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

func flattenEC2TagFilters(list []types.EC2TagFilter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, tf := range list {
		l := make(map[string]interface{})
		if v := tf.Key; aws.ToString(v) != "" {
			l[names.AttrKey] = aws.ToString(v)
		}
		if v := tf.Value; aws.ToString(v) != "" {
			l[names.AttrValue] = aws.ToString(v)
		}
		if v := tf.Type; v != "" {
			l[names.AttrType] = string(v)
		}
		result = append(result, l)
	}
	return result
}

func flattenTagFilters(list []types.TagFilter) []map[string]string {
	result := make([]map[string]string, 0, len(list))
	for _, tf := range list {
		l := make(map[string]string)
		if v := tf.Key; aws.ToString(v) != "" {
			l[names.AttrKey] = aws.ToString(v)
		}
		if v := tf.Value; aws.ToString(v) != "" {
			l[names.AttrValue] = aws.ToString(v)
		}
		if v := tf.Type; string(v) != "" {
			l[names.AttrType] = string(v)
		}
		result = append(result, l)
	}
	return result
}

func flattenEC2TagSet(tagSet *types.EC2TagSet) []map[string]interface{} {
	var result []map[string]interface{}
	if tagSet == nil {
		result = make([]map[string]interface{}, 0)
	} else {
		result = make([]map[string]interface{}, 0, len(tagSet.Ec2TagSetList))
		for _, filterSet := range tagSet.Ec2TagSetList {
			filters := flattenEC2TagFilters(filterSet)
			filtersAsIntfSlice := make([]interface{}, 0, len(filters))
			for _, item := range filters {
				filtersAsIntfSlice = append(filtersAsIntfSlice, item)
			}
			tagFilters := map[string]interface{}{
				"ec2_tag_filter": filtersAsIntfSlice,
			}
			result = append(result, tagFilters)
		}
	}
	return result
}

func flattenTriggerConfigs(list []types.TriggerConfig) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, tc := range list {
		item := make(map[string]interface{})
		item["trigger_events"] = tc.TriggerEvents
		item["trigger_name"] = aws.ToString(tc.TriggerName)
		item["trigger_target_arn"] = aws.ToString(tc.TriggerTargetArn)
		result = append(result, item)
	}
	return result
}

func flattenAutoRollbackConfiguration(config *types.AutoRollbackConfiguration) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	// only create configurations that are enabled or temporarily disabled (retaining events)
	// otherwise empty configurations will be created
	if config != nil && (config.Enabled || len(config.Events) > 0) {
		item := make(map[string]interface{})
		item[names.AttrEnabled] = config.Enabled
		item["events"] = config.Events
		result = append(result, item)
	}

	return result
}

func flattenAlarmConfiguration(config *types.AlarmConfiguration) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	// only create configurations that are enabled or temporarily disabled (retaining alarms)
	// otherwise empty configurations will be created
	if config != nil && (config.Enabled || len(config.Alarms) > 0) {
		n := make([]*string, 0, len(config.Alarms))
		for _, alarm := range config.Alarms {
			n = append(n, alarm.Name)
		}

		item := make(map[string]interface{})
		item["alarms"] = flex.FlattenStringSet(n)
		item[names.AttrEnabled] = config.Enabled
		item["ignore_poll_alarm_failure"] = config.IgnorePollAlarmFailure

		result = append(result, item)
	}

	return result
}

func flattenECSServices(ecsServices []types.ECSService) []interface{} {
	l := make([]interface{}, 0)

	for _, ecsService := range ecsServices {
		m := map[string]interface{}{
			names.AttrClusterName: aws.ToString(ecsService.ClusterName),
			names.AttrServiceName: aws.ToString(ecsService.ServiceName),
		}

		l = append(l, m)
	}

	return l
}

func flattenELBInfos(elbInfos []types.ELBInfo) []interface{} {
	l := make([]interface{}, 0)

	for _, elbInfo := range elbInfos {
		m := map[string]interface{}{
			names.AttrName: aws.ToString(elbInfo.Name),
		}

		l = append(l, m)
	}

	return l
}

func flattenTargetGroupInfos(targetGroupInfos []types.TargetGroupInfo) []interface{} {
	l := make([]interface{}, 0)

	for _, targetGroupInfo := range targetGroupInfos {
		m := map[string]interface{}{
			names.AttrName: aws.ToString(targetGroupInfo.Name),
		}

		l = append(l, m)
	}

	return l
}

func flattenTargetGroupPairInfos(targetGroupPairInfos []types.TargetGroupPairInfo) []interface{} {
	l := make([]interface{}, 0)

	for _, targetGroupPairInfo := range targetGroupPairInfos {
		m := map[string]interface{}{
			"prod_traffic_route": flattenTrafficRoute(targetGroupPairInfo.ProdTrafficRoute),
			"target_group":       flattenTargetGroupInfos(targetGroupPairInfo.TargetGroups),
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

func flattenDeploymentStyle(style *types.DeploymentStyle) []map[string]interface{} {
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

func flattenLoadBalancerInfo(loadBalancerInfo *types.LoadBalancerInfo) []interface{} {
	if loadBalancerInfo == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"elb_info":               flattenELBInfos(loadBalancerInfo.ElbInfoList),
		"target_group_info":      flattenTargetGroupInfos(loadBalancerInfo.TargetGroupInfoList),
		"target_group_pair_info": flattenTargetGroupPairInfos(loadBalancerInfo.TargetGroupPairInfoList),
	}

	return []interface{}{m}
}

func flattenBlueGreenDeploymentConfiguration(config *types.BlueGreenDeploymentConfiguration) []map[string]interface{} {
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
			deploymentReadyOption["wait_time_in_minutes"] = v
		}

		m["deployment_ready_option"] = append(a, deploymentReadyOption)
	}

	if config.GreenFleetProvisioningOption != nil {
		b := make([]map[string]interface{}, 0)
		greenFleetProvisioningOption := make(map[string]interface{})

		if v := string(config.GreenFleetProvisioningOption.Action); v != "" {
			greenFleetProvisioningOption[names.AttrAction] = v
		}

		m["green_fleet_provisioning_option"] = append(b, greenFleetProvisioningOption)
	}

	if config.TerminateBlueInstancesOnDeploymentSuccess != nil {
		c := make([]map[string]interface{}, 0)
		blueInstanceTerminationOption := make(map[string]interface{})

		if v := string(config.TerminateBlueInstancesOnDeploymentSuccess.Action); v != "" {
			blueInstanceTerminationOption[names.AttrAction] = v
		}
		if v := config.TerminateBlueInstancesOnDeploymentSuccess.TerminationWaitTimeInMinutes; v != 0 {
			blueInstanceTerminationOption["termination_wait_time_in_minutes"] = v
		}

		m["terminate_blue_instances_on_deployment_success"] = append(c, blueInstanceTerminationOption)
	}

	list := make([]map[string]interface{}, 0)
	list = append(list, m)
	return list
}
