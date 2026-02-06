// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_arcregionswitch_plan", name="Plan")
// @Tags(identifierAttribute="arn")
// @Region(overrideDeprecated=true)
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types;awstypes;awstypes.Plan")
// @Testing(altRegionTfVars=true)
// @Testing(preIdentityVersion="6.30.0")
// @Testing(existsTakesT=false, destroyTakesT=false)
// @Testing(preCheck="testAccPreCheck")
func newResourcePlan(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePlan{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type resourcePlan struct {
	framework.ResourceWithModel[resourcePlanModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func arcRoutingControlConfigBlock(ctx context.Context) fwschema.Block {
	return fwschema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[arcRoutingControlConfigModel](ctx),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"cross_account_role": fwschema.StringAttribute{
					Optional: true,
				},
				names.AttrExternalID: fwschema.StringAttribute{
					Optional: true,
				},
				"timeout_minutes": fwschema.Int32Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]fwschema.Block{
				"region_and_routing_controls": fwschema.SetNestedBlock{
					CustomType: fwtypes.NewSetNestedObjectTypeOf[regionAndRoutingControlsModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							names.AttrRegion: fwschema.StringAttribute{
								Required: true,
							},
						},
						Blocks: map[string]fwschema.Block{
							"routing_control": fwschema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[routingControlModel](ctx),
								NestedObject: fwschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"routing_control_arn": fwschema.StringAttribute{
											CustomType: fwtypes.ARNType,
											Required:   true,
										},
										names.AttrState: fwschema.StringAttribute{
											CustomType: fwtypes.StringEnumType[awstypes.RoutingControlStateChange](),
											Required:   true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func customActionLambdaConfigBlock(ctx context.Context) fwschema.Block {
	return fwschema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[customActionLambdaConfigModel](ctx),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"region_to_run": fwschema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.RegionToRunIn](),
					Required:   true,
				},
				"retry_interval_minutes": fwschema.Float32Attribute{
					Required: true,
				},
				"timeout_minutes": fwschema.Int32Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]fwschema.Block{
				"lambda": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							names.AttrARN: fwschema.StringAttribute{
								Required: true,
							},
							"cross_account_role": fwschema.StringAttribute{
								Optional: true,
							},
							names.AttrExternalID: fwschema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
				"ungraceful": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[ungracefulModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							"behavior": fwschema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.LambdaUngracefulBehavior](),
								Required:   true,
							},
						},
					},
				},
			},
		},
	}
}

func documentDBConfigBlock(ctx context.Context) fwschema.Block {
	return fwschema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[documentDbConfigModel](ctx),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"behavior": fwschema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.DocumentDbDefaultBehavior](),
					Required:   true,
				},
				"database_cluster_arns": fwschema.ListAttribute{
					CustomType: fwtypes.ListOfARNType,
					Required:   true,
				},
				"global_cluster_identifier": fwschema.StringAttribute{
					Required: true,
				},
				"cross_account_role": fwschema.StringAttribute{
					Optional: true,
				},
				names.AttrExternalID: fwschema.StringAttribute{
					Optional: true,
				},
				"timeout_minutes": fwschema.Int32Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]fwschema.Block{
				"ungraceful": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[documentDbUngracefulModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							"ungraceful": fwschema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.DocumentDbUngracefulBehavior](),
								Required:   true,
							},
						},
					},
				},
			},
		},
	}
}

func ec2ASGCapacityIncreaseConfigBlock(ctx context.Context) fwschema.Block {
	return fwschema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ec2ASGCapacityIncreaseConfigModel](ctx),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"capacity_monitoring_approach": fwschema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.Ec2AsgCapacityMonitoringApproach](),
					Required:   true,
				},
				"target_percent": fwschema.Int64Attribute{
					Optional: true,
				},
				"timeout_minutes": fwschema.Int32Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]fwschema.Block{
				"asg": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[asgModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							names.AttrARN: fwschema.StringAttribute{
								Required: true,
							},
							"cross_account_role": fwschema.StringAttribute{
								Optional: true,
							},
							names.AttrExternalID: fwschema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
				"ungraceful": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[ec2UngracefulModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							"minimum_success_percentage": fwschema.Int64Attribute{
								Required: true,
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
			},
		},
	}
}

func ecsCapacityIncreaseConfigBlock(ctx context.Context) fwschema.Block {
	return fwschema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ecsCapacityIncreaseConfigModel](ctx),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"capacity_monitoring_approach": fwschema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.EcsCapacityMonitoringApproach](),
					Required:   true,
				},
				"target_percent": fwschema.Int64Attribute{
					Optional: true,
				},
				"timeout_minutes": fwschema.Int32Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]fwschema.Block{
				"service": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[serviceModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							"cluster_arn": fwschema.StringAttribute{
								Required: true,
							},
							"service_arn": fwschema.StringAttribute{
								Required: true,
							},
							"cross_account_role": fwschema.StringAttribute{
								Optional: true,
							},
							names.AttrExternalID: fwschema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
				"ungraceful": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[ecsUngracefulModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							"minimum_success_percentage": fwschema.Int64Attribute{
								Required: true,
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
			},
		},
	}
}

