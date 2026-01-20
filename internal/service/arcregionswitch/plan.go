// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_arcregionswitch_plan", name="Plan")
// @Tags(identifierAttribute="arn")
func newResourcePlan(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePlan{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type resourcePlan struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourcePlan) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"execution_role": fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
			"recovery_approach": fwschema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RecoveryApproach](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("activeActive", "activePassive"),
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
							Validators: []validator.String{
								stringvalidator.OneOf("applicationHealth", "trigger"),
							},
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
			"workflow": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[workflowModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"workflow_target_action": fwschema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.WorkflowTargetAction](),
							Required:   true,
							Validators: []validator.String{
								stringvalidator.OneOf("activate", "deactivate"),
							},
						},
						"workflow_target_region": fwschema.StringAttribute{
							Optional: true,
						},
						"workflow_description": fwschema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"step": fwschema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[stepModel](ctx),
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
									"execution_block_configuration": fwschema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[executionBlockConfigurationModel](ctx),
										NestedObject: fwschema.NestedBlockObject{
											Blocks: map[string]fwschema.Block{
												"execution_approval_config": fwschema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[executionApprovalConfigModel](ctx),
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"approval_role": fwschema.StringAttribute{
																Required: true,
															},
															"timeout_minutes": fwschema.Int64Attribute{
																Optional: true,
															},
														},
													},
												},
												"route53_health_check_config": fwschema.ListNestedBlock{
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
															"timeout_minutes": fwschema.Int64Attribute{
																Optional: true,
															},
														},
														Blocks: map[string]fwschema.Block{
															"record_sets": fwschema.ListNestedBlock{
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
												},
												"custom_action_lambda_config": fwschema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[customActionLambdaConfigModel](ctx),
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"region_to_run": fwschema.StringAttribute{
																Required: true,
															},
															"retry_interval_minutes": fwschema.Float64Attribute{
																Required: true,
															},
															"timeout_minutes": fwschema.Int64Attribute{
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
												},
												"global_aurora_config": fwschema.ListNestedBlock{
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
															"timeout_minutes": fwschema.Int64Attribute{
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
												},
												"ec2_asg_capacity_increase_config": fwschema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[ec2AsgCapacityIncreaseConfigModel](ctx),
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"capacity_monitoring_approach": fwschema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.Ec2AsgCapacityMonitoringApproach](),
																Required:   true,
															},
															"target_percent": fwschema.Int64Attribute{
																Required: true,
															},
															"timeout_minutes": fwschema.Int64Attribute{
																Optional: true,
															},
														},
														Blocks: map[string]fwschema.Block{
															"asgs": fwschema.ListNestedBlock{
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
															},
														},
													},
												},
												"ecs_capacity_increase_config": fwschema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[ecsCapacityIncreaseConfigModel](ctx),
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"capacity_monitoring_approach": fwschema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.EcsCapacityMonitoringApproach](),
																Required:   true,
															},
															"target_percent": fwschema.Int64Attribute{
																Required: true,
															},
															"timeout_minutes": fwschema.Int64Attribute{
																Optional: true,
															},
														},
														Blocks: map[string]fwschema.Block{
															"services": fwschema.ListNestedBlock{
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
															},
														},
													},
												},
												"eks_resource_scaling_config": fwschema.ListNestedBlock{
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
															"timeout_minutes": fwschema.Int64Attribute{
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
												},
												"arc_routing_control_config": fwschema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[arcRoutingControlConfigModel](ctx),
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"cross_account_role": fwschema.StringAttribute{
																Optional: true,
															},
															names.AttrExternalID: fwschema.StringAttribute{
																Optional: true,
															},
															"timeout_minutes": fwschema.Int64Attribute{
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
																		"routing_control_arns": fwschema.ListAttribute{
																			CustomType: fwtypes.ListOfARNType,
																			Required:   true,
																		},
																	},
																},
															},
														},
													},
												},
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
																		"execution_approval_config": fwschema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[executionApprovalConfigModel](ctx),
																			NestedObject: fwschema.NestedBlockObject{
																				Attributes: map[string]fwschema.Attribute{
																					"approval_role": fwschema.StringAttribute{
																						Required: true,
																					},
																					"timeout_minutes": fwschema.Int64Attribute{
																						Optional: true,
																					},
																				},
																			},
																		},
																		"custom_action_lambda_config": fwschema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[customActionLambdaConfigModel](ctx),
																			NestedObject: fwschema.NestedBlockObject{
																				Attributes: map[string]fwschema.Attribute{
																					"region_to_run": fwschema.StringAttribute{
																						Required: true,
																					},
																					"retry_interval_minutes": fwschema.Float64Attribute{
																						Required: true,
																					},
																					"timeout_minutes": fwschema.Int64Attribute{
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
							},
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
							Validators: []validator.String{
								stringvalidator.OneOf("activate", "deactivate"),
							},
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
										Validators: []validator.String{
											stringvalidator.OneOf("red", "green"),
										},
									},
								},
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
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	// Use custom Expand method for resourcePlanModel
	expanded, diags := plan.Expand(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	input := expanded.(arcregionswitch.CreatePlanInput)

	// Handle tags - use getTagsIn to get all tags including provider defaults
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreatePlan(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("creating ARC Region Switch Plan", err.Error())
		return
	}

	plan.ARN = types.StringValue(aws.ToString(output.Plan.Arn))
	plan.ID = types.StringValue(aws.ToString(output.Plan.Arn))

	// Wait for plan to be available (eventual consistency)
	planOutput, err := waitPlanCreated(ctx, conn, plan.ID.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError("waiting for ARC Region Switch Plan create", err.Error())
		return
	}

	diags = plan.Flatten(ctx, planOutput)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	plan.ID = types.StringValue(aws.ToString(planOutput.Arn))
	plan.ARN = types.StringValue(aws.ToString(planOutput.Arn))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePlan) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourcePlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	plan, err := findPlanByARN(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("reading ARC Region Switch Plan", err.Error())
		return
	}

	// Use custom Flatten method for resourcePlanModel
	diags := state.Flatten(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Set ID and ARN explicitly since they might need special handling
	state.ID = types.StringValue(aws.ToString(plan.Arn))
	state.ARN = types.StringValue(aws.ToString(plan.Arn))

	// Handle tags
	tags, err := listTags(ctx, conn, aws.ToString(plan.Arn))
	if err != nil {
		resp.Diagnostics.AddError("listing tags for ARC Region Switch Plan", err.Error())
		return
	}
	setTagsOut(ctx, tags.Map())

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourcePlan) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourcePlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	// Use custom expand logic (similar to Create)
	apiObject, diags := plan.Expand(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	createInput := apiObject.(arcregionswitch.CreatePlanInput)

	// Convert CreatePlanInput to UpdatePlanInput (only updatable fields)
	var input arcregionswitch.UpdatePlanInput
	input.Arn = state.ID.ValueStringPointer()
	input.ExecutionRole = createInput.ExecutionRole
	input.Description = createInput.Description
	input.RecoveryTimeObjectiveMinutes = createInput.RecoveryTimeObjectiveMinutes
	input.Workflows = createInput.Workflows
	input.AssociatedAlarms = createInput.AssociatedAlarms
	input.Triggers = createInput.Triggers

	_, err := conn.UpdatePlan(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("updating ARC Region Switch Plan", err.Error())
		return
	}

	// Handle tags update
	if !plan.TagsAll.Equal(state.TagsAll) {
		if err := updateTags(ctx, conn, plan.ID.ValueString(), state.TagsAll, plan.TagsAll); err != nil {
			resp.Diagnostics.AddError("updating tags", err.Error())
			return
		}
	}

	// Read after update to refresh state
	planOutput, err := findPlanByARN(ctx, conn, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("reading ARC Region Switch Plan after update", err.Error())
		return
	}

	diags = plan.Flatten(ctx, planOutput)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	plan.ID = types.StringValue(aws.ToString(planOutput.Arn))
	plan.ARN = types.StringValue(aws.ToString(planOutput.Arn))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePlan) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourcePlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	input := arcregionswitch.DeletePlanInput{
		Arn: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeletePlan(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		// Retry if health check allocation is in progress (check error message generically)
		if errs.Contains(err, "health check allocation is in progress") {
			_, err = waitPlanDeletable(ctx, conn, state.ID.ValueString(), r.DeleteTimeout(ctx, state.Timeouts))
			if err != nil {
				resp.Diagnostics.AddError("waiting for ARC Region Switch Plan to be deletable", err.Error())
				return
			}
			// Retry delete
			_, err = conn.DeletePlan(ctx, &input)
			if err != nil {
				if errors.As(err, &nfe) {
					return
				}
				resp.Diagnostics.AddError("deleting ARC Region Switch Plan", err.Error())
				return
			}
			return
		}
		resp.Diagnostics.AddError("deleting ARC Region Switch Plan", err.Error())
		return
	}
}

