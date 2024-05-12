// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
					Set: resourceScalingAdjustmentHash,
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

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	name := d.Get(names.AttrName).(string)
	input, err := expandPutScalingPolicyInput(d)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Policy (%s): %s", name, err)
	}

	_, err = conn.PutScalingPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Policy (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	d.Set(names.AttrName, p.PolicyName)
	d.Set("policy_type", p.PolicyType)
	d.Set("min_adjustment_magnitude", p.MinAdjustmentMagnitude)

	d.Set("scaling_adjustment", p.ScalingAdjustment)
	if err := d.Set("predictive_scaling_configuration", flattenPredictiveScalingConfig(p.PredictiveScalingConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting predictive_scaling_configuration: %s", err)
	}
	if err := d.Set("step_adjustment", flattenStepAdjustments(p.StepAdjustments)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting step_adjustment: %s", err)
	}
	if err := d.Set("target_tracking_configuration", flattenTargetTrackingConfiguration(p.TargetTrackingConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_tracking_configuration: %s", err)
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	input, err := expandPutScalingPolicyInput(d)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Auto Scaling Policy (%s): %s", d.Id(), err)
	}

	_, err = conn.PutScalingPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Auto Scaling Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	log.Printf("[INFO] Deleting Auto Scaling Policy: %s", d.Id())
	_, err := conn.DeletePolicy(ctx, &autoscaling.DeletePolicyInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		PolicyName:           aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func resourcePolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	if predictiveScalingConfigFlat := d.Get("predictive_scaling_configuration").([]interface{}); len(predictiveScalingConfigFlat) > 0 {
		input.PredictiveScalingConfiguration = expandPredictiveScalingConfig(predictiveScalingConfigFlat)
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
		input.TargetTrackingConfiguration = expandTargetTrackingConfiguration(v.([]interface{}))
		if policyType != policyTypeTargetTrackingScaling {
			return input, fmt.Errorf("target_tracking_configuration is only supported for policy type TargetTrackingScaling")
		}
	} else if !ok && policyType == policyTypeTargetTrackingScaling {
		return input, fmt.Errorf("target_tracking_configuration is required for policy type TargetTrackingScaling")
	}

	return input, nil
}

func resourceScalingAdjustmentHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["metric_interval_lower_bound"]; ok {
		buf.WriteString(fmt.Sprintf("%f-", v))
	}
	if v, ok := m["metric_interval_upper_bound"]; ok {
		buf.WriteString(fmt.Sprintf("%f-", v))
	}
	buf.WriteString(fmt.Sprintf("%d-", m["scaling_adjustment"].(int)))

	return create.StringHashcode(buf.String())
}

func expandTargetTrackingConfiguration(configs []interface{}) *awstypes.TargetTrackingConfiguration {
	if len(configs) < 1 {
		return nil
	}

	config := configs[0].(map[string]interface{})

	result := &awstypes.TargetTrackingConfiguration{}

	result.TargetValue = aws.Float64(config["target_value"].(float64))
	if v, ok := config["disable_scale_in"]; ok {
		result.DisableScaleIn = aws.Bool(v.(bool))
	}
	if v, ok := config["predefined_metric_specification"]; ok && len(v.([]interface{})) > 0 {
		spec := v.([]interface{})[0].(map[string]interface{})
		predSpec := &awstypes.PredefinedMetricSpecification{
			PredefinedMetricType: awstypes.MetricType(spec["predefined_metric_type"].(string)),
		}
		if val, ok := spec["resource_label"]; ok && val.(string) != "" {
			predSpec.ResourceLabel = aws.String(val.(string))
		}
		result.PredefinedMetricSpecification = predSpec
	}
	if v, ok := config["customized_metric_specification"]; ok && len(v.([]interface{})) > 0 {
		spec := v.([]interface{})[0].(map[string]interface{})
		customSpec := &awstypes.CustomizedMetricSpecification{}
		if val, ok := spec["metrics"].(*schema.Set); ok && val.Len() > 0 {
			customSpec.Metrics = expandTargetTrackingMetricDataQueries(val.List())
		} else {
			customSpec.Namespace = aws.String(spec[names.AttrNamespace].(string))
			customSpec.MetricName = aws.String(spec[names.AttrMetricName].(string))
			customSpec.Statistic = awstypes.MetricStatistic(spec["statistic"].(string))
			if val, ok := spec[names.AttrUnit]; ok && len(val.(string)) > 0 {
				customSpec.Unit = aws.String(val.(string))
			}
			if val, ok := spec["metric_dimension"]; ok {
				dims := val.([]interface{})
				metDimList := make([]awstypes.MetricDimension, len(dims))
				for i := range metDimList {
					dim := dims[i].(map[string]interface{})
					md := awstypes.MetricDimension{
						Name:  aws.String(dim[names.AttrName].(string)),
						Value: aws.String(dim[names.AttrValue].(string)),
					}
					metDimList[i] = md
				}
				customSpec.Dimensions = metDimList
			}
		}
		result.CustomizedMetricSpecification = customSpec
	}
	return result
}

func expandTargetTrackingMetricDataQueries(metricDataQuerySlices []interface{}) []awstypes.TargetTrackingMetricDataQuery {
	if metricDataQuerySlices == nil || len(metricDataQuerySlices) < 1 {
		return nil
	}
	metricDataQueries := make([]awstypes.TargetTrackingMetricDataQuery, len(metricDataQuerySlices))

	for i := range metricDataQueries {
		metricDataQueryFlat := metricDataQuerySlices[i].(map[string]interface{})
		metricDataQuery := awstypes.TargetTrackingMetricDataQuery{
			Id: aws.String(metricDataQueryFlat[names.AttrID].(string)),
		}
		if val, ok := metricDataQueryFlat["metric_stat"]; ok && len(val.([]interface{})) > 0 {
			metricStatSpec := val.([]interface{})[0].(map[string]interface{})
			metricSpec := metricStatSpec["metric"].([]interface{})[0].(map[string]interface{})
			metric := &awstypes.Metric{
				MetricName: aws.String(metricSpec[names.AttrMetricName].(string)),
				Namespace:  aws.String(metricSpec[names.AttrNamespace].(string)),
			}
			if v, ok := metricSpec["dimensions"]; ok {
				dims := v.(*schema.Set).List()
				dimList := make([]awstypes.MetricDimension, len(dims))
				for i := range dimList {
					dim := dims[i].(map[string]interface{})
					md := awstypes.MetricDimension{
						Name:  aws.String(dim[names.AttrName].(string)),
						Value: aws.String(dim[names.AttrValue].(string)),
					}
					dimList[i] = md
				}
				metric.Dimensions = dimList
			}
			metricStat := &awstypes.TargetTrackingMetricStat{
				Metric: metric,
				Stat:   aws.String(metricStatSpec["stat"].(string)),
			}
			if v, ok := metricStatSpec[names.AttrUnit]; ok && len(v.(string)) > 0 {
				metricStat.Unit = aws.String(v.(string))
			}
			metricDataQuery.MetricStat = metricStat
		}
		if val, ok := metricDataQueryFlat[names.AttrExpression]; ok && val.(string) != "" {
			metricDataQuery.Expression = aws.String(val.(string))
		}
		if val, ok := metricDataQueryFlat["label"]; ok && val.(string) != "" {
			metricDataQuery.Label = aws.String(val.(string))
		}
		if val, ok := metricDataQueryFlat["return_data"]; ok {
			metricDataQuery.ReturnData = aws.Bool(val.(bool))
		}
		metricDataQueries[i] = metricDataQuery
	}
	return metricDataQueries
}

func expandPredictiveScalingConfig(predictiveScalingConfigSlice []interface{}) *awstypes.PredictiveScalingConfiguration {
	if predictiveScalingConfigSlice == nil || len(predictiveScalingConfigSlice) < 1 {
		return nil
	}
	predictiveScalingConfigFlat := predictiveScalingConfigSlice[0].(map[string]interface{})
	predictiveScalingConfig := &awstypes.PredictiveScalingConfiguration{
		MetricSpecifications:      expandPredictiveScalingMetricSpecifications(predictiveScalingConfigFlat["metric_specification"].([]interface{})),
		MaxCapacityBreachBehavior: awstypes.PredictiveScalingMaxCapacityBreachBehavior(predictiveScalingConfigFlat["max_capacity_breach_behavior"].(string)),
		Mode:                      awstypes.PredictiveScalingMode(predictiveScalingConfigFlat[names.AttrMode].(string)),
	}
	if v, null, _ := nullable.Int(predictiveScalingConfigFlat["max_capacity_buffer"].(string)).ValueInt32(); !null {
		predictiveScalingConfig.MaxCapacityBuffer = aws.Int32(v)
	}
	if v, null, _ := nullable.Int(predictiveScalingConfigFlat["scheduling_buffer_time"].(string)).ValueInt32(); !null {
		predictiveScalingConfig.SchedulingBufferTime = aws.Int32(v)
	}
	return predictiveScalingConfig
}

func expandPredictiveScalingMetricSpecifications(metricSpecificationsSlice []interface{}) []awstypes.PredictiveScalingMetricSpecification {
	if metricSpecificationsSlice == nil || len(metricSpecificationsSlice) < 1 {
		return nil
	}
	metricSpecificationsFlat := metricSpecificationsSlice[0].(map[string]interface{})
	metricSpecification := awstypes.PredictiveScalingMetricSpecification{
		CustomizedCapacityMetricSpecification: expandCustomizedCapacityMetricSpecification(metricSpecificationsFlat["customized_capacity_metric_specification"].([]interface{})),
		CustomizedLoadMetricSpecification:     expandCustomizedLoadMetricSpecification(metricSpecificationsFlat["customized_load_metric_specification"].([]interface{})),
		CustomizedScalingMetricSpecification:  expandCustomizedScalingMetricSpecification(metricSpecificationsFlat["customized_scaling_metric_specification"].([]interface{})),
		PredefinedLoadMetricSpecification:     expandPredefinedLoadMetricSpecification(metricSpecificationsFlat["predefined_load_metric_specification"].([]interface{})),
		PredefinedMetricPairSpecification:     expandPredefinedMetricPairSpecification(metricSpecificationsFlat["predefined_metric_pair_specification"].([]interface{})),
		PredefinedScalingMetricSpecification:  expandPredefinedScalingMetricSpecification(metricSpecificationsFlat["predefined_scaling_metric_specification"].([]interface{})),
		TargetValue:                           aws.Float64(metricSpecificationsFlat["target_value"].(float64)),
	}
	return []awstypes.PredictiveScalingMetricSpecification{metricSpecification}
}

func expandPredefinedLoadMetricSpecification(predefinedLoadMetricSpecificationSlice []interface{}) *awstypes.PredictiveScalingPredefinedLoadMetric {
	if predefinedLoadMetricSpecificationSlice == nil || len(predefinedLoadMetricSpecificationSlice) < 1 {
		return nil
	}
	predefinedLoadMetricSpecificationFlat := predefinedLoadMetricSpecificationSlice[0].(map[string]interface{})
	predefinedLoadMetricSpecification := &awstypes.PredictiveScalingPredefinedLoadMetric{
		PredefinedMetricType: awstypes.PredefinedLoadMetricType(predefinedLoadMetricSpecificationFlat["predefined_metric_type"].(string)),
	}
	if label, ok := predefinedLoadMetricSpecificationFlat["resource_label"].(string); ok && label != "" {
		predefinedLoadMetricSpecification.ResourceLabel = aws.String(label)
	}
	return predefinedLoadMetricSpecification
}

func expandPredefinedMetricPairSpecification(predefinedMetricPairSpecificationSlice []interface{}) *awstypes.PredictiveScalingPredefinedMetricPair {
	if predefinedMetricPairSpecificationSlice == nil || len(predefinedMetricPairSpecificationSlice) < 1 {
		return nil
	}
	predefinedMetricPairSpecificationFlat := predefinedMetricPairSpecificationSlice[0].(map[string]interface{})
	predefinedMetricPairSpecification := &awstypes.PredictiveScalingPredefinedMetricPair{
		PredefinedMetricType: awstypes.PredefinedMetricPairType(predefinedMetricPairSpecificationFlat["predefined_metric_type"].(string)),
	}
	if label, ok := predefinedMetricPairSpecificationFlat["resource_label"].(string); ok && label != "" {
		predefinedMetricPairSpecification.ResourceLabel = aws.String(label)
	}
	return predefinedMetricPairSpecification
}

func expandPredefinedScalingMetricSpecification(predefinedScalingMetricSpecificationSlice []interface{}) *awstypes.PredictiveScalingPredefinedScalingMetric {
	if predefinedScalingMetricSpecificationSlice == nil || len(predefinedScalingMetricSpecificationSlice) < 1 {
		return nil
	}
	predefinedScalingMetricSpecificationFlat := predefinedScalingMetricSpecificationSlice[0].(map[string]interface{})
	predefinedScalingMetricSpecification := &awstypes.PredictiveScalingPredefinedScalingMetric{
		PredefinedMetricType: awstypes.PredefinedScalingMetricType(predefinedScalingMetricSpecificationFlat["predefined_metric_type"].(string)),
	}
	if label, ok := predefinedScalingMetricSpecificationFlat["resource_label"].(string); ok && label != "" {
		predefinedScalingMetricSpecification.ResourceLabel = aws.String(label)
	}
	return predefinedScalingMetricSpecification
}

func expandCustomizedScalingMetricSpecification(customizedScalingMetricSpecificationSlice []interface{}) *awstypes.PredictiveScalingCustomizedScalingMetric {
	if customizedScalingMetricSpecificationSlice == nil || len(customizedScalingMetricSpecificationSlice) < 1 {
		return nil
	}
	customizedScalingMetricSpecificationFlat := customizedScalingMetricSpecificationSlice[0].(map[string]interface{})
	customizedScalingMetricSpecification := &awstypes.PredictiveScalingCustomizedScalingMetric{
		MetricDataQueries: expandMetricDataQueries(customizedScalingMetricSpecificationFlat["metric_data_queries"].([]interface{})),
	}
	return customizedScalingMetricSpecification
}

func expandCustomizedLoadMetricSpecification(customizedLoadMetricSpecificationSlice []interface{}) *awstypes.PredictiveScalingCustomizedLoadMetric {
	if customizedLoadMetricSpecificationSlice == nil || len(customizedLoadMetricSpecificationSlice) < 1 {
		return nil
	}
	customizedLoadMetricSpecificationSliceFlat := customizedLoadMetricSpecificationSlice[0].(map[string]interface{})
	customizedLoadMetricSpecification := &awstypes.PredictiveScalingCustomizedLoadMetric{
		MetricDataQueries: expandMetricDataQueries(customizedLoadMetricSpecificationSliceFlat["metric_data_queries"].([]interface{})),
	}
	return customizedLoadMetricSpecification
}

func expandCustomizedCapacityMetricSpecification(customizedCapacityMetricSlice []interface{}) *awstypes.PredictiveScalingCustomizedCapacityMetric {
	if customizedCapacityMetricSlice == nil || len(customizedCapacityMetricSlice) < 1 {
		return nil
	}
	customizedCapacityMetricSliceFlat := customizedCapacityMetricSlice[0].(map[string]interface{})
	customizedCapacityMetricSpecification := &awstypes.PredictiveScalingCustomizedCapacityMetric{
		MetricDataQueries: expandMetricDataQueries(customizedCapacityMetricSliceFlat["metric_data_queries"].([]interface{})),
	}
	return customizedCapacityMetricSpecification
}

func expandMetricDataQueries(metricDataQuerySlices []interface{}) []awstypes.MetricDataQuery {
	if metricDataQuerySlices == nil || len(metricDataQuerySlices) < 1 {
		return nil
	}
	metricDataQueries := make([]awstypes.MetricDataQuery, len(metricDataQuerySlices))

	for i := range metricDataQueries {
		metricDataQueryFlat := metricDataQuerySlices[i].(map[string]interface{})
		metricDataQuery := awstypes.MetricDataQuery{
			Id: aws.String(metricDataQueryFlat[names.AttrID].(string)),
		}
		if val, ok := metricDataQueryFlat["metric_stat"]; ok && len(val.([]interface{})) > 0 {
			metricStatSpec := val.([]interface{})[0].(map[string]interface{})
			metricSpec := metricStatSpec["metric"].([]interface{})[0].(map[string]interface{})
			metric := &awstypes.Metric{
				MetricName: aws.String(metricSpec[names.AttrMetricName].(string)),
				Namespace:  aws.String(metricSpec[names.AttrNamespace].(string)),
			}
			if v, ok := metricSpec["dimensions"]; ok {
				dims := v.(*schema.Set).List()
				dimList := make([]awstypes.MetricDimension, len(dims))
				for i := range dimList {
					dim := dims[i].(map[string]interface{})
					md := awstypes.MetricDimension{
						Name:  aws.String(dim[names.AttrName].(string)),
						Value: aws.String(dim[names.AttrValue].(string)),
					}
					dimList[i] = md
				}
				metric.Dimensions = dimList
			}
			metricStat := &awstypes.MetricStat{
				Metric: metric,
				Stat:   aws.String(metricStatSpec["stat"].(string)),
			}
			if v, ok := metricStatSpec[names.AttrUnit]; ok && len(v.(string)) > 0 {
				metricStat.Unit = aws.String(v.(string))
			}
			metricDataQuery.MetricStat = metricStat
		}
		if val, ok := metricDataQueryFlat[names.AttrExpression]; ok && val.(string) != "" {
			metricDataQuery.Expression = aws.String(val.(string))
		}
		if val, ok := metricDataQueryFlat["label"]; ok && val.(string) != "" {
			metricDataQuery.Label = aws.String(val.(string))
		}
		if val, ok := metricDataQueryFlat["return_data"]; ok {
			metricDataQuery.ReturnData = aws.Bool(val.(bool))
		}
		metricDataQueries[i] = metricDataQuery
	}
	return metricDataQueries
}

func flattenTargetTrackingConfiguration(config *awstypes.TargetTrackingConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{}
	result["disable_scale_in"] = aws.ToBool(config.DisableScaleIn)
	result["target_value"] = aws.ToFloat64(config.TargetValue)
	if config.PredefinedMetricSpecification != nil {
		spec := map[string]interface{}{}
		spec["predefined_metric_type"] = string(config.PredefinedMetricSpecification.PredefinedMetricType)
		if config.PredefinedMetricSpecification.ResourceLabel != nil {
			spec["resource_label"] = aws.ToString(config.PredefinedMetricSpecification.ResourceLabel)
		}
		result["predefined_metric_specification"] = []map[string]interface{}{spec}
	}
	if config.CustomizedMetricSpecification != nil {
		spec := map[string]interface{}{}
		if config.CustomizedMetricSpecification.Metrics != nil {
			spec["metrics"] = flattenTargetTrackingMetricDataQueries(config.CustomizedMetricSpecification.Metrics)
		} else {
			spec[names.AttrMetricName] = aws.ToString(config.CustomizedMetricSpecification.MetricName)
			spec[names.AttrNamespace] = aws.ToString(config.CustomizedMetricSpecification.Namespace)
			spec["statistic"] = string(config.CustomizedMetricSpecification.Statistic)
			if config.CustomizedMetricSpecification.Unit != nil {
				spec[names.AttrUnit] = aws.ToString(config.CustomizedMetricSpecification.Unit)
			}
			if config.CustomizedMetricSpecification.Dimensions != nil {
				dimSpec := make([]interface{}, len(config.CustomizedMetricSpecification.Dimensions))
				for i := range dimSpec {
					dim := map[string]interface{}{}
					rawDim := config.CustomizedMetricSpecification.Dimensions[i]
					dim[names.AttrName] = aws.ToString(rawDim.Name)
					dim[names.AttrValue] = aws.ToString(rawDim.Value)
					dimSpec[i] = dim
				}
				spec["metric_dimension"] = dimSpec
			}
		}
		result["customized_metric_specification"] = []map[string]interface{}{spec}
	}
	return []interface{}{result}
}

func flattenTargetTrackingMetricDataQueries(metricDataQueries []awstypes.TargetTrackingMetricDataQuery) []interface{} {
	metricDataQueriesSpec := make([]interface{}, len(metricDataQueries))
	for i := range metricDataQueriesSpec {
		metricDataQuery := map[string]interface{}{}
		rawMetricDataQuery := metricDataQueries[i]
		metricDataQuery[names.AttrID] = aws.ToString(rawMetricDataQuery.Id)
		if rawMetricDataQuery.Expression != nil {
			metricDataQuery[names.AttrExpression] = aws.ToString(rawMetricDataQuery.Expression)
		}
		if rawMetricDataQuery.Label != nil {
			metricDataQuery["label"] = aws.ToString(rawMetricDataQuery.Label)
		}
		if rawMetricDataQuery.MetricStat != nil {
			metricStatSpec := map[string]interface{}{}
			rawMetricStat := rawMetricDataQuery.MetricStat
			rawMetric := rawMetricStat.Metric
			metricSpec := map[string]interface{}{}
			if rawMetric.Dimensions != nil {
				dimSpec := make([]interface{}, len(rawMetric.Dimensions))
				for i := range dimSpec {
					dim := map[string]interface{}{}
					rawDim := rawMetric.Dimensions[i]
					dim[names.AttrName] = aws.ToString(rawDim.Name)
					dim[names.AttrValue] = aws.ToString(rawDim.Value)
					dimSpec[i] = dim
				}
				metricSpec["dimensions"] = dimSpec
			}
			metricSpec[names.AttrMetricName] = aws.ToString(rawMetric.MetricName)
			metricSpec[names.AttrNamespace] = aws.ToString(rawMetric.Namespace)
			metricStatSpec["metric"] = []map[string]interface{}{metricSpec}
			metricStatSpec["stat"] = aws.ToString(rawMetricStat.Stat)
			if rawMetricStat.Unit != nil {
				metricStatSpec[names.AttrUnit] = aws.ToString(rawMetricStat.Unit)
			}
			metricDataQuery["metric_stat"] = []map[string]interface{}{metricStatSpec}
		}
		if rawMetricDataQuery.ReturnData != nil {
			metricDataQuery["return_data"] = aws.ToBool(rawMetricDataQuery.ReturnData)
		}
		metricDataQueriesSpec[i] = metricDataQuery
	}
	return metricDataQueriesSpec
}

func flattenPredictiveScalingConfig(predictiveScalingConfig *awstypes.PredictiveScalingConfiguration) []map[string]interface{} {
	predictiveScalingConfigFlat := map[string]interface{}{}
	if predictiveScalingConfig == nil {
		return nil
	}
	if predictiveScalingConfig.MetricSpecifications != nil && len(predictiveScalingConfig.MetricSpecifications) > 0 {
		predictiveScalingConfigFlat["metric_specification"] = flattenPredictiveScalingMetricSpecifications(predictiveScalingConfig.MetricSpecifications)
	}
	predictiveScalingConfigFlat[names.AttrMode] = string(predictiveScalingConfig.Mode)
	if predictiveScalingConfig.SchedulingBufferTime != nil {
		predictiveScalingConfigFlat["scheduling_buffer_time"] = strconv.FormatInt(int64(aws.ToInt32(predictiveScalingConfig.SchedulingBufferTime)), 10)
	}
	predictiveScalingConfigFlat["max_capacity_breach_behavior"] = string(predictiveScalingConfig.MaxCapacityBreachBehavior)
	if predictiveScalingConfig.MaxCapacityBuffer != nil {
		predictiveScalingConfigFlat["max_capacity_buffer"] = strconv.FormatInt(int64(aws.ToInt32(predictiveScalingConfig.MaxCapacityBuffer)), 10)
	}
	return []map[string]interface{}{predictiveScalingConfigFlat}
}

func flattenPredictiveScalingMetricSpecifications(metricSpecification []awstypes.PredictiveScalingMetricSpecification) []map[string]interface{} {
	metricSpecificationFlat := map[string]interface{}{}
	if metricSpecification == nil || len(metricSpecification) < 1 {
		return []map[string]interface{}{metricSpecificationFlat}
	}
	if metricSpecification[0].TargetValue != nil {
		metricSpecificationFlat["target_value"] = aws.ToFloat64(metricSpecification[0].TargetValue)
	}
	if metricSpecification[0].CustomizedCapacityMetricSpecification != nil {
		metricSpecificationFlat["customized_capacity_metric_specification"] = flattenCustomizedCapacityMetricSpecification(metricSpecification[0].CustomizedCapacityMetricSpecification)
	}
	if metricSpecification[0].CustomizedLoadMetricSpecification != nil {
		metricSpecificationFlat["customized_load_metric_specification"] = flattenCustomizedLoadMetricSpecification(metricSpecification[0].CustomizedLoadMetricSpecification)
	}
	if metricSpecification[0].CustomizedScalingMetricSpecification != nil {
		metricSpecificationFlat["customized_scaling_metric_specification"] = flattenCustomizedScalingMetricSpecification(metricSpecification[0].CustomizedScalingMetricSpecification)
	}
	if metricSpecification[0].PredefinedLoadMetricSpecification != nil {
		metricSpecificationFlat["predefined_load_metric_specification"] = flattenPredefinedLoadMetricSpecification(metricSpecification[0].PredefinedLoadMetricSpecification)
	}
	if metricSpecification[0].PredefinedMetricPairSpecification != nil {
		metricSpecificationFlat["predefined_metric_pair_specification"] = flattenPredefinedMetricPairSpecification(metricSpecification[0].PredefinedMetricPairSpecification)
	}
	if metricSpecification[0].PredefinedScalingMetricSpecification != nil {
		metricSpecificationFlat["predefined_scaling_metric_specification"] = flattenPredefinedScalingMetricSpecification(metricSpecification[0].PredefinedScalingMetricSpecification)
	}
	return []map[string]interface{}{metricSpecificationFlat}
}

func flattenPredefinedScalingMetricSpecification(predefinedScalingMetricSpecification *awstypes.PredictiveScalingPredefinedScalingMetric) []map[string]interface{} {
	predefinedScalingMetricSpecificationFlat := map[string]interface{}{}
	if predefinedScalingMetricSpecification == nil {
		return []map[string]interface{}{predefinedScalingMetricSpecificationFlat}
	}
	predefinedScalingMetricSpecificationFlat["predefined_metric_type"] = string(predefinedScalingMetricSpecification.PredefinedMetricType)
	predefinedScalingMetricSpecificationFlat["resource_label"] = aws.ToString(predefinedScalingMetricSpecification.ResourceLabel)
	return []map[string]interface{}{predefinedScalingMetricSpecificationFlat}
}

func flattenPredefinedLoadMetricSpecification(predefinedLoadMetricSpecification *awstypes.PredictiveScalingPredefinedLoadMetric) []map[string]interface{} {
	predefinedLoadMetricSpecificationFlat := map[string]interface{}{}
	if predefinedLoadMetricSpecification == nil {
		return []map[string]interface{}{predefinedLoadMetricSpecificationFlat}
	}
	predefinedLoadMetricSpecificationFlat["predefined_metric_type"] = string(predefinedLoadMetricSpecification.PredefinedMetricType)
	predefinedLoadMetricSpecificationFlat["resource_label"] = aws.ToString(predefinedLoadMetricSpecification.ResourceLabel)
	return []map[string]interface{}{predefinedLoadMetricSpecificationFlat}
}

func flattenPredefinedMetricPairSpecification(predefinedMetricPairSpecification *awstypes.PredictiveScalingPredefinedMetricPair) []map[string]interface{} {
	predefinedMetricPairSpecificationFlat := map[string]interface{}{}
	if predefinedMetricPairSpecification == nil {
		return []map[string]interface{}{predefinedMetricPairSpecificationFlat}
	}
	predefinedMetricPairSpecificationFlat["predefined_metric_type"] = string(predefinedMetricPairSpecification.PredefinedMetricType)
	predefinedMetricPairSpecificationFlat["resource_label"] = aws.ToString(predefinedMetricPairSpecification.ResourceLabel)
	return []map[string]interface{}{predefinedMetricPairSpecificationFlat}
}

func flattenCustomizedScalingMetricSpecification(customizedScalingMetricSpecification *awstypes.PredictiveScalingCustomizedScalingMetric) []map[string]interface{} {
	customizedScalingMetricSpecificationFlat := map[string]interface{}{}
	if customizedScalingMetricSpecification == nil {
		return []map[string]interface{}{customizedScalingMetricSpecificationFlat}
	}
	customizedScalingMetricSpecificationFlat["metric_data_queries"] = flattenMetricDataQueries(customizedScalingMetricSpecification.MetricDataQueries)
	return []map[string]interface{}{customizedScalingMetricSpecificationFlat}
}

func flattenCustomizedLoadMetricSpecification(customizedLoadMetricSpecification *awstypes.PredictiveScalingCustomizedLoadMetric) []map[string]interface{} {
	customizedLoadMetricSpecificationFlat := map[string]interface{}{}
	if customizedLoadMetricSpecification == nil {
		return []map[string]interface{}{customizedLoadMetricSpecificationFlat}
	}
	customizedLoadMetricSpecificationFlat["metric_data_queries"] = flattenMetricDataQueries(customizedLoadMetricSpecification.MetricDataQueries)
	return []map[string]interface{}{customizedLoadMetricSpecificationFlat}
}

func flattenCustomizedCapacityMetricSpecification(customizedCapacityMetricSpecification *awstypes.PredictiveScalingCustomizedCapacityMetric) []map[string]interface{} {
	customizedCapacityMetricSpecificationFlat := map[string]interface{}{}
	if customizedCapacityMetricSpecification == nil {
		return []map[string]interface{}{customizedCapacityMetricSpecificationFlat}
	}
	customizedCapacityMetricSpecificationFlat["metric_data_queries"] = flattenMetricDataQueries(customizedCapacityMetricSpecification.MetricDataQueries)

	return []map[string]interface{}{customizedCapacityMetricSpecificationFlat}
}

func flattenMetricDataQueries(metricDataQueries []awstypes.MetricDataQuery) []interface{} {
	metricDataQueriesSpec := make([]interface{}, len(metricDataQueries))
	for i := range metricDataQueriesSpec {
		metricDataQuery := map[string]interface{}{}
		rawMetricDataQuery := metricDataQueries[i]
		metricDataQuery[names.AttrID] = aws.ToString(rawMetricDataQuery.Id)
		if rawMetricDataQuery.Expression != nil {
			metricDataQuery[names.AttrExpression] = aws.ToString(rawMetricDataQuery.Expression)
		}
		if rawMetricDataQuery.Label != nil {
			metricDataQuery["label"] = aws.ToString(rawMetricDataQuery.Label)
		}
		if rawMetricDataQuery.MetricStat != nil {
			metricStatSpec := map[string]interface{}{}
			rawMetricStat := rawMetricDataQuery.MetricStat
			rawMetric := rawMetricStat.Metric
			metricSpec := map[string]interface{}{}
			if rawMetric.Dimensions != nil {
				dimSpec := make([]interface{}, len(rawMetric.Dimensions))
				for i := range dimSpec {
					dim := map[string]interface{}{}
					rawDim := rawMetric.Dimensions[i]
					dim[names.AttrName] = aws.ToString(rawDim.Name)
					dim[names.AttrValue] = aws.ToString(rawDim.Value)
					dimSpec[i] = dim
				}
				metricSpec["dimensions"] = dimSpec
			}
			metricSpec[names.AttrMetricName] = aws.ToString(rawMetric.MetricName)
			metricSpec[names.AttrNamespace] = aws.ToString(rawMetric.Namespace)
			metricStatSpec["metric"] = []map[string]interface{}{metricSpec}
			metricStatSpec["stat"] = aws.ToString(rawMetricStat.Stat)
			if rawMetricStat.Unit != nil {
				metricStatSpec[names.AttrUnit] = aws.ToString(rawMetricStat.Unit)
			}
			metricDataQuery["metric_stat"] = []map[string]interface{}{metricStatSpec}
		}
		if rawMetricDataQuery.ReturnData != nil {
			metricDataQuery["return_data"] = aws.ToBool(rawMetricDataQuery.ReturnData)
		}
		metricDataQueriesSpec[i] = metricDataQuery
	}
	return metricDataQueriesSpec
}

func expandStepAdjustments(tfList []interface{}) []awstypes.StepAdjustment {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.StepAdjustment

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
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

func flattenStepAdjustments(apiObjects []awstypes.StepAdjustment) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"scaling_adjustment": aws.ToInt32(apiObject.ScalingAdjustment),
		}

		if v := apiObject.MetricIntervalUpperBound; v != nil {
			tfMap["metric_interval_upper_bound"] = flex.Float64ToStringValue(v)
		}

		if v := apiObject.MetricIntervalLowerBound; v != nil {
			tfMap["metric_interval_lower_bound"] = flex.Float64ToStringValue(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