func eksResourceScalingConfigBlock(ctx context.Context) fwschema.Block {
	return fwschema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[eksResourceScalingConfigModel](ctx),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"capacity_monitoring_approach": fwschema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.EksCapacityMonitoringApproach](),
					Required:   true,
				},
				"target_percent": fwschema.Int64Attribute{
					Required: true,
				},
				"timeout_minutes": fwschema.Int32Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]fwschema.Block{
				"kubernetes_resource_type": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[kubernetesResourceTypeModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							"api_version": fwschema.StringAttribute{
								Required: true,
							},
							"kind": fwschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
				"eks_clusters": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[eksClusterModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							"cluster_arn": fwschema.StringAttribute{
								Required: true,
							},
							"cross_account_role": fwschema.StringAttribute{
								Optional: true,
							},
							names.AttrExternalID: fwschema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
				"scaling_resources": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[scalingResourcesModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							names.AttrNamespace: fwschema.StringAttribute{
								Required: true,
							},
						},
						Blocks: map[string]fwschema.Block{
							names.AttrResources: fwschema.SetNestedBlock{
								CustomType: fwtypes.NewSetNestedObjectTypeOf[kubernetesScalingResourceModel](ctx),
								NestedObject: fwschema.NestedBlockObject{
									Attributes: map[string]fwschema.Attribute{
										"resource_name": fwschema.StringAttribute{
											Required: true,
										},
										names.AttrName: fwschema.StringAttribute{
											Required: true,
										},
										names.AttrNamespace: fwschema.StringAttribute{
											Required: true,
										},
										"hpa_name": fwschema.StringAttribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
				"ungraceful": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[eksUngracefulModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							"minimum_success_percentage": fwschema.Int64Attribute{
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

func executionApprovalConfigBlock(ctx context.Context) fwschema.Block {
	return fwschema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[executionApprovalConfigModel](ctx),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"approval_role": fwschema.StringAttribute{
					Required: true,
				},
				"timeout_minutes": fwschema.Int32Attribute{
					Optional: true,
				},
			},
		},
	}
}

func globalAuroraConfigBlock(ctx context.Context) fwschema.Block {
	return fwschema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[globalAuroraConfigModel](ctx),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"behavior": fwschema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.GlobalAuroraDefaultBehavior](),
					Required:   true,
				},
				"global_cluster_identifier": fwschema.StringAttribute{
					Required: true,
				},
				"database_cluster_arns": fwschema.ListAttribute{
					CustomType: fwtypes.ListOfARNType,
					Required:   true,
				},
				"cross_account_role": fwschema.StringAttribute{
					Optional: true,
				},
				names.AttrExternalID: fwschema.StringAttribute{
					Optional: true,
				},
				"timeout_minutes": fwschema.Int32Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]fwschema.Block{
				"ungraceful": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[globalAuroraUngracefulModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							"ungraceful": fwschema.StringAttribute{
								CustomType: fwtypes.StringEnumType[awstypes.GlobalAuroraUngracefulBehavior](),
								Required:   true,
							},
						},
					},
				},
			},
		},
	}
}

func route53HealthCheckConfigBlock(ctx context.Context) fwschema.Block {
	return fwschema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[route53HealthCheckConfigModel](ctx),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				names.AttrHostedZoneID: fwschema.StringAttribute{
					Required: true,
				},
				"record_name": fwschema.StringAttribute{
					Required: true,
				},
				"cross_account_role": fwschema.StringAttribute{
					Optional: true,
				},
				names.AttrExternalID: fwschema.StringAttribute{
					Optional: true,
				},
				"timeout_minutes": fwschema.Int32Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]fwschema.Block{
				"record_set": fwschema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[recordSetModel](ctx),
					NestedObject: fwschema.NestedBlockObject{
						Attributes: map[string]fwschema.Attribute{
							"record_set_identifier": fwschema.StringAttribute{
								Required: true,
							},
							names.AttrRegion: fwschema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

func regionSwitchPlanConfigBlock(ctx context.Context) fwschema.Block {
	return fwschema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[regionSwitchPlanConfigModel](ctx),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				names.AttrARN: fwschema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Required:   true,
				},
				"cross_account_role": fwschema.StringAttribute{
					Optional: true,
				},
				names.AttrExternalID: fwschema.StringAttribute{
					Optional: true,
				},
			},
		},
	}
}