func (r *resourcePlan) ValidateModel(ctx context.Context, schema *fwschema.Schema) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	// Basic validation is handled by the schema validators
	return diags
}

// Custom expand to handle complex nested transformations
// Expand converts the Terraform resource model to AWS API input.
// Custom expansion is required because:
// 1. Union Type Handling: ExecutionBlockConfiguration uses AWS SDK union types that AutoFlex cannot handle automatically
// 2. Complex Nested Transformations: ScalingResources (list→map[string]map[string]) and RegionAndRoutingControls (list→map[string][]) require manual conversion
// 3. Conditional Logic: Different execution block types require different field mappings and validations
// AutoFlex works well for simple field mappings but cannot handle these complex structural transformations.
func (m resourcePlanModel) Expand(ctx context.Context) (result any, diags fwdiag.Diagnostics) {
	var apiObject arcregionswitch.CreatePlanInput

	// Manually expand all fields except AssociatedAlarms and complex nested structures
	diags.Append(flex.Expand(ctx, m.Name, &apiObject.Name)...)
	diags.Append(flex.Expand(ctx, m.ExecutionRole, &apiObject.ExecutionRole)...)
	diags.Append(flex.Expand(ctx, m.RecoveryApproach, &apiObject.RecoveryApproach)...)
	diags.Append(flex.Expand(ctx, m.Regions, &apiObject.Regions)...)
	diags.Append(flex.Expand(ctx, m.Description, &apiObject.Description)...)
	diags.Append(flex.Expand(ctx, m.PrimaryRegion, &apiObject.PrimaryRegion)...)
	diags.Append(flex.Expand(ctx, m.RecoveryTimeObjectiveMinutes, &apiObject.RecoveryTimeObjectiveMinutes)...)
	diags.Append(flex.Expand(ctx, m.AssociatedAlarms, &apiObject.AssociatedAlarms)...)
	diags.Append(flex.Expand(ctx, m.Triggers, &apiObject.Triggers)...)
	diags.Append(flex.Expand(ctx, m.Tags, &apiObject.Tags)...)

	// Handle Workflows with complex nested ScalingResources transformation
	if !m.Workflows.IsNull() && !m.Workflows.IsUnknown() {
		var workflows []workflowModel
		diags.Append(m.Workflows.ElementsAs(ctx, &workflows, false)...)
		if diags.HasError() {
			return nil, diags
		}

		apiObject.Workflows = make([]awstypes.Workflow, len(workflows))
		for i, workflow := range workflows {
			var apiWorkflow awstypes.Workflow
			diags.Append(flex.Expand(ctx, workflow.WorkflowTargetAction, &apiWorkflow.WorkflowTargetAction)...)
			diags.Append(flex.Expand(ctx, workflow.WorkflowTargetRegion, &apiWorkflow.WorkflowTargetRegion)...)
			diags.Append(flex.Expand(ctx, workflow.WorkflowDescription, &apiWorkflow.WorkflowDescription)...)

			// Handle Steps with EKS ScalingResources transformation
			if !workflow.Steps.IsNull() && !workflow.Steps.IsUnknown() {
				var steps []stepModel
				diags.Append(workflow.Steps.ElementsAs(ctx, &steps, false)...)
				if diags.HasError() {
					return nil, diags
				}

				apiWorkflow.Steps = make([]awstypes.Step, len(steps))
				for j, step := range steps {
					var apiStep awstypes.Step
					diags.Append(flex.Expand(ctx, step.Name, &apiStep.Name)...)
					diags.Append(flex.Expand(ctx, step.Description, &apiStep.Description)...)
					diags.Append(flex.Expand(ctx, step.ExecutionBlockType, &apiStep.ExecutionBlockType)...)

					// Handle ExecutionBlockConfiguration with ScalingResources
					if !step.ExecutionBlockConfiguration.IsNull() && !step.ExecutionBlockConfiguration.IsUnknown() {
						var execConfigs []executionBlockConfigurationModel
						diags.Append(step.ExecutionBlockConfiguration.ElementsAs(ctx, &execConfigs, false)...)
						if diags.HasError() {
							return nil, diags
						}

						if len(execConfigs) > 0 {
							execConfig := execConfigs[0]

							// Handle EksResourceScalingConfig specifically
							if !execConfig.EksResourceScalingConfig.IsNull() {
								data, d := execConfig.EksResourceScalingConfig.ToPtr(ctx)
								diags.Append(d...)
								if diags.HasError() {
									return nil, diags
								}

								var apiEksConfig awstypes.EksResourceScalingConfiguration
								diags.Append(flex.Expand(ctx, data.CapacityMonitoringApproach, &apiEksConfig.CapacityMonitoringApproach)...)
								diags.Append(flex.Expand(ctx, data.TargetPercent, &apiEksConfig.TargetPercent)...)
								diags.Append(flex.Expand(ctx, data.TimeoutMinutes, &apiEksConfig.TimeoutMinutes)...)
								diags.Append(flex.Expand(ctx, data.KubernetesResourceType, &apiEksConfig.KubernetesResourceType)...)
								diags.Append(flex.Expand(ctx, data.EksClusters, &apiEksConfig.EksClusters)...)
								diags.Append(flex.Expand(ctx, data.Ungraceful, &apiEksConfig.Ungraceful)...)

								// Handle complex ScalingResources conversion
								if !data.ScalingResources.IsNull() && !data.ScalingResources.IsUnknown() {
									var scalingResources []scalingResourcesModel
									diags.Append(data.ScalingResources.ElementsAs(ctx, &scalingResources, false)...)
									if diags.HasError() {
										return nil, diags
									}

									apiEksConfig.ScalingResources = make([]map[string]map[string]awstypes.KubernetesScalingResource, len(scalingResources))
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
												var kubeRes awstypes.KubernetesScalingResource
												diags.Append(flex.Expand(ctx, res.Name, &kubeRes.Name)...)
												diags.Append(flex.Expand(ctx, res.Namespace, &kubeRes.Namespace)...)
												diags.Append(flex.Expand(ctx, res.HpaName, &kubeRes.HpaName)...)

												resourceMap[res.ResourceName.ValueString()] = kubeRes
											}

											namespaceMap[sr.Namespace.ValueString()] = resourceMap
										}

										apiEksConfig.ScalingResources[k] = namespaceMap
									}
								}

								r := awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig{
									Value: apiEksConfig,
								}
								apiStep.ExecutionBlockConfiguration = &r
							} else if !execConfig.ArcRoutingControlConfig.IsNull() {
								// Handle ArcRoutingControlConfig specifically
								data, d := execConfig.ArcRoutingControlConfig.ToPtr(ctx)
								diags.Append(d...)
								if diags.HasError() {
									return nil, diags
								}

								var apiArcConfig awstypes.ArcRoutingControlConfiguration
								diags.Append(flex.Expand(ctx, data.CrossAccountRole, &apiArcConfig.CrossAccountRole)...)
								diags.Append(flex.Expand(ctx, data.ExternalID, &apiArcConfig.ExternalId)...)
								diags.Append(flex.Expand(ctx, data.TimeoutMinutes, &apiArcConfig.TimeoutMinutes)...)

								// Handle complex RegionAndRoutingControls conversion
								if !data.RegionAndRoutingControls.IsNull() && !data.RegionAndRoutingControls.IsUnknown() {
									var regionControls []regionAndRoutingControlsModel
									diags.Append(data.RegionAndRoutingControls.ElementsAs(ctx, &regionControls, false)...)
									if diags.HasError() {
										return nil, diags
									}

									apiArcConfig.RegionAndRoutingControls = make(map[string][]awstypes.ArcRoutingControlState, len(regionControls))
									for _, rc := range regionControls {
										region := rc.Region.ValueString()

										if !rc.RoutingControlARNs.IsNull() && !rc.RoutingControlARNs.IsUnknown() {
											var arns []string
											diags.Append(rc.RoutingControlARNs.ElementsAs(ctx, &arns, false)...)
											if diags.HasError() {
												return nil, diags
											}

											states := make([]awstypes.ArcRoutingControlState, len(arns))
											for i, arn := range arns {
												states[i] = awstypes.ArcRoutingControlState{
													RoutingControlArn: &arn,
													// Default state - could be configurable
													State: awstypes.RoutingControlStateChangeOn,
												}
											}

											apiArcConfig.RegionAndRoutingControls[region] = states
										}
									}
								}

								r := awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig{
									Value: apiArcConfig,
								}
								apiStep.ExecutionBlockConfiguration = &r
							} else if !execConfig.ExecutionApprovalConfig.IsNull() {
								data, d := execConfig.ExecutionApprovalConfig.ToPtr(ctx)
								diags.Append(d...)
								if diags.HasError() {
									return nil, diags
								}
								var r awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig
								diags.Append(flex.Expand(ctx, data, &r.Value)...)
								apiStep.ExecutionBlockConfiguration = &r
							} else if !execConfig.EcsCapacityIncreaseConfig.IsNull() {
								data, d := execConfig.EcsCapacityIncreaseConfig.ToPtr(ctx)
								diags.Append(d...)
								if diags.HasError() {
									return nil, diags
								}
								var r awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig
								diags.Append(flex.Expand(ctx, data, &r.Value)...)
								apiStep.ExecutionBlockConfiguration = &r
							} else if !execConfig.Route53HealthCheckConfig.IsNull() {
								data, d := execConfig.Route53HealthCheckConfig.ToPtr(ctx)
								diags.Append(d...)
								if diags.HasError() {
									return nil, diags
								}
								var r awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig
								diags.Append(flex.Expand(ctx, data, &r.Value)...)
								apiStep.ExecutionBlockConfiguration = &r
							} else if !execConfig.CustomActionLambdaConfig.IsNull() {
								data, d := execConfig.CustomActionLambdaConfig.ToPtr(ctx)
								diags.Append(d...)
								if diags.HasError() {
									return nil, diags
								}
								var r awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig
								diags.Append(flex.Expand(ctx, data, &r.Value)...)
								apiStep.ExecutionBlockConfiguration = &r
							} else if !execConfig.ParallelConfig.IsNull() {
								data, d := execConfig.ParallelConfig.ToPtr(ctx)
								diags.Append(d...)
								if diags.HasError() {
									return nil, diags
								}

								// Handle ParallelConfig with nested step execution block configurations
								var apiParallelConfig awstypes.ParallelExecutionBlockConfiguration
								if !data.Step.IsNull() && !data.Step.IsUnknown() {
									var parallelSteps []parallelStepModel
									diags.Append(data.Step.ElementsAs(ctx, &parallelSteps, false)...)
									if diags.HasError() {
										return nil, diags
									}

									apiParallelConfig.Steps = make([]awstypes.Step, len(parallelSteps))
									for k, pStep := range parallelSteps {
										var apiParallelStep awstypes.Step
										diags.Append(flex.Expand(ctx, pStep.Name, &apiParallelStep.Name)...)
										diags.Append(flex.Expand(ctx, pStep.ExecutionBlockType, &apiParallelStep.ExecutionBlockType)...)
										diags.Append(flex.Expand(ctx, pStep.Description, &apiParallelStep.Description)...)

										// Handle execution block configuration for parallel steps
										if !pStep.ExecutionApprovalConfig.IsNull() {
											pData, pD := pStep.ExecutionApprovalConfig.ToPtr(ctx)
											diags.Append(pD...)
											if diags.HasError() {
												return nil, diags
											}
											var pR awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig
											diags.Append(flex.Expand(ctx, pData, &pR.Value)...)
											apiParallelStep.ExecutionBlockConfiguration = &pR
										} else if !pStep.CustomActionLambdaConfig.IsNull() {
											pData, pD := pStep.CustomActionLambdaConfig.ToPtr(ctx)
											diags.Append(pD...)
											if diags.HasError() {
												return nil, diags
											}
											var pR awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig
											diags.Append(flex.Expand(ctx, pData, &pR.Value)...)
											apiParallelStep.ExecutionBlockConfiguration = &pR
										}

										apiParallelConfig.Steps[k] = apiParallelStep
									}
								}

								var r awstypes.ExecutionBlockConfigurationMemberParallelConfig
								r.Value = apiParallelConfig
								apiStep.ExecutionBlockConfiguration = &r
							} else if !execConfig.GlobalAuroraConfig.IsNull() {
								data, d := execConfig.GlobalAuroraConfig.ToPtr(ctx)
								diags.Append(d...)
								if diags.HasError() {
									return nil, diags
								}
								var r awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig
								diags.Append(flex.Expand(ctx, data, &r.Value)...)
								apiStep.ExecutionBlockConfiguration = &r
							} else if !execConfig.Ec2AsgCapacityIncreaseConfig.IsNull() {
								data, d := execConfig.Ec2AsgCapacityIncreaseConfig.ToPtr(ctx)
								diags.Append(d...)
								if diags.HasError() {
									return nil, diags
								}
								var r awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig
								diags.Append(flex.Expand(ctx, data, &r.Value)...)
								apiStep.ExecutionBlockConfiguration = &r
							}
						}
					}

					apiWorkflow.Steps[j] = apiStep
				}
			}

			apiObject.Workflows[i] = apiWorkflow
		}
	}

	return apiObject, diags
}

