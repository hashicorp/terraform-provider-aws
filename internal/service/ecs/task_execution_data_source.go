// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ecs_task_execution", name="Task Execution")
func dataSourceTaskExecution() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTaskExecutionRead,

		Schema: map[string]*schema.Schema{
			names.AttrCapacityProviderStrategy: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
						},
						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrWeight: {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},
					},
				},
			},
			"client_token": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cluster": {
				Type:     schema.TypeString,
				Required: true,
			},
			"desired_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"enable_ecs_managed_tags": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enable_execute_command": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"group": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"launch_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.LaunchType](),
			},
			names.AttrNetworkConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"overrides": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_overrides": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"command": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"cpu": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrEnvironment: {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrKey: {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrValue: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"memory": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"memory_reservation": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
									},
									"resource_requirements": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.ResourceType](),
												},
												names.AttrValue: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"cpu": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrExecutionRoleARN: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"inference_accelerator_overrides": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDeviceName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"device_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"memory": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"task_role_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"placement_constraints": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrExpression: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PlacementConstraintType](),
						},
					},
				},
			},
			"placement_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrField: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrPropagateTags: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PropagateTags](),
			},
			"reference_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"started_by": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags: tftags.TagsSchema(),
			"task_arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTaskExecutionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	cluster := d.Get("cluster").(string)
	taskDefinition := d.Get("task_definition").(string)
	id := strings.Join([]string{cluster, taskDefinition}, ",")
	input := &ecs.RunTaskInput{
		Cluster:        aws.String(cluster),
		TaskDefinition: aws.String(taskDefinition),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})))
	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk(names.AttrCapacityProviderStrategy); ok {
		input.CapacityProviderStrategy = expandCapacityProviderStrategyItems(v.(*schema.Set))
	}
	if v, ok := d.GetOk("client_token"); ok {
		input.ClientToken = aws.String(v.(string))
	}
	if v, ok := d.GetOk("desired_count"); ok {
		input.Count = aws.Int32(int32(v.(int)))
	}
	if v, ok := d.GetOk("enable_ecs_managed_tags"); ok {
		input.EnableECSManagedTags = v.(bool)
	}
	if v, ok := d.GetOk("enable_execute_command"); ok {
		input.EnableExecuteCommand = v.(bool)
	}
	if v, ok := d.GetOk("group"); ok {
		input.Group = aws.String(v.(string))
	}
	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = awstypes.LaunchType(v.(string))
	}
	if v, ok := d.GetOk(names.AttrNetworkConfiguration); ok {
		input.NetworkConfiguration = expandNetworkConfiguration(v.([]interface{}))
	}
	if v, ok := d.GetOk("overrides"); ok {
		input.Overrides = expandTaskOverride(v.([]interface{}))
	}
	if v, ok := d.GetOk("placement_constraints"); ok {
		apiObject, err := expandPlacementConstraints(v.(*schema.Set).List())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.PlacementConstraints = apiObject
	}
	if v, ok := d.GetOk("placement_strategy"); ok {
		apiObject, err := expandPlacementStrategy(v.([]interface{}))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.PlacementStrategy = apiObject
	}
	if v, ok := d.GetOk("platform_version"); ok {
		input.PlatformVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk(names.AttrPropagateTags); ok {
		input.PropagateTags = awstypes.PropagateTags(v.(string))
	}
	if v, ok := d.GetOk("reference_id"); ok {
		input.ReferenceId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("started_by"); ok {
		input.StartedBy = aws.String(v.(string))
	}

	output, err := conn.RunTask(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "running ECS Task (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("task_arns", tfslices.ApplyToAll(output.Tasks, func(v awstypes.Task) string {
		return aws.ToString(v.TaskArn)
	}))

	return diags
}