func (r *resourcePlan) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"execution_role": fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"recovery_approach": fwschema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RecoveryApproach](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"regions": fwschema.ListAttribute{
				Required:   true,
				CustomType: fwtypes.ListOfStringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
			},
			"primary_region": fwschema.StringAttribute{
				Optional: true,
			},
			"recovery_time_objective_minutes": fwschema.Int64Attribute{
				Optional: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{
			"associated_alarms": fwschema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[associatedAlarmModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{ // nosemgrep:ci.semgrep.framework.map_block_key-meaningful-names
						"map_block_key": fwschema.StringAttribute{
							Required: true,
						},
						"alarm_type": fwschema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AlarmType](),
							Required:   true,
						},
						"resource_identifier": fwschema.StringAttribute{
							Required: true,
						},
						"cross_account_role": fwschema.StringAttribute{
							Optional: true,
						},
						names.AttrExternalID: fwschema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			names.AttrTriggers: fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[triggerModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						names.AttrAction: fwschema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.WorkflowTargetAction](),
							Required:   true,
						},
						names.AttrDescription: fwschema.StringAttribute{
							Optional: true,
						},
						"min_delay_minutes_between_executions": fwschema.Int64Attribute{
							Required: true,
						},
						"target_region": fwschema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"conditions": fwschema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[conditionModel](ctx),
							NestedObject: fwschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"associated_alarm_name": fwschema.StringAttribute{
										Required: true,
									},
									names.AttrCondition: fwschema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.AlarmCondition](),
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			"workflow": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[workflowModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"workflow_description": fwschema.StringAttribute{
							Optional: true,
						},
						"workflow_target_action": fwschema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.WorkflowTargetAction](),
							Required:   true,
						},
						"workflow_target_region": fwschema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"step": fwschema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[stepModel](ctx),
							NestedObject: fwschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									names.AttrDescription: fwschema.StringAttribute{
										Optional: true,
									},
									"execution_block_type": fwschema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.ExecutionBlockType](),
										Required:   true,
									},
									names.AttrName: fwschema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]fwschema.Block{
									"arc_routing_control_config":       arcRoutingControlConfigBlock(ctx),
									"custom_action_lambda_config":      customActionLambdaConfigBlock(ctx),
									"document_db_config":               documentDBConfigBlock(ctx),
									"ec2_asg_capacity_increase_config": ec2ASGCapacityIncreaseConfigBlock(ctx),
									"ecs_capacity_increase_config":     ecsCapacityIncreaseConfigBlock(ctx),
									"eks_resource_scaling_config":      eksResourceScalingConfigBlock(ctx),
									"execution_approval_config":        executionApprovalConfigBlock(ctx),
									"global_aurora_config":             globalAuroraConfigBlock(ctx),
									"region_switch_plan_config":        regionSwitchPlanConfigBlock(ctx),
									"parallel_config": fwschema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[parallelConfigModel](ctx),
										NestedObject: fwschema.NestedBlockObject{
											Blocks: map[string]fwschema.Block{
												"step": fwschema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[parallelStepModel](ctx),
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															names.AttrName: fwschema.StringAttribute{
																Required: true,
															},
															"execution_block_type": fwschema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.ExecutionBlockType](),
																Required:   true,
															},
															names.AttrDescription: fwschema.StringAttribute{
																Optional: true,
															},
														},
														Blocks: map[string]fwschema.Block{
															"arc_routing_control_config":       arcRoutingControlConfigBlock(ctx),
															"custom_action_lambda_config":      customActionLambdaConfigBlock(ctx),
															"document_db_config":               documentDBConfigBlock(ctx),
															"ec2_asg_capacity_increase_config": ec2ASGCapacityIncreaseConfigBlock(ctx),
															"ecs_capacity_increase_config":     ecsCapacityIncreaseConfigBlock(ctx),
															"eks_resource_scaling_config":      eksResourceScalingConfigBlock(ctx),
															"execution_approval_config":        executionApprovalConfigBlock(ctx),
															"global_aurora_config":             globalAuroraConfigBlock(ctx),
															"region_switch_plan_config":        regionSwitchPlanConfigBlock(ctx),
															"route53_health_check_config":      route53HealthCheckConfigBlock(ctx),
														},
													},
												},
											},
										},
									},
									"route53_health_check_config": route53HealthCheckConfigBlock(ctx),
								}, // blocks
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourcePlan) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourcePlanModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	var input arcregionswitch.CreatePlanInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle tags - use getTagsIn to get all tags including provider defaults
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreatePlan(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	plan.ARN = types.StringValue(aws.ToString(output.Plan.Arn))

	// Wait for plan to be available (eventual consistency)
	planOutput, err := waitPlanCreated(ctx, conn, plan.ARN.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, planOutput, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, sortWorkflows(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ARN = types.StringValue(aws.ToString(planOutput.Arn))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourcePlan) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourcePlanModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	plan, err := findPlanByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, plan, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, sortWorkflows(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ARN explicitly
	state.ARN = types.StringValue(aws.ToString(plan.Arn))

	// Handle tags
	tags, err := listTags(ctx, conn, aws.ToString(plan.Arn))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
	setTagsOut(ctx, tags.Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *resourcePlan) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourcePlanModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	var createInput arcregionswitch.CreatePlanInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &createInput))
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert CreatePlanInput to UpdatePlanInput (only updatable fields)
	var input arcregionswitch.UpdatePlanInput
	input.Arn = state.ARN.ValueStringPointer()
	input.ExecutionRole = createInput.ExecutionRole
	input.Description = createInput.Description
	input.RecoveryTimeObjectiveMinutes = createInput.RecoveryTimeObjectiveMinutes
	input.Workflows = createInput.Workflows
	input.AssociatedAlarms = createInput.AssociatedAlarms
	input.Triggers = createInput.Triggers

	_, err := conn.UpdatePlan(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
		return
	}

	// Handle tags update
	if !plan.TagsAll.Equal(state.TagsAll) {
		if err := updateTags(ctx, conn, plan.ARN.ValueString(), state.TagsAll, plan.TagsAll); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
			return
		}
	}

	// Read after update to refresh state
	planOutput, err := findPlanByARN(ctx, conn, plan.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, planOutput, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, sortWorkflows(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ARN = types.StringValue(aws.ToString(planOutput.Arn))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourcePlan) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourcePlanModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	input := arcregionswitch.DeletePlanInput{
		Arn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeletePlan(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		// Retry if health check allocation is in progress (check error message generically)
		if errs.Contains(err, "health check allocation is in progress") {
			_, err = waitPlanDeletable(ctx, conn, state.ARN.ValueString(), r.DeleteTimeout(ctx, state.Timeouts))
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
				return
			}
			// Retry delete
			_, err = conn.DeletePlan(ctx, &input)
			if err != nil {
				if errors.As(err, &nfe) {
					return
				}
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
				return
			}
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
}

func findPlanByARN(ctx context.Context, conn *arcregionswitch.Client, arn string) (*awstypes.Plan, error) {
	input := arcregionswitch.GetPlanInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetPlan(ctx, &input)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.Plan == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.Plan, nil
}

func findRoute53HealthChecksByARN(ctx context.Context, conn *arcregionswitch.Client, planARN string) ([]awstypes.Route53HealthCheck, error) {
	input := arcregionswitch.ListRoute53HealthChecksInput{
		Arn: aws.String(planARN),
	}

	output, err := conn.ListRoute53HealthChecks(ctx, &input)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return output.HealthChecks, nil
}

// sortWorkflows sorts workflows by target action (activate before deactivate) for consistent ordering
func sortWorkflows(ctx context.Context, m *resourcePlanModel) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	if m.Workflows.IsNull() || m.Workflows.IsUnknown() {
		return diags
	}

	workflows, d := m.Workflows.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() || len(workflows) <= 1 {
		return diags
	}

	slices.SortFunc(workflows, func(a, b *workflowModel) int {
		return cmp.Compare(a.WorkflowTargetAction.ValueString(), b.WorkflowTargetAction.ValueString())
	})

	m.Workflows = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, workflows)
	return diags
}

type resourcePlanModel struct {
	framework.WithRegionModel
	ARN                          types.String                                         `tfsdk:"arn"`
	AssociatedAlarms             fwtypes.SetNestedObjectValueOf[associatedAlarmModel] `tfsdk:"associated_alarms"`
	Description                  types.String                                         `tfsdk:"description"`
	ExecutionRole                types.String                                         `tfsdk:"execution_role"`
	Name                         types.String                                         `tfsdk:"name"`
	PrimaryRegion                types.String                                         `tfsdk:"primary_region"`
	RecoveryApproach             fwtypes.StringEnum[awstypes.RecoveryApproach]        `tfsdk:"recovery_approach"`
	RecoveryTimeObjectiveMinutes types.Int64                                          `tfsdk:"recovery_time_objective_minutes"`
	Regions                      fwtypes.ListOfString                                 `tfsdk:"regions"`
	Tags                         tftags.Map                                           `tfsdk:"tags"`
	TagsAll                      tftags.Map                                           `tfsdk:"tags_all"`
	Timeouts                     timeouts.Value                                       `tfsdk:"timeouts"`
	Triggers                     fwtypes.ListNestedObjectValueOf[triggerModel]        `tfsdk:"triggers"`
	Workflows                    fwtypes.ListNestedObjectValueOf[workflowModel]       `tfsdk:"workflow"`
}

type associatedAlarmModel struct {
	AlarmType          fwtypes.StringEnum[awstypes.AlarmType] `tfsdk:"alarm_type"`
	CrossAccountRole   types.String                           `tfsdk:"cross_account_role"`
	ExternalID         types.String                           `tfsdk:"external_id"`
	MapBlockKey        types.String                           `tfsdk:"map_block_key"`
	ResourceIdentifier types.String                           `tfsdk:"resource_identifier"`
}