// Custom flatten to handle reverse transformations
// Flatten converts AWS API output to Terraform resource model.
// Custom flattening is required because:
// 1. Union Type Handling: ExecutionBlockConfiguration union types need manual type switching and field extraction
// 2. Reverse Complex Transformations: Converting AWS API maps back to Terraform list structures for ScalingResources and RegionAndRoutingControls
// 3. Workflow Ordering: AWS API returns workflows in non-deterministic order, requiring sorting for consistent Terraform state
// 4. Nested Parallel Steps: Parallel execution block configurations require recursive flattening with proper initialization
// AutoFlex cannot handle these reverse transformations and complex nested structures automatically.
func (m *resourcePlanModel) Flatten(ctx context.Context, v any) (diags fwdiag.Diagnostics) {
	plan, ok := v.(*awstypes.Plan)
	if !ok {
		diags.AddError(
			"Unexpected Type",
			"Expected *awstypes.Plan",
		)
		return diags
	}

	if plan == nil {
		diags.AddError(
			"Unexpected Response",
			"Plan is nil",
		)
		return diags
	}

	// Handle simple fields with AutoFlex
	diags.Append(flex.Flatten(ctx, plan.Name, &m.Name)...)
	diags.Append(flex.Flatten(ctx, plan.ExecutionRole, &m.ExecutionRole)...)
	diags.Append(flex.Flatten(ctx, plan.RecoveryApproach, &m.RecoveryApproach)...)
	diags.Append(flex.Flatten(ctx, plan.Regions, &m.Regions)...)
	diags.Append(flex.Flatten(ctx, plan.Description, &m.Description)...)
	diags.Append(flex.Flatten(ctx, plan.PrimaryRegion, &m.PrimaryRegion)...)
	diags.Append(flex.Flatten(ctx, plan.RecoveryTimeObjectiveMinutes, &m.RecoveryTimeObjectiveMinutes)...)
	diags.Append(flex.Flatten(ctx, plan.Triggers, &m.Triggers)...)

	diags.Append(flex.Flatten(ctx, plan.AssociatedAlarms, &m.AssociatedAlarms)...)

	// Handle Workflows with complex nested transformations
	if len(plan.Workflows) > 0 {
		// Sort workflows by target action for consistent ordering (activate before deactivate)
		slices.SortFunc(plan.Workflows, func(i, j awstypes.Workflow) int {
			if string(i.WorkflowTargetAction) < string(j.WorkflowTargetAction) {
				return -1
			}
			if string(i.WorkflowTargetAction) > string(j.WorkflowTargetAction) {
				return 1
			}
			return 0
		})

		workflows := make([]workflowModel, len(plan.Workflows))
		for i, workflow := range plan.Workflows {
			diags.Append(flex.Flatten(ctx, workflow.WorkflowTargetAction, &workflows[i].WorkflowTargetAction)...)
			diags.Append(flex.Flatten(ctx, workflow.WorkflowTargetRegion, &workflows[i].WorkflowTargetRegion)...)
			diags.Append(flex.Flatten(ctx, workflow.WorkflowDescription, &workflows[i].WorkflowDescription)...)

			// Handle Steps - this is where the complex logic will go
			if len(workflow.Steps) > 0 {
				steps := make([]stepModel, len(workflow.Steps))
				for j, step := range workflow.Steps {
					diags.Append(flex.Flatten(ctx, step.Name, &steps[j].Name)...)
					diags.Append(flex.Flatten(ctx, step.Description, &steps[j].Description)...)
					diags.Append(flex.Flatten(ctx, step.ExecutionBlockType, &steps[j].ExecutionBlockType)...)

					// Handle ExecutionBlockConfiguration - reverse of our complex expand logic
					if step.ExecutionBlockConfiguration != nil {
						// Initialize with empty values for all fields to avoid nil pointer issues
						execConfig := executionBlockConfigurationModel{
							ExecutionApprovalConfig:      fwtypes.NewListNestedObjectValueOfNull[executionApprovalConfigModel](ctx),
							Route53HealthCheckConfig:     fwtypes.NewListNestedObjectValueOfNull[route53HealthCheckConfigModel](ctx),
							CustomActionLambdaConfig:     fwtypes.NewListNestedObjectValueOfNull[customActionLambdaConfigModel](ctx),
							GlobalAuroraConfig:           fwtypes.NewListNestedObjectValueOfNull[globalAuroraConfigModel](ctx),
							Ec2AsgCapacityIncreaseConfig: fwtypes.NewListNestedObjectValueOfNull[ec2AsgCapacityIncreaseConfigModel](ctx),
							EcsCapacityIncreaseConfig:    fwtypes.NewListNestedObjectValueOfNull[ecsCapacityIncreaseConfigModel](ctx),
							EksResourceScalingConfig:     fwtypes.NewListNestedObjectValueOfNull[eksResourceScalingConfigModel](ctx),
							ArcRoutingControlConfig:      fwtypes.NewListNestedObjectValueOfNull[arcRoutingControlConfigModel](ctx),
							ParallelConfig:               fwtypes.NewListNestedObjectValueOfNull[parallelConfigModel](ctx),
						}

						// Handle union type flattening manually (similar to expand logic)
						switch t := step.ExecutionBlockConfiguration.(type) {
						case *awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig:
							// Handle EKS ScalingResources complex transformation manually
							var eksConfig eksResourceScalingConfigModel
							diags.Append(flex.Flatten(ctx, t.Value.CapacityMonitoringApproach, &eksConfig.CapacityMonitoringApproach)...)
							diags.Append(flex.Flatten(ctx, t.Value.EksClusters, &eksConfig.EksClusters)...)
							diags.Append(flex.Flatten(ctx, t.Value.KubernetesResourceType, &eksConfig.KubernetesResourceType)...)
							diags.Append(flex.Flatten(ctx, t.Value.TargetPercent, &eksConfig.TargetPercent)...)
							diags.Append(flex.Flatten(ctx, t.Value.TimeoutMinutes, &eksConfig.TimeoutMinutes)...)
							diags.Append(flex.Flatten(ctx, t.Value.Ungraceful, &eksConfig.Ungraceful)...)

							// Handle ScalingResources: []map[string]map[string]KubernetesScalingResource → []scalingResourcesModel
							if len(t.Value.ScalingResources) > 0 {
								scalingResources := make([]scalingResourcesModel, len(t.Value.ScalingResources))
								for i, sr := range t.Value.ScalingResources {
									for namespace, resourceMap := range sr {
										scalingResources[i].Namespace = types.StringValue(namespace)

										// Convert map[string]KubernetesScalingResource → []kubernetesScalingResourceModel
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
								eksConfig.ScalingResources, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, scalingResources)
								diags.Append(d...)
							}

							execConfig.EksResourceScalingConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &eksConfig)
						case *awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig:
							// Handle ARC RegionAndRoutingControls complex transformation manually
							var arcConfig arcRoutingControlConfigModel
							diags.Append(flex.Flatten(ctx, t.Value.CrossAccountRole, &arcConfig.CrossAccountRole)...)
							diags.Append(flex.Flatten(ctx, t.Value.ExternalId, &arcConfig.ExternalID)...)
							diags.Append(flex.Flatten(ctx, t.Value.TimeoutMinutes, &arcConfig.TimeoutMinutes)...)

							// Handle RegionAndRoutingControls: map[string][]ArcRoutingControlState → []regionAndRoutingControlsModel
							if len(t.Value.RegionAndRoutingControls) > 0 {
								regionControls := make([]regionAndRoutingControlsModel, 0, len(t.Value.RegionAndRoutingControls))
								for region, controlStates := range t.Value.RegionAndRoutingControls {
									var regionModel regionAndRoutingControlsModel
									regionModel.Region = types.StringValue(region)

									// Extract ARNs from ArcRoutingControlState slice
									arns := make([]string, len(controlStates))
									for i, state := range controlStates {
										if state.RoutingControlArn != nil {
											arns[i] = aws.ToString(state.RoutingControlArn)
										}
									}

									diags.Append(flex.Flatten(ctx, arns, &regionModel.RoutingControlARNs)...)

									regionControls = append(regionControls, regionModel)
								}

								var d fwdiag.Diagnostics
								arcConfig.RegionAndRoutingControls, d = fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, regionControls)
								diags.Append(d...)
							}

							execConfig.ArcRoutingControlConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &arcConfig)
						case *awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig:
							diags.Append(flex.Flatten(ctx, &t.Value, &execConfig.ExecutionApprovalConfig)...)
						case *awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig:
							diags.Append(flex.Flatten(ctx, &t.Value, &execConfig.EcsCapacityIncreaseConfig)...)
						case *awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig:
							diags.Append(flex.Flatten(ctx, &t.Value, &execConfig.Route53HealthCheckConfig)...)
						case *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig:
							diags.Append(flex.Flatten(ctx, &t.Value, &execConfig.CustomActionLambdaConfig)...)
						case *awstypes.ExecutionBlockConfigurationMemberParallelConfig:
							// Handle ParallelConfig with nested step execution block configurations manually
							var parallelConfig parallelConfigModel

							if len(t.Value.Steps) > 0 {
								parallelSteps := make([]parallelStepModel, len(t.Value.Steps))
								for i, step := range t.Value.Steps {
									// Initialize with empty values for all fields to avoid nil pointer issues
									parallelSteps[i] = parallelStepModel{
										CustomActionLambdaConfig: fwtypes.NewListNestedObjectValueOfNull[customActionLambdaConfigModel](ctx),
										ExecutionApprovalConfig:  fwtypes.NewListNestedObjectValueOfNull[executionApprovalConfigModel](ctx),
									}

									diags.Append(flex.Flatten(ctx, step.Name, &parallelSteps[i].Name)...)
									diags.Append(flex.Flatten(ctx, step.Description, &parallelSteps[i].Description)...)
									diags.Append(flex.Flatten(ctx, step.ExecutionBlockType, &parallelSteps[i].ExecutionBlockType)...)

									// Handle parallel step execution block configurations
									if step.ExecutionBlockConfiguration != nil {
										switch pType := step.ExecutionBlockConfiguration.(type) {
										case *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig:
											diags.Append(flex.Flatten(ctx, &pType.Value, &parallelSteps[i].CustomActionLambdaConfig)...)
										case *awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig:
											diags.Append(flex.Flatten(ctx, &pType.Value, &parallelSteps[i].ExecutionApprovalConfig)...)
										}
									}
								}

								var d fwdiag.Diagnostics
								parallelConfig.Step, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, parallelSteps)
								diags.Append(d...)
							}

							execConfig.ParallelConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &parallelConfig)
						case *awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig:
							diags.Append(flex.Flatten(ctx, &t.Value, &execConfig.GlobalAuroraConfig)...)
						case *awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig:
							diags.Append(flex.Flatten(ctx, &t.Value, &execConfig.Ec2AsgCapacityIncreaseConfig)...)
						}

						var d fwdiag.Diagnostics
						steps[j].ExecutionBlockConfiguration, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []executionBlockConfigurationModel{execConfig})
						diags.Append(d...)
					} else {
						// Set empty list if no execution block configuration
						var d fwdiag.Diagnostics
						steps[j].ExecutionBlockConfiguration, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []executionBlockConfigurationModel{})
						diags.Append(d...)
					}
				}

				var d fwdiag.Diagnostics
				workflows[i].Steps, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, steps)
				diags.Append(d...)
			} else {
				// Set empty list if no steps
				var d fwdiag.Diagnostics
				workflows[i].Steps, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []stepModel{})
				diags.Append(d...)
			}
		}

		var d fwdiag.Diagnostics
		m.Workflows, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, workflows)
		diags.Append(d...)
	} else {
		// Set empty list if no workflows
		var d fwdiag.Diagnostics
		m.Workflows, d = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []workflowModel{})
		diags.Append(d...)
	}

	return diags
}

