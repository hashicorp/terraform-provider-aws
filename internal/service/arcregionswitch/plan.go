package arcregionswitch

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// baseStepConfigSchemas returns the common execution block configuration schemas
func baseStepConfigSchemas() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// Execution Approval Config
		"execution_approval_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"approval_role": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: verify.ValidARN,
					},
					"timeout_minutes": {
						Type:     schema.TypeInt,
						Optional: true,
					},
				},
			},
		},
		// Custom Action Lambda Config
		"custom_action_lambda_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"region_to_run": {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							"activatingRegion", "deactivatingRegion",
						}, false),
					},
					"retry_interval_minutes": {
						Type:     schema.TypeFloat,
						Required: true,
					},
					"timeout_minutes": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"lambda": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"arn": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: verify.ValidARN,
								},
								"cross_account_role": {
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: verify.ValidARN,
								},
								"external_id": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
					"ungraceful": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"behavior": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice([]string{"skip"}, false),
								},
							},
						},
					},
				},
			},
		},
		// ARC Routing Control Config
		"arc_routing_control_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"region_and_routing_controls": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"region": {
									Type:     schema.TypeString,
									Required: true,
								},
								"routing_control_arns": {
									Type:     schema.TypeList,
									Required: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
							},
						},
					},
					"cross_account_role": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: verify.ValidARN,
					},
					"external_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"timeout_minutes": {
						Type:     schema.TypeInt,
						Optional: true,
					},
				},
			},
		},
		// EC2 ASG Capacity Increase Config
		"ec2_asg_capacity_increase_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"asgs": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"arn": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: verify.ValidARN,
								},
								"cross_account_role": {
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: verify.ValidARN,
								},
								"external_id": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
					"capacity_monitoring_approach": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							"sampledMaxInLast24Hours",
							"autoscalingMaxInLast24Hours",
						}, false),
					},
					"target_percent": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"timeout_minutes": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"ungraceful": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"minimum_success_percentage": {
									Type:     schema.TypeInt,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		// Global Aurora Config
		"global_aurora_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"behavior": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"switchoverOnly", "failover"}, false),
					},
					"global_cluster_identifier": {
						Type:     schema.TypeString,
						Required: true,
					},
					"database_cluster_arns": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"cross_account_role": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: verify.ValidARN,
					},
					"external_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"timeout_minutes": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"ungraceful": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"ungraceful": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice([]string{"failover"}, false),
								},
							},
						},
					},
				},
			},
		},
		// ECS Capacity Increase Config
		"ecs_capacity_increase_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"services": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"cluster_arn": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: verify.ValidARN,
								},
								"service_arn": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: verify.ValidARN,
								},
								"cross_account_role": {
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: verify.ValidARN,
								},
								"external_id": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
					"capacity_monitoring_approach": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							"sampledMaxInLast24Hours",
							"containerInsightsMaxInLast24Hours",
						}, false),
					},
					"target_percent": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"timeout_minutes": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"ungraceful": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"minimum_success_percentage": {
									Type:     schema.TypeInt,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		// EKS Resource Scaling Config
		"eks_resource_scaling_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"kubernetes_resource_type": {
						Type:     schema.TypeList,
						Required: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"api_version": {
									Type:     schema.TypeString,
									Required: true,
								},
								"kind": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					"eks_clusters": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"cluster_arn": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: verify.ValidARN,
								},
								"cross_account_role": {
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: verify.ValidARN,
								},
								"external_id": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
					"scaling_resources": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"namespace": {
									Type:     schema.TypeString,
									Required: true,
								},
								"resources": {
									Type:     schema.TypeSet,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"resource_name": {
												Type:     schema.TypeString,
												Required: true,
											},
											"name": {
												Type:     schema.TypeString,
												Required: true,
											},
											"namespace": {
												Type:     schema.TypeString,
												Required: true,
											},
											"hpa_name": {
												Type:     schema.TypeString,
												Optional: true,
											},
										},
									},
								},
							},
						},
					},
					"capacity_monitoring_approach": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							"sampledMaxInLast24Hours",
							"autoscalingMaxInLast24Hours",
						}, false),
					},
					"target_percent": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"timeout_minutes": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"ungraceful": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"minimum_success_percentage": {
									Type:     schema.TypeInt,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		// Route53 Health Check Config
		"route53_health_check_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"hosted_zone_id": {
						Type:     schema.TypeString,
						Required: true,
					},
					"record_name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"record_sets": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"record_set_identifier": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"region": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
					"cross_account_role": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: verify.ValidARN,
					},
					"external_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"timeout_minutes": {
						Type:     schema.TypeInt,
						Optional: true,
					},
				},
			},
		},
		// Region Switch Plan Config
		"region_switch_plan_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"arn": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: verify.ValidARN,
					},
					"cross_account_role": {
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: verify.ValidARN,
					},
					"external_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}
}