type workflowModel struct {
	Steps                fwtypes.ListNestedObjectValueOf[stepModel]        `tfsdk:"step"`
	WorkflowDescription  types.String                                      `tfsdk:"workflow_description"`
	WorkflowTargetAction fwtypes.StringEnum[awstypes.WorkflowTargetAction] `tfsdk:"workflow_target_action"`
	WorkflowTargetRegion types.String                                      `tfsdk:"workflow_target_region"`
}

type stepModel struct {
	ArcRoutingControlConfig      fwtypes.ListNestedObjectValueOf[arcRoutingControlConfigModel]      `tfsdk:"arc_routing_control_config" autoflex:"-"`
	CustomActionLambdaConfig     fwtypes.ListNestedObjectValueOf[customActionLambdaConfigModel]     `tfsdk:"custom_action_lambda_config" autoflex:"-"`
	Description                  types.String                                                       `tfsdk:"description"`
	DocumentDbConfig             fwtypes.ListNestedObjectValueOf[documentDbConfigModel]             `tfsdk:"document_db_config" autoflex:"-"`
	EC2ASGCapacityIncreaseConfig fwtypes.ListNestedObjectValueOf[ec2ASGCapacityIncreaseConfigModel] `tfsdk:"ec2_asg_capacity_increase_config" autoflex:"-"`
	ECSCapacityIncreaseConfig    fwtypes.ListNestedObjectValueOf[ecsCapacityIncreaseConfigModel]    `tfsdk:"ecs_capacity_increase_config" autoflex:"-"`
	EKSResourceScalingConfig     fwtypes.ListNestedObjectValueOf[eksResourceScalingConfigModel]     `tfsdk:"eks_resource_scaling_config" autoflex:"-"`
	ExecutionApprovalConfig      fwtypes.ListNestedObjectValueOf[executionApprovalConfigModel]      `tfsdk:"execution_approval_config" autoflex:"-"`
	ExecutionBlockType           fwtypes.StringEnum[awstypes.ExecutionBlockType]                    `tfsdk:"execution_block_type"`
	GlobalAuroraConfig           fwtypes.ListNestedObjectValueOf[globalAuroraConfigModel]           `tfsdk:"global_aurora_config" autoflex:"-"`
	Name                         types.String                                                       `tfsdk:"name"`
	ParallelConfig               fwtypes.ListNestedObjectValueOf[parallelConfigModel]               `tfsdk:"parallel_config" autoflex:"-"`
	RegionSwitchPlanConfig       fwtypes.ListNestedObjectValueOf[regionSwitchPlanConfigModel]       `tfsdk:"region_switch_plan_config" autoflex:"-"`
	Route53HealthCheckConfig     fwtypes.ListNestedObjectValueOf[route53HealthCheckConfigModel]     `tfsdk:"route53_health_check_config" autoflex:"-"`
}

var (
	_ flex.Expander  = stepModel{}
	_ flex.Flattener = &stepModel{}
)

func (m stepModel) Expand(ctx context.Context) (any, fwdiag.Diagnostics) {
	var result awstypes.Step
	var diags fwdiag.Diagnostics

	// Expand basic step fields first
	diags.Append(flex.Expand(ctx, m.Name, &result.Name)...)
	diags.Append(flex.Expand(ctx, m.Description, &result.Description)...)
	diags.Append(flex.Expand(ctx, m.ExecutionBlockType, &result.ExecutionBlockType)...)
	if diags.HasError() {
		return nil, diags
	}

	// Handle ExecutionBlockConfiguration union type
	switch {
	case !m.ArcRoutingControlConfig.IsNull():
		config, d := m.ArcRoutingControlConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.CustomActionLambdaConfig.IsNull():
		config, d := m.CustomActionLambdaConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.DocumentDbConfig.IsNull():
		config, d := m.DocumentDbConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberDocumentDbConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.EC2ASGCapacityIncreaseConfig.IsNull():
		config, d := m.EC2ASGCapacityIncreaseConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.ECSCapacityIncreaseConfig.IsNull():
		config, d := m.ECSCapacityIncreaseConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.EKSResourceScalingConfig.IsNull():
		config, d := m.EKSResourceScalingConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.ExecutionApprovalConfig.IsNull():
		config, d := m.ExecutionApprovalConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.GlobalAuroraConfig.IsNull():
		config, d := m.GlobalAuroraConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.ParallelConfig.IsNull():
		config, d := m.ParallelConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberParallelConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.RegionSwitchPlanConfig.IsNull():
		config, d := m.RegionSwitchPlanConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberRegionSwitchPlanConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.Route53HealthCheckConfig.IsNull():
		config, d := m.Route53HealthCheckConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	}

	return &result, diags
}

func (m *stepModel) Flatten(ctx context.Context, v any) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	step, ok := v.(awstypes.Step)
	if !ok {
		diags.AddError("Unexpected Type", "Expected awstypes.Step")
		return diags
	}

	// Flatten basic step fields
	diags.Append(flex.Flatten(ctx, step.Name, &m.Name)...)
	diags.Append(flex.Flatten(ctx, step.Description, &m.Description)...)
	diags.Append(flex.Flatten(ctx, step.ExecutionBlockType, &m.ExecutionBlockType)...)
	if diags.HasError() {
		return diags
	}

	// Handle ExecutionBlockConfiguration union type
	if step.ExecutionBlockConfiguration != nil {
		switch v := step.ExecutionBlockConfiguration.(type) {
		case *awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.ArcRoutingControlConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.CustomActionLambdaConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberDocumentDbConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.DocumentDbConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.EC2ASGCapacityIncreaseConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.ECSCapacityIncreaseConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.EKSResourceScalingConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.ExecutionApprovalConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.GlobalAuroraConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberParallelConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.ParallelConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberRegionSwitchPlanConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.RegionSwitchPlanConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.Route53HealthCheckConfig)...)
		}
	}

	return diags
}

// ARC Routing Control Configuration Models
type arcRoutingControlConfigModel struct {
	CrossAccountRole         types.String                                                  `tfsdk:"cross_account_role"`
	ExternalID               types.String                                                  `tfsdk:"external_id"`
	RegionAndRoutingControls fwtypes.SetNestedObjectValueOf[regionAndRoutingControlsModel] `tfsdk:"region_and_routing_controls" autoflex:"-"`
	TimeoutMinutes           types.Int32                                                   `tfsdk:"timeout_minutes"`
}