func findPlanByARN(ctx context.Context, conn *arcregionswitch.Client, arn string) (*awstypes.Plan, error) {
	input := arcregionswitch.GetPlanInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetPlan(ctx, &input)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		return nil, err
	}

	if output == nil || output.Plan == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Plan, nil
}

func findRoute53HealthChecks(ctx context.Context, conn *arcregionswitch.Client, planARN string) ([]awstypes.Route53HealthCheck, error) {
	input := arcregionswitch.ListRoute53HealthChecksInput{
		Arn: aws.String(planARN),
	}

	output, err := conn.ListRoute53HealthChecks(ctx, &input)
	if err != nil {
		return nil, err
	}

	return output.HealthChecks, nil
}

type resourcePlanModel struct {
	framework.WithRegionModel
	ARN                          types.String                                         `tfsdk:"arn"`
	ID                           types.String                                         `tfsdk:"id"`
	Name                         types.String                                         `tfsdk:"name"`
	ExecutionRole                types.String                                         `tfsdk:"execution_role"`
	RecoveryApproach             fwtypes.StringEnum[awstypes.RecoveryApproach]        `tfsdk:"recovery_approach"`
	Regions                      fwtypes.ListOfString                                 `tfsdk:"regions"`
	Description                  types.String                                         `tfsdk:"description"`
	PrimaryRegion                types.String                                         `tfsdk:"primary_region"`
	RecoveryTimeObjectiveMinutes types.Int64                                          `tfsdk:"recovery_time_objective_minutes"`
	AssociatedAlarms             fwtypes.SetNestedObjectValueOf[associatedAlarmModel] `tfsdk:"associated_alarms"`
	Workflows                    fwtypes.ListNestedObjectValueOf[workflowModel]       `tfsdk:"workflow"`
	Triggers                     fwtypes.ListNestedObjectValueOf[triggerModel]        `tfsdk:"triggers"`
	Tags                         tftags.Map                                           `tfsdk:"tags"`
	TagsAll                      tftags.Map                                           `tfsdk:"tags_all"`
	Timeouts                     timeouts.Value                                       `tfsdk:"timeouts"`
}