// stepSchema returns the complete schema for a step, including parallel config support
func stepSchema() map[string]*schema.Schema {
	stepSchemaMap := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"execution_block_type": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				"CustomActionLambda",
				"ManualApproval",
				"AuroraGlobalDatabase",
				"EC2AutoScaling",
				"ARCRoutingControl",
				"ARCRegionSwitchPlan",
				"Parallel",
				"ECSServiceScaling",
				"EKSResourceScaling",
				"Route53HealthCheck",
			}, false),
		},
		"description": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	// Add all base config schemas
	for k, v := range baseStepConfigSchemas() {
		stepSchemaMap[k] = v
	}

	// Add parallel config with a simplified step schema to avoid infinite recursion
	// The parallel config steps will support all the same configurations as regular steps
	stepSchemaMap["parallel_config"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"step": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Resource{
						Schema: parallelStepSchema(),
					},
				},
			},
		},
	}

	return stepSchemaMap
}

// parallelStepSchema returns the schema for steps within parallel configs
// This is identical to stepSchema but without the parallel_config to avoid infinite recursion
// Note: AWS SDK does support nested parallel configs, but we limit to one level for simplicity
func parallelStepSchema() map[string]*schema.Schema {
	stepSchemaMap := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"execution_block_type": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				"CustomActionLambda",
				"ManualApproval",
				"AuroraGlobalDatabase",
				"EC2AutoScaling",
				"ARCRoutingControl",
				"ARCRegionSwitchPlan",
				// "Parallel", // Removed to prevent infinite recursion
				"ECSServiceScaling",
				"EKSResourceScaling",
				"Route53HealthCheck",
			}, false),
		},
		"description": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	// Add all base config schemas
	for k, v := range baseStepConfigSchemas() {
		stepSchemaMap[k] = v
	}

	// Note: parallel_config is intentionally omitted to prevent infinite recursion
	// At the API level - it is also limited to a depth of 1.

	return stepSchemaMap
}