var (
	_ flex.Expander  = arcRoutingControlConfigModel{}
	_ flex.Flattener = &arcRoutingControlConfigModel{}
)

func (m arcRoutingControlConfigModel) Expand(ctx context.Context) (any, fwdiag.Diagnostics) {
	var result awstypes.ArcRoutingControlConfiguration
	var diags fwdiag.Diagnostics

	diags.Append(flex.Expand(ctx, m.CrossAccountRole, &result.CrossAccountRole)...)
	diags.Append(flex.Expand(ctx, m.ExternalID, &result.ExternalId)...)
	diags.Append(flex.Expand(ctx, m.TimeoutMinutes, &result.TimeoutMinutes)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.RegionAndRoutingControls.IsNull() && !m.RegionAndRoutingControls.IsUnknown() {
		var regionControls []regionAndRoutingControlsModel
		diags.Append(m.RegionAndRoutingControls.ElementsAs(ctx, &regionControls, false)...)
		if diags.HasError() {
			return nil, diags
		}

		result.RegionAndRoutingControls = make(map[string][]awstypes.ArcRoutingControlState, len(regionControls))
		for _, rc := range regionControls {
			region := rc.Region.ValueString()

			if !rc.RoutingControls.IsNull() && !rc.RoutingControls.IsUnknown() {
				var controls []routingControlModel
				diags.Append(rc.RoutingControls.ElementsAs(ctx, &controls, false)...)
				if diags.HasError() {
					return nil, diags
				}

				states := make([]awstypes.ArcRoutingControlState, len(controls))
				for i, control := range controls {
					arn := control.RoutingControlArn.ValueString()
					states[i] = awstypes.ArcRoutingControlState{
						RoutingControlArn: &arn,
						State:             control.State.ValueEnum(),
					}
				}
				result.RegionAndRoutingControls[region] = states
			}
		}
	}

	return &result, diags
}

func (m *arcRoutingControlConfigModel) Flatten(ctx context.Context, v any) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	config, ok := v.(awstypes.ArcRoutingControlConfiguration)
	if !ok {
		diags.AddError("Unexpected Type", "Expected awstypes.ArcRoutingControlConfiguration")
		return diags
	}

	diags.Append(flex.Flatten(ctx, config.CrossAccountRole, &m.CrossAccountRole)...)
	diags.Append(flex.Flatten(ctx, config.ExternalId, &m.ExternalID)...)
	diags.Append(flex.Flatten(ctx, config.TimeoutMinutes, &m.TimeoutMinutes)...)
	if diags.HasError() {
		return diags
	}

	if len(config.RegionAndRoutingControls) > 0 {
		regionControls := make([]regionAndRoutingControlsModel, 0, len(config.RegionAndRoutingControls))
		for region, controlStates := range config.RegionAndRoutingControls {
			var regionModel regionAndRoutingControlsModel
			regionModel.Region = types.StringValue(region)

			controls := make([]routingControlModel, len(controlStates))
			for i, state := range controlStates {
				controls[i] = routingControlModel{
					RoutingControlArn: fwtypes.ARNValue(aws.ToString(state.RoutingControlArn)),
					State:             fwtypes.StringEnumValue(state.State),
				}
			}

			var d fwdiag.Diagnostics
			regionModel.RoutingControls, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, controls)
			diags.Append(d...)

			regionControls = append(regionControls, regionModel)
		}

		var d fwdiag.Diagnostics
		m.RegionAndRoutingControls, d = fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, regionControls)
		diags.Append(d...)
	}

	return diags
}

type regionAndRoutingControlsModel struct {
	Region          types.String                                         `tfsdk:"region"`
	RoutingControls fwtypes.ListNestedObjectValueOf[routingControlModel] `tfsdk:"routing_control"`
}

type routingControlModel struct {
	RoutingControlArn fwtypes.ARN                                            `tfsdk:"routing_control_arn"`
	State             fwtypes.StringEnum[awstypes.RoutingControlStateChange] `tfsdk:"state"`
}

type customActionLambdaConfigModel struct {
	Lambdas              fwtypes.ListNestedObjectValueOf[lambdaModel]     `tfsdk:"lambda"`
	RegionToRun          fwtypes.StringEnum[awstypes.RegionToRunIn]       `tfsdk:"region_to_run"`
	RetryIntervalMinutes types.Float32                                    `tfsdk:"retry_interval_minutes"`
	TimeoutMinutes       types.Int32                                      `tfsdk:"timeout_minutes"`
	Ungraceful           fwtypes.ListNestedObjectValueOf[ungracefulModel] `tfsdk:"ungraceful"`
}

type lambdaModel struct {
	ARN              types.String `tfsdk:"arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalID       types.String `tfsdk:"external_id"`
}

type ungracefulModel struct {
	Behavior fwtypes.StringEnum[awstypes.LambdaUngracefulBehavior] `tfsdk:"behavior"`
}

// DocumentDB Configuration Models
type documentDbConfigModel struct {
	Behavior                fwtypes.StringEnum[awstypes.DocumentDbDefaultBehavior]     `tfsdk:"behavior"`
	CrossAccountRole        types.String                                               `tfsdk:"cross_account_role"`
	DatabaseClusterArns     fwtypes.ListOfARN                                          `tfsdk:"database_cluster_arns"`
	ExternalID              types.String                                               `tfsdk:"external_id"`
	GlobalClusterIdentifier types.String                                               `tfsdk:"global_cluster_identifier"`
	TimeoutMinutes          types.Int32                                                `tfsdk:"timeout_minutes"`
	Ungraceful              fwtypes.ListNestedObjectValueOf[documentDbUngracefulModel] `tfsdk:"ungraceful"`
}

type documentDbUngracefulModel struct {
	Ungraceful fwtypes.StringEnum[awstypes.DocumentDbUngracefulBehavior] `tfsdk:"ungraceful"`
}

// EC2 ASG Configuration Models
type ec2ASGCapacityIncreaseConfigModel struct {
	ASGs                       fwtypes.ListNestedObjectValueOf[asgModel]                     `tfsdk:"asg"`
	CapacityMonitoringApproach fwtypes.StringEnum[awstypes.Ec2AsgCapacityMonitoringApproach] `tfsdk:"capacity_monitoring_approach"`
	TargetPercent              types.Int64                                                   `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int32                                                   `tfsdk:"timeout_minutes"`
	Ungraceful                 fwtypes.ListNestedObjectValueOf[ec2UngracefulModel]           `tfsdk:"ungraceful"`
}