type associatedAlarmModel struct {
	MapBlockKey        types.String                           `tfsdk:"map_block_key"`
	AlarmType          fwtypes.StringEnum[awstypes.AlarmType] `tfsdk:"alarm_type"`
	ResourceIdentifier types.String                           `tfsdk:"resource_identifier"`
	CrossAccountRole   types.String                           `tfsdk:"cross_account_role"`
	ExternalID         types.String                           `tfsdk:"external_id"`
}

type workflowModel struct {
	WorkflowTargetAction fwtypes.StringEnum[awstypes.WorkflowTargetAction] `tfsdk:"workflow_target_action"`
	WorkflowTargetRegion types.String                                      `tfsdk:"workflow_target_region"`
	WorkflowDescription  types.String                                      `tfsdk:"workflow_description"`
	Steps                fwtypes.ListNestedObjectValueOf[stepModel]        `tfsdk:"step"`
}

type route53HealthCheckConfigModel struct {
	HostedZoneID     types.String                                    `tfsdk:"hosted_zone_id"`
	RecordName       types.String                                    `tfsdk:"record_name"`
	CrossAccountRole types.String                                    `tfsdk:"cross_account_role"`
	ExternalID       types.String                                    `tfsdk:"external_id"`
	TimeoutMinutes   types.Int64                                     `tfsdk:"timeout_minutes"`
	RecordSets       fwtypes.ListNestedObjectValueOf[recordSetModel] `tfsdk:"record_sets"`
}