func expandTaskOverride(tfList []interface{}) *awstypes.TaskOverride {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &awstypes.TaskOverride{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["cpu"]; ok {
		apiObject.Cpu = aws.String(v.(string))
	}
	if v, ok := tfMap["memory"]; ok {
		apiObject.Memory = aws.String(v.(string))
	}
	if v, ok := tfMap[names.AttrExecutionRoleARN]; ok {
		apiObject.ExecutionRoleArn = aws.String(v.(string))
	}
	if v, ok := tfMap["task_role_arn"]; ok {
		apiObject.TaskRoleArn = aws.String(v.(string))
	}
	if v, ok := tfMap["inference_accelerator_overrides"]; ok {
		apiObject.InferenceAcceleratorOverrides = expandInferenceAcceleratorOverrides(v.(*schema.Set))
	}
	if v, ok := tfMap["container_overrides"]; ok {
		apiObject.ContainerOverrides = expandContainerOverride(v.([]interface{}))
	}

	return apiObject
}

func expandInferenceAcceleratorOverrides(tfSet *schema.Set) []awstypes.InferenceAcceleratorOverride {
	if tfSet.Len() == 0 {
		return nil
	}
	apiObject := make([]awstypes.InferenceAcceleratorOverride, 0)

	for _, item := range tfSet.List() {
		tfMap := item.(map[string]interface{})
		iao := awstypes.InferenceAcceleratorOverride{
			DeviceName: aws.String(tfMap[names.AttrDeviceName].(string)),
			DeviceType: aws.String(tfMap["device_type"].(string)),
		}
		apiObject = append(apiObject, iao)
	}

	return apiObject
}

func expandContainerOverride(tfList []interface{}) []awstypes.ContainerOverride {
	if len(tfList) == 0 {
		return nil
	}
	apiObject := make([]awstypes.ContainerOverride, 0)

	for _, item := range tfList {
		tfMap := item.(map[string]interface{})
		co := awstypes.ContainerOverride{
			Name: aws.String(tfMap[names.AttrName].(string)),
		}
		if v, ok := tfMap["command"]; ok {
			commandStrings := v.([]interface{})
			co.Command = flex.ExpandStringValueList(commandStrings)
		}
		if v, ok := tfMap["cpu"]; ok {
			co.Cpu = aws.Int32(int32(v.(int)))
		}
		if v, ok := tfMap[names.AttrEnvironment]; ok {
			co.Environment = expandTaskEnvironment(v.(*schema.Set))
		}
		if v, ok := tfMap["memory"]; ok {
			co.Memory = aws.Int32(int32(v.(int)))
		}
		if v, ok := tfMap["memory_reservation"]; ok {
			co.MemoryReservation = aws.Int32(int32(v.(int)))
		}
		if v, ok := tfMap["resource_requirements"]; ok {
			co.ResourceRequirements = expandResourceRequirements(v.(*schema.Set))
		}
		apiObject = append(apiObject, co)
	}

	return apiObject
}

func expandTaskEnvironment(tfSet *schema.Set) []awstypes.KeyValuePair {
	if tfSet.Len() == 0 {
		return nil
	}
	apiObject := make([]awstypes.KeyValuePair, 0)

	for _, item := range tfSet.List() {
		tfMap := item.(map[string]interface{})
		te := awstypes.KeyValuePair{
			Name:  aws.String(tfMap[names.AttrKey].(string)),
			Value: aws.String(tfMap[names.AttrValue].(string)),
		}
		apiObject = append(apiObject, te)
	}

	return apiObject
}

func expandResourceRequirements(tfSet *schema.Set) []awstypes.ResourceRequirement {
	if tfSet.Len() == 0 {
		return nil
	}

	apiObject := make([]awstypes.ResourceRequirement, 0)
	for _, item := range tfSet.List() {
		tfMap := item.(map[string]interface{})
		rr := awstypes.ResourceRequirement{
			Type:  awstypes.ResourceType(tfMap[names.AttrType].(string)),
			Value: aws.String(tfMap[names.AttrValue].(string)),
		}
		apiObject = append(apiObject, rr)
	}

	return apiObject
}