type asgModel struct {
	ARN              types.String `tfsdk:"arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalID       types.String `tfsdk:"external_id"`
}

type ec2UngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

// ECS Configuration Models
type ecsCapacityIncreaseConfigModel struct {
	CapacityMonitoringApproach fwtypes.StringEnum[awstypes.EcsCapacityMonitoringApproach] `tfsdk:"capacity_monitoring_approach"`
	Services                   fwtypes.ListNestedObjectValueOf[serviceModel]              `tfsdk:"service"`
	TargetPercent              types.Int64                                                `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int32                                                `tfsdk:"timeout_minutes"`
	Ungraceful                 fwtypes.ListNestedObjectValueOf[ecsUngracefulModel]        `tfsdk:"ungraceful"`
}

type serviceModel struct {
	ClusterARN       types.String `tfsdk:"cluster_arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalID       types.String `tfsdk:"external_id"`
	ServiceARN       types.String `tfsdk:"service_arn"`
}

type ecsUngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

// EKS Configuration Models
type eksResourceScalingConfigModel struct {
	CapacityMonitoringApproach fwtypes.StringEnum[awstypes.EksCapacityMonitoringApproach]   `tfsdk:"capacity_monitoring_approach"`
	EKSClusters                fwtypes.ListNestedObjectValueOf[eksClusterModel]             `tfsdk:"eks_clusters"`
	KubernetesResourceType     fwtypes.ListNestedObjectValueOf[kubernetesResourceTypeModel] `tfsdk:"kubernetes_resource_type"`
	ScalingResources           fwtypes.ListNestedObjectValueOf[scalingResourcesModel]       `tfsdk:"scaling_resources" autoflex:"-"`
	TargetPercent              types.Int64                                                  `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int32                                                  `tfsdk:"timeout_minutes"`
	Ungraceful                 fwtypes.ListNestedObjectValueOf[eksUngracefulModel]          `tfsdk:"ungraceful"`
}

var (
	_ flex.Expander  = eksResourceScalingConfigModel{}
	_ flex.Flattener = &eksResourceScalingConfigModel{}
)

func (m eksResourceScalingConfigModel) Expand(ctx context.Context) (any, fwdiag.Diagnostics) {
	var result awstypes.EksResourceScalingConfiguration
	var diags fwdiag.Diagnostics

	diags.Append(flex.Expand(ctx, m.CapacityMonitoringApproach, &result.CapacityMonitoringApproach)...)
	diags.Append(flex.Expand(ctx, m.TargetPercent, &result.TargetPercent)...)
	diags.Append(flex.Expand(ctx, m.TimeoutMinutes, &result.TimeoutMinutes)...)
	diags.Append(flex.Expand(ctx, m.KubernetesResourceType, &result.KubernetesResourceType)...)
	diags.Append(flex.Expand(ctx, m.EKSClusters, &result.EksClusters)...)
	diags.Append(flex.Expand(ctx, m.Ungraceful, &result.Ungraceful)...)
	if diags.HasError() {
		return nil, diags
	}

	if !m.ScalingResources.IsNull() && !m.ScalingResources.IsUnknown() {
		var scalingResources []scalingResourcesModel
		diags.Append(m.ScalingResources.ElementsAs(ctx, &scalingResources, false)...)
		if diags.HasError() {
			return nil, diags
		}

		result.ScalingResources = make([]map[string]map[string]awstypes.KubernetesScalingResource, len(scalingResources))
		for k, sr := range scalingResources {
			namespaceMap := make(map[string]map[string]awstypes.KubernetesScalingResource)

			if !sr.Resources.IsNull() && !sr.Resources.IsUnknown() {
				var resources []kubernetesScalingResourceModel
				diags.Append(sr.Resources.ElementsAs(ctx, &resources, false)...)
				if diags.HasError() {
					return nil, diags
				}

				resourceMap := make(map[string]awstypes.KubernetesScalingResource)
				for _, res := range resources {
					resourceMap[res.ResourceName.ValueString()] = awstypes.KubernetesScalingResource{
						Name:      res.Name.ValueStringPointer(),
						Namespace: res.Namespace.ValueStringPointer(),
						HpaName:   res.HpaName.ValueStringPointer(),
					}
				}
				namespaceMap[sr.Namespace.ValueString()] = resourceMap
			}
			result.ScalingResources[k] = namespaceMap
		}
	}

	return &result, diags
}

func (m *eksResourceScalingConfigModel) Flatten(ctx context.Context, v any) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	config, ok := v.(awstypes.EksResourceScalingConfiguration)
	if !ok {
		diags.AddError("Unexpected Type", "Expected awstypes.EksResourceScalingConfiguration")
		return diags
	}

	diags.Append(flex.Flatten(ctx, config.CapacityMonitoringApproach, &m.CapacityMonitoringApproach)...)
	diags.Append(flex.Flatten(ctx, config.EksClusters, &m.EKSClusters)...)
	diags.Append(flex.Flatten(ctx, config.KubernetesResourceType, &m.KubernetesResourceType)...)
	diags.Append(flex.Flatten(ctx, config.TargetPercent, &m.TargetPercent)...)
	diags.Append(flex.Flatten(ctx, config.TimeoutMinutes, &m.TimeoutMinutes)...)
	diags.Append(flex.Flatten(ctx, config.Ungraceful, &m.Ungraceful)...)
	if diags.HasError() {
		return diags
	}

	if len(config.ScalingResources) > 0 {
		scalingResources := make([]scalingResourcesModel, len(config.ScalingResources))
		for i, sr := range config.ScalingResources {
			for namespace, resourceMap := range sr {
				scalingResources[i].Namespace = types.StringValue(namespace)

				resources := make([]kubernetesScalingResourceModel, 0, len(resourceMap))
				for resourceName, resource := range resourceMap {
					var resourceModel kubernetesScalingResourceModel
					resourceModel.ResourceName = types.StringValue(resourceName)
					diags.Append(flex.Flatten(ctx, resource.Name, &resourceModel.Name)...)
					diags.Append(flex.Flatten(ctx, resource.Namespace, &resourceModel.Namespace)...)
					diags.Append(flex.Flatten(ctx, resource.HpaName, &resourceModel.HpaName)...)
					resources = append(resources, resourceModel)
				}

				var d fwdiag.Diagnostics
				scalingResources[i].Resources, d = fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, resources)
				diags.Append(d...)
			}
		}

		var d fwdiag.Diagnostics
		m.ScalingResources, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, scalingResources)
		diags.Append(d...)
	}

	return diags
}