type recordSetModel struct {
	RecordSetIdentifier types.String `tfsdk:"record_set_identifier"`
	Region              types.String `tfsdk:"region"`
}

type stepModel struct {
	Name                        types.String                                                      `tfsdk:"name"`
	ExecutionBlockType          fwtypes.StringEnum[awstypes.ExecutionBlockType]                   `tfsdk:"execution_block_type"`
	Description                 types.String                                                      `tfsdk:"description"`
	ExecutionBlockConfiguration fwtypes.ListNestedObjectValueOf[executionBlockConfigurationModel] `tfsdk:"execution_block_configuration"`
}

type executionBlockConfigurationModel struct {
	ExecutionApprovalConfig      fwtypes.ListNestedObjectValueOf[executionApprovalConfigModel]      `tfsdk:"execution_approval_config"`
	Route53HealthCheckConfig     fwtypes.ListNestedObjectValueOf[route53HealthCheckConfigModel]     `tfsdk:"route53_health_check_config"`
	CustomActionLambdaConfig     fwtypes.ListNestedObjectValueOf[customActionLambdaConfigModel]     `tfsdk:"custom_action_lambda_config"`
	GlobalAuroraConfig           fwtypes.ListNestedObjectValueOf[globalAuroraConfigModel]           `tfsdk:"global_aurora_config"`
	Ec2AsgCapacityIncreaseConfig fwtypes.ListNestedObjectValueOf[ec2AsgCapacityIncreaseConfigModel] `tfsdk:"ec2_asg_capacity_increase_config"`
	EcsCapacityIncreaseConfig    fwtypes.ListNestedObjectValueOf[ecsCapacityIncreaseConfigModel]    `tfsdk:"ecs_capacity_increase_config"`
	EksResourceScalingConfig     fwtypes.ListNestedObjectValueOf[eksResourceScalingConfigModel]     `tfsdk:"eks_resource_scaling_config"`
	ArcRoutingControlConfig      fwtypes.ListNestedObjectValueOf[arcRoutingControlConfigModel]      `tfsdk:"arc_routing_control_config"`
	ParallelConfig               fwtypes.ListNestedObjectValueOf[parallelConfigModel]               `tfsdk:"parallel_config"`
}

