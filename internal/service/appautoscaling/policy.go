// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appautoscaling

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appautoscaling_policy", name="Scaling Policy")
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyPut,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyPut,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourcePolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"alarm_arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// https://github.com/boto/botocore/blob/9f322b1/botocore/data/autoscaling/2011-01-01/service-2.json#L1862-L1873
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"policy_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "StepScaling",
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scalable_dimension": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_namespace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"step_scaling_policy_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"adjustment_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cooldown": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"metric_aggregation_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"min_adjustment_magnitude": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"step_adjustment": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metric_interval_lower_bound": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"metric_interval_upper_bound": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"scaling_adjustment": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"target_tracking_scaling_policy_configuration": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"customized_metric_specification": {
							Type:          schema.TypeList,
							MaxItems:      1,
							Optional:      true,
							ConflictsWith: []string{"target_tracking_scaling_policy_configuration.0.predefined_metric_specification"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dimensions": {
										Type:          schema.TypeSet,
										Optional:      true,
										ConflictsWith: []string{"target_tracking_scaling_policy_configuration.0.customized_metric_specification.0.metrics"},
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
										ConflictsWith: []string{"target_tracking_scaling_policy_configuration.0.customized_metric_specification.0.metrics"},
									},
									"metrics": {
										Type:          schema.TypeSet,
										Optional:      true,
										ConflictsWith: []string{"target_tracking_scaling_policy_configuration.0.customized_metric_specification.0.dimensions", "target_tracking_scaling_policy_configuration.0.customized_metric_specification.0.metric_name", "target_tracking_scaling_policy_configuration.0.customized_metric_specification.0.namespace", "target_tracking_scaling_policy_configuration.0.customized_metric_specification.0.statistic", "target_tracking_scaling_policy_configuration.0.customized_metric_specification.0.unit"},
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
										ConflictsWith: []string{"target_tracking_scaling_policy_configuration.0.customized_metric_specification.0.metrics"},
									},
									"statistic": {
										Type:             schema.TypeString,
										Optional:         true,
										ConflictsWith:    []string{"target_tracking_scaling_policy_configuration.0.customized_metric_specification.0.metrics"},
										ValidateDiagFunc: enum.Validate[awstypes.MetricStatistic](),
									},
									names.AttrUnit: {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"target_tracking_scaling_policy_configuration.0.customized_metric_specification.0.metrics"},
									},
								},
							},
						},
						"disable_scale_in": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"predefined_metric_specification": {
							Type:          schema.TypeList,
							MaxItems:      1,
							Optional:      true,
							ConflictsWith: []string{"target_tracking_scaling_policy_configuration.0.customized_metric_specification"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"predefined_metric_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"resource_label": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1023),
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
							Type:     schema.TypeFloat,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourcePolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingClient(ctx)

	id := d.Get(names.AttrName).(string)
	input := expandPutScalingPolicyInput(d)

	_, err := tfresource.RetryWhenIsA[*awstypes.FailedResourceAccessException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.PutScalingPolicy(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Application Auto Scaling Scaling Policy (%s): %s", id, err)
	}

	if d.IsNewResource() {
		d.SetId(id)
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingClient(ctx)

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.FailedResourceAccessException](ctx, propagationTimeout, func() (interface{}, error) {
		return findScalingPolicyByFourPartKey(ctx, conn, d.Get(names.AttrName).(string), d.Get("service_namespace").(string), d.Get(names.AttrResourceID).(string), d.Get("scalable_dimension").(string))
	})

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Application Auto Scaling Scaling Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Application Auto Scaling Scaling Policy (%s): %s", d.Id(), err)
	}

	output := outputRaw.(*awstypes.ScalingPolicy)
	d.Set("alarm_arns", tfslices.ApplyToAll(output.Alarms, func(v awstypes.Alarm) string {
		return aws.ToString(v.AlarmARN)
	}))
	d.Set(names.AttrARN, output.PolicyARN)
	d.Set(names.AttrName, output.PolicyName)
	d.Set("policy_type", output.PolicyType)
	d.Set(names.AttrResourceID, output.ResourceId)
	d.Set("scalable_dimension", output.ScalableDimension)
	d.Set("service_namespace", output.ServiceNamespace)
	if err := d.Set("step_scaling_policy_configuration", flattenStepScalingPolicyConfiguration(output.StepScalingPolicyConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting step_scaling_policy_configuration: %s", err)
	}
	if err := d.Set("target_tracking_scaling_policy_configuration", flattenTargetTrackingScalingPolicyConfiguration(output.TargetTrackingScalingPolicyConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_tracking_scaling_policy_configuration: %s", err)
	}

	return diags
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppAutoScalingClient(ctx)

	input := &applicationautoscaling.DeleteScalingPolicyInput{
		PolicyName:        aws.String(d.Get(names.AttrName).(string)),
		ResourceId:        aws.String(d.Get(names.AttrResourceID).(string)),
		ScalableDimension: awstypes.ScalableDimension(d.Get("scalable_dimension").(string)),
		ServiceNamespace:  awstypes.ServiceNamespace(d.Get("service_namespace").(string)),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.FailedResourceAccessException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.DeleteScalingPolicy(ctx, input)
	})

	if errs.IsA[*awstypes.ObjectNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Application Auto Scaling Scaling Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func resourcePolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts, err := validPolicyImportInput(d.Id())
	if err != nil {
		return nil, err
	}

	serviceNamespace := parts[0]
	resourceID := parts[1]
	scalableDimension := parts[2]
	name := parts[3]

	d.SetId(name)
	d.Set(names.AttrName, name)
	d.Set(names.AttrResourceID, resourceID)
	d.Set("scalable_dimension", scalableDimension)
	d.Set("service_namespace", serviceNamespace)

	return []*schema.ResourceData{d}, nil
}

func findScalingPolicyByFourPartKey(ctx context.Context, conn *applicationautoscaling.Client, name, serviceNamespace, resourceID, scalableDimension string) (*awstypes.ScalingPolicy, error) {
	input := &applicationautoscaling.DescribeScalingPoliciesInput{
		PolicyNames:       []string{name},
		ResourceId:        aws.String(resourceID),
		ScalableDimension: awstypes.ScalableDimension(scalableDimension),
		ServiceNamespace:  awstypes.ServiceNamespace(serviceNamespace),
	}

	return findScalingPolicy(ctx, conn, input, func(v awstypes.ScalingPolicy) bool {
		return aws.ToString(v.PolicyName) == name && string(v.ScalableDimension) == scalableDimension
	})
}

func findScalingPolicy(ctx context.Context, conn *applicationautoscaling.Client, input *applicationautoscaling.DescribeScalingPoliciesInput, filter tfslices.Predicate[awstypes.ScalingPolicy]) (*awstypes.ScalingPolicy, error) {
	output, err := findScalingPolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findScalingPolicies(ctx context.Context, conn *applicationautoscaling.Client, input *applicationautoscaling.DescribeScalingPoliciesInput, filter tfslices.Predicate[awstypes.ScalingPolicy]) ([]awstypes.ScalingPolicy, error) {
	var output []awstypes.ScalingPolicy

	pages := applicationautoscaling.NewDescribeScalingPoliciesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ScalingPolicies {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func validPolicyImportInput(id string) ([]string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) < 4 {
		return nil, fmt.Errorf("unexpected format (%q), expected <service-namespace>/<resource-id>/<scalable-dimension>/<policy-name>", id)
	}

	var serviceNamespace, resourceID, scalableDimension, name string
	switch idParts[0] {
	case "dynamodb":
		serviceNamespace = idParts[0]

		dimensionIdx := 3
		// DynamoDB resource ID can be "/table/tableName" or "/table/tableName/index/indexName"
		if idParts[dimensionIdx] == "index" {
			dimensionIdx = 5
		}

		resourceID = strings.Join(idParts[1:dimensionIdx], "/")
		scalableDimension = idParts[dimensionIdx]
		name = strings.Join(idParts[dimensionIdx+1:], "/")
	case "kafka":
		serviceNamespace = idParts[0]
		// MSK resource ID contains three sections, separated by '/', so scalableDimension is at index 4
		dimensionIdx := 4

		resourceID = strings.Join(idParts[1:dimensionIdx], "/")
		scalableDimension = idParts[dimensionIdx]
		name = strings.Join(idParts[dimensionIdx+1:], "/")
	default:
		serviceNamespace = idParts[0]
		resourceID = strings.Join(idParts[1:len(idParts)-2], "/")
		scalableDimension = idParts[len(idParts)-2]
		name = idParts[len(idParts)-1]
	}

	if serviceNamespace == "" || resourceID == "" || scalableDimension == "" || name == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <service-namespace>/<resource-id>/<scalable-dimension>/<policy-name>", id)
	}

	return []string{serviceNamespace, resourceID, scalableDimension, name}, nil
}

// Takes the result of flatmap.Expand for an array of step adjustments and
// returns a []*awstypes.StepAdjustment.
func expandStepAdjustments(configured []interface{}) ([]awstypes.StepAdjustment, error) {
	var adjustments []awstypes.StepAdjustment

	// Loop over our configured step adjustments and create an array
	// of aws-sdk-go compatible objects. We're forced to convert strings
	// to floats here because there's no way to detect whether or not
	// an uninitialized, optional schema element is "0.0" deliberately.
	// With strings, we can test for "", which is definitely an empty
	// struct value.
	for _, raw := range configured {
		data := raw.(map[string]interface{})
		a := awstypes.StepAdjustment{
			ScalingAdjustment: aws.Int32(int32(data["scaling_adjustment"].(int))),
		}
		if data["metric_interval_lower_bound"] != "" {
			bound := data["metric_interval_lower_bound"]
			switch bound := bound.(type) {
			case string:
				f, err := strconv.ParseFloat(bound, 64)
				if err != nil {
					return nil, errors.New("metric_interval_lower_bound must be a float value represented as a string")
				}
				a.MetricIntervalLowerBound = aws.Float64(f)
			default:
				return nil, errors.New("metric_interval_lower_bound isn't a string")
			}
		}
		if data["metric_interval_upper_bound"] != "" {
			bound := data["metric_interval_upper_bound"]
			switch bound := bound.(type) {
			case string:
				f, err := strconv.ParseFloat(bound, 64)
				if err != nil {
					return nil, errors.New("metric_interval_upper_bound must be a float value represented as a string")
				}
				a.MetricIntervalUpperBound = aws.Float64(f)
			default:
				return nil, errors.New("metric_interval_upper_bound isn't a string")
			}
		}
		adjustments = append(adjustments, a)
	}

	return adjustments, nil
}

func expandCustomizedMetricSpecification(configured []interface{}) *awstypes.CustomizedMetricSpecification {
	spec := &awstypes.CustomizedMetricSpecification{}

	for _, raw := range configured {
		data := raw.(map[string]interface{})
		if val, ok := data["metrics"].(*schema.Set); ok && val.Len() > 0 {
			spec.Metrics = expandTargetTrackingMetricDataQueries(val.List())
		} else {
			if v, ok := data[names.AttrMetricName]; ok {
				spec.MetricName = aws.String(v.(string))
			}

			if v, ok := data[names.AttrNamespace]; ok {
				spec.Namespace = aws.String(v.(string))
			}

			if v, ok := data[names.AttrUnit].(string); ok && v != "" {
				spec.Unit = aws.String(v)
			}

			if v, ok := data["statistic"]; ok {
				spec.Statistic = awstypes.MetricStatistic(v.(string))
			}

			if s, ok := data["dimensions"].(*schema.Set); ok && s.Len() > 0 {
				dimensions := make([]awstypes.MetricDimension, s.Len())
				for i, d := range s.List() {
					dimension := d.(map[string]interface{})
					dimensions[i] = awstypes.MetricDimension{
						Name:  aws.String(dimension[names.AttrName].(string)),
						Value: aws.String(dimension[names.AttrValue].(string)),
					}
				}
				spec.Dimensions = dimensions
			}
		}
	}
	return spec
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
			metric := &awstypes.TargetTrackingMetric{
				MetricName: aws.String(metricSpec[names.AttrMetricName].(string)),
				Namespace:  aws.String(metricSpec[names.AttrNamespace].(string)),
			}
			if v, ok := metricSpec["dimensions"]; ok {
				dims := v.(*schema.Set).List()
				dimList := make([]awstypes.TargetTrackingMetricDimension, len(dims))
				for i := range dimList {
					dim := dims[i].(map[string]interface{})
					md := awstypes.TargetTrackingMetricDimension{
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

func expandPredefinedMetricSpecification(configured []interface{}) *awstypes.PredefinedMetricSpecification {
	spec := &awstypes.PredefinedMetricSpecification{}

	for _, raw := range configured {
		data := raw.(map[string]interface{})

		if v, ok := data["predefined_metric_type"]; ok {
			spec.PredefinedMetricType = awstypes.MetricType(v.(string))
		}

		if v, ok := data["resource_label"].(string); ok && v != "" {
			spec.ResourceLabel = aws.String(v)
		}
	}
	return spec
}

func expandPutScalingPolicyInput(d *schema.ResourceData) *applicationautoscaling.PutScalingPolicyInput {
	apiObject := &applicationautoscaling.PutScalingPolicyInput{
		PolicyName: aws.String(d.Get(names.AttrName).(string)),
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
	}

	if v, ok := d.GetOk("policy_type"); ok {
		apiObject.PolicyType = awstypes.PolicyType(v.(string))
	}

	if v, ok := d.GetOk("scalable_dimension"); ok {
		apiObject.ScalableDimension = awstypes.ScalableDimension(v.(string))
	}

	if v, ok := d.GetOk("service_namespace"); ok {
		apiObject.ServiceNamespace = awstypes.ServiceNamespace(v.(string))
	}

	if v, ok := d.GetOk("step_scaling_policy_configuration"); ok {
		apiObject.StepScalingPolicyConfiguration = expandStepScalingPolicyConfiguration(v.([]interface{}))
	}

	if l, ok := d.GetOk("target_tracking_scaling_policy_configuration"); ok {
		v := l.([]interface{})
		if len(v) == 1 {
			ttspCfg := v[0].(map[string]interface{})
			cfg := &awstypes.TargetTrackingScalingPolicyConfiguration{
				TargetValue: aws.Float64(ttspCfg["target_value"].(float64)),
			}

			if v, ok := ttspCfg["scale_in_cooldown"]; ok {
				cfg.ScaleInCooldown = aws.Int32(int32(v.(int)))
			}

			if v, ok := ttspCfg["scale_out_cooldown"]; ok {
				cfg.ScaleOutCooldown = aws.Int32(int32(v.(int)))
			}

			if v, ok := ttspCfg["disable_scale_in"]; ok {
				cfg.DisableScaleIn = aws.Bool(v.(bool))
			}

			if v, ok := ttspCfg["customized_metric_specification"].([]interface{}); ok && len(v) > 0 {
				cfg.CustomizedMetricSpecification = expandCustomizedMetricSpecification(v)
			}

			if v, ok := ttspCfg["predefined_metric_specification"].([]interface{}); ok && len(v) > 0 {
				cfg.PredefinedMetricSpecification = expandPredefinedMetricSpecification(v)
			}

			apiObject.TargetTrackingScalingPolicyConfiguration = cfg
		}
	}

	return apiObject
}

func expandStepScalingPolicyConfiguration(cfg []interface{}) *awstypes.StepScalingPolicyConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	out := &awstypes.StepScalingPolicyConfiguration{}

	m := cfg[0].(map[string]interface{})
	if v, ok := m["adjustment_type"]; ok {
		out.AdjustmentType = awstypes.AdjustmentType(v.(string))
	}
	if v, ok := m["cooldown"]; ok {
		out.Cooldown = aws.Int32(int32(v.(int)))
	}
	if v, ok := m["metric_aggregation_type"]; ok {
		out.MetricAggregationType = awstypes.MetricAggregationType(v.(string))
	}
	if v, ok := m["min_adjustment_magnitude"].(int); ok && v > 0 {
		out.MinAdjustmentMagnitude = aws.Int32(int32(v))
	}
	if v, ok := m["step_adjustment"].(*schema.Set); ok && v.Len() > 0 {
		out.StepAdjustments, _ = expandStepAdjustments(v.List())
	}

	return out
}

func flattenStepScalingPolicyConfiguration(cfg *awstypes.StepScalingPolicyConfiguration) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["adjustment_type"] = string(cfg.AdjustmentType)

	if cfg.Cooldown != nil {
		m["cooldown"] = aws.ToInt32(cfg.Cooldown)
	}

	m["metric_aggregation_type"] = string(cfg.MetricAggregationType)

	if cfg.MinAdjustmentMagnitude != nil {
		m["min_adjustment_magnitude"] = aws.ToInt32(cfg.MinAdjustmentMagnitude)
	}
	if cfg.StepAdjustments != nil {
		stepAdjustmentsResource := &schema.Resource{
			Schema: map[string]*schema.Schema{
				"metric_interval_lower_bound": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"metric_interval_upper_bound": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"scaling_adjustment": {
					Type:     schema.TypeInt,
					Required: true,
				},
			},
		}
		m["step_adjustment"] = schema.NewSet(schema.HashResource(stepAdjustmentsResource), flattenStepAdjustments(cfg.StepAdjustments))
	}

	return []interface{}{m}
}

func flattenStepAdjustments(adjs []awstypes.StepAdjustment) []interface{} {
	out := make([]interface{}, len(adjs))

	for i, adj := range adjs {
		m := make(map[string]interface{})

		m["scaling_adjustment"] = int(aws.ToInt32(adj.ScalingAdjustment))

		if adj.MetricIntervalLowerBound != nil {
			m["metric_interval_lower_bound"] = fmt.Sprintf("%g", aws.ToFloat64(adj.MetricIntervalLowerBound))
		}
		if adj.MetricIntervalUpperBound != nil {
			m["metric_interval_upper_bound"] = fmt.Sprintf("%g", aws.ToFloat64(adj.MetricIntervalUpperBound))
		}

		out[i] = m
	}

	return out
}

func flattenTargetTrackingScalingPolicyConfiguration(cfg *awstypes.TargetTrackingScalingPolicyConfiguration) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if v := cfg.CustomizedMetricSpecification; v != nil {
		m["customized_metric_specification"] = flattenCustomizedMetricSpecification(v)
	}

	if v := cfg.DisableScaleIn; v != nil {
		m["disable_scale_in"] = aws.ToBool(v)
	}

	if v := cfg.PredefinedMetricSpecification; v != nil {
		m["predefined_metric_specification"] = flattenPredefinedMetricSpecification(v)
	}

	if v := cfg.ScaleInCooldown; v != nil {
		m["scale_in_cooldown"] = aws.ToInt32(v)
	}

	if v := cfg.ScaleOutCooldown; v != nil {
		m["scale_out_cooldown"] = aws.ToInt32(v)
	}

	if v := cfg.TargetValue; v != nil {
		m["target_value"] = aws.ToFloat64(v)
	}

	return []interface{}{m}
}

func flattenCustomizedMetricSpecification(cfg *awstypes.CustomizedMetricSpecification) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if cfg.Metrics != nil {
		m["metrics"] = flattenTargetTrackingMetricDataQueries(cfg.Metrics)
	} else {
		if v := cfg.Dimensions; len(v) > 0 {
			m["dimensions"] = flattenMetricDimensions(cfg.Dimensions)
		}

		if v := cfg.MetricName; v != nil {
			m[names.AttrMetricName] = aws.ToString(v)
		}

		if v := cfg.Namespace; v != nil {
			m[names.AttrNamespace] = aws.ToString(v)
		}

		m["statistic"] = string(cfg.Statistic)

		if v := cfg.Unit; v != nil {
			m[names.AttrUnit] = aws.ToString(v)
		}
	}

	return []interface{}{m}
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

func flattenMetricDimensions(ds []awstypes.MetricDimension) []interface{} {
	l := make([]interface{}, len(ds))
	for i, d := range ds {
		if ds == nil {
			continue
		}

		m := map[string]interface{}{}

		if v := d.Name; v != nil {
			m[names.AttrName] = aws.ToString(v)
		}

		if v := d.Value; v != nil {
			m[names.AttrValue] = aws.ToString(v)
		}

		l[i] = m
	}
	return l
}

func flattenPredefinedMetricSpecification(cfg *awstypes.PredefinedMetricSpecification) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	m["predefined_metric_type"] = string(cfg.PredefinedMetricType)

	if v := cfg.ResourceLabel; v != nil {
		m["resource_label"] = aws.ToString(v)
	}

	return []interface{}{m}
}