type kubernetesResourceTypeModel struct {
	ApiVersion types.String `tfsdk:"api_version"`
	Kind       types.String `tfsdk:"kind"`
}

type eksClusterModel struct {
	ClusterARN       types.String `tfsdk:"cluster_arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalID       types.String `tfsdk:"external_id"`
}

type scalingResourcesModel struct {
	Namespace types.String                                                   `tfsdk:"namespace"`
	Resources fwtypes.SetNestedObjectValueOf[kubernetesScalingResourceModel] `tfsdk:"resources"`
}

type kubernetesScalingResourceModel struct {
	HpaName      types.String `tfsdk:"hpa_name"`
	Name         types.String `tfsdk:"name"`
	Namespace    types.String `tfsdk:"namespace"`
	ResourceName types.String `tfsdk:"resource_name"`
}

type eksUngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

type executionApprovalConfigModel struct {
	ApprovalRole   types.String `tfsdk:"approval_role"`
	TimeoutMinutes types.Int32  `tfsdk:"timeout_minutes"`
}

// Global Aurora Configuration Models
type globalAuroraConfigModel struct {
	Behavior                fwtypes.StringEnum[awstypes.GlobalAuroraDefaultBehavior]     `tfsdk:"behavior"`
	CrossAccountRole        types.String                                                 `tfsdk:"cross_account_role"`
	DatabaseClusterARNs     fwtypes.ListOfARN                                            `tfsdk:"database_cluster_arns"`
	ExternalID              types.String                                                 `tfsdk:"external_id"`
	GlobalClusterIdentifier types.String                                                 `tfsdk:"global_cluster_identifier"`
	TimeoutMinutes          types.Int32                                                  `tfsdk:"timeout_minutes"`
	Ungraceful              fwtypes.ListNestedObjectValueOf[globalAuroraUngracefulModel] `tfsdk:"ungraceful"`
}

type globalAuroraUngracefulModel struct {
	Ungraceful fwtypes.StringEnum[awstypes.GlobalAuroraUngracefulBehavior] `tfsdk:"ungraceful"`
}

// Parallel Configuration Models
type parallelConfigModel struct {
	Step fwtypes.ListNestedObjectValueOf[parallelStepModel] `tfsdk:"step"`
}

type parallelStepModel struct {
	ArcRoutingControlConfig      fwtypes.ListNestedObjectValueOf[arcRoutingControlConfigModel]      `tfsdk:"arc_routing_control_config" autoflex:"-"`
	CustomActionLambdaConfig     fwtypes.ListNestedObjectValueOf[customActionLambdaConfigModel]     `tfsdk:"custom_action_lambda_config" autoflex:"-"`
	Description                  types.String                                                       `tfsdk:"description"`
	DocumentDbConfig             fwtypes.ListNestedObjectValueOf[documentDbConfigModel]             `tfsdk:"document_db_config" autoflex:"-"`
	EC2ASGCapacityIncreaseConfig fwtypes.ListNestedObjectValueOf[ec2ASGCapacityIncreaseConfigModel] `tfsdk:"ec2_asg_capacity_increase_config" autoflex:"-"`
	ECSCapacityIncreaseConfig    fwtypes.ListNestedObjectValueOf[ecsCapacityIncreaseConfigModel]    `tfsdk:"ecs_capacity_increase_config" autoflex:"-"`
	EKSResourceScalingConfig     fwtypes.ListNestedObjectValueOf[eksResourceScalingConfigModel]     `tfsdk:"eks_resource_scaling_config" autoflex:"-"`
	ExecutionApprovalConfig      fwtypes.ListNestedObjectValueOf[executionApprovalConfigModel]      `tfsdk:"execution_approval_config" autoflex:"-"`
	ExecutionBlockType           fwtypes.StringEnum[awstypes.ExecutionBlockType]                    `tfsdk:"execution_block_type"`
	GlobalAuroraConfig           fwtypes.ListNestedObjectValueOf[globalAuroraConfigModel]           `tfsdk:"global_aurora_config" autoflex:"-"`
	Name                         types.String                                                       `tfsdk:"name"`
	RegionSwitchPlanConfig       fwtypes.ListNestedObjectValueOf[regionSwitchPlanConfigModel]       `tfsdk:"region_switch_plan_config" autoflex:"-"`
	Route53HealthCheckConfig     fwtypes.ListNestedObjectValueOf[route53HealthCheckConfigModel]     `tfsdk:"route53_health_check_config" autoflex:"-"`
}

var (
	_ flex.Expander  = parallelStepModel{}
	_ flex.Flattener = &parallelStepModel{}
)

func (m parallelStepModel) Expand(ctx context.Context) (any, fwdiag.Diagnostics) {
	var result awstypes.Step
	var diags fwdiag.Diagnostics

	// Expand basic step fields first
	diags.Append(flex.Expand(ctx, m.Name, &result.Name)...)
	diags.Append(flex.Expand(ctx, m.Description, &result.Description)...)
	diags.Append(flex.Expand(ctx, m.ExecutionBlockType, &result.ExecutionBlockType)...)
	if diags.HasError() {
		return nil, diags
	}

	// Handle ExecutionBlockConfiguration union type
	switch {
	case !m.ArcRoutingControlConfig.IsNull():
		config, d := m.ArcRoutingControlConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.CustomActionLambdaConfig.IsNull():
		config, d := m.CustomActionLambdaConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.DocumentDbConfig.IsNull():
		config, d := m.DocumentDbConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberDocumentDbConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.EC2ASGCapacityIncreaseConfig.IsNull():
		config, d := m.EC2ASGCapacityIncreaseConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.ECSCapacityIncreaseConfig.IsNull():
		config, d := m.ECSCapacityIncreaseConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.EKSResourceScalingConfig.IsNull():
		config, d := m.EKSResourceScalingConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.ExecutionApprovalConfig.IsNull():
		config, d := m.ExecutionApprovalConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.GlobalAuroraConfig.IsNull():
		config, d := m.GlobalAuroraConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.RegionSwitchPlanConfig.IsNull():
		config, d := m.RegionSwitchPlanConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberRegionSwitchPlanConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	case !m.Route53HealthCheckConfig.IsNull():
		config, d := m.Route53HealthCheckConfig.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig
		diags.Append(flex.Expand(ctx, config, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		result.ExecutionBlockConfiguration = &r
	}

	return &result, diags
}

func (m *parallelStepModel) Flatten(ctx context.Context, v any) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	step, ok := v.(awstypes.Step)
	if !ok {
		diags.AddError("Unexpected Type", "Expected awstypes.Step")
		return diags
	}

	// Flatten basic step fields
	diags.Append(flex.Flatten(ctx, step.Name, &m.Name)...)
	diags.Append(flex.Flatten(ctx, step.Description, &m.Description)...)
	diags.Append(flex.Flatten(ctx, step.ExecutionBlockType, &m.ExecutionBlockType)...)
	if diags.HasError() {
		return diags
	}

	// Handle ExecutionBlockConfiguration union type
	if step.ExecutionBlockConfiguration != nil {
		switch v := step.ExecutionBlockConfiguration.(type) {
		case *awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.ArcRoutingControlConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.CustomActionLambdaConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberDocumentDbConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.DocumentDbConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.EC2ASGCapacityIncreaseConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.ECSCapacityIncreaseConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.EKSResourceScalingConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.ExecutionApprovalConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.GlobalAuroraConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberRegionSwitchPlanConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.RegionSwitchPlanConfig)...)
		case *awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig:
			diags.Append(flex.Flatten(ctx, &v.Value, &m.Route53HealthCheckConfig)...)
		}
	}

	return diags
}

type route53HealthCheckConfigModel struct {
	CrossAccountRole types.String                                    `tfsdk:"cross_account_role"`
	ExternalID       types.String                                    `tfsdk:"external_id"`
	HostedZoneID     types.String                                    `tfsdk:"hosted_zone_id"`
	RecordName       types.String                                    `tfsdk:"record_name"`
	RecordSets       fwtypes.ListNestedObjectValueOf[recordSetModel] `tfsdk:"record_set"`
	TimeoutMinutes   types.Int32                                     `tfsdk:"timeout_minutes"`
}

type recordSetModel struct {
	RecordSetIdentifier types.String `tfsdk:"record_set_identifier"`
	Region              types.String `tfsdk:"region"`
}

type regionSwitchPlanConfigModel struct {
	ARN              fwtypes.ARN  `tfsdk:"arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalID       types.String `tfsdk:"external_id"`
}