type executionApprovalConfigModel struct {
	ApprovalRole   types.String `tfsdk:"approval_role"`
	TimeoutMinutes types.Int64  `tfsdk:"timeout_minutes"`
}

type customActionLambdaConfigModel struct {
	RegionToRun          types.String                                     `tfsdk:"region_to_run"`
	RetryIntervalMinutes types.Float64                                    `tfsdk:"retry_interval_minutes"`
	TimeoutMinutes       types.Int64                                      `tfsdk:"timeout_minutes"`
	Lambda               fwtypes.ListNestedObjectValueOf[lambdaModel]     `tfsdk:"lambda"`
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

// Global Aurora Configuration Models
type globalAuroraConfigModel struct {
	Behavior                fwtypes.StringEnum[awstypes.GlobalAuroraDefaultBehavior]     `tfsdk:"behavior"`
	GlobalClusterIdentifier types.String                                                 `tfsdk:"global_cluster_identifier"`
	DatabaseClusterARNs     fwtypes.ListOfARN                                            `tfsdk:"database_cluster_arns"`
	CrossAccountRole        types.String                                                 `tfsdk:"cross_account_role"`
	ExternalID              types.String                                                 `tfsdk:"external_id"`
	TimeoutMinutes          types.Int64                                                  `tfsdk:"timeout_minutes"`
	Ungraceful              fwtypes.ListNestedObjectValueOf[globalAuroraUngracefulModel] `tfsdk:"ungraceful"`
}

type globalAuroraUngracefulModel struct {
	Ungraceful fwtypes.StringEnum[awstypes.GlobalAuroraUngracefulBehavior] `tfsdk:"ungraceful"`
}

// EC2 ASG Configuration Models
type ec2AsgCapacityIncreaseConfigModel struct {
	CapacityMonitoringApproach fwtypes.StringEnum[awstypes.Ec2AsgCapacityMonitoringApproach] `tfsdk:"capacity_monitoring_approach"`
	TargetPercent              types.Int64                                                   `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int64                                                   `tfsdk:"timeout_minutes"`
	Asgs                       fwtypes.ListNestedObjectValueOf[asgModel]                     `tfsdk:"asgs"`
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
	TargetPercent              types.Int64                                                `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int64                                                `tfsdk:"timeout_minutes"`
	Services                   fwtypes.ListNestedObjectValueOf[serviceModel]              `tfsdk:"services"`
	Ungraceful                 fwtypes.ListNestedObjectValueOf[ecsUngracefulModel]        `tfsdk:"ungraceful"`
}

type serviceModel struct {
	ClusterARN       types.String `tfsdk:"cluster_arn"`
	ServiceARN       types.String `tfsdk:"service_arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalID       types.String `tfsdk:"external_id"`
}

type ecsUngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

// EKS Configuration Models
type eksResourceScalingConfigModel struct {
	CapacityMonitoringApproach fwtypes.StringEnum[awstypes.EksCapacityMonitoringApproach]   `tfsdk:"capacity_monitoring_approach"`
	TargetPercent              types.Int64                                                  `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int64                                                  `tfsdk:"timeout_minutes"`
	KubernetesResourceType     fwtypes.ListNestedObjectValueOf[kubernetesResourceTypeModel] `tfsdk:"kubernetes_resource_type"`
	EksClusters                fwtypes.ListNestedObjectValueOf[eksClusterModel]             `tfsdk:"eks_clusters"`
	ScalingResources           fwtypes.ListNestedObjectValueOf[scalingResourcesModel]       `tfsdk:"scaling_resources"`
	Ungraceful                 fwtypes.ListNestedObjectValueOf[eksUngracefulModel]          `tfsdk:"ungraceful"`
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
	ResourceName types.String `tfsdk:"resource_name"`
	Name         types.String `tfsdk:"name"`
	Namespace    types.String `tfsdk:"namespace"`
	HpaName      types.String `tfsdk:"hpa_name"`
}

type eksUngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

// ARC Routing Control Configuration Models
type arcRoutingControlConfigModel struct {
	CrossAccountRole         types.String                                                  `tfsdk:"cross_account_role"`
	ExternalID               types.String                                                  `tfsdk:"external_id"`
	TimeoutMinutes           types.Int64                                                   `tfsdk:"timeout_minutes"`
	RegionAndRoutingControls fwtypes.SetNestedObjectValueOf[regionAndRoutingControlsModel] `tfsdk:"region_and_routing_controls"`
}

type regionAndRoutingControlsModel struct {
	Region             types.String      `tfsdk:"region"`
	RoutingControlARNs fwtypes.ListOfARN `tfsdk:"routing_control_arns"`
}

// Parallel Configuration Models
type parallelConfigModel struct {
	Step fwtypes.ListNestedObjectValueOf[parallelStepModel] `tfsdk:"step"`
}

type parallelStepModel struct {
	Name                     types.String                                                   `tfsdk:"name"`
	ExecutionBlockType       fwtypes.StringEnum[awstypes.ExecutionBlockType]                `tfsdk:"execution_block_type"`
	Description              types.String                                                   `tfsdk:"description"`
	ExecutionApprovalConfig  fwtypes.ListNestedObjectValueOf[executionApprovalConfigModel]  `tfsdk:"execution_approval_config"`
	CustomActionLambdaConfig fwtypes.ListNestedObjectValueOf[customActionLambdaConfigModel] `tfsdk:"custom_action_lambda_config"`
}

// Trigger Configuration Models
type triggerModel struct {
	Action                           fwtypes.StringEnum[awstypes.WorkflowTargetAction] `tfsdk:"action"`
	Description                      types.String                                      `tfsdk:"description"`
	MinDelayMinutesBetweenExecutions types.Int64                                       `tfsdk:"min_delay_minutes_between_executions"`
	TargetRegion                     types.String                                      `tfsdk:"target_region"`
	Conditions                       fwtypes.ListNestedObjectValueOf[conditionModel]   `tfsdk:"conditions"`
}

type conditionModel struct {
	AssociatedAlarmName types.String                                `tfsdk:"associated_alarm_name"`
	Condition           fwtypes.StringEnum[awstypes.AlarmCondition] `tfsdk:"condition"`
}

func (r *resourcePlan) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Basic validation is handled by the schema validators
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
		return nil, err
	}

	plan, ok := outputRaw.(*awstypes.Plan)
	if !ok {
		return nil, nil
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
			return nil, err
		}
	}

	return plan, nil
}

func statusRoute53HealthChecks(ctx context.Context, conn *arcregionswitch.Client, arn string, expectedCount int) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		healthChecks, err := findRoute53HealthChecks(ctx, conn, arn)
		if err != nil {
			return nil, "", err
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
	return func(ctx context.Context) (any, string, error) {
		plan, err := findPlanByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
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
		return output, err
	}

	return nil, err
}

func statusPlanDeletable(ctx context.Context, conn *arcregionswitch.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		plan, err := findPlanByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
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
		return nil, "", err
	}
}