// @SDKResource("aws_arcregionswitch_plan", name="Plan")
// @Tags(identifierAttribute="arn")
func ResourcePlan() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePlanCreate,
		ReadWithoutTimeout:   resourcePlanRead,
		UpdateWithoutTimeout: resourcePlanUpdate,
		DeleteWithoutTimeout: resourcePlanDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"execution_role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"recovery_approach": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"activeActive", "activePassive"}, false),
			},
			"regions": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"workflow": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"workflow_target_action": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"activate", "deactivate"}, false),
						},
						"workflow_target_region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"workflow_description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"step": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: stepSchema(),
							},
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"primary_region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"recovery_time_objective_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"associated_alarms": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"alarm_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"applicationHealth", "trigger"}, false),
						},
						"resource_identifier": {
							Type:     schema.TypeString,
							Required: true,
						},
						"cross_account_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"external_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"trigger": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"activate", "deactivate"}, false),
						},
						"conditions": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"associated_alarm_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"condition": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"red", "green"}, false),
									},
								},
							},
						},
						"min_delay_minutes_between_executions": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"target_region": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ARCRegionSwitchClient(ctx)

	name := d.Get("name").(string)
	input := &arcregionswitch.CreatePlanInput{
		Name:             aws.String(name),
		ExecutionRole:    aws.String(d.Get("execution_role").(string)),
		RecoveryApproach: types.RecoveryApproach(d.Get("recovery_approach").(string)),
		Regions:          flex.ExpandStringValueList(d.Get("regions").([]interface{})),
		Workflows:        expandWorkflows(d.Get("workflow").([]interface{})),
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = tftags.New(ctx, v.(map[string]interface{})).Map()
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("primary_region"); ok {
		input.PrimaryRegion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("recovery_time_objective_minutes"); ok {
		input.RecoveryTimeObjectiveMinutes = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("associated_alarms"); ok {
		input.AssociatedAlarms = expandAssociatedAlarms(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("trigger"); ok {
		input.Triggers = expandTriggers(v.([]interface{}))
	}

	output, err := conn.CreatePlan(ctx, input)

	if err != nil {
		return diag.Errorf("creating ARC Region Switch Plan (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Plan.Arn))

	return resourcePlanRead(ctx, d, meta)
}

func resourcePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ARCRegionSwitchClient(ctx)

	plan, err := FindPlanByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading ARC Region Switch Plan (%s): %s", d.Id(), err)
	}

	d.Set("arn", plan.Arn)
	d.Set("name", plan.Name)
	d.Set("execution_role", plan.ExecutionRole)
	d.Set("recovery_approach", plan.RecoveryApproach)
	d.Set("regions", plan.Regions)

	if err := d.Set("workflow", flattenWorkflows(plan.Workflows)); err != nil {
		return diag.Errorf("setting workflow: %s", err)
	}

	d.Set("description", plan.Description)
	d.Set("primary_region", plan.PrimaryRegion)
	d.Set("recovery_time_objective_minutes", plan.RecoveryTimeObjectiveMinutes)

	if err := d.Set("associated_alarms", flattenAssociatedAlarms(plan.AssociatedAlarms)); err != nil {
		return diag.Errorf("setting associated_alarms: %s", err)
	}

	if err := d.Set("trigger", flattenTriggers(plan.Triggers)); err != nil {
		return diag.Errorf("setting trigger: %s", err)
	}

	d.Set("owner", plan.Owner)
	if plan.UpdatedAt != nil {
		d.Set("updated_at", plan.UpdatedAt.Format(time.RFC3339))
	}
	d.Set("version", plan.Version)

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("listing tags for ARC Region Switch Plan (%s): %s", d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	if err := d.Set("tags", tags.Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourcePlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ARCRegionSwitchClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &arcregionswitch.UpdatePlanInput{
			Arn:           aws.String(d.Id()),
			ExecutionRole: aws.String(d.Get("execution_role").(string)),
			Workflows:     expandWorkflows(d.Get("workflow").([]interface{})),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("recovery_time_objective_minutes") {
			input.RecoveryTimeObjectiveMinutes = aws.Int32(int32(d.Get("recovery_time_objective_minutes").(int)))
		}

		if d.HasChange("associated_alarms") {
			input.AssociatedAlarms = expandAssociatedAlarms(d.Get("associated_alarms").(*schema.Set).List())
		}

		if d.HasChange("trigger") {
			input.Triggers = expandTriggers(d.Get("trigger").([]interface{}))
		}

		_, err := conn.UpdatePlan(ctx, input)

		if err != nil {
			return diag.Errorf("updating ARC Region Switch Plan (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating ARC Region Switch Plan (%s) tags: %s", d.Id(), err)
		}
	}

	return resourcePlanRead(ctx, d, meta)
}

func resourcePlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ARCRegionSwitchClient(ctx)

	_, err := conn.DeletePlan(ctx, &arcregionswitch.DeletePlanInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting ARC Region Switch Plan (%s): %s", d.Id(), err)
	}

	return nil
}

func FindPlanByARN(ctx context.Context, conn *arcregionswitch.Client, arn string) (*types.Plan, error) {
	input := &arcregionswitch.GetPlanInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetPlan(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.Plan == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Plan, nil
}

// Tags returns arcregionswitch service tags.
func Tags(tags tftags.KeyValueTags) map[string]string {
	return tags.Map()
}

// ListTags lists arcregionswitch service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(ctx context.Context, conn *arcregionswitch.Client, identifier string) (tftags.KeyValueTags, error) {
	input := &arcregionswitch.ListTagsForResourceInput{
		Arn: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(ctx, input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return tftags.New(ctx, output.ResourceTags), nil
}

// UpdateTags updates arcregionswitch service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(ctx context.Context, conn *arcregionswitch.Client, identifier string, oldTagsMap, newTagsMap interface{}) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &arcregionswitch.UntagResourceInput{
			Arn:             aws.String(identifier),
			ResourceTagKeys: removedTags.Keys(),
		}

		_, err := conn.UntagResource(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &arcregionswitch.TagResourceInput{
			Arn:  aws.String(identifier),
			Tags: Tags(updatedTags),
		}

		_, err := conn.TagResource(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