// Trigger Configuration Models
type triggerModel struct {
	Action                           fwtypes.StringEnum[awstypes.WorkflowTargetAction] `tfsdk:"action"`
	Conditions                       fwtypes.ListNestedObjectValueOf[conditionModel]   `tfsdk:"conditions"`
	Description                      types.String                                      `tfsdk:"description"`
	MinDelayMinutesBetweenExecutions types.Int64                                       `tfsdk:"min_delay_minutes_between_executions"`
	TargetRegion                     types.String                                      `tfsdk:"target_region"`
}

type conditionModel struct {
	AssociatedAlarmName types.String                                `tfsdk:"associated_alarm_name"`
	Condition           fwtypes.StringEnum[awstypes.AlarmCondition] `tfsdk:"condition"`
}

func waitPlanCreated(ctx context.Context, conn *arcregionswitch.Client, arn string, timeout time.Duration) (*awstypes.Plan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{"exists"},
		Refresh: statusPlan(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	plan, ok := outputRaw.(*awstypes.Plan)
	if !ok {
		return nil, nil // nosemgrep:ci.semgrep.smarterr.go-no-bare-return-err
	}

	// Check if plan has Route53HealthCheck steps
	hasRoute53HealthChecks := false
	expectedCount := 0
	for _, workflow := range plan.Workflows {
		for _, step := range workflow.Steps {
			if step.ExecutionBlockType == awstypes.ExecutionBlockTypeRoute53HealthCheck {
				hasRoute53HealthChecks = true
				expectedCount++
			}
		}
	}

	// If plan has Route53 health checks, wait for them to be allocated
	if hasRoute53HealthChecks {
		healthCheckConf := &retry.StateChangeConf{
			Pending: []string{"pending"},
			Target:  []string{"allocated"},
			Refresh: statusRoute53HealthChecks(ctx, conn, arn, expectedCount),
			Timeout: timeout,
		}

		_, err = healthCheckConf.WaitForStateContext(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}
	}

	return plan, nil
}

func statusRoute53HealthChecks(ctx context.Context, conn *arcregionswitch.Client, arn string, expectedCount int) retry.StateRefreshFunc {
	return func(_ context.Context) (any, string, error) {
		healthChecks, err := findRoute53HealthChecksByARN(ctx, conn, arn)
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		// Wait for expected number of health checks to exist
		if len(healthChecks) < expectedCount {
			return healthChecks, "pending", nil
		}

		// Wait for all health check IDs to be populated
		for _, hc := range healthChecks {
			if aws.ToString(hc.HealthCheckId) == "" {
				return healthChecks, "pending", nil
			}
		}

		// All health checks exist with IDs populated
		return healthChecks, "allocated", nil
	}
}

func statusPlan(ctx context.Context, conn *arcregionswitch.Client, arn string) retry.StateRefreshFunc {
	return func(_ context.Context) (any, string, error) {
		plan, err := findPlanByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return plan, "exists", nil
	}
}

func waitPlanDeletable(ctx context.Context, conn *arcregionswitch.Client, arn string, timeout time.Duration) (*awstypes.Plan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"health_check_allocation_in_progress"},
		Target:  []string{"deletable"},
		Refresh: statusPlanDeletable(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Plan); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusPlanDeletable(ctx context.Context, conn *arcregionswitch.Client, arn string) retry.StateRefreshFunc {
	return func(_ context.Context) (any, string, error) {
		plan, err := findPlanByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		// Try to delete to check if it's ready
		input := arcregionswitch.DeletePlanInput{
			Arn: &arn,
		}
		_, err = conn.DeletePlan(ctx, &input)
		if err == nil {
			// Delete succeeded, plan is gone
			return plan, "deletable", nil
		}

		if errs.Contains(err, "health check allocation is in progress") {
			// Still in progress
			return plan, "health_check_allocation_in_progress", nil
		}

		// Other error
		return nil, "", smarterr.NewError(err)
	}
}
