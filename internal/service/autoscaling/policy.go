// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_autoscaling_policy", name="Policy")
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourcePolicyImport,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			// All predictive scaling customized metrics shares same metric data query schema
			customizedMetricDataQuerySchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 10,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrExpression: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 1023),
							},
							names.AttrID: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
							"label": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 2047),
							},
							"metric_stat": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"metric": {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"dimensions": {
														Type:     schema.TypeSet,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrName: {
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
													names.AttrMetricName: {
														Type:     schema.TypeString,
														Required: true,
													},
													names.AttrNamespace: {
														Type:     schema.TypeString,
														Required: true,
													},
												},
											},
										},
										"stat": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 100),
										},
										names.AttrUnit: {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
							"return_data": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  true,
							},
						},
					},
				}
			}

			return map[string]*schema.Schema{
				"adjustment_type": {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"autoscaling_group_name": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"cooldown": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				names.AttrEnabled: {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"estimated_instance_warmup": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"metric_aggregation_type": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"min_adjustment_magnitude": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntAtLeast(1),
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"policy_type": {
					Type:             schema.TypeString,
					Optional:         true,
					Default:          policyTypeSimpleScaling, // preserve AWS's default to make validation easier.
					ValidateDiagFunc: enum.Validate[policyType](),
				},
				"predictive_scaling_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"max_capacity_breach_behavior": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          awstypes.PredictiveScalingMaxCapacityBreachBehaviorHonorMaxCapacity,
								ValidateDiagFunc: enum.Validate[awstypes.PredictiveScalingMaxCapacityBreachBehavior](),
							},
							"max_capacity_buffer": {
								Type:         nullable.TypeNullableInt,
								Optional:     true,
								ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(0, 100),
							},
							"metric_specification": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"customized_capacity_metric_specification": {
											Type:          schema.TypeList,
											Optional:      true,
											MaxItems:      1,
											ConflictsWith: []string{"predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification"},
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"metric_data_queries": customizedMetricDataQuerySchema(),
												},
											},
										},
										"customized_load_metric_specification": {
											Type:          schema.TypeList,
											Optional:      true,
											MaxItems:      1,
											ConflictsWith: []string{"predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification"},
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"metric_data_queries": customizedMetricDataQuerySchema(),
												},
											},
										},
										"customized_scaling_metric_specification": {
											Type:          schema.TypeList,
											Optional:      true,
											MaxItems:      1,
											ConflictsWith: []string{"predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification"},
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"metric_data_queries": customizedMetricDataQuerySchema(),
												},
											},
										},
										"predefined_load_metric_specification": {
											Type:          schema.TypeList,
											Optional:      true,
											MaxItems:      1,
											ConflictsWith: []string{"predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification"},
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"predefined_metric_type": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.PredefinedLoadMetricType](),
													},
													"resource_label": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
										"predefined_metric_pair_specification": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"predefined_metric_type": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.PredefinedMetricPairType](),
													},
													"resource_label": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
										"predefined_scaling_metric_specification": {
											Type:          schema.TypeList,
											Optional:      true,
											MaxItems:      1,
											ConflictsWith: []string{"predictive_scaling_configuration.0.metric_specification.0.customized_scaling_metric_specification"},
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"predefined_metric_type": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.PredefinedScalingMetricType](),
													},
													"resource_label": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
										"target_value": {
											Type:     schema.TypeFloat,
											Required: true,
										},
									},
								},
							},
							names.AttrMode: {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          awstypes.PredictiveScalingModeForecastOnly,
								ValidateDiagFunc: enum.Validate[awstypes.PredictiveScalingMode](),
							},
							"scheduling_buffer_time": {
								Type:         nullable.TypeNullableInt,
								Optional:     true,
								ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(0),
							},
						},
					},
				},
				"scaling_adjustment": {
					Type:          schema.TypeInt,
					Optional:      true,
					ConflictsWith: []string{"step_adjustment"},
				},
				"step_adjustment": {
					Type:          schema.TypeSet,
					Optional:      true,
					ConflictsWith: []string{"scaling_adjustment"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"metric_interval_lower_bound": {
								Type:         nullable.TypeNullableFloat,
								Optional:     true,
								ValidateFunc: nullable.ValidateTypeStringNullableFloat,
							},
							"metric_interval_upper_bound": {
								Type:         nullable.TypeNullableFloat,
								Optional:     true,
								ValidateFunc: nullable.ValidateTypeStringNullableFloat,
							},
							"scaling_adjustment": {
								Type:     schema.TypeInt,
								Required: true,
							},
						},
					},
				},
				"target_tracking_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"customized_metric_specification": {
								Type:          schema.TypeList,
								Optional:      true,
								MaxItems:      1,
								ConflictsWith: []string{"target_tracking_configuration.0.predefined_metric_specification"},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"metric_dimension": {
											Type:          schema.TypeList,
											Optional:      true,
											ConflictsWith: []string{"target_tracking_configuration.0.customized_metric_specification.0.metrics"},
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrName: {
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
										names.AttrMetricName: {
											Type:          schema.TypeString,
											Optional:      true,
											ConflictsWith: []string{"target_tracking_configuration.0.customized_metric_specification.0.metrics"},
										},
										"metrics": {
											Type:          schema.TypeSet,
											Optional:      true,
											ConflictsWith: []string{"target_tracking_configuration.0.customized_metric_specification.0.metric_dimension", "target_tracking_configuration.0.customized_metric_specification.0.metric_name", "target_tracking_configuration.0.customized_metric_specification.0.namespace", "target_tracking_configuration.0.customized_metric_specification.0.statistic", "target_tracking_configuration.0.customized_metric_specification.0.unit"},
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrExpression: {
														Type:         schema.TypeString,
														Optional:     true,
														ValidateFunc: validation.StringLenBetween(1, 2047),
													},
													names.AttrID: {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringLenBetween(1, 255),
													},
													"label": {
														Type:         schema.TypeString,
														Optional:     true,
														ValidateFunc: validation.StringLenBetween(1, 2047),
													},
													"metric_stat": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"metric": {
																	Type:     schema.TypeList,
																	Required: true,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"dimensions": {
																				Type:     schema.TypeSet,
																				Optional: true,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						names.AttrName: {
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
																			names.AttrMetricName: {
																				Type:     schema.TypeString,
																				Required: true,
																			},
																			names.AttrNamespace: {
																				Type:     schema.TypeString,
																				Required: true,
																			},
																		},
																	},
																},
																"period": {
																	Type:         schema.TypeInt,
																	Optional:     true,
																	ValidateFunc: validation.IntInSlice([]int{10, 30, 60}),
																},
																"stat": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringLenBetween(1, 100),
																},
																names.AttrUnit: {
																	Type:     schema.TypeString,
																	Optional: true,
																},
															},
														},
													},
													"return_data": {
														Type:     schema.TypeBool,
														Optional: true,
														Default:  true,
													},
												},
											},
										},
										names.AttrNamespace: {
											Type:          schema.TypeString,
											Optional:      true,
											ConflictsWith: []string{"target_tracking_configuration.0.customized_metric_specification.0.metrics"},
										},
										"period": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntInSlice([]int{10, 30, 60}),
										},
										"statistic": {
											Type:          schema.TypeString,
											Optional:      true,
											ConflictsWith: []string{"target_tracking_configuration.0.customized_metric_specification.0.metrics"},
										},
										names.AttrUnit: {
											Type:          schema.TypeString,
											Optional:      true,
											ConflictsWith: []string{"target_tracking_configuration.0.customized_metric_specification.0.metrics"},
										},
									},
								},
							},
							"disable_scale_in": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							"predefined_metric_specification": {
								Type:          schema.TypeList,
								Optional:      true,
								MaxItems:      1,
								ConflictsWith: []string{"target_tracking_configuration.0.customized_metric_specification"},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"predefined_metric_type": {
											Type:     schema.TypeString,
											Required: true,
										},
										"resource_label": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
							"target_value": {
								Type:     schema.TypeFloat,
								Required: true,
							},
						},
					},
				},
			}
		},
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	name := d.Get(names.AttrName).(string)
	input, err := expandPutScalingPolicyInput(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.PutScalingPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Policy (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	p, err := findScalingPolicyByTwoPartKey(ctx, conn, d.Get("autoscaling_group_name").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Policy (%s): %s", d.Id(), err)
	}

	d.Set("adjustment_type", p.AdjustmentType)
	d.Set(names.AttrARN, p.PolicyARN)
	d.Set("autoscaling_group_name", p.AutoScalingGroupName)
	d.Set("cooldown", p.Cooldown)
	d.Set(names.AttrEnabled, p.Enabled)
	d.Set("estimated_instance_warmup", p.EstimatedInstanceWarmup)
	d.Set("metric_aggregation_type", p.MetricAggregationType)
	d.Set("min_adjustment_magnitude", p.MinAdjustmentMagnitude)
	d.Set(names.AttrName, p.PolicyName)
	d.Set("policy_type", p.PolicyType)
	if err := d.Set("predictive_scaling_configuration", flattenPredictiveScalingConfiguration(p.PredictiveScalingConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting predictive_scaling_configuration: %s", err)
	}
	d.Set("scaling_adjustment", p.ScalingAdjustment)
	if err := d.Set("step_adjustment", flattenStepAdjustments(p.StepAdjustments)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting step_adjustment: %s", err)
	}
	if err := d.Set("target_tracking_configuration", flattenTargetTrackingConfiguration(p.TargetTrackingConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_tracking_configuration: %s", err)
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	input, err := expandPutScalingPolicyInput(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.PutScalingPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Auto Scaling Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	log.Printf("[INFO] Deleting Auto Scaling Policy: %s", d.Id())
	input := autoscaling.DeletePolicyInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		PolicyName:           aws.String(d.Id()),
	}
	_, err := conn.DeletePolicy(ctx, &input)

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func resourcePolicyImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <asg-name>/<policy-name>", d.Id())
	}

	asgName := idParts[0]
	policyName := idParts[1]

	d.Set(names.AttrName, policyName)
	d.Set("autoscaling_group_name", asgName)
	d.SetId(policyName)

	return []*schema.ResourceData{d}, nil
}

func findScalingPolicy(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribePoliciesInput) (*awstypes.ScalingPolicy, error) {
	output, err := findScalingPolicies(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findScalingPolicies(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribePoliciesInput) ([]awstypes.ScalingPolicy, error) {
	var output []awstypes.ScalingPolicy

	pages := autoscaling.NewDescribePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ScalingPolicies...)
	}

	return output, nil
}

func findScalingPolicyByTwoPartKey(ctx context.Context, conn *autoscaling.Client, asgName, policyName string) (*awstypes.ScalingPolicy, error) {
	input := &autoscaling.DescribePoliciesInput{
		AutoScalingGroupName: aws.String(asgName),
		PolicyNames:          []string{policyName},
	}

	return findScalingPolicy(ctx, conn, input)
}

// PutScalingPolicy can safely resend all parameters without destroying the
// resource, so create and update can share this common function. It will error
// if certain mutually exclusive values are set.
func expandPutScalingPolicyInput(d *schema.ResourceData) (*autoscaling.PutScalingPolicyInput, error) {
	input := &autoscaling.PutScalingPolicyInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		Enabled:              aws.Bool(d.Get(names.AttrEnabled).(bool)),
		PolicyName:           aws.String(d.Get(names.AttrName).(string)),
	}

	// get policy_type first as parameter support depends on policy type
	policyType := policyType(d.Get("policy_type").(string))
	input.PolicyType = aws.String(string(policyType))

	// This parameter is supported if the policy type is SimpleScaling or StepScaling.
	if v, ok := d.GetOk("adjustment_type"); ok && (policyType == policyTypeSimpleScaling || policyType == policyTypeStepScaling) {
		input.AdjustmentType = aws.String(v.(string))
	}

	// This parameter is supported if the policy type is SimpleScaling.
	if v, ok := d.GetOkExists("cooldown"); ok {
		// 0 is allowed as placeholder even if policyType is not supported
		input.Cooldown = aws.Int32(int32(v.(int)))
		if v.(int) != 0 && policyType != policyTypeSimpleScaling {
			return input, fmt.Errorf("cooldown is only supported for policy type SimpleScaling")
		}
	}

	// This parameter is supported if the policy type is StepScaling or TargetTrackingScaling.
	if v, ok := d.GetOkExists("estimated_instance_warmup"); ok {
		// 0 is NOT allowed as placeholder if policyType is not supported
		if policyType == policyTypeStepScaling || policyType == policyTypeTargetTrackingScaling {
			input.EstimatedInstanceWarmup = aws.Int32(int32(v.(int)))
		}
		if v.(int) != 0 && policyType != policyTypeStepScaling && policyType != policyTypeTargetTrackingScaling {
			return input, fmt.Errorf("estimated_instance_warmup is only supported for policy type StepScaling and TargetTrackingScaling")
		}
	}

	// This parameter is supported if the policy type is StepScaling.
	if v, ok := d.GetOk("metric_aggregation_type"); ok && policyType == policyTypeStepScaling {
		input.MetricAggregationType = aws.String(v.(string))
	}

	// MinAdjustmentMagnitude is supported if the policy type is SimpleScaling or StepScaling.
	if v, ok := d.GetOkExists("min_adjustment_magnitude"); ok && v.(int) != 0 && (policyType == policyTypeSimpleScaling || policyType == policyTypeStepScaling) {
		input.MinAdjustmentMagnitude = aws.Int32(int32(v.(int)))
	}

	if v := d.Get("predictive_scaling_configuration").([]any); len(v) > 0 {
		input.PredictiveScalingConfiguration = expandPredictiveScalingConfiguration(v)
	}

	// This parameter is required if the policy type is SimpleScaling and not supported otherwise.
	//if policy_type=="SimpleScaling" then scaling_adjustment is required and 0 is allowed
	if v, ok := d.GetOkExists("scaling_adjustment"); ok {
		// 0 is NOT allowed as placeholder if policyType is not supported
		if policyType == policyTypeSimpleScaling {
			input.ScalingAdjustment = aws.Int32(int32(v.(int)))
		}
		if v.(int) != 0 && policyType != policyTypeSimpleScaling {
			return input, fmt.Errorf("scaling_adjustment is only supported for policy type SimpleScaling")
		}
	} else if !ok && policyType == policyTypeSimpleScaling {
		return input, fmt.Errorf("scaling_adjustment is required for policy type SimpleScaling")
	}

	// This parameter is required if the policy type is StepScaling and not supported otherwise.
	if v, ok := d.GetOk("step_adjustment"); ok && v.(*schema.Set).Len() > 0 {
		steps := expandStepAdjustments(v.(*schema.Set).List())
		if len(steps) != 0 && policyType != policyTypeStepScaling {
			return input, fmt.Errorf("step_adjustment is only supported for policy type StepScaling")
		}

		input.StepAdjustments = expandStepAdjustments(v.(*schema.Set).List())
	} else if !ok && policyType == policyTypeStepScaling {
		return input, fmt.Errorf("step_adjustment is required for policy type StepScaling")
	}

	// This parameter is required if the policy type is TargetTrackingScaling and not supported otherwise.
	if v, ok := d.GetOk("target_tracking_configuration"); ok {
		input.TargetTrackingConfiguration = expandTargetTrackingConfiguration(v.([]any))
		if policyType != policyTypeTargetTrackingScaling {
			return input, fmt.Errorf("target_tracking_configuration is only supported for policy type TargetTrackingScaling")
		}
	} else if !ok && policyType == policyTypeTargetTrackingScaling {
		return input, fmt.Errorf("target_tracking_configuration is required for policy type TargetTrackingScaling")
	}

	return input, nil
}

func expandTargetTrackingConfiguration(tfList []any) *awstypes.TargetTrackingConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.TargetTrackingConfiguration{}

	apiObject.TargetValue = aws.Float64(tfMap["target_value"].(float64))
	if v, ok := tfMap["disable_scale_in"]; ok {
		apiObject.DisableScaleIn = aws.Bool(v.(bool))
	}
	if v, ok := tfMap["predefined_metric_specification"]; ok && len(v.([]any)) > 0 {
		tfMap := v.([]any)[0].(map[string]any)
		predefinedMetricSpecification := &awstypes.PredefinedMetricSpecification{
			PredefinedMetricType: awstypes.MetricType(tfMap["predefined_metric_type"].(string)),
		}
		if v, ok := tfMap["resource_label"]; ok && v.(string) != "" {
			predefinedMetricSpecification.ResourceLabel = aws.String(v.(string))
		}
		apiObject.PredefinedMetricSpecification = predefinedMetricSpecification
	}
	if v, ok := tfMap["customized_metric_specification"]; ok && len(v.([]any)) > 0 {
		tfMap := v.([]any)[0].(map[string]any)
		customizedMetricSpecification := &awstypes.CustomizedMetricSpecification{}
		if v, ok := tfMap["metrics"].(*schema.Set); ok && v.Len() > 0 {
			customizedMetricSpecification.Metrics = expandTargetTrackingMetricDataQueries(v.List())
		} else {
			if v, ok := tfMap["metric_dimension"]; ok {
				tfList := v.([]any)
				metricDimensions := make([]awstypes.MetricDimension, len(tfList))
				for i := range metricDimensions {
					tfMap := tfList[i].(map[string]any)
					metricDimensions[i] = awstypes.MetricDimension{
						Name:  aws.String(tfMap[names.AttrName].(string)),
						Value: aws.String(tfMap[names.AttrValue].(string)),
					}
				}
				customizedMetricSpecification.Dimensions = metricDimensions
			}
			customizedMetricSpecification.MetricName = aws.String(tfMap[names.AttrMetricName].(string))
			customizedMetricSpecification.Namespace = aws.String(tfMap[names.AttrNamespace].(string))
			if v, ok := tfMap["period"].(int); ok && v != 0 {
				customizedMetricSpecification.Period = aws.Int32(int32(v))
			}
			customizedMetricSpecification.Statistic = awstypes.MetricStatistic(tfMap["statistic"].(string))
			if v, ok := tfMap[names.AttrUnit]; ok && len(v.(string)) > 0 {
				customizedMetricSpecification.Unit = aws.String(v.(string))
			}
		}
		apiObject.CustomizedMetricSpecification = customizedMetricSpecification
	}

	return apiObject
}

func expandTargetTrackingMetricDataQueries(tfList []any) []awstypes.TargetTrackingMetricDataQuery {
	if len(tfList) < 1 {
		return nil
	}

	apiObjects := make([]awstypes.TargetTrackingMetricDataQuery, len(tfList))

	for i := range apiObjects {
		tfMap := tfList[i].(map[string]any)
		apiObject := awstypes.TargetTrackingMetricDataQuery{
			Id: aws.String(tfMap[names.AttrID].(string)),
		}
		if v, ok := tfMap[names.AttrExpression]; ok && v.(string) != "" {
			apiObject.Expression = aws.String(v.(string))
		}
		if v, ok := tfMap["label"]; ok && v.(string) != "" {
			apiObject.Label = aws.String(v.(string))
		}
		if v, ok := tfMap["metric_stat"]; ok && len(v.([]any)) > 0 {
			tfMapMetricStat := v.([]any)[0].(map[string]any)
			tfMapMetric := tfMapMetricStat["metric"].([]any)[0].(map[string]any)
			metric := &awstypes.Metric{
				MetricName: aws.String(tfMapMetric[names.AttrMetricName].(string)),
				Namespace:  aws.String(tfMapMetric[names.AttrNamespace].(string)),
			}
			if v, ok := tfMapMetric["dimensions"]; ok {
				tfList := v.(*schema.Set).List()
				metricDimensions := make([]awstypes.MetricDimension, len(tfList))
				for i := range metricDimensions {
					tfMap := tfList[i].(map[string]any)
					metricDimensions[i] = awstypes.MetricDimension{
						Name:  aws.String(tfMap[names.AttrName].(string)),
						Value: aws.String(tfMap[names.AttrValue].(string)),
					}
				}
				metric.Dimensions = metricDimensions
			}
			targetTrackingMetricStat := &awstypes.TargetTrackingMetricStat{
				Metric: metric,
				Stat:   aws.String(tfMapMetricStat["stat"].(string)),
			}
			if v, ok := tfMapMetricStat["period"].(int); ok && v != 0 {
				targetTrackingMetricStat.Period = aws.Int32(int32(v))
			}
			if v, ok := tfMapMetricStat[names.AttrUnit]; ok && len(v.(string)) > 0 {
				targetTrackingMetricStat.Unit = aws.String(v.(string))
			}
			apiObject.MetricStat = targetTrackingMetricStat
		}
		if v, ok := tfMap["return_data"]; ok {
			apiObject.ReturnData = aws.Bool(v.(bool))
		}
		apiObjects[i] = apiObject
	}

	return apiObjects
}

func expandPredictiveScalingConfiguration(tfList []any) *awstypes.PredictiveScalingConfiguration {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.PredictiveScalingConfiguration{
		MaxCapacityBreachBehavior: awstypes.PredictiveScalingMaxCapacityBreachBehavior(tfMap["max_capacity_breach_behavior"].(string)),
		MetricSpecifications:      expandPredictiveScalingMetricSpecifications(tfMap["metric_specification"].([]any)),
		Mode:                      awstypes.PredictiveScalingMode(tfMap[names.AttrMode].(string)),
	}
	if v, null, _ := nullable.Int(tfMap["max_capacity_buffer"].(string)).ValueInt32(); !null {
		apiObject.MaxCapacityBuffer = aws.Int32(v)
	}
	if v, null, _ := nullable.Int(tfMap["scheduling_buffer_time"].(string)).ValueInt32(); !null {
		apiObject.SchedulingBufferTime = aws.Int32(v)
	}

	return apiObject
}

func expandPredictiveScalingMetricSpecifications(tfList []any) []awstypes.PredictiveScalingMetricSpecification {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := awstypes.PredictiveScalingMetricSpecification{
		CustomizedCapacityMetricSpecification: expandPredictiveScalingCustomizedCapacityMetric(tfMap["customized_capacity_metric_specification"].([]any)),
		CustomizedLoadMetricSpecification:     expandPredictiveScalingCustomizedLoadMetric(tfMap["customized_load_metric_specification"].([]any)),
		CustomizedScalingMetricSpecification:  expandPredictiveScalingCustomizedScalingMetric(tfMap["customized_scaling_metric_specification"].([]any)),
		PredefinedLoadMetricSpecification:     expandPredictiveScalingPredefinedLoadMetric(tfMap["predefined_load_metric_specification"].([]any)),
		PredefinedMetricPairSpecification:     expandPredictiveScalingPredefinedMetricPair(tfMap["predefined_metric_pair_specification"].([]any)),
		PredefinedScalingMetricSpecification:  expandPredictiveScalingPredefinedScalingMetric(tfMap["predefined_scaling_metric_specification"].([]any)),
		TargetValue:                           aws.Float64(tfMap["target_value"].(float64)),
	}

	return []awstypes.PredictiveScalingMetricSpecification{apiObject}
}

func expandPredictiveScalingPredefinedLoadMetric(tfList []any) *awstypes.PredictiveScalingPredefinedLoadMetric {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.PredictiveScalingPredefinedLoadMetric{
		PredefinedMetricType: awstypes.PredefinedLoadMetricType(tfMap["predefined_metric_type"].(string)),
	}
	if v, ok := tfMap["resource_label"].(string); ok && v != "" {
		apiObject.ResourceLabel = aws.String(v)
	}

	return apiObject
}

func expandPredictiveScalingPredefinedMetricPair(tfList []any) *awstypes.PredictiveScalingPredefinedMetricPair {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.PredictiveScalingPredefinedMetricPair{
		PredefinedMetricType: awstypes.PredefinedMetricPairType(tfMap["predefined_metric_type"].(string)),
	}
	if v, ok := tfMap["resource_label"].(string); ok && v != "" {
		apiObject.ResourceLabel = aws.String(v)
	}

	return apiObject
}

func expandPredictiveScalingPredefinedScalingMetric(tfList []any) *awstypes.PredictiveScalingPredefinedScalingMetric {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.PredictiveScalingPredefinedScalingMetric{
		PredefinedMetricType: awstypes.PredefinedScalingMetricType(tfMap["predefined_metric_type"].(string)),
	}
	if v, ok := tfMap["resource_label"].(string); ok && v != "" {
		apiObject.ResourceLabel = aws.String(v)
	}

	return apiObject
}

func expandPredictiveScalingCustomizedScalingMetric(tfList []any) *awstypes.PredictiveScalingCustomizedScalingMetric {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.PredictiveScalingCustomizedScalingMetric{
		MetricDataQueries: expandMetricDataQueries(tfMap["metric_data_queries"].([]any)),
	}

	return apiObject
}

func expandPredictiveScalingCustomizedLoadMetric(tfList []any) *awstypes.PredictiveScalingCustomizedLoadMetric {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.PredictiveScalingCustomizedLoadMetric{
		MetricDataQueries: expandMetricDataQueries(tfMap["metric_data_queries"].([]any)),
	}

	return apiObject
}

func expandPredictiveScalingCustomizedCapacityMetric(tfList []any) *awstypes.PredictiveScalingCustomizedCapacityMetric {
	if len(tfList) < 1 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.PredictiveScalingCustomizedCapacityMetric{
		MetricDataQueries: expandMetricDataQueries(tfMap["metric_data_queries"].([]any)),
	}

	return apiObject
}

func expandMetricDataQueries(tfList []any) []awstypes.MetricDataQuery {
	if len(tfList) < 1 {
		return nil
	}

	apiObjects := make([]awstypes.MetricDataQuery, len(tfList))

	for i := range apiObjects {
		tfMap := tfList[i].(map[string]any)
		apiObject := awstypes.MetricDataQuery{
			Id: aws.String(tfMap[names.AttrID].(string)),
		}
		if v, ok := tfMap[names.AttrExpression]; ok && v.(string) != "" {
			apiObject.Expression = aws.String(v.(string))
		}
		if v, ok := tfMap["label"]; ok && v.(string) != "" {
			apiObject.Label = aws.String(v.(string))
		}
		if v, ok := tfMap["metric_stat"]; ok && len(v.([]any)) > 0 {
			tfMapMetricStat := v.([]any)[0].(map[string]any)
			tfMapMetric := tfMapMetricStat["metric"].([]any)[0].(map[string]any)
			metric := &awstypes.Metric{
				MetricName: aws.String(tfMapMetric[names.AttrMetricName].(string)),
				Namespace:  aws.String(tfMapMetric[names.AttrNamespace].(string)),
			}
			if v, ok := tfMapMetric["dimensions"]; ok {
				tfList := v.(*schema.Set).List()
				metricDimensions := make([]awstypes.MetricDimension, len(tfList))
				for i := range metricDimensions {
					tfMap := tfList[i].(map[string]any)
					metricDimensions[i] = awstypes.MetricDimension{
						Name:  aws.String(tfMap[names.AttrName].(string)),
						Value: aws.String(tfMap[names.AttrValue].(string)),
					}
				}
				metric.Dimensions = metricDimensions
			}
			metricStat := &awstypes.MetricStat{
				Metric: metric,
				Stat:   aws.String(tfMapMetricStat["stat"].(string)),
			}
			if v, ok := tfMapMetricStat[names.AttrUnit]; ok && len(v.(string)) > 0 {
				metricStat.Unit = aws.String(v.(string))
			}
			apiObject.MetricStat = metricStat
		}
		if v, ok := tfMap["return_data"]; ok {
			apiObject.ReturnData = aws.Bool(v.(bool))
		}
		apiObjects[i] = apiObject
	}

	return apiObjects
}

func expandStepAdjustments(tfList []any) []awstypes.StepAdjustment {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.StepAdjustment

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.StepAdjustment{
			ScalingAdjustment: aws.Int32(int32(tfMap["scaling_adjustment"].(int))),
		}

		if v, ok := tfMap["metric_interval_lower_bound"].(string); ok {
			if v, null, _ := nullable.Float(v).ValueFloat64(); !null {
				apiObject.MetricIntervalLowerBound = aws.Float64(v)
			}
		}

		if v, ok := tfMap["metric_interval_upper_bound"].(string); ok {
			if v, null, _ := nullable.Float(v).ValueFloat64(); !null {
				apiObject.MetricIntervalUpperBound = aws.Float64(v)
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenTargetTrackingConfiguration(apiObject *awstypes.TargetTrackingConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}
	if apiObject := apiObject.CustomizedMetricSpecification; apiObject != nil {
		tfMapCustomizedMetricSpecification := map[string]any{}
		if v := apiObject.Metrics; v != nil {
			tfMapCustomizedMetricSpecification["metrics"] = flattenTargetTrackingMetricDataQueries(v)
		} else {
			if apiObjects := apiObject.Dimensions; apiObjects != nil {
				tfList := make([]any, len(apiObjects))
				for i := range tfList {
					tfMap := map[string]any{}
					apiObject := apiObjects[i]
					tfMap[names.AttrName] = aws.ToString(apiObject.Name)
					tfMap[names.AttrValue] = aws.ToString(apiObject.Value)
					tfList[i] = tfMap
				}
				tfMapCustomizedMetricSpecification["metric_dimension"] = tfList
			}
			tfMapCustomizedMetricSpecification[names.AttrMetricName] = aws.ToString(apiObject.MetricName)
			tfMapCustomizedMetricSpecification[names.AttrNamespace] = aws.ToString(apiObject.Namespace)
			if v := apiObject.Period; v != nil {
				tfMapCustomizedMetricSpecification["period"] = aws.ToInt32(v)
			}
			tfMapCustomizedMetricSpecification["statistic"] = apiObject.Statistic
			if v := apiObject.Unit; v != nil {
				tfMapCustomizedMetricSpecification[names.AttrUnit] = aws.ToString(v)
			}
		}
		tfMap["customized_metric_specification"] = []map[string]any{tfMapCustomizedMetricSpecification}
	}
	tfMap["disable_scale_in"] = aws.ToBool(apiObject.DisableScaleIn)
	if apiObject := apiObject.PredefinedMetricSpecification; apiObject != nil {
		tfMapPredefinedMetricSpecification := map[string]any{}
		tfMapPredefinedMetricSpecification["predefined_metric_type"] = apiObject.PredefinedMetricType
		if v := apiObject.ResourceLabel; v != nil {
			tfMapPredefinedMetricSpecification["resource_label"] = aws.ToString(v)
		}
		tfMap["predefined_metric_specification"] = []map[string]any{tfMapPredefinedMetricSpecification}
	}
	tfMap["target_value"] = aws.ToFloat64(apiObject.TargetValue)

	return []any{tfMap}
}

func flattenTargetTrackingMetricDataQueries(apiObjects []awstypes.TargetTrackingMetricDataQuery) []any {
	tfList := make([]any, len(apiObjects))

	for i := range tfList {
		tfMap := map[string]any{}
		apiObject := apiObjects[i]
		if v := apiObject.Expression; v != nil {
			tfMap[names.AttrExpression] = aws.ToString(v)
		}
		tfMap[names.AttrID] = aws.ToString(apiObject.Id)
		if v := apiObject.Label; v != nil {
			tfMap["label"] = aws.ToString(v)
		}
		if apiObject := apiObject.MetricStat; apiObject != nil {
			tfMapMetricStat := map[string]any{}
			if apiObject := apiObject.Metric; apiObject != nil {
				tfMapMetric := map[string]any{}
				if apiObjects := apiObject.Dimensions; apiObjects != nil {
					tfList := make([]any, len(apiObjects))
					for i := range tfList {
						tfMap := map[string]any{}
						apiObject := apiObject.Dimensions[i]
						tfMap[names.AttrName] = aws.ToString(apiObject.Name)
						tfMap[names.AttrValue] = aws.ToString(apiObject.Value)
						tfList[i] = tfMap
					}
					tfMapMetric["dimensions"] = tfList
				}
				tfMapMetric[names.AttrMetricName] = aws.ToString(apiObject.MetricName)
				tfMapMetric[names.AttrNamespace] = aws.ToString(apiObject.Namespace)
				tfMapMetricStat["metric"] = []map[string]any{tfMapMetric}
			}
			if v := apiObject.Period; v != nil {
				tfMapMetricStat["period"] = aws.ToInt32(v)
			}
			tfMapMetricStat["stat"] = aws.ToString(apiObject.Stat)
			if v := apiObject.Unit; v != nil {
				tfMapMetricStat[names.AttrUnit] = aws.ToString(v)
			}
			tfMap["metric_stat"] = []map[string]any{tfMapMetricStat}
		}
		if v := apiObject.ReturnData; v != nil {
			tfMap["return_data"] = aws.ToBool(v)
		}
		tfList[i] = tfMap
	}

	return tfList
}

func flattenPredictiveScalingConfiguration(apiObject *awstypes.PredictiveScalingConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}
	tfMap["max_capacity_breach_behavior"] = string(apiObject.MaxCapacityBreachBehavior)
	if v := apiObject.MaxCapacityBuffer; v != nil {
		tfMap["max_capacity_buffer"] = flex.Int32ToStringValue(v)
	}
	if v := apiObject.MetricSpecifications; len(v) > 0 {
		tfMap["metric_specification"] = flattenPredictiveScalingMetricSpecifications(v)
	}
	tfMap[names.AttrMode] = apiObject.Mode
	if v := apiObject.SchedulingBufferTime; v != nil {
		tfMap["scheduling_buffer_time"] = flex.Int32ToStringValue(v)
	}

	return []any{tfMap}
}

func flattenPredictiveScalingMetricSpecifications(apiObjects []awstypes.PredictiveScalingMetricSpecification) []any {
	tfMap := map[string]any{}
	if len(apiObjects) < 1 {
		return []any{tfMap}
	}

	apiObject := apiObjects[0]
	if v := apiObject.CustomizedCapacityMetricSpecification; v != nil {
		tfMap["customized_capacity_metric_specification"] = flattenPredictiveScalingCustomizedCapacityMetric(v)
	}
	if v := apiObject.CustomizedLoadMetricSpecification; v != nil {
		tfMap["customized_load_metric_specification"] = flattenPredictiveScalingCustomizedLoadMetric(v)
	}
	if v := apiObject.CustomizedScalingMetricSpecification; v != nil {
		tfMap["customized_scaling_metric_specification"] = flattenPredictiveScalingCustomizedScalingMetric(v)
	}
	if v := apiObject.PredefinedLoadMetricSpecification; v != nil {
		tfMap["predefined_load_metric_specification"] = flattenPredictiveScalingPredefinedLoadMetric(v)
	}
	if v := apiObject.PredefinedMetricPairSpecification; v != nil {
		tfMap["predefined_metric_pair_specification"] = flattenPredictiveScalingPredefinedMetricPair(v)
	}
	if v := apiObject.PredefinedScalingMetricSpecification; v != nil {
		tfMap["predefined_scaling_metric_specification"] = flattenPredictiveScalingPredefinedScalingMetric(v)
	}
	if v := apiObject.TargetValue; v != nil {
		tfMap["target_value"] = aws.ToFloat64(v)
	}

	return []any{tfMap}
}

func flattenPredictiveScalingPredefinedScalingMetric(apiObject *awstypes.PredictiveScalingPredefinedScalingMetric) []any {
	tfMap := map[string]any{}
	if apiObject == nil {
		return []any{tfMap}
	}

	tfMap["predefined_metric_type"] = apiObject.PredefinedMetricType
	tfMap["resource_label"] = aws.ToString(apiObject.ResourceLabel)

	return []any{tfMap}
}

func flattenPredictiveScalingPredefinedLoadMetric(apiObject *awstypes.PredictiveScalingPredefinedLoadMetric) []any {
	tfMap := map[string]any{}
	if apiObject == nil {
		return []any{tfMap}
	}

	tfMap["predefined_metric_type"] = apiObject.PredefinedMetricType
	tfMap["resource_label"] = aws.ToString(apiObject.ResourceLabel)

	return []any{tfMap}
}

func flattenPredictiveScalingPredefinedMetricPair(apiObject *awstypes.PredictiveScalingPredefinedMetricPair) []any {
	tfMap := map[string]any{}
	if apiObject == nil {
		return []any{tfMap}
	}

	tfMap["predefined_metric_type"] = apiObject.PredefinedMetricType
	tfMap["resource_label"] = aws.ToString(apiObject.ResourceLabel)

	return []any{tfMap}
}

func flattenPredictiveScalingCustomizedScalingMetric(apiObject *awstypes.PredictiveScalingCustomizedScalingMetric) []any {
	tfMap := map[string]any{}
	if apiObject == nil {
		return []any{tfMap}
	}

	tfMap["metric_data_queries"] = flattenMetricDataQueries(apiObject.MetricDataQueries)

	return []any{tfMap}
}

func flattenPredictiveScalingCustomizedLoadMetric(apiObject *awstypes.PredictiveScalingCustomizedLoadMetric) []any {
	tfMap := map[string]any{}
	if apiObject == nil {
		return []any{tfMap}
	}

	tfMap["metric_data_queries"] = flattenMetricDataQueries(apiObject.MetricDataQueries)

	return []any{tfMap}
}

func flattenPredictiveScalingCustomizedCapacityMetric(apiObject *awstypes.PredictiveScalingCustomizedCapacityMetric) []any {
	tfMap := map[string]any{}
	if apiObject == nil {
		return []any{tfMap}
	}

	tfMap["metric_data_queries"] = flattenMetricDataQueries(apiObject.MetricDataQueries)

	return []any{tfMap}
}

func flattenMetricDataQueries(apiObjects []awstypes.MetricDataQuery) []any {
	tfList := make([]any, len(apiObjects))

	for i := range tfList {
		tfMap := map[string]any{}
		apiObject := apiObjects[i]

		if v := apiObject.Expression; v != nil {
			tfMap[names.AttrExpression] = aws.ToString(v)
		}
		tfMap[names.AttrID] = aws.ToString(apiObject.Id)
		if v := apiObject.Label; v != nil {
			tfMap["label"] = aws.ToString(v)
		}
		if apiObject := apiObject.MetricStat; apiObject != nil {
			tfMapMetricStat := map[string]any{}
			if apiObject := apiObject.Metric; apiObject != nil {
				tfMapMetric := map[string]any{}
				if apiObjects := apiObject.Dimensions; apiObjects != nil {
					tfList := make([]any, len(apiObjects))
					for i := range tfList {
						tfMap := map[string]any{}
						apiObject := apiObject.Dimensions[i]
						tfMap[names.AttrName] = aws.ToString(apiObject.Name)
						tfMap[names.AttrValue] = aws.ToString(apiObject.Value)
						tfList[i] = tfMap
					}
					tfMapMetric["dimensions"] = tfList
				}
				tfMapMetric[names.AttrMetricName] = aws.ToString(apiObject.MetricName)
				tfMapMetric[names.AttrNamespace] = aws.ToString(apiObject.Namespace)
				tfMapMetricStat["metric"] = []map[string]any{tfMapMetric}
			}
			tfMapMetricStat["stat"] = aws.ToString(apiObject.Stat)
			if v := apiObject.Unit; v != nil {
				tfMapMetricStat[names.AttrUnit] = aws.ToString(v)
			}
			tfMap["metric_stat"] = []map[string]any{tfMapMetricStat}
		}
		if v := apiObject.ReturnData; v != nil {
			tfMap["return_data"] = aws.ToBool(v)
		}
		tfList[i] = tfMap
	}

	return tfList
}

func flattenStepAdjustments(apiObjects []awstypes.StepAdjustment) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"scaling_adjustment": aws.ToInt32(apiObject.ScalingAdjustment),
		}

		if v := apiObject.MetricIntervalLowerBound; v != nil {
			tfMap["metric_interval_lower_bound"] = flex.Float64ToStringValue(v)
		}

		if v := apiObject.MetricIntervalUpperBound; v != nil {
			tfMap["metric_interval_upper_bound"] = flex.Float64ToStringValue(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
