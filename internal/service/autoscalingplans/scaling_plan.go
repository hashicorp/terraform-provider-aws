// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscalingplans

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscalingplans"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscalingplans/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_autoscalingplans_scaling_plan")
func ResourceScalingPlan() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScalingPlanCreate,
		ReadWithoutTimeout:   resourceScalingPlanRead,
		UpdateWithoutTimeout: resourceScalingPlanUpdate,
		DeleteWithoutTimeout: resourceScalingPlanDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceScalingPlanImport,
		},

		Schema: map[string]*schema.Schema{
			"application_source": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudformation_stack_arn": {
							Type:          schema.TypeString,
							Optional:      true,
							ValidateFunc:  verify.ValidARN,
							ConflictsWith: []string{"application_source.0.tag_filter"},
						},

						"tag_filter": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 0,
							MaxItems: 50,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
										Type:     schema.TypeString,
										Required: true,
									},

									names.AttrValues: {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 0,
										MaxItems: 50,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
							ConflictsWith: []string{"application_source.0.cloudformation_stack_arn"},
						},
					},
				},
			},

			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^[[:print:]]+$`), "must be printable"),
					validation.StringDoesNotContainAny("|:/"),
				),
			},

			"scaling_instruction": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"customized_load_metric_specification": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dimensions": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},

									names.AttrMetricName: {
										Type:     schema.TypeString,
										Required: true,
									},

									names.AttrNamespace: {
										Type:     schema.TypeString,
										Required: true,
									},

									"statistic": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(enum.Slice(awstypes.MetricStatisticSum), false),
									},

									names.AttrUnit: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},

						"disable_dynamic_scaling": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						names.AttrMaxCapacity: {
							Type:     schema.TypeInt,
							Required: true,
						},

						"min_capacity": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"predefined_load_metric_specification": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"predefined_load_metric_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.LoadMetricType](),
									},

									"resource_label": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1023),
									},
								},
							},
						},

						"predictive_scaling_max_capacity_behavior": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PredictiveScalingMaxCapacityBehavior](),
						},

						"predictive_scaling_max_capacity_buffer": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(1, 100),
						},

						"predictive_scaling_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PredictiveScalingMode](),
						},

						names.AttrResourceID: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1600),
						},

						"scalable_dimension": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ScalableDimension](),
						},

						"scaling_policy_update_behavior": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.ScalingPolicyUpdateBehaviorKeepExternalPolicies,
							ValidateDiagFunc: enum.Validate[awstypes.ScalingPolicyUpdateBehavior](),
						},

						"scheduled_action_buffer_time": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"service_namespace": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ServiceNamespace](),
						},

						"target_tracking_configuration": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							MaxItems: 10,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"customized_scaling_metric_specification": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"dimensions": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},

												names.AttrMetricName: {
													Type:     schema.TypeString,
													Required: true,
												},

												names.AttrNamespace: {
													Type:     schema.TypeString,
													Required: true,
												},

												"statistic": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.MetricStatistic](),
												},

												names.AttrUnit: {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},

									"disable_scale_in": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},

									"estimated_instance_warmup": {
										Type:     schema.TypeInt,
										Optional: true,
									},

									"predefined_scaling_metric_specification": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"predefined_scaling_metric_type": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.ScalingMetricType](),
												},

												"resource_label": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(1, 1023),
												},
											},
										},
									},

									"scale_in_cooldown": {
										Type:     schema.TypeInt,
										Optional: true,
									},

									"scale_out_cooldown": {
										Type:     schema.TypeInt,
										Optional: true,
									},

									"target_value": {
										Type:         schema.TypeFloat,
										Required:     true,
										ValidateFunc: validation.FloatBetween(8.515920e-109, 1.174271e+108),
									},
								},
							},
						},
					},
				},
			},

			"scaling_plan_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceScalingPlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingPlansClient(ctx)

	scalingPlanName := d.Get(names.AttrName).(string)
	input := &autoscalingplans.CreateScalingPlanInput{
		ApplicationSource:   expandApplicationSource(d.Get("application_source").([]interface{})),
		ScalingInstructions: expandScalingInstructions(d.Get("scaling_instruction").(*schema.Set)),
		ScalingPlanName:     aws.String(scalingPlanName),
	}

	log.Printf("[DEBUG] Creating Auto Scaling Scaling Plan: %+v", input)
	output, err := conn.CreateScalingPlan(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Scaling Plan (%s): %s", scalingPlanName, err)
	}

	scalingPlanVersion := int(aws.ToInt64(output.ScalingPlanVersion))
	d.SetId(scalingPlanCreateResourceID(scalingPlanName, scalingPlanVersion))
	d.Set("scaling_plan_version", scalingPlanVersion)

	_, err = waitScalingPlanCreated(ctx, conn, scalingPlanName, scalingPlanVersion)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Scaling Plan (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceScalingPlanRead(ctx, d, meta)...)
}

func resourceScalingPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingPlansClient(ctx)

	scalingPlanName, scalingPlanVersion, err := scalingPlanParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Scaling Plan (%s): %s", d.Id(), err)
	}

	scalingPlan, err := FindScalingPlanByNameAndVersion(ctx, conn, scalingPlanName, scalingPlanVersion)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Scaling Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Scaling Plan (%s): %s", d.Id(), err)
	}

	err = d.Set("application_source", flattenApplicationSource(scalingPlan.ApplicationSource))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting application_source: %s", err)
	}
	d.Set(names.AttrName, scalingPlan.ScalingPlanName)
	err = d.Set("scaling_instruction", flattenScalingInstructions(scalingPlan.ScalingInstructions))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting scaling_instruction: %s", err)
	}
	d.Set("scaling_plan_version", scalingPlan.ScalingPlanVersion)

	return diags
}

func resourceScalingPlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingPlansClient(ctx)

	scalingPlanName, scalingPlanVersion, err := scalingPlanParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Auto Scaling Scaling Plan (%s): %s", d.Id(), err)
	}

	input := &autoscalingplans.UpdateScalingPlanInput{
		ApplicationSource:   expandApplicationSource(d.Get("application_source").([]interface{})),
		ScalingInstructions: expandScalingInstructions(d.Get("scaling_instruction").(*schema.Set)),
		ScalingPlanName:     aws.String(scalingPlanName),
		ScalingPlanVersion:  aws.Int64(int64(scalingPlanVersion)),
	}

	log.Printf("[DEBUG] Updating Auto Scaling Scaling Plan: %+v", input)
	_, err = conn.UpdateScalingPlan(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Auto Scaling Scaling Plan (%s): %s", d.Id(), err)
	}

	_, err = waitScalingPlanUpdated(ctx, conn, scalingPlanName, scalingPlanVersion)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Auto Scaling Scaling Plan (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceScalingPlanRead(ctx, d, meta)...)
}

func resourceScalingPlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingPlansClient(ctx)

	scalingPlanName, scalingPlanVersion, err := scalingPlanParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Scaling Plan (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Auto Scaling Scaling Plan: %s", d.Id())
	_, err = conn.DeleteScalingPlan(ctx, &autoscalingplans.DeleteScalingPlanInput{
		ScalingPlanName:    aws.String(scalingPlanName),
		ScalingPlanVersion: aws.Int64(int64(scalingPlanVersion)),
	})

	if errs.IsA[*awstypes.ObjectNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Scaling Plan (%s): %s", d.Id(), err)
	}

	_, err = waitScalingPlanDeleted(ctx, conn, scalingPlanName, scalingPlanVersion)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Scaling Plan (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func resourceScalingPlanImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	scalingPlanName := d.Id()
	scalingPlanVersion := 1

	d.SetId(scalingPlanCreateResourceID(scalingPlanName, scalingPlanVersion))
	d.Set(names.AttrName, scalingPlanName)
	d.Set("scaling_plan_version", scalingPlanVersion)

	return []*schema.ResourceData{d}, nil
}

//
// ApplicationSource functions.
//

func expandApplicationSource(vApplicationSource []interface{}) *awstypes.ApplicationSource {
	if len(vApplicationSource) == 0 || vApplicationSource[0] == nil {
		return nil
	}
	mApplicationSource := vApplicationSource[0].(map[string]interface{})

	applicationSource := &awstypes.ApplicationSource{}

	if v, ok := mApplicationSource["cloudformation_stack_arn"].(string); ok && v != "" {
		applicationSource.CloudFormationStackARN = aws.String(v)
	}

	if vTagFilters, ok := mApplicationSource["tag_filter"].(*schema.Set); ok && vTagFilters.Len() > 0 {
		tagFilters := []awstypes.TagFilter{}

		for _, vTagFilter := range vTagFilters.List() {
			tagFilter := awstypes.TagFilter{}

			mTagFilter := vTagFilter.(map[string]interface{})

			if v, ok := mTagFilter[names.AttrKey].(string); ok && v != "" {
				tagFilter.Key = aws.String(v)
			}

			if vValues, ok := mTagFilter[names.AttrValues].(*schema.Set); ok && vValues.Len() > 0 {
				tagFilter.Values = flex.ExpandStringValueSet(vValues)
			}

			tagFilters = append(tagFilters, tagFilter)
		}

		applicationSource.TagFilters = tagFilters
	}

	return applicationSource
}

func flattenApplicationSource(applicationSource *awstypes.ApplicationSource) []interface{} {
	if applicationSource == nil {
		return []interface{}{}
	}

	mApplicationSource := map[string]interface{}{
		"cloudformation_stack_arn": aws.ToString(applicationSource.CloudFormationStackARN),
	}

	if tagFilters := applicationSource.TagFilters; tagFilters != nil {
		vTagFilters := []interface{}{}

		for _, tagFilter := range tagFilters {
			mTagFilter := map[string]interface{}{
				names.AttrKey:    aws.ToString(tagFilter.Key),
				names.AttrValues: flex.FlattenStringValueSet(tagFilter.Values),
			}

			vTagFilters = append(vTagFilters, mTagFilter)
		}

		mApplicationSource["tag_filter"] = vTagFilters
	}

	return []interface{}{mApplicationSource}
}

//
// ScalingInstruction functions.
//

func expandScalingInstructions(vScalingInstructions *schema.Set) []awstypes.ScalingInstruction {
	scalingInstructions := []awstypes.ScalingInstruction{}

	for _, vScalingInstruction := range vScalingInstructions.List() {
		mScalingInstruction := vScalingInstruction.(map[string]interface{})

		scalingInstruction := awstypes.ScalingInstruction{}

		if v, ok := mScalingInstruction["service_namespace"].(string); ok && v != "" {
			scalingInstruction.ServiceNamespace = awstypes.ServiceNamespace(v)
		} else {
			// https://github.com/hashicorp/terraform-provider-aws/issues/17929
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			continue
		}

		if v, ok := mScalingInstruction["disable_dynamic_scaling"].(bool); ok {
			scalingInstruction.DisableDynamicScaling = aws.Bool(v)
		}
		if v, ok := mScalingInstruction[names.AttrMaxCapacity].(int); ok {
			scalingInstruction.MaxCapacity = aws.Int32(int32(v))
		}
		if v, ok := mScalingInstruction["min_capacity"].(int); ok {
			scalingInstruction.MinCapacity = aws.Int32(int32(v))
		}
		if v, ok := mScalingInstruction["predictive_scaling_max_capacity_behavior"].(string); ok && v != "" {
			scalingInstruction.PredictiveScalingMaxCapacityBehavior = awstypes.PredictiveScalingMaxCapacityBehavior(v)
		}
		if v, ok := mScalingInstruction["predictive_scaling_max_capacity_buffer"].(int); ok && v > 0 {
			scalingInstruction.PredictiveScalingMaxCapacityBuffer = aws.Int32(int32(v))
		}
		if v, ok := mScalingInstruction["predictive_scaling_mode"].(string); ok && v != "" {
			scalingInstruction.PredictiveScalingMode = awstypes.PredictiveScalingMode(v)
		}
		if v, ok := mScalingInstruction[names.AttrResourceID].(string); ok && v != "" {
			scalingInstruction.ResourceId = aws.String(v)
		}
		if v, ok := mScalingInstruction["scalable_dimension"].(string); ok && v != "" {
			scalingInstruction.ScalableDimension = awstypes.ScalableDimension(v)
		}
		if v, ok := mScalingInstruction["scaling_policy_update_behavior"].(string); ok && v != "" {
			scalingInstruction.ScalingPolicyUpdateBehavior = awstypes.ScalingPolicyUpdateBehavior(v)
		}
		if v, ok := mScalingInstruction["scheduled_action_buffer_time"].(int); ok && v > 0 {
			scalingInstruction.ScheduledActionBufferTime = aws.Int32(int32(v))
		}

		if vCustomizedLoadMetricSpecification, ok := mScalingInstruction["customized_load_metric_specification"].([]interface{}); ok && len(vCustomizedLoadMetricSpecification) > 0 && vCustomizedLoadMetricSpecification[0] != nil {
			mCustomizedLoadMetricSpecification := vCustomizedLoadMetricSpecification[0].(map[string]interface{})

			customizedLoadMetricSpecification := &awstypes.CustomizedLoadMetricSpecification{}

			if v, ok := mCustomizedLoadMetricSpecification["dimensions"].(map[string]interface{}); ok {
				dimensions := []awstypes.MetricDimension{}

				for key, value := range v {
					dimension := awstypes.MetricDimension{}

					dimension.Name = aws.String(key)
					dimension.Value = aws.String(value.(string))

					dimensions = append(dimensions, dimension)
				}

				customizedLoadMetricSpecification.Dimensions = dimensions
			}
			if v, ok := mCustomizedLoadMetricSpecification[names.AttrMetricName].(string); ok && v != "" {
				customizedLoadMetricSpecification.MetricName = aws.String(v)
			}
			if v, ok := mCustomizedLoadMetricSpecification[names.AttrNamespace].(string); ok && v != "" {
				customizedLoadMetricSpecification.Namespace = aws.String(v)
			}
			if v, ok := mCustomizedLoadMetricSpecification["statistic"].(string); ok && v != "" {
				customizedLoadMetricSpecification.Statistic = awstypes.MetricStatistic(v)
			}
			if v, ok := mCustomizedLoadMetricSpecification[names.AttrUnit].(string); ok && v != "" {
				customizedLoadMetricSpecification.Unit = aws.String(v)
			}

			scalingInstruction.CustomizedLoadMetricSpecification = customizedLoadMetricSpecification
		}

		if vPredefinedLoadMetricSpecification, ok := mScalingInstruction["predefined_load_metric_specification"].([]interface{}); ok && len(vPredefinedLoadMetricSpecification) > 0 && vPredefinedLoadMetricSpecification[0] != nil {
			mPredefinedLoadMetricSpecification := vPredefinedLoadMetricSpecification[0].(map[string]interface{})

			predefinedLoadMetricSpecification := &awstypes.PredefinedLoadMetricSpecification{}

			if v, ok := mPredefinedLoadMetricSpecification["predefined_load_metric_type"].(string); ok && v != "" {
				predefinedLoadMetricSpecification.PredefinedLoadMetricType = awstypes.LoadMetricType(v)
			}
			if v, ok := mPredefinedLoadMetricSpecification["resource_label"].(string); ok && v != "" {
				predefinedLoadMetricSpecification.ResourceLabel = aws.String(v)
			}

			scalingInstruction.PredefinedLoadMetricSpecification = predefinedLoadMetricSpecification
		}

		if vTargetTrackingConfigurations, ok := mScalingInstruction["target_tracking_configuration"].(*schema.Set); ok && vTargetTrackingConfigurations.Len() > 0 {
			targetTrackingConfigurations := []awstypes.TargetTrackingConfiguration{}

			for _, vTargetTrackingConfiguration := range vTargetTrackingConfigurations.List() {
				targetTrackingConfiguration := awstypes.TargetTrackingConfiguration{}

				mTargetTrackingConfiguration := vTargetTrackingConfiguration.(map[string]interface{})

				if v, ok := mTargetTrackingConfiguration["disable_scale_in"].(bool); ok {
					targetTrackingConfiguration.DisableScaleIn = aws.Bool(v)
				}
				if v, ok := mTargetTrackingConfiguration["estimated_instance_warmup"].(int); ok && v > 0 {
					targetTrackingConfiguration.EstimatedInstanceWarmup = aws.Int32(int32(v))
				}
				if v, ok := mTargetTrackingConfiguration["scale_in_cooldown"].(int); ok && v > 0 {
					targetTrackingConfiguration.ScaleInCooldown = aws.Int32(int32(v))
				}
				if v, ok := mTargetTrackingConfiguration["scale_out_cooldown"].(int); ok && v > 0 {
					targetTrackingConfiguration.ScaleOutCooldown = aws.Int32(int32(v))
				}
				if v, ok := mTargetTrackingConfiguration["target_value"].(float64); ok && v > 0.0 {
					targetTrackingConfiguration.TargetValue = aws.Float64(v)
				}

				if vCustomizedScalingMetricSpecification, ok := mTargetTrackingConfiguration["customized_scaling_metric_specification"].([]interface{}); ok && len(vCustomizedScalingMetricSpecification) > 0 && vCustomizedScalingMetricSpecification[0] != nil {
					mCustomizedScalingMetricSpecification := vCustomizedScalingMetricSpecification[0].(map[string]interface{})

					customizedScalingMetricSpecification := &awstypes.CustomizedScalingMetricSpecification{}

					if v, ok := mCustomizedScalingMetricSpecification["dimensions"].(map[string]interface{}); ok {
						dimensions := []awstypes.MetricDimension{}

						for key, value := range v {
							dimension := awstypes.MetricDimension{}

							dimension.Name = aws.String(key)
							dimension.Value = aws.String(value.(string))

							dimensions = append(dimensions, dimension)
						}

						customizedScalingMetricSpecification.Dimensions = dimensions
					}
					if v, ok := mCustomizedScalingMetricSpecification[names.AttrMetricName].(string); ok && v != "" {
						customizedScalingMetricSpecification.MetricName = aws.String(v)
					}
					if v, ok := mCustomizedScalingMetricSpecification[names.AttrNamespace].(string); ok && v != "" {
						customizedScalingMetricSpecification.Namespace = aws.String(v)
					}
					if v, ok := mCustomizedScalingMetricSpecification["statistic"].(string); ok && v != "" {
						customizedScalingMetricSpecification.Statistic = awstypes.MetricStatistic(v)
					}
					if v, ok := mCustomizedScalingMetricSpecification[names.AttrUnit].(string); ok && v != "" {
						customizedScalingMetricSpecification.Unit = aws.String(v)
					}

					targetTrackingConfiguration.CustomizedScalingMetricSpecification = customizedScalingMetricSpecification
				}

				if vPredefinedScalingMetricSpecification, ok := mTargetTrackingConfiguration["predefined_scaling_metric_specification"].([]interface{}); ok && len(vPredefinedScalingMetricSpecification) > 0 && vPredefinedScalingMetricSpecification[0] != nil {
					mPredefinedScalingMetricSpecification := vPredefinedScalingMetricSpecification[0].(map[string]interface{})

					predefinedScalingMetricSpecification := &awstypes.PredefinedScalingMetricSpecification{}

					if v, ok := mPredefinedScalingMetricSpecification["predefined_scaling_metric_type"].(string); ok && v != "" {
						predefinedScalingMetricSpecification.PredefinedScalingMetricType = awstypes.ScalingMetricType(v)
					}
					if v, ok := mPredefinedScalingMetricSpecification["resource_label"].(string); ok && v != "" {
						predefinedScalingMetricSpecification.ResourceLabel = aws.String(v)
					}

					targetTrackingConfiguration.PredefinedScalingMetricSpecification = predefinedScalingMetricSpecification
				}

				targetTrackingConfigurations = append(targetTrackingConfigurations, targetTrackingConfiguration)
			}

			scalingInstruction.TargetTrackingConfigurations = targetTrackingConfigurations
		}

		scalingInstructions = append(scalingInstructions, scalingInstruction)
	}

	return scalingInstructions
}

func flattenScalingInstructions(scalingInstructions []awstypes.ScalingInstruction) []interface{} {
	vScalingInstructions := []interface{}{}

	for _, scalingInstruction := range scalingInstructions {
		mScalingInstruction := map[string]interface{}{
			"disable_dynamic_scaling":                  aws.ToBool(scalingInstruction.DisableDynamicScaling),
			names.AttrMaxCapacity:                      int(aws.ToInt32(scalingInstruction.MaxCapacity)),
			"min_capacity":                             int(aws.ToInt32(scalingInstruction.MinCapacity)),
			"predictive_scaling_max_capacity_behavior": scalingInstruction.PredictiveScalingMaxCapacityBehavior,
			"predictive_scaling_max_capacity_buffer":   int(aws.ToInt32(scalingInstruction.PredictiveScalingMaxCapacityBuffer)),
			"predictive_scaling_mode":                  string(scalingInstruction.PredictiveScalingMode),
			names.AttrResourceID:                       aws.ToString(scalingInstruction.ResourceId),
			"scalable_dimension":                       string(scalingInstruction.ScalableDimension),
			"scaling_policy_update_behavior":           string(scalingInstruction.ScalingPolicyUpdateBehavior),
			"scheduled_action_buffer_time":             int(aws.ToInt32(scalingInstruction.ScheduledActionBufferTime)),
			"service_namespace":                        string(scalingInstruction.ServiceNamespace),
		}

		if customizedLoadMetricSpecification := scalingInstruction.CustomizedLoadMetricSpecification; customizedLoadMetricSpecification != nil {
			mDimensions := map[string]interface{}{}
			for _, dimension := range customizedLoadMetricSpecification.Dimensions {
				mDimensions[aws.ToString(dimension.Name)] = aws.ToString(dimension.Value)
			}

			mScalingInstruction["customized_load_metric_specification"] = []interface{}{
				map[string]interface{}{
					"dimensions":         mDimensions,
					names.AttrMetricName: aws.ToString(customizedLoadMetricSpecification.MetricName),
					names.AttrNamespace:  aws.ToString(customizedLoadMetricSpecification.Namespace),
					"statistic":          string(customizedLoadMetricSpecification.Statistic),
					names.AttrUnit:       aws.ToString(customizedLoadMetricSpecification.Unit),
				},
			}
		}

		if predefinedLoadMetricSpecification := scalingInstruction.PredefinedLoadMetricSpecification; predefinedLoadMetricSpecification != nil {
			mScalingInstruction["predefined_load_metric_specification"] = []interface{}{
				map[string]interface{}{
					"predefined_load_metric_type": string(predefinedLoadMetricSpecification.PredefinedLoadMetricType),
					"resource_label":              aws.ToString(predefinedLoadMetricSpecification.ResourceLabel),
				},
			}
		}

		if targetTrackingConfigurations := scalingInstruction.TargetTrackingConfigurations; targetTrackingConfigurations != nil {
			vTargetTrackingConfigurations := []interface{}{}

			for _, targetTrackingConfiguration := range targetTrackingConfigurations {
				mTargetTrackingConfiguration := map[string]interface{}{
					"disable_scale_in":          aws.ToBool(targetTrackingConfiguration.DisableScaleIn),
					"estimated_instance_warmup": int(aws.ToInt32(targetTrackingConfiguration.EstimatedInstanceWarmup)),
					"scale_in_cooldown":         int(aws.ToInt32(targetTrackingConfiguration.ScaleInCooldown)),
					"scale_out_cooldown":        int(aws.ToInt32(targetTrackingConfiguration.ScaleOutCooldown)),
					"target_value":              aws.ToFloat64(targetTrackingConfiguration.TargetValue),
				}

				if customizedScalingMetricSpecification := targetTrackingConfiguration.CustomizedScalingMetricSpecification; customizedScalingMetricSpecification != nil {
					mDimensions := map[string]interface{}{}
					for _, dimension := range customizedScalingMetricSpecification.Dimensions {
						mDimensions[aws.ToString(dimension.Name)] = aws.ToString(dimension.Value)
					}

					mTargetTrackingConfiguration["customized_scaling_metric_specification"] = []interface{}{
						map[string]interface{}{
							"dimensions":         mDimensions,
							names.AttrMetricName: aws.ToString(customizedScalingMetricSpecification.MetricName),
							names.AttrNamespace:  aws.ToString(customizedScalingMetricSpecification.Namespace),
							"statistic":          string(customizedScalingMetricSpecification.Statistic),
							names.AttrUnit:       aws.ToString(customizedScalingMetricSpecification.Unit),
						},
					}
				}

				if predefinedScalingMetricSpecification := targetTrackingConfiguration.PredefinedScalingMetricSpecification; predefinedScalingMetricSpecification != nil {
					mTargetTrackingConfiguration["predefined_scaling_metric_specification"] = []interface{}{
						map[string]interface{}{
							"predefined_scaling_metric_type": string(predefinedScalingMetricSpecification.PredefinedScalingMetricType),
							"resource_label":                 aws.ToString(predefinedScalingMetricSpecification.ResourceLabel),
						},
					}
				}

				vTargetTrackingConfigurations = append(vTargetTrackingConfigurations, mTargetTrackingConfiguration)
			}

			mScalingInstruction["target_tracking_configuration"] = vTargetTrackingConfigurations
		}

		vScalingInstructions = append(vScalingInstructions, mScalingInstruction)
	}

	return vScalingInstructions
}
